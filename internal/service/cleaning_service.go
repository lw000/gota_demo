package service

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/judwhite/go-svc"
	"data-cleaning-service/internal/cleaner"
	"data-cleaning-service/internal/config"
	"data-cleaning-service/internal/loader"
	"data-cleaning-service/internal/reporter"
	"data-cleaning-service/pkg/logger"
	"data-cleaning-service/pkg/kafka"
	"go.uber.org/zap"
)

// CleaningService 数据清洗服务
type CleaningService struct {
	config     *config.Config
	loader     *loader.BatchLoader
	cleaner    *cleaner.DataCleaner
	reporter   *reporter.ReportGenerator
	producer   *kafka.Producer // 可选的Kafka输出
	stopChan   chan struct{}
}

// NewCleaningService 创建清洗服务
func NewCleaningService(cfg *config.Config) *CleaningService {
	return &CleaningService{
		config:   cfg,
		stopChan: make(chan struct{}),
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

	// 创建数据加载器
	var err error
	s.loader, err = loader.NewBatchLoader(&s.config.Data)
	if err != nil {
		return fmt.Errorf("failed to create loader: %w", err)
	}

	// 创建数据清洗器
	s.cleaner = cleaner.NewDataCleaner(&s.config.CleaningRules)

	// 创建报告生成器
	s.reporter = reporter.NewReportGenerator(s.config.Data.ReportPath)

	// 如果启用了Kafka输出，创建生产者
	if s.config.Data.KafkaOutput.Enabled {
		s.producer, err = kafka.NewProducer(&s.config.Data.KafkaOutput)
		if err != nil {
			return fmt.Errorf("failed to create kafka producer: %w", err)
		}
		logger.Info("Kafka生产者已启用")
	}

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

	// 关闭资源
	if s.loader != nil {
		if err := s.loader.Close(); err != nil {
			logger.Error("关闭加载器失败", zap.Error(err))
		}
	}

	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			logger.Error("关闭Kafka生产者失败", zap.Error(err))
		}
	}

	logger.Info("数据清洗服务已停止")
	return nil
}

// run 运行清洗服务
func (s *CleaningService) run() {
	startTime := time.Now()
	
	logger.Info("开始执行数据清洗任务")

	// 1. 加载数据
	logger.Info("步骤1: 加载数据")
	originalDF, err := s.loader.Load()
	if err != nil {
		logger.Error("加载数据失败", zap.Error(err))
		return
	}

	logger.Info("数据加载成功",
		zap.Int("rows", originalDF.Nrow()),
		zap.Int("columns", originalDF.Ncol()),
	)

	// 2. 清洗数据
	logger.Info("步骤2: 清洗数据")
	cleanedDF, err := s.cleaner.Clean(originalDF)
	if err != nil {
		logger.Error("清洗数据失败", zap.Error(err))
		return
	}

	logger.Info("数据清洗成功",
		zap.Int("cleaned_rows", cleanedDF.Nrow()),
		zap.Int("removed_rows", s.cleaner.GetOriginalRows()-s.cleaner.GetCleanedRows()),
	)

	// 3. 保存清洗后的数据
	logger.Info("步骤3: 保存清洗后的数据")
	if err := s.saveOutputData(cleanedDF); err != nil {
		logger.Error("保存数据失败", zap.Error(err))
		return
	}

	// 4. 推送到Kafka（如果启用）
	if s.producer != nil {
		logger.Info("步骤4: 推送清洗后的数据到Kafka")
		if err := s.pushToKafka(cleanedDF); err != nil {
			logger.Error("推送Kafka失败", zap.Error(err))
		}
	}

	// 5. 生成报告
	logger.Info("步骤5: 生成清洗报告")
	processingTime := time.Since(startTime)
	report, err := s.reporter.GenerateReport(
		originalDF,
		cleanedDF,
		s.loader,
		s.cleaner,
		processingTime,
	)
	if err != nil {
		logger.Error("生成报告失败", zap.Error(err))
		return
	}

	// 6. 保存报告
	if err := s.reporter.SaveReport(report); err != nil {
		logger.Error("保存报告失败", zap.Error(err))
		return
	}

	// 打印报告到控制台
	s.reporter.PrintReport(report)

	logger.Info("数据清洗任务完成",
		zap.Duration("total_time", processingTime),
	)

	// 如果是CSV模式，任务完成后退出
	if s.config.Data.CSV.Enabled && !s.config.Data.Kafka.Enabled {
		logger.Info("CSV模式下任务完成，退出服务")
		s.Stop()
		return
	}

	// 如果是Kafka模式，继续监听
	if s.config.Data.Kafka.Enabled {
		logger.Info("Kafka模式下继续监听消息...")
		s.runKafkaLoop()
	}
}

// runKafkaLoop 运行Kafka循环
func (s *CleaningService) runKafkaLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			logger.Info("Kafka循环停止")
			return
		case <-ticker.C:
			logger.Info("Kafka心跳检查")
			// 可以在这里添加健康检查逻辑
		}
	}
}

// saveOutputData 保存输出数据
func (s *CleaningService) saveOutputData(df dataframe.DataFrame) error {
	logger.Info("保存清洗后的数据",
		zap.String("path", s.config.Data.OutputPath),
	)

	// 确保目录存在
	if err := os.MkdirAll(getDirPath(s.config.Data.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 创建输出文件
	file, err := os.Create(s.config.Data.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// 写入CSV
	if err := df.WriteCSV(file); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	logger.Info("数据保存成功")
	return nil
}

// pushToKafka 推送数据到Kafka
func (s *CleaningService) pushToKafka(df dataframe.DataFrame) error {
	if s.producer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}

	logger.Info("推送数据到Kafka",
		zap.String("topic", s.config.Data.KafkaOutput.Topic),
		zap.Int("rows", df.Nrow()),
	)

	// 遍历每一行，发送到Kafka
	colNames := df.Names()
	for i := 0; i < df.Nrow(); i++ {
		// 构建消息
		rowData := make(map[string]interface{})
		for j := 0; j < df.Ncol(); j++ {
			s := df.Col(colNames[j])
			elem := s.Elem(i)
			if elem.IsNA() {
				rowData[colNames[j]] = nil
			} else {
				// 根据类型获取值
				switch s.Type() {
				case series.String:
					rowData[colNames[j]] = elem.String()
				case series.Int:
					if val, err := elem.Int(); err == nil {
						rowData[colNames[j]] = val
					} else {
						rowData[colNames[j]] = nil
					}
				case series.Float:
					rowData[colNames[j]] = elem.Float()
				case series.Bool:
					if val, err := elem.Bool(); err == nil {
						rowData[colNames[j]] = val
					} else {
						rowData[colNames[j]] = nil
					}
				default:
					rowData[colNames[j]] = nil
				}
			}
		}

		// 序列化
		jsonData, err := json.Marshal(rowData)
		if err != nil {
			logger.Error("序列化数据失败",
				zap.Int("row", i),
				zap.Error(err),
			)
			continue
		}

		// 发送
		_, _, err = s.producer.SendMessage([]byte(colNames[0]), jsonData)
		if err != nil {
			logger.Error("发送Kafka消息失败",
				zap.Int("row", i),
				zap.Error(err),
			)
		}
	}

	logger.Info("数据推送Kafka完成")
	return nil
}

// getDirPath 获取目录路径
func getDirPath(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			return filePath[:i]
		}
	}
	return "."
}

// dfToString 简化版的JSON序列化（已删除，使用json.Marshal替代）
// 保留此函数以保持接口一致性
func dfToString(data map[string]interface{}) ([]byte, error) {
	return json.Marshal(data)
}
