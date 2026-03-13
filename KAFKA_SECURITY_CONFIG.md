# Kafka安全配置示例

本文档提供了Kafka不同安全场景的配置示例。

## 1. 无加密（开发/测试环境）

适用于本地开发或测试环境，Kafka服务器不启用加密。

```toml
[data.kafka]
enabled = true
brokers = ["localhost:9092"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000
auto_commit = true

[data.kafka.security]
security_protocol = "PLAINTEXT"
sasl_mechanism = ""
```

## 2. SASL_SSL + PLAIN认证

最常用的生产环境配置，使用SASL PLAIN机制和SSL加密。

```toml
[data.kafka]
enabled = true
brokers = ["kafka1.example.com:9093", "kafka2.example.com:9093"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000
auto_commit = true

[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "PLAIN"
sasl_username = "your_username"
sasl_password = "your_password"
ssl_ca_location = "/path/to/ca.pem"
ssl_endpoint_identification_algorithm = "https"
```

## 3. SASL_SSL + SCRAM-SHA-256认证

更安全的SASL认证方式，Kafka服务器配置为SCRAM-SHA-256。

```toml
[data.kafka]
enabled = true
brokers = ["kafka1.example.com:9093"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000
auto_commit = true

[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "SCRAM-SHA-256"
sasl_username = "your_username"
sasl_password = "your_password"
ssl_ca_location = "/path/to/ca.pem"
ssl_endpoint_identification_algorithm = "https"
```

## 4. SASL_SSL + SCRAM-SHA-512认证

最安全的SASL认证方式，Kafka服务器配置为SCRAM-SHA-512。

```toml
[data.kafka]
enabled = true
brokers = ["kafka1.example.com:9093"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000
auto_commit = true

[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "SCRAM-SHA-512"
sasl_username = "your_username"
sasl_password = "your_password"
ssl_ca_location = "/path/to/ca.pem"
ssl_endpoint_identification_algorithm = "https"
```

## 5. 双向SSL认证（客户端证书）

最高安全级别，需要客户端证书和私钥。

```toml
[data.kafka]
enabled = true
brokers = ["kafka1.example.com:9093"]
topic = "raw_data"
consumer_group = "data_cleaning_group"
batch_size = 100
batch_timeout_ms = 5000
auto_commit = true

[data.kafka.security]
security_protocol = "SSL"
ssl_ca_location = "/path/to/ca.pem"
ssl_key_location = "/path/to/client.key"
ssl_certificate_location = "/path/to/client.crt"
ssl_endpoint_identification_algorithm = "https"
```

## 配置参数说明

### security_protocol（安全协议）
- `PLAINTEXT` - 无加密，明文传输
- `SASL_SSL` - SASL认证 + SSL加密（推荐）
- `SSL` - 仅SSL加密，无SASL认证
- `SASL_PLAINTEXT` - SASL认证，无SSL加密（不推荐）

### sasl_mechanism（SASL认证机制）
当 `security_protocol` 为 `SASL_SSL` 或 `SASL_PLAINTEXT` 时有效：
- `PLAIN` - 明文传输用户名密码（适用于SASL_SSL）
- `SCRAM-SHA-256` - SHA-256哈希认证（更安全）
- `SCRAM-SHA-512` - SHA-512哈希认证（最安全）

### ssl_endpoint_identification_algorithm（SSL主机名验证）
- `none` - 跳过主机名验证（适用于自签名证书）
- `https` - 验证主机名（推荐，防止中间人攻击）

### 证书文件路径
- `ssl_ca_location` - CA证书文件路径（.pem格式）
- `ssl_key_location` - 客户端私钥文件路径（.key格式）
- `ssl_certificate_location` - 客户端证书文件路径（.crt或.pem格式）

## 证书文件获取

### 从Kafka管理员获取
1. CA证书（`.pem`文件）
2. 如果需要双向认证：
   - 客户端证书（`.crt`或`.pem`）
   - 客户端私钥（`.key`）

### 格式说明
- CA证书通常以 `-----BEGIN CERTIFICATE-----` 开头
- 私钥以 `-----BEGIN PRIVATE KEY-----` 或 `-----BEGIN RSA PRIVATE KEY-----` 开头

## 故障排查

### 连接超时
1. 检查 `brokers` 地址和端口是否正确
2. 确认防火墙规则允许访问Kafka端口
3. 检查网络连接

### 认证失败
1. 确认 `security_protocol` 与Kafka服务器配置一致
2. 验证用户名和密码是否正确
3. 检查SASL机制是否匹配

### SSL/TLS错误
1. 确认CA证书路径正确且文件存在
2. 检查证书是否过期
3. 如使用自签名证书，设置 `ssl_endpoint_identification_algorithm = "none"`

### 权限问题
1. 检查证书文件读取权限
2. 确认用户对Topic有读取权限
3. 验证Consumer Group名称是否可用

## 云服务商配置参考

### 阿里云Kafka
```toml
[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "PLAIN"
sasl_username = "your_instance_username"
sasl_password = "your_instance_password"
ssl_ca_location = ""  # 阿里云通常不需要
ssl_endpoint_identification_algorithm = "https"
```

### 腾讯云Kafka
```toml
[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "PLAIN"
sasl_username = "your_instance_username"
sasl_password = "your_instance_password"
ssl_ca_location = ""  # 腾讯云通常不需要
ssl_endpoint_identification_algorithm = "https"
```

### AWS MSK
```toml
[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "SCRAM-SHA-512"
sasl_username = "your_iam_user"
sasl_password = "your_iam_password"
ssl_ca_location = "/path/to/aws-ca.pem"
ssl_endpoint_identification_algorithm = "https"
```

### Confluent Cloud
```toml
[data.kafka.security]
security_protocol = "SASL_SSL"
sasl_mechanism = "PLAIN"
sasl_username = "your_api_key"
sasl_password = "your_api_secret"
ssl_ca_location = ""  # Confluent自带CA证书
ssl_endpoint_identification_algorithm = "https"
```

## 安全建议

1. **生产环境务必使用加密**
   - 使用 `SASL_SSL` 或 `SSL`
   - 启用主机名验证

2. **定期更换凭证**
   - 定期更换用户名密码
   - 轮换客户端证书

3. **最小权限原则**
   - Consumer只授予所需Topic的读取权限
   - 使用专用的Consumer Group

4. **监控告警**
   - 监控认证失败日志
   - 设置连接异常告警

5. **证书管理**
   - 及时更新即将过期的证书
   - 安全存储私钥文件
