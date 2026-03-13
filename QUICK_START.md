# 快速入门指南

## 5分钟快速上手

### 步骤1: 安装依赖

```bash
go mod download
```

### 步骤2: 配置数据源

#### 使用CSV文件（默认配置）

编辑 `config.toml`，确保CSV配置已启用：

```toml
[data.csv]
enabled = true
input_path = "./data/input.csv"
```

#### 使用Kafka（可选）

编辑 `config.toml`：

```toml
[data.csv]
enabled = false

[data.kafka]
enabled = true
brokers = ["localhost:9092"]
topic = "your_topic"
consumer_group = "data_cleaning_group"
```

详细的Kafka配置请参考 [KAFKA_SECURITY_CONFIG.md](./KAFKA_SECURITY_CONFIG.md)

### 步骤3: 准备输入数据

**CSV格式示例：**
```csv
id,name,age,email,salary,department
1,张三,25,zhangsan@example.com,8000,研发部
2,李四,30,lisi@example.com,12000,市场部
3,王五,28,wangwu@example.com,10000,研发部
```

**Kafka消息格式：**
```json
{
  "data": {
    "id": 1,
    "name": "张三",
    "age": 25
  },
  "metadata": {
    "timestamp": "2026-03-13T10:00:00Z"
  }
}
```

### 步骤4: 配置清洗规则

编辑 `config.toml` 中的清洗规则：

```toml
[cleaning_rules]
remove_duplicates = true              # 去除重复数据
handle_missing_values = "delete"      # 删除包含缺失值的行
string_normalization = true           # 字符串规范化

# 数值范围校验
[[cleaning_rules.numeric_ranges]]
column = "age"
min = 0
max = 120

[[cleaning_rules.numeric_ranges]]
column = "salary"
min = 0
max = 1000000
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

运行完成后，查看以下文件：

1. **清洗后的数据**: `data/output.csv`
2. **清洗报告**: `data/report.json`
3. **运行日志**: `logs/service.log`

报告会在控制台输出，格式如下：

```
========== 数据清洗报告 ==========
时间: 2026-03-13 10:00:00
数据源: csv

---------- 清洗摘要 ----------
原始行数: 20
清洗后行数: 17
删除行数: 3

---------- 数据质量 ----------
completeness: 85.00%

---------- 清洗操作 ----------
- remove_duplicates: 去除重复行 (影响 3 行)

---------- 统计信息 ----------
处理时间: 150.50 ms
================================
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

### Q: CSV文件读取失败
**A:** 确保CSV文件：
1. 格式正确（逗号分隔）
2. 第一行是表头
3. 文件编码为UTF-8
4. 路径配置正确

### Q: 清洗后数据为空
**A:** 可能原因：
1. 所有数据都不符合清洗规则（如数值范围）
2. 所有行都有缺失值且配置为删除
3. 检查日志了解详情

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

### 场景1: 清洗客户数据

输入CSV包含：重复客户、无效邮箱、错误年龄

**配置：**
```toml
[cleaning_rules]
remove_duplicates = true
handle_missing_values = "delete"
string_normalization = true

[[cleaning_rules.numeric_ranges]]
column = "age"
min = 0
max = 120
```

**结果：**
- 删除重复记录
- 删除年龄不合理的记录
- 统一邮箱格式（小写）
- 删除缺失邮箱或年龄的记录

### 场景2: 实时Kafka数据清洗

**配置：**
```toml
[data.kafka]
enabled = true
batch_size = 100
batch_timeout_ms = 5000

[cleaning_rules]
remove_duplicates = true
handle_missing_values = "delete"
```

**结果：**
- 实时消费Kafka消息
- 批量处理数据（每100条或5秒）
- 持续运行，适合生产环境

### 场景3: 数据质量报告

**配置：**
```toml
[cleaning_rules]
# 不进行清洗，仅生成报告
remove_duplicates = false
handle_missing_values = "none"
string_normalization = false
```

**结果：**
- 生成完整的数据质量报告
- 包含每列的统计信息
- 可用于数据审计
