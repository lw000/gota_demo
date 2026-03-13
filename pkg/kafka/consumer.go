package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/IBM/sarama"
	"data-cleaning-service/internal/config"
	"data-cleaning-service/pkg/logger"
	"go.uber.org/zap"
)

// Consumer Kafka消费者
type Consumer struct {
	config     *config.KafkaConfig
	group      sarama.ConsumerGroup
	handler    *ConsumerHandler
	ctx        context.Context
	cancel     context.CancelFunc
}

// ConsumerHandler 消费者处理器
type ConsumerHandler struct {
	messages chan<- *sarama.ConsumerMessage
	logger   *zap.Logger
}

// NewConsumer 创建Kafka消费者
func NewConsumer(kafkaConfig *config.KafkaConfig) (*Consumer, error) {
	logger.Info("创建Kafka消费者",
		zap.Strings("brokers", kafkaConfig.Brokers),
		zap.String("topic", kafkaConfig.Topic),
		zap.String("group", kafkaConfig.ConsumerGroup),
	)

	// 创建Sarama配置
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_5_0_0
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConfig.Consumer.Group.Session.Timeout = 10 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	// 自动提交offset
	if kafkaConfig.AutoCommit {
		saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
		saramaConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	} else {
		saramaConfig.Consumer.Offsets.AutoCommit.Enable = false
	}

	// 配置安全认证
	if err := configureSASL(saramaConfig, &kafkaConfig.Security); err != nil {
		return nil, fmt.Errorf("failed to configure SASL: %w", err)
	}

	if err := configureTLS(saramaConfig, &kafkaConfig.Security); err != nil {
		return nil, fmt.Errorf("failed to configure TLS: %w", err)
	}

	// 创建消费者组
	group, err := sarama.NewConsumerGroup(kafkaConfig.Brokers, kafkaConfig.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	consumer := &Consumer{
		config: kafkaConfig,
		group:  group,
		ctx:    ctx,
		cancel: cancel,
	}

	logger.Info("Kafka消费者创建成功")
	return consumer, nil
}

// configureSASL 配置SASL认证
func configureSASL(config *sarama.Config, security *config.KafkaSecurity) error {
	if security.SecurityProtocol == "PLAINTEXT" || security.SecurityProtocol == "" {
		return nil
	}

	// 设置安全协议
	switch security.SecurityProtocol {
	case "SASL_SSL":
		config.Net.SASL.Enable = true
		config.Net.TLS.Enable = true
	case "SASL_PLAINTEXT":
		config.Net.SASL.Enable = true
	case "SSL":
		config.Net.TLS.Enable = true
	default:
		logger.Warn("未知的安全协议，使用默认配置",
			zap.String("protocol", security.SecurityProtocol),
		)
	}

	// 配置SASL机制
	if config.Net.SASL.Enable {
		switch security.SASLMechanism {
		case "PLAIN":
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
			config.Net.SASL.User = security.SASLUsername
			config.Net.SASL.Password = security.SASLPassword
		case "SCRAM-SHA-256":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			config.Net.SASL.User = security.SASLUsername
			config.Net.SASL.Password = security.SASLPassword
		case "SCRAM-SHA-512":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			config.Net.SASL.User = security.SASLUsername
			config.Net.SASL.Password = security.SASLPassword
		default:
			logger.Warn("未知的SASL机制，使用PLAIN",
				zap.String("mechanism", security.SASLMechanism),
			)
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
			config.Net.SASL.User = security.SASLUsername
			config.Net.SASL.Password = security.SASLPassword
		}
	}

	return nil
}

// configureTLS 配置TLS
func configureTLS(config *sarama.Config, security *config.KafkaSecurity) error {
	if security.SecurityProtocol == "PLAINTEXT" || security.SecurityProtocol == "" {
		return nil
	}

	// 配置TLS
	config.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: security.SSLEndpointIdentificationAlgorithm == "none",
	}

	// 加载CA证书
	if security.SSLCALocation != "" {
		caCert, err := os.ReadFile(security.SSLCALocation)
		if err != nil {
			return fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		config.Net.TLS.Config.RootCAs = caCertPool
	}

	// 加载客户端证书和密钥
	if security.SSLCertificateLocation != "" && security.SSLKeyLocation != "" {
		cert, err := tls.LoadX509KeyPair(security.SSLCertificateLocation, security.SSLKeyLocation)
		if err != nil {
			return fmt.Errorf("failed to load client certificate: %w", err)
		}
		config.Net.TLS.Config.Certificates = []tls.Certificate{cert}
	}

	return nil
}

// Start 开始消费
func (c *Consumer) Start(messages chan<- *sarama.ConsumerMessage) error {
	c.handler = &ConsumerHandler{
		messages: messages,
		logger:   logger.With(zap.String("component", "kafka_consumer")),
	}

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				logger.Info("停止消费Kafka消息")
				return
			default:
				if err := c.group.Consume(c.ctx, []string{c.config.Topic}, c.handler); err != nil {
					logger.Error("消费错误", zap.Error(err))
					// 等待一段时间后重试
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	logger.Info("Kafka消费者已启动")
	return nil
}

// Stop 停止消费
func (c *Consumer) Stop() error {
	logger.Info("正在停止Kafka消费者")
	c.cancel()
	if c.group != nil {
		return c.group.Close()
	}
	return nil
}

// Setup 消费者组设置
func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	logger.Debug("消费者组Setup")
	return nil
}

// Cleanup 消费者组清理
func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Debug("消费者组Cleanup")
	return nil
}

// ConsumeClaim 消费消息
func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			// 发送消息到通道
			h.messages <- msg
			// 标记消息已处理
			session.MarkMessage(msg, "")
			
			logger.Debug("接收到Kafka消息",
				zap.String("topic", msg.Topic),
				zap.Int32("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
				zap.Int("key_length", len(msg.Key)),
				zap.Int("value_length", len(msg.Value)),
			)
		case <-session.Context().Done():
			return nil
		}
	}
}
