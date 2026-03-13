# 数据清洗服务 - 项目总结

## 项目概述

基于Go语言开发的数据清洗服务，支持CSV文件和Kafka数据源，提供数据清洗、报告生成等功能。

## 已实现功能

### ✅ 核心功能

1. **多数据源支持**
   - ✅ CSV文件加载
   - ✅ Kafka流式数据消费
   - ✅ 可扩展的数据加载器接口

2. **数据清洗功能**
   - ✅ 去除重复数据
   - ✅ 删除包含缺失值的行
   - ✅ 数值范围校验
   - ✅ 字符串规范化（去空格、转小写）
   - 🔲 缺失值填充（均值、零值、自定义值）- 框架已建立，待实现

3. **日志记录**
   - ✅ 基于uber-go/zap的高性能结构化日志
   - ✅ 支持多种日志级别（Debug/Info/Warn/Error）
   - ✅ 日志文件自动轮转
   - ✅ 同时输出到文件和控制台

4. **清洗报告**
   - ✅ JSON格式详细报告
   - ✅ 清洗前后数据对比
   - ✅ 操作记录追踪
   - ✅ 数据质量指标
   - ✅ 列级统计信息

5. **Kafka支持**
   - ✅ Kafka消费者（数据输入）
   - ✅ Kafka生产者（数据输出）
   - ✅ 批量数据处理
   - ✅ SASL认证（PLAIN, SCRAM-SHA-256, SCRAM-SHA-512）
   - ✅ SSL/TLS加密
   - ✅ 双向证书认证
   - ✅ 消费者组支持
   - ✅ 自动提交offset

## 项目结构

```
data-cleaning-service/
├── cmd/service/main.go              # 服务入口
├── internal/
│   ├── config/config.go             # 配置管理（TOML）
│   ├── loader/
│   │   ├── csv_loader.go           # CSV加载器
│   │   ├── kafka_loader.go         # Kafka加载器
│   │   └── batch_loader.go         # 批量加载器
│   ├── cleaner/data_cleaner.go      # 数据清洗器
│   ├── reporter/
│   │   └── report_generator.go     # 报告生成器
│   └── service/
│       └── cleaning_service.go     # 服务核心逻辑
├── pkg/
│   ├── logger/logger.go             # 日志封装
│   └── kafka/
│       ├── consumer.go             # Kafka消费者
│       └── producer.go             # Kafka生产者
├── data/                         # 数据目录
│   ├── input.csv                 # 示例输入数据
│   ├── output.csv                # 输出数据
│   └── report.json              # 清洗报告
├── logs/                         # 日志目录
├── config.toml                   # 配置文件
├── build.bat                     # Windows编译脚本
├── run.bat                       # Windows运行脚本
├── build_linux.sh               # Linux编译脚本
├── build_cross_platform.bat      # 交叉编译脚本
└── README.md                     # 项目文档
```

## 技术栈

| 库 | 版本 | 用途 |
|------|------|------|
| github.com/go-gota/gota | v0.12.0 | DataFrame数据处理 |
| github.com/judwhite/go-svc | v1.2.1 | Windows/Linux服务框架 |
| go.uber.org/zap | v1.27.1 | 高性能结构化日志 |
| github.com/BurntSushi/toml | v1.6.0 | TOML配置文件解析 |
| github.com/IBM/sarama | v1.47.0 | Kafka客户端 |
| gopkg.in/natefinch/lumberjack.v2 | v2.2.1 | 日志文件轮转 |

## 配置说明

### 数据源配置

#### CSV模式
```toml
[data.csv]
enabled = true
input_path = "./data/input.csv"
```

#### Kafka模式
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

### Kafka安全配置

**无加密（开发环境）：**
```toml
[data.kafka.security]
security_protocol = "PLAINTEXT"
```

**SASL_SSL + PLAIN认证（生产环境）：**
```toml
[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "PLAIN"
sasl_username = "your_username"
sasl_password = "your_password"
ssl_ca_location = "/path/to/ca.pem"
ssl_endpoint_identification_algorithm = "https"
```

详细配置请参考 [KAFKA_SECURITY_CONFIG.md](./KAFKA_SECURITY_CONFIG.md)

## 使用方法

### 1. 编译

**Windows:**
```bash
# 使用编译脚本
build.bat

# 或手动编译
go build -o data-cleaning-service.exe cmd/service/main.go
```

**Linux:**
```bash
# 使用编译脚本
chmod +x build_linux.sh
./build_linux.sh

# 或手动编译
go build -o data-cleaning-service cmd/service/main.go
```

### 2. 运行

**Windows:**
```bash
# 使用运行脚本
run.bat

# 或直接运行
data-cleaning-service.exe -config config.toml
```

**Linux:**
```bash
./data-cleaning-service -config config.toml
```

### 3. 查看结果

- **清洗后的数据**: `./data/output.csv`
- **清洗报告**: `./data/report.json`
- **运行日志**: `./logs/service.log`

## 清洗报告示例

```json
{
  "timestamp": "2026-03-13T10:00:00Z",
  "data_source": "csv",
  "data_source_info": {
    "type": "csv",
    "file_path": "./data/input.csv",
    "file_size": 1234,
    "modified_time": "2026-03-13T09:00:00Z"
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
    "processing_time_ms": 150.5
  },
  "columns_info": [
    {
      "name": "id",
      "type": "String",
      "null_count": 0,
      "unique": 17,
      "statistics": {
        "avg_length": 1.0,
        "max_length": 2,
        "count": 17
      }
    }
  ]
}
```

## 已知问题和限制

### 1. Kafka安全配置
- ✅ 已修复：安全协议配置现在正确支持 `SASL_SSL`、`SSL`、`PLAINTEXT`
- ✅ 已实现：完整的SASL和TLS配置支持

### 2. 缺失值处理
- 🔲 均值填充：框架已建立，待实现计算逻辑
- 🔲 零值填充：框架已建立，待实现
- 🔲 自定义值填充：框架已建立，待实现

### 3. 字符串规范化
- ✅ 当前实现：去除前后空格、统一转小写
- 🔲 可扩展：正则表达式清洗、特殊字符处理

## 扩展建议

### 1. 新增数据源
实现 `DataLoader` 接口：
```go
type DataLoader interface {
    Load() (dataframe.DataFrame, error)
    Close() error
    GetSourceInfo() map[string]interface{}
}
```

### 2. 新增清洗规则
在 `DataCleaner` 中添加新方法：
```go
func (c *DataCleaner) customClean(df dataframe.DataFrame) (dataframe.DataFrame, error) {
    // 实现自定义清洗逻辑
    return df, nil
}
```

### 3. 新增报告格式
在 `ReportGenerator` 中添加：
```go
func (g *ReportGenerator) ExportToExcel(report *Report, filePath string) error {
    // 实现Excel导出逻辑
}
```

## 性能优化建议

1. **大文件处理**
   - 使用流式读取CSV
   - 分批处理数据
   - 考虑使用内存映射

2. **Kafka消费**
   - 调整 `batch_size` 和 `batch_timeout_ms`
   - 使用多个消费者组提高吞吐量
   - 启用自动提交减少网络开销

3. **日志优化**
   - 生产环境使用 `warn` 或 `error` 级别
   - 调整日志轮转策略
   - 考虑使用日志聚合系统

## 部署建议

### Windows服务部署
```powershell
# 安装服务
sc create DataCleaningService binPath= "C:\path\to\data-cleaning-service.exe"
sc start DataCleaningService
```

### Linux systemd服务
```bash
# 创建服务文件
sudo vim /etc/systemd/system/data-cleaning.service
```

```ini
[Unit]
Description=Data Cleaning Service
After=network.target

[Service]
Type=simple
User=datacleaning
WorkingDirectory=/opt/data-cleaning-service
ExecStart=/opt/data-cleaning-service/data-cleaning-service -config /opt/data-cleaning-service/config.toml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```bash
# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable data-cleaning
sudo systemctl start data-cleaning
```

## 监控和运维

### 日志监控
- 检查 `logs/service.log`
- 监控清洗失败率
- 追踪处理时间

### 数据质量监控
- 监控 `completeness` 指标
- 设置阈值告警（低于90%告警）
- 定期检查清洗报告

### 性能监控
- 监控 `processing_time_ms`
- 设置超时告警（>10秒告警）
- 追踪资源使用情况

## 安全建议

1. **配置文件**
   - 不要在配置文件中硬编码密码
   - 使用环境变量或密钥管理工具
   - 设置适当的文件权限

2. **Kafka凭证**
   - 定期轮换用户名密码
   - 使用最小权限原则
   - 启用SSL加密

3. **日志安全**
   - 避免记录敏感数据
   - 定期清理旧日志
   - 加密日志存储

## 版本历史

### v1.0.0 (2026-03-13)
- ✅ 初始版本发布
- ✅ CSV和Kafka数据源支持
- ✅ 基础数据清洗功能
- ✅ 日志和报告功能
- ✅ Kafka安全配置支持

## 联系方式

如有问题或建议，请联系开发团队。

---

**注意**: 本项目基于Go 1.21开发，请确保Go版本 >= 1.21。
