package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"data-cleaning-service/internal/config"
	"data-cleaning-service/pkg/logger"
	"go.uber.org/zap"
)

// Producer Kafka生产者
type Producer struct {
	config   *config.KafkaConfig
	topic    string
	producer sarama.SyncProducer
}

// NewProducer 创建Kafka生产者
func NewProducer(kafkaConfig *config.KafkaConfig, topic string) (*Producer, error) {
	logger.Info("创建Kafka生产者",
		zap.Strings("brokers", kafkaConfig.Brokers),
		zap.String("topic", topic),
	)

	// 创建Sarama配置
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_5_0_0
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Retry.Backoff = 100 // 100ms

	// 配置安全认证（复用consumer的配置逻辑）
	if err := configureSASL(saramaConfig, &kafkaConfig.Security); err != nil {
		return nil, fmt.Errorf("failed to configure SASL: %w", err)
	}

	if err := configureTLS(saramaConfig, &kafkaConfig.Security); err != nil {
		return nil, fmt.Errorf("failed to configure TLS: %w", err)
	}

	// 创建生产者
	producer, err := sarama.NewSyncProducer(kafkaConfig.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	logger.Info("Kafka生产者创建成功")
	return &Producer{
		config:   kafkaConfig,
		topic:    topic,
		producer: producer,
	}, nil
}

// SendMessage 发送消息
func (p *Producer) SendMessage(key, value []byte) (partition int32, offset int64, err error) {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err = p.producer.SendMessage(msg)
	if err != nil {
		logger.Error("发送Kafka消息失败",
			zap.String("topic", p.topic),
			zap.Error(err),
		)
		return 0, 0, err
	}

	logger.Debug("Kafka消息发送成功",
		zap.String("topic", p.topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return partition, offset, nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	logger.Info("关闭Kafka生产者")
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
