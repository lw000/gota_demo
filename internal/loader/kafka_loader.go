package loader

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"data-cleaning-service/internal/config"
	"data-cleaning-service/pkg/kafka"
	"data-cleaning-service/pkg/logger"
	"go.uber.org/zap"
)

// KafkaLoader Kafka数据加载器
type KafkaLoader struct {
	config       *config.KafkaConfig
	consumer     *kafka.Consumer
	batchSize    int
	batchTimeout time.Duration
	messages     chan *kafka.KafkaMessageWithTopic
}

// NewKafkaLoader 创建Kafka加载器
func NewKafkaLoader(kafkaConfig *config.KafkaConfig) (*KafkaLoader, error) {
	logger.Info("创建Kafka加载器",
		zap.Strings("brokers", kafkaConfig.Brokers),
		zap.Strings("topics", kafkaConfig.Topics),
	)

	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	loader := &KafkaLoader{
		config:       kafkaConfig,
		consumer:     consumer,
		batchSize:    kafkaConfig.BatchSize,
		batchTimeout: time.Duration(kafkaConfig.BatchTimeoutMS) * time.Millisecond,
		messages:     make(chan *kafka.KafkaMessageWithTopic, 1000),
	}

	// 启动消费者
	if err := loader.consumer.Start(loader.messages); err != nil {
		return nil, fmt.Errorf("failed to start kafka consumer: %w", err)
	}

	return loader, nil
}

// Load 加载Kafka数据（批量）
func (l *KafkaLoader) Load() (map[string]dataframe.DataFrame, error) {
	logger.Info("开始批量加载Kafka数据",
		zap.Int("batch_size", l.batchSize),
	)

	ctx := context.Background()
	topicBatches := make(map[string][]KafkaMessageData)
	timeout := time.After(l.batchTimeout)

	startTime := time.Now()
	var messagesConsumed int

	for {
		select {
		case msgWithTopic := <-l.messages:
			topic := msgWithTopic.Topic

			// 初始化主题批次
			if _, exists := topicBatches[topic]; !exists {
				topicBatches[topic] = make([]KafkaMessageData, 0, l.batchSize)
			}

			// 添加消息到对应主题批次
			topicBatches[topic] = append(topicBatches[topic], msgWithTopic.Data)
			messagesConsumed++

			// 检查所有主题的总消息数是否达到批次大小
			totalMessages := 0
			for _, batch := range topicBatches {
				totalMessages += len(batch)
			}

			// 达到批次大小
			if totalMessages >= l.batchSize {
				logger.Info("达到批次大小，停止加载",
					zap.Int("total_messages", totalMessages),
					zap.Int("messages_consumed", messagesConsumed),
				)
				goto PROCESS
			}

		case <-timeout:
			totalMessages := 0
			for _, batch := range topicBatches {
				totalMessages += len(batch)
			}

			if totalMessages > 0 {
				logger.Info("批次超时，停止加载",
					zap.Int("total_messages", totalMessages),
					zap.Int("messages_consumed", messagesConsumed),
				)
				goto PROCESS
			}

			// 如果超时但还没有数据，继续等待
			logger.Warn("批次超时但无数据，继续等待")
			timeout = time.After(l.batchTimeout)

		case <-ctx.Done():
			logger.Info("上下文取消，停止加载")
			goto PROCESS
		}
	}

PROCESS:
	if len(topicBatches) == 0 {
		return nil, fmt.Errorf("no data received from kafka")
	}

	// 将各主题批次数据转换为DataFrame
	result := make(map[string]dataframe.DataFrame)
	for topic, batch := range topicBatches {
		df, err := l.batchToDataFrame(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to convert topic %s batch to dataframe: %w", topic, err)
		}
		result[topic] = df
	}

	duration := time.Since(startTime)
	totalRows := 0
	for _, df := range result {
		totalRows += df.Nrow()
	}

	logger.Info("Kafka数据加载完成",
		zap.Int("topics", len(result)),
		zap.Int("total_rows", totalRows),
		zap.Int("messages_consumed", messagesConsumed),
		zap.Duration("duration", duration),
	)

	return result, nil
}

// batchToDataFrame 将批次数据转换为DataFrame
func (l *KafkaLoader) batchToDataFrame(batch []KafkaMessageData) (dataframe.DataFrame, error) {
	if len(batch) == 0 {
		return dataframe.DataFrame{}, nil
	}

	// 收集所有列名
	columns := make(map[string][]interface{})
	for _, row := range batch {
		for key := range row {
			if _, exists := columns[key]; !exists {
				columns[key] = make([]interface{}, 0, len(batch))
			}
		}
		_ = row // 使用row避免unused变量警告
	}

	// 填充数据
	for _, row := range batch {
		for key := range columns {
			if value, exists := row[key]; exists {
				columns[key] = append(columns[key], value)
			} else {
				columns[key] = append(columns[key], nil)
			}
		}
		_ = row // 使用row避免unused变量警告
	}

	// 创建Series数组
	seriesArray := make([]series.Series, 0, len(columns))
	for colName, values := range columns {
		seriesArray = append(seriesArray, series.New(values, series.String, colName))
	}

	// 创建DataFrame
	df := dataframe.New(seriesArray...)
	return df, nil
}

// LoadWithCallback 加载Kafka数据并使用回调处理
func (l *KafkaLoader) LoadWithCallback(callback func(map[string]dataframe.DataFrame) error, maxBatches int) error {
	logger.Info("开始加载Kafka数据（回调模式）",
		zap.Int("max_batches", maxBatches),
	)

	batchCount := 0
	var wg sync.WaitGroup

	for {
		if maxBatches > 0 && batchCount >= maxBatches {
			logger.Info("达到最大批次限制",
				zap.Int("batch_count", batchCount),
			)
			break
		}

		dfMap, err := l.Load()
		if err != nil {
			logger.Error("加载数据失败", zap.Error(err))
			if strings.Contains(err.Error(), "no data received") {
				time.Sleep(1 * time.Second)
				continue
			}
			return err
		}

		batchCount++

		// 异步处理批次
		wg.Add(1)
		go func(batchDFMap map[string]dataframe.DataFrame) {
			defer wg.Done()
			if err := callback(batchDFMap); err != nil {
				logger.Error("处理批次失败", zap.Error(err))
			}
		}(dfMap)
	}

	// 等待所有批次处理完成
	wg.Wait()

	logger.Info("所有批次处理完成",
		zap.Int("total_batches", batchCount),
	)

	return nil
}

// Close 关闭加载器
func (l *KafkaLoader) Close() error {
	logger.Info("关闭Kafka加载器")
	if l.consumer != nil {
		return l.consumer.Stop()
	}
	return nil
}

// GetSourceInfo 获取数据源信息
func (l *KafkaLoader) GetSourceInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":           "kafka",
		"brokers":        l.config.Brokers,
		"topics":         l.config.Topics,
		"consumer_group": l.config.ConsumerGroup,
		"batch_size":     l.batchSize,
		"batch_timeout":  l.batchTimeout.Milliseconds(),
	}
}
