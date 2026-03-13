package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/judwhite/go-svc"
	"data-cleaning-service/internal/cleaner"
	"data-cleaning-service/internal/config"
	"data-cleaning-service/internal/loader"
	"data-cleaning-service/pkg/logger"
	"go.uber.org/zap"
)

// CleaningService 数据清洗服务
type CleaningService struct {
	config     *config.Config
	kafkaLoader *loader.KafkaLoader
	cleaner    *cleaner.DataCleaner
	stopChan   chan struct{}
	csvWriters map[string]*CSVWriter // 按主题维护CSV写入器
}

// CSVWriter CSV写入器
type CSVWriter struct {
	file      *os.File
	csvWriter *csv.Writer
	mu        sync.Mutex
	path      string
	rowsCount int
}

// NewCleaningService 创建清洗服务
func NewCleaningService(cfg *config.Config) *CleaningService {
	return &CleaningService{
		config:     cfg,
		stopChan:   make(chan struct{}),
		csvWriters: make(map[string]*CSVWriter),
	}
}

// Init 初始化服务
func (s *CleaningService) Init(env svc.Environment) error {
	logger.Info("初始化数据清洗服务",
		zap.String("name", s.config.Service.Name),
	)

	// 确保必要的目录存在
	if err := s.config.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	// 创建Kafka加载器
	var err error
	s.kafkaLoader, err = loader.NewKafkaLoader(&s.config.Data.Kafka)
	if err != nil {
		return fmt.Errorf("failed to create kafka loader: %w", err)
	}

	// 创建数据清洗器
	s.cleaner = cleaner.NewDataCleaner(&s.config.CleaningRules)

	logger.Info("数据清洗服务初始化完成")
	return nil
}

// Start 启动服务
func (s *CleaningService) Start() error {
	logger.Info("启动数据清洗服务")

	// 在协程中执行清洗任务
	go s.run()

	return nil
}

// Stop 停止服务
func (s *CleaningService) Stop() error {
	logger.Info("停止数据清洗服务")
	close(s.stopChan)

	// 关闭所有CSV写入器
	for topic, writer := range s.csvWriters {
		if writer != nil {
			writer.Close()
			logger.Info("关闭CSV写入器", zap.String("topic", topic))
		}
	}

	// 关闭资源
	if s.kafkaLoader != nil {
		if err := s.kafkaLoader.Close(); err != nil {
			logger.Error("关闭加载器失败", zap.Error(err))
		}
	}

	logger.Info("数据清洗服务已停止")
	return nil
}

// run 运行清洗服务
func (s *CleaningService) run() {
	logger.Info("开始持续消费Kafka数据",
		zap.Strings("topics", s.config.Data.Kafka.Topics),
		zap.Int("max_rows_per_file", s.config.Data.MaxRowsPerFile),
	)

	for {
		select {
		case <-s.stopChan:
			logger.Info("接收到停止信号")
			return

		default:
			// 加载一批数据（按主题分组）
			dfMap, loadErr := s.kafkaLoader.Load()
			if loadErr != nil {
				if loadErr.Error() == "no data received from kafka" {
					time.Sleep(1 * time.Second)
					continue
				}
				logger.Error("加载数据失败", zap.Error(loadErr))
				time.Sleep(5 * time.Second)
				continue
			}

			if len(dfMap) == 0 {
				continue
			}

			// 按主题处理数据
			for topic, df := range dfMap {
				if df.Nrow() == 0 {
					continue
				}

				// 获取该主题的专用清洗规则
				topicCleaner := s.getTopicCleaner(topic)

				// 清洗数据
				cleanedDF, cleanErr := topicCleaner.Clean(df)
				if cleanErr != nil {
					logger.Error("清洗数据失败",
						zap.String("topic", topic),
						zap.Error(cleanErr),
					)
					continue
				}

				// 获取或创建该主题的CSV写入器
				csvWriter, err := s.getOrCreateCSVWriter(topic)
				if err != nil {
					logger.Error("获取CSV写入器失败",
						zap.String("topic", topic),
						zap.Error(err),
					)
					continue
				}

				// 如果需要切换文件，创建新的写入器
				if csvWriter.rowsCount+cleanedDF.Nrow() > s.config.Data.MaxRowsPerFile {
					// 关闭旧的写入器
					csvWriter.Close()
					delete(s.csvWriters, topic)

					// 创建新的写入器
					csvWriter, err = s.newCSVWriter(topic)
					if err != nil {
						logger.Error("创建CSV写入器失败",
							zap.String("topic", topic),
							zap.Error(err),
						)
						continue
					}
					s.csvWriters[topic] = csvWriter
				}

				// 写入数据到CSV
				if err := s.writeDataToCSV(csvWriter, cleanedDF); err != nil {
					logger.Error("写入CSV失败",
						zap.String("topic", topic),
						zap.Error(err),
					)
					continue
				}

				logger.Info("批次处理完成",
					zap.String("topic", topic),
					zap.Int("rows", cleanedDF.Nrow()),
					zap.Int("total_rows", csvWriter.rowsCount),
				)
			}
		}
	}
}

// getTopicCleaner 获取主题专用清洗器
func (s *CleaningService) getTopicCleaner(topic string) *cleaner.DataCleaner {
	// 获取主题专用规则
	if topicRule, exists := s.config.Data.Kafka.TopicRules[topic]; exists && len(topicRule.NumericRanges) > 0 {
		// 创建合并的清洗规则
		mergedRules := s.config.CleaningRules
		mergedRules.NumericRanges = topicRule.NumericRanges // 使用主题专用规则覆盖默认规则
		return cleaner.NewDataCleaner(&mergedRules)
	}

	// 使用默认清洗规则
	return s.cleaner
}

// getOrCreateCSVWriter 获取或创建CSV写入器
func (s *CleaningService) getOrCreateCSVWriter(topic string) (*CSVWriter, error) {
	if writer, exists := s.csvWriters[topic]; exists {
		return writer, nil
	}

	writer, err := s.newCSVWriter(topic)
	if err != nil {
		return nil, err
	}

	s.csvWriters[topic] = writer
	return writer, nil
}

// newCSVWriter 创建新的CSV写入器
func (s *CleaningService) newCSVWriter(topic string) (*CSVWriter, error) {
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%d.csv", topic, timestamp)
	path := filepath.Join(s.config.Data.OutputDir, filename)

	logger.Info("创建新的CSV文件",
		zap.String("topic", topic),
		zap.String("path", path),
	)

	// 创建文件
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create csv file: %w", err)
	}

	csvWriter := &CSVWriter{
		file:      file,
		csvWriter: csv.NewWriter(file),
		path:      path,
		rowsCount: 0,
	}

	return csvWriter, nil
}

// writeDataToCSV 写入数据到CSV
func (s *CleaningService) writeDataToCSV(csvWriter *CSVWriter, df dataframe.DataFrame) error {
	csvWriter.mu.Lock()
	defer csvWriter.mu.Unlock()

	// 获取列名并按字典序排序
	colNames := df.Names()
	sortedColNames := make([]string, len(colNames))
	copy(sortedColNames, colNames)
	sort.Strings(sortedColNames)

	isFirstWrite := csvWriter.rowsCount == 0

	// 如果是第一次写入，写入表头
	if isFirstWrite {
		if err := csvWriter.csvWriter.Write(sortedColNames); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		csvWriter.rowsCount++
	}

	// 写入数据行
	for i := 0; i < df.Nrow(); i++ {
		row := make([]string, len(sortedColNames))
		for j := 0; j < len(sortedColNames); j++ {
			colName := sortedColNames[j]
			s := df.Col(colName)
			elem := s.Elem(i)
			if elem.IsNA() {
				row[j] = ""
			} else {
				// 根据类型获取值
				switch s.Type() {
				case series.String:
					row[j] = elem.String()
				case series.Int:
					if val, err := elem.Int(); err == nil {
						row[j] = fmt.Sprintf("%d", val)
					} else {
						row[j] = ""
					}
				case series.Float:
					row[j] = fmt.Sprintf("%f", elem.Float())
				case series.Bool:
					if val, err := elem.Bool(); err == nil {
						row[j] = fmt.Sprintf("%t", val)
					} else {
						row[j] = ""
					}
				default:
					row[j] = ""
				}
			}
		}

		if err := csvWriter.csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i, err)
		}
		csvWriter.rowsCount++
	}

	csvWriter.csvWriter.Flush()
	if err := csvWriter.csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to flush csv writer: %w", err)
	}

	return nil
}

// Close 关闭CSV写入器
func (w *CSVWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.csvWriter != nil {
		w.csvWriter.Flush()
	}
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
	}

	logger.Info("CSV文件已关闭",
		zap.String("path", w.path),
		zap.Int("total_rows", w.rowsCount),
	)

	return nil
}
