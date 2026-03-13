# 快速入门指南

## 5分钟快速上手

### 步骤1: 安装依赖

```bash
go mod download
```

### 步骤2: 配置Kafka数据源

编辑 `config.toml`，配置Kafka连接：

```toml
[data.kafka]
enabled = true
brokers = ["localhost:9092"]
topics = ["your_topic", "another_topic"]  # 订阅多个主题
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000
```

详细的Kafka配置请参考 [KAFKA_SECURITY_CONFIG.md](./KAFKA_SECURITY_CONFIG.md)

### 步骤3: 配置输出文件

编辑 `config.toml`，配置输出目录和文件分割：

```toml
[data]
output_dir = "./data"
max_rows_per_file = 1000000  # 每个CSV文件最多100万行
```

### 步骤4: 准备Kafka消息

**Kafka消息格式（扁平JSON）：**
```json
{
  "TAG_001": 111,
  "TAG_002": 11.31,
  "TAG_003": true
}
```

### 步骤5: 配置清洗规则

编辑 `config.toml` 中的清洗规则：

```toml
[cleaning_rules]
remove_duplicates = true              # 去除重复数据
handle_missing_values = "delete"      # 删除包含缺失值的行
string_normalization = true           # 字符串规范化

# 全局数值范围校验（适用于所有主题）
[[cleaning_rules.numeric_ranges]]
column = "TAG_001"
min = 0
max = 1000

# 主题专用数值范围校验（覆盖全局规则）
[data.kafka.topic_rules]
[data.kafka.topic_rules.your_topic]
[[data.kafka.topic_rules.your_topic.numeric_ranges]]
column = "TAG_001"
min = 0
max = 1000
[[data.kafka.topic_rules.your_topic.numeric_ranges]]
column = "TAG_002"
min = 0.0
max = 100.0
```

### 步骤5: 编译程序

**Windows:**
```bash
build.bat
```

**Linux/Mac:**
```bash
chmod +x build_linux.sh
./build_linux.sh
```

### 步骤6: 运行服务

**Windows:**
```bash
run.bat
```

**Linux/Mac:**
```bash
./data-cleaning-service -config config.toml
```

### 步骤7: 查看结果

运行后，持续消费Kafka数据并输出到CSV文件：

1. **清洗后的数据**: `./data/{topic}_{timestamp}.csv`
   - 文件命名: `{topic}_{timestamp}.csv`
   - 每个主题独立保存到不同的文件
   - 示例: `your_topic_1710345600.csv`, `another_topic_1710345700.csv`
   - 当达到最大行数时自动创建新文件

2. **运行日志**: `./logs/service.log`

3. **监控日志输出**:
```
INFO 开始持续消费Kafka数据 topics=["your_topic", "another_topic"] max_rows_per_file=1000000
INFO 创建新的CSV文件 topic=your_topic path=./data/your_topic_1710345600.csv
INFO 批次处理完成 topic=your_topic rows=100 total_rows=100
INFO 批次处理完成 topic=your_topic rows=100 total_rows=200
INFO 创建新的CSV文件 topic=another_topic path=./data/another_topic_1710345700.csv
INFO 批次处理完成 topic=another_topic rows=100 total_rows=100
...
```

## 常见问题

### Q: 编译失败，提示找不到依赖
**A:** 运行以下命令：
```bash
go mod tidy
```

### Q: 运行时提示配置文件不存在
**A:** 确保 `config.toml` 在当前目录，或使用 `-config` 参数指定完整路径：
```bash
./data-cleaning-service -config /path/to/config.toml
```

### Q: Kafka连接失败
**A:** 检查以下几点：
1. Kafka服务是否运行
2. brokers地址和端口是否正确
3. 防火墙是否允许连接
4. 安全配置是否正确（见 KAFKA_SECURITY_CONFIG.md）

### Q: 没有生成CSV文件
**A:** 可能原因：
1. Kafka topic中没有消息
2. 所有数据都不符合清洗规则（如数值范围）
3. 所有行都有缺失值且配置为删除
4. 检查日志了解详情

## 下一步

- 详细配置: [README.md](./README.md)
- Kafka安全配置: [KAFKA_SECURITY_CONFIG.md](./KAFKA_SECURITY_CONFIG.md)
- 项目总结: [PROJECT_SUMMARY.md](./PROJECT_SUMMARY.md)

## 获取帮助

如遇问题，请：
1. 查看 `logs/service.log` 日志文件
2. 将日志级别设置为 `debug` 查看详细信息
3. 参考项目文档

## 示例场景

### 场景1: 多主题IoT数据清洗

同时订阅多个传感器主题，进行清洗和存储。

**配置：**
```toml
[data.kafka]
enabled = true
topics = ["sensor-data", "gateway-data"]  # 订阅两个主题
batch_size = 100

[data]
max_rows_per_file = 1000000

# 全局清洗规则
[cleaning_rules]
remove_duplicates = true
handle_missing_values = "delete"
string_normalization = true

# sensor-data主题专用规则
[data.kafka.topic_rules.sensor-data]
[[data.kafka.topic_rules.sensor-data.numeric_ranges]]
column = "temperature"
min = -50
max = 150

# gateway-data主题专用规则
[data.kafka.topic_rules.gateway-data]
[[data.kafka.topic_rules.gateway-data.numeric_ranges]]
column = "signal_strength"
min = -100
max = 0
```

**Kafka消息（sensor-data）：**
```json
{
  "sensor_id": "S001",
  "temperature": 25.5,
  "humidity": 60.2,
  "timestamp": 1710345600
}
```

**Kafka消息（gateway-data）：**
```json
{
  "gateway_id": "G001",
  "signal_strength": -60,
  "device_count": 10,
  "timestamp": 1710345600
}
```

**结果：**
- 同时消费两个主题的数据
- 每个主题应用不同的数值范围校验
- 两个主题的数据分别保存到独立文件：
  - `sensor-data_1710345600.csv`
  - `gateway-data_1710345700.csv`
- 每100万条数据自动分割文件

### 场景2: 日志数据持续清洗

清洗应用日志数据，输出到CSV文件。

**配置：**
```toml
[data.kafka]
enabled = true
topics = ["app-logs", "system-logs"]  # 同时订阅应用日志和系统日志
batch_size = 500
batch_timeout_ms = 10000

[data]
max_rows_per_file = 500000

[cleaning_rules]
remove_duplicates = true
handle_missing_values = "delete"
string_normalization = true

# app-logs主题专用规则
[data.kafka.topic_rules.app-logs]
[[data.kafka.topic_rules.app-logs.numeric_ranges]]
column = "user_id"
min = 1
max = 999999
```

**Kafka消息（app-logs）：**
```json
{
  "level": "INFO",
  "message": "User login",
  "user_id": 12345,
  "timestamp": 1710345600
}
```

**Kafka消息（system-logs）：**
```json
{
  "level": "ERROR",
  "message": "Disk space low",
  "disk_usage": 95.5,
  "timestamp": 1710345600
}
```

**结果：**
- 批量处理日志数据（每500条或10秒）
- 每个文件最多50万条记录
- app-logs主题会校验user_id范围
- 两个主题的数据分别存储：
  - `app-logs_1710345600.csv`
  - `system-logs_1710345700.csv`
