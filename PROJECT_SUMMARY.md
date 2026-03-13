# 数据清洗服务 - 项目总结

## 项目概述

基于Go语言开发的Kafka数据清洗服务，支持持续消费Kafka主题数据，进行数据清洗，并输出到分割的CSV文件中。

## 已实现功能

### ✅ 核心功能

1. **多主题订阅**
   - ✅ 同时订阅多个Kafka主题
   - ✅ 每个主题独立处理和存储
   - ✅ 支持主题专用清洗规则
   - ✅ 共享基础清洗规则

2. **Kafka持续消费**
   - ✅ 实时消费Kafka主题数据
   - ✅ 批量处理机制（可配置批次大小和超时）
   - ✅ 消费者组支持
   - ✅ 自动提交offset
   - ✅ 完整的Kafka安全配置支持

3. **数据清洗功能**
   - ✅ 去除重复数据
   - ✅ 删除包含缺失值的行
   - ✅ 数值范围校验（支持全局和主题专用规则）
   - ✅ 字符串规范化（去空格、转小写）
   - 🔲 缺失值填充（均值、零值、自定义值）- 框架已建立，待实现

4. **CSV文件输出**
   - ✅ 自动分割文件（可配置每文件最大行数）
   - ✅ 按主题分别存储
   - ✅ 列名按字典序排序，确保数据一致性
   - ✅ 文件命名规则：{topic}_{timestamp}.csv
   - ✅ 支持大量数据持续写入

5. **日志记录**
   - ✅ 基于uber-go/zap的高性能结构化日志
   - ✅ 支持多种日志级别（Debug/Info/Warn/Error）
   - ✅ 日志文件自动轮转
   - ✅ 同时输出到文件和控制台

6. **Kafka安全支持**
   - ✅ SASL认证（PLAIN, SCRAM-SHA-256, SCRAM-SHA-512）
   - ✅ SSL/TLS加密
   - ✅ 双向证书认证
   - ✅ 完整的安全配置选项

## 项目结构

```
data-cleaning-service/
├── cmd/service/main.go              # 服务入口
├── internal/
│   ├── config/config.go             # 配置管理（TOML）
│   ├── loader/
│   │   └── kafka_loader.go         # Kafka加载器
│   ├── cleaner/data_cleaner.go      # 数据清洗器
│   └── service/
│       └── cleaning_service.go     # 服务核心逻辑
├── pkg/
│   ├── logger/logger.go             # 日志封装
│   └── kafka/
│       └── consumer.go             # Kafka消费者
├── data/                         # 数据输出目录
│   ├── test-topic-1_1710345600.csv # 输出CSV文件
│   └── ...                       # 更多分割文件
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

### Kafka数据源配置（必须启用）

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

### 输出文件配置

```toml
[data]
output_dir = "./data"                    # 输出目录
max_rows_per_file = 1000000              # 每个CSV文件最大行数
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

- **清洗后的数据**: `./data/{topic}_{timestamp}.csv`
- **运行日志**: `./logs/service.log`

### 输出文件说明

服务会持续消费Kafka数据并输出到CSV文件：
- 文件命名: `{topic}_{timestamp}.csv`
- 当单个文件达到 `max_rows_per_file` 行时自动创建新文件
- 所有文件的列名按字典序排列，确保数据一致性

### Kafka消息格式

输入消息格式（扁平JSON）：

```json
{
  "TAG_001": 111,
  "TAG_002": 11.31,
  "TAG_003": true
}
```

### 输出CSV示例

```csv
TAG_001,TAG_002,TAG_003
111,11.31,true
222,22.62,false
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

### v3.0.0 (2026-03-13)
- ✅ 多主题订阅：支持同时订阅多个Kafka主题
- ✅ 主题专用规则：支持每个主题配置独立的数值范围校验规则
- ✅ 规则合并：主题专用规则可覆盖全局规则
- ✅ 数据隔离：每个主题的数据保存到独立的CSV文件
- ✅ 架构优化：重构Kafka消息处理流程，避免循环导入

### v2.0.0 (2026-03-13)
- ✅ 架构重构：移除CSV文件加载，专注Kafka持续消费
- ✅ CSV文件自动分割：支持按最大行数分割文件
- ✅ 列名字典序排序：确保数据一致性
- ✅ 移除批量报告功能：改为持续消费模式
- ✅ 简化项目结构：删除不必要文件

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
