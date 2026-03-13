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
topic = "your_topic"
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

# 数值范围校验
[[cleaning_rules.numeric_ranges]]
column = "TAG_001"
min = 0
max = 1000
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
   - 示例: `your_topic_1710345600.csv`
   - 当达到最大行数时自动创建新文件

2. **运行日志**: `./logs/service.log`

3. **监控日志输出**:
```
INFO 开始持续消费Kafka数据 topic=your_topic max_rows_per_file=1000000
INFO 创建新的CSV文件 path=./data/your_topic_1710345600.csv
INFO 批次处理完成 rows=100 total_rows=100
INFO 批次处理完成 rows=100 total_rows=200
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

### 场景1: IoT传感器数据清洗

持续消费传感器数据，进行清洗和存储。

**配置：**
```toml
[data.kafka]
enabled = true
topic = "sensor-data"
batch_size = 100

[data]
max_rows_per_file = 1000000

[cleaning_rules]
remove_duplicates = true
handle_missing_values = "delete"
string_normalization = true

[[cleaning_rules.numeric_ranges]]
column = "temperature"
min = -50
max = 150
```

**Kafka消息：**
```json
{
  "sensor_id": "S001",
  "temperature": 25.5,
  "humidity": 60.2,
  "timestamp": 1710345600
}
```

**结果：**
- 实时消费传感器数据
- 去除重复和异常数据
- 每100万条数据创建一个新文件
- 列名按字典序排列（humidity, sensor_id, temperature, timestamp）

### 场景2: 日志数据持续清洗

清洗应用日志数据，输出到CSV文件。

**配置：**
```toml
[data.kafka]
enabled = true
topic = "app-logs"
batch_size = 500
batch_timeout_ms = 10000

[data]
max_rows_per_file = 500000

[cleaning_rules]
remove_duplicates = true
handle_missing_values = "delete"
string_normalization = true
```

**Kafka消息：**
```json
{
  "level": "INFO",
  "message": "User login",
  "user_id": 12345,
  "timestamp": 1710345600
}
```

**结果：**
- 批量处理日志数据（每500条或10秒）
- 每个文件最多50万条记录
- 自动分割文件便于管理
