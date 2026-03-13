package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config 全局配置
type Config struct {
	Service      ServiceConfig      `toml:"service"`
	Data         DataConfig         `toml:"data"`
	CleaningRules CleaningRules     `toml:"cleaning_rules"`
	Logging      LoggingConfig      `toml:"logging"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string `toml:"name"`
	DisplayName string `toml:"display_name"`
	Description string `toml:"description"`
}

// DataConfig 数据源配置
type DataConfig struct {
	OutputDir       string      `toml:"output_dir"`
	ReportPath      string      `toml:"report_path"`
	CSV             CSVConfig   `toml:"csv"`
	Kafka           KafkaConfig `toml:"kafka"`
	KafkaOutput     KafkaConfig `toml:"kafka_output"`
	MaxRowsPerFile  int         `toml:"max_rows_per_file"`
}

// CSVConfig CSV配置
type CSVConfig struct {
	Enabled    bool   `toml:"enabled"`
	InputPath  string `toml:"input_path"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Enabled        bool                    `toml:"enabled"`
	Brokers        []string                `toml:"brokers"`
	Topics         []string                `toml:"topics"`
	ConsumerGroup  string                  `toml:"consumer_group"`
	BatchSize      int                     `toml:"batch_size"`
	BatchTimeoutMS int                     `toml:"batch_timeout_ms"`
	AutoCommit     bool                    `toml:"auto_commit"`
	Security       KafkaSecurity           `toml:"security"`
	TopicRules     map[string]TopicRule    `toml:"topic_rules"`
}

// TopicRule 主题专用清洗规则
type TopicRule struct {
	NumericRanges []NumericRange `toml:"numeric_ranges"`
}

// KafkaSecurity Kafka安全配置
type KafkaSecurity struct {
	SASLUsername                      string `toml:"sasl_username"`
	SASLPassword                      string `toml:"sasl_password"`
	SSLCALocation                     string `toml:"ssl_ca_location"`
	SSLKeyLocation                    string `toml:"ssl_key_location"`
	SSLCertificateLocation            string `toml:"ssl_certificate_location"`
	SecurityProtocol                  string `toml:"security_protocol"`
	SASLMechanism                      string `toml:"sasl_mechanism"`
	SSLEndpointIdentificationAlgorithm string `toml:"ssl_endpoint_identification_algorithm"`
}

// CleaningRules 清洗规则配置
type CleaningRules struct {
	RemoveDuplicates      bool             `toml:"remove_duplicates"`
	HandleMissingValues   string           `toml:"handle_missing_values"`
	DateFormat            string           `toml:"date_format"`
	StringNormalization   bool             `toml:"string_normalization"`
	NumericRanges         []NumericRange   `toml:"numeric_ranges"`
}

// NumericRange 数值范围配置
type NumericRange struct {
	Column string  `toml:"column"`
	Min    float64 `toml:"min"`
	Max    float64 `toml:"max"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level       string `toml:"level"`
	FilePath    string `toml:"file_path"`
	MaxSizeMB   int    `toml:"max_size_mb"`
	MaxBackups  int    `toml:"max_backups"`
	MaxAgeDays  int    `toml:"max_age_days"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	var cfg Config
	
	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)
	
	// 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	if cfg.Data.CSV.InputPath == "" {
		cfg.Data.CSV.InputPath = "./data/input.csv"
	}
	if cfg.Data.OutputDir == "" {
		cfg.Data.OutputDir = "./data"
	}
	if cfg.Data.ReportPath == "" {
		cfg.Data.ReportPath = "./data/report.json"
	}
	if cfg.Data.Kafka.BatchSize <= 0 {
		cfg.Data.Kafka.BatchSize = 100
	}
	if cfg.Data.Kafka.BatchTimeoutMS <= 0 {
		cfg.Data.Kafka.BatchTimeoutMS = 5000
	}
	if cfg.Data.MaxRowsPerFile <= 0 {
		cfg.Data.MaxRowsPerFile = 1000000
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	// 必须启用Kafka数据源
	if !cfg.Data.Kafka.Enabled {
		return fmt.Errorf("kafka must be enabled")
	}

	// 验证Kafka配置
	if len(cfg.Data.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers cannot be empty")
	}
	if len(cfg.Data.Kafka.Topics) == 0 {
		return fmt.Errorf("kafka topics cannot be empty")
	}
	if cfg.Data.Kafka.ConsumerGroup == "" {
		return fmt.Errorf("kafka consumer group cannot be empty")
	}

	return nil
}

// EnsureDirectories 确保必要的目录存在
func (cfg *Config) EnsureDirectories() error {
	dirs := []string{
		cfg.Data.OutputDir,
		"logs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
