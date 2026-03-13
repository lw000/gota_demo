# 数据清洗服务

基于Go语言开发的数据清洗服务，支持CSV文件和Kafka数据源，提供数据清洗、报告生成等功能。

## 功能特性

- **多数据源支持**
  - CSV文件加载
  - Kafka流式数据消费
  - 可扩展支持其他数据源

- **数据清洗功能**
  - 去除重复数据
  - 缺失值处理（删除/填充）
  - 数值范围校验
  - 字符串规范化

- **日志记录**
  - 基于uber-go/zap的高性能结构化日志
  - 支持多种日志级别
  - 日志文件自动轮转

- **清洗报告**
  - JSON格式详细报告
  - 清洗前后数据对比
  - 操作记录追踪
  - 数据质量指标

- **Kafka支持**
  - 支持SASL认证（PLAIN、SCRAM-SHA-256、SCRAM-SHA-512）
  - 支持SSL/TLS加密
  - 支持双向证书认证
  - 支持消费者组

## 技术栈

- `github.com/go-gota/gota` - DataFrame数据处理
- `github.com/judwhite/go-svc` - Windows/Linux服务框架
- `go.uber.org/zap` - 高性能日志库
- `github.com/BurntSushi/toml` - TOML配置文件解析
- `github.com/IBM/sarama` - Kafka客户端

## 项目结构

```
data-cleaning-service/
├── cmd/
│   └── service/
│       └── main.go              # 服务入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── loader/
│   │   ├── csv_loader.go        # CSV加载器
│   │   ├── kafka_loader.go      # Kafka加载器
│   │   └── batch_loader.go      # 批量加载器
│   ├── cleaner/
│   │   └── data_cleaner.go      # 数据清洗器
│   ├── reporter/
│   │   └── report_generator.go  # 报告生成器
│   └── service/
│       └── cleaning_service.go  # 服务核心逻辑
├── pkg/
│   ├── logger/
│   │   └── logger.go            # 日志封装
│   └── kafka/
│       ├── consumer.go          # Kafka消费者
│       └── producer.go          # Kafka生产者
├── data/                        # 数据目录
├── logs/                        # 日志目录
├── config.toml                  # 配置文件
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置文件

编辑 `config.toml` 配置文件：

```toml
# CSV模式配置
[data.csv]
enabled = true
input_path = "./data/input.csv"

# Kafka模式配置
[data.kafka]
enabled = false
brokers = ["localhost:9092"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
```

### 3. 运行服务

```bash
# 使用默认配置文件
go run cmd/service/main.go

# 指定配置文件
go run cmd/service/main.go -config /path/to/config.toml
```

### 4. 编译部署

```bash
# Windows
go build -o data-cleaning-service.exe cmd/service/main.go

# Linux/Mac
go build -o data-cleaning-service cmd/service/main.go

# 运行编译后的程序
./data-cleaning-service
```

## 配置说明

### 服务配置

```toml
[service]
name = "DataCleaningService"
display_name = "数据清洗服务"
description = "CSV数据清洗服务"
```

### 数据源配置

#### CSV数据源

```toml
[data.csv]
enabled = true
input_path = "./data/input.csv"
```

#### Kafka数据源

```toml
[data.kafka]
enabled = false
brokers = ["localhost:9092"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000
auto_commit = true
```

#### Kafka安全配置

详细的Kafka安全配置示例请参考 [KAFKA_SECURITY_CONFIG.md](./KAFKA_SECURITY_CONFIG.md)

**常用配置示例：**

**无加密（开发环境）：**
```toml
[data.kafka.security]
security_protocol = "PLAINTEXT"
```

**SASL_SSL + PLAIN认证（生产环境常用）：**
```toml
[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "PLAIN"
sasl_username = "your_username"
sasl_password = "your_password"
ssl_ca_location = "/path/to/ca.pem"
ssl_endpoint_identification_algorithm = "https"
```

**参数说明：**
- `security_protocol`: 安全协议
  - `PLAINTEXT` - 无加密（开发环境）
  - `SASL_SSL` - SASL认证 + SSL加密（生产环境推荐）
  - `SSL` - 仅SSL加密
- `sasl_mechanism`: SASL认证机制（SASL_SSL时有效）
  - `PLAIN` - 明文认证
  - `SCRAM-SHA-256` - SHA-256哈希认证
  - `SCRAM-SHA-512` - SHA-512哈希认证
- `ssl_ca_location`: CA证书路径
- `ssl_endpoint_identification_algorithm`: SSL主机名验证
  - `none` - 跳过验证（自签名证书）
  - `https` - 验证主机名（推荐）

### 清洗规则配置

```toml
[cleaning_rules]
remove_duplicates = true
handle_missing_values = "delete"  # delete|fill_mean|fill_zero|fill_custom
date_format = "2006-01-02"
string_normalization = true

[[cleaning_rules.numeric_ranges]]
column = "age"
min = 0
max = 120
```

### 日志配置

```toml
[logging]
level = "info"  # debug|info|warn|error
file_path = "./logs/service.log"
max_size_mb = 100
max_backups = 3
max_age_days = 30
```

## Kafka消息格式

输入消息格式（JSON）：

```json
{
  "data": {
    "id": 1,
    "name": "张三",
    "age": 25,
    "email": "zhangsan@example.com",
    "salary": 8000,
    "department": "研发部"
  },
  "metadata": {
    "timestamp": "2026-03-13T10:00:00Z",
    "source": "system_a"
  }
}
```

## 清洗报告示例

```json
{
  "timestamp": "2026-03-13T10:00:00Z",
  "data_source": "csv",
  "data_source_info": {
    "type": "csv",
    "file_path": "./data/input.csv"
  },
  "cleaning_summary": {
    "original_rows": 20,
    "cleaned_rows": 17,
    "removed_rows": 3,
    "original_columns": 7,
    "cleaned_columns": 7
  },
  "operations_performed": [
    {
      "operation": "remove_duplicates",
      "affected_rows": 3,
      "description": "去除重复行",
      "timestamp": "2026-03-13T10:00:01Z"
    }
  ],
  "data_quality": {
    "completeness": 85.0,
    "data_loss": 15.0
  },
  "statistics": {
    "processing_time_ms": 150
  }
}
```

## 开发说明

### 添加新的数据源

1. 实现 `DataLoader` 接口：

```go
type DataLoader interface {
    Load() (dataframe.DataFrame, error)
    Close() error
    GetSourceInfo() map[string]interface{}
}
```

2. 在 `config.go` 中添加配置结构

3. 在 `batch_loader.go` 中添加数据源选择逻辑

### 添加新的清洗规则

1. 在 `cleaning_rules` 中添加规则配置

2. 在 `data_cleaner.go` 中实现清洗逻辑

## License

MIT License
