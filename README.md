# 数据清洗服务

基于Go语言开发的Kafka数据清洗服务，支持持续消费Kafka数据，进行数据清洗，并输出到CSV文件。

## 功能特性

- **Kafka持续消费**
  - 实时消费Kafka主题数据
  - 批量处理机制
  - 支持消费者组
  - 自动提交offset

- **数据清洗功能**
  - 去除重复数据
  - 缺失值处理（删除/填充）
  - 数值范围校验
  - 字符串规范化

- **CSV文件输出**
  - 自动分割文件（可配置每文件最大行数）
  - 列名按字典序排序，确保数据一致性
  - 文件命名规则：{topic}_{timestamp}.csv

- **日志记录**
  - 基于uber-go/zap的高性能结构化日志
  - 支持多种日志级别
  - 日志文件自动轮转

- **Kafka安全支持**
  - 支持SASL认证（PLAIN、SCRAM-SHA-256、SCRAM-SHA-512）
  - 支持SSL/TLS加密
  - 支持双向证书认证

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
│   │   └── kafka_loader.go      # Kafka加载器
│   ├── cleaner/
│   │   └── data_cleaner.go      # 数据清洗器
│   └── service/
│       └── cleaning_service.go  # 服务核心逻辑
├── pkg/
│   ├── logger/
│   │   └── logger.go            # 日志封装
│   └── kafka/
│       └── consumer.go          # Kafka消费者
├── data/                        # 数据输出目录
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
# Kafka配置
[data.kafka]
enabled = true
brokers = ["localhost:9092"]
topic = "your_topic"
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000

# 输出配置
[data]
output_dir = "./data"
max_rows_per_file = 1000000
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
description = "Kafka数据清洗服务"
```

### 数据源配置

#### Kafka数据源（必须启用）

```toml
[data.kafka]
enabled = true
brokers = ["localhost:9092"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
batch_size = 100              # 每批次处理的消息数
batch_timeout_ms = 5000        # 批次超时时间（毫秒）
auto_commit = true             # 自动提交offset
```

#### 输出文件配置

```toml
[data]
output_dir = "./data"                    # 输出目录
max_rows_per_file = 1000000              # 每个CSV文件最大行数
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

输入消息格式（扁平JSON）：

```json
{
  "TAG_001": 111,
  "TAG_002": 11.31,
  "TAG_003": true
}
```

## 输出文件说明

### CSV文件命名规则

- 格式: `{topic}_{timestamp}.csv`
- 示例: `raw_data_1710345600.csv`
- 时间戳: Unix时间戳（秒）

### CSV文件分割

- 当单个文件行数达到 `max_rows_per_file` 时，自动创建新文件
- 新文件使用当前时间戳命名
- 所有文件的列名按字典序排列，保证数据一致性

### 输出示例

假设topic为`test-topic-1`，输出文件如下：

```
data/
├── test-topic-1_1710345600.csv  # 第一个文件（100万条）
├── test-topic-1_1710346600.csv  # 第二个文件（100万条）
└── test-topic-1_1710347600.csv  # 第三个文件（不满100万条）
```

CSV内容示例：
```csv
TAG_001,TAG_002,TAG_003
111,11.31,true
222,22.62,false
333,33.93,true
```

## 开发说明

### 添加新的清洗规则

1. 在 `cleaning_rules` 中添加规则配置

2. 在 `data_cleaner.go` 中实现清洗逻辑

## License

MIT License
