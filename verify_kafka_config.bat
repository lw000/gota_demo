@echo off
chcp 65001 >nul
echo ========================================
echo Kafka配置验证脚本
echo ========================================
echo.

if not exist "config.toml" (
    echo [错误] config.toml 不存在
    pause
    exit /b 1
)

echo 检查Kafka配置...
echo.

:: 使用go程序验证配置
go run -mod=mod -c "package main; import(`fmt`); import(`github.com/BurntSushi/toml`); import(`os`); func main() { type Config struct { Kafka struct { Enabled bool `toml:\"enabled\"` Security struct { Protocol string `toml:\"security_protocol\"` SASLMechanism string `toml:\"sasl_mechanism\"` } `toml:\"security\"` } `toml:\"kafka\"` }; data, _ := os.ReadFile(`config.toml`); var cfg Config; toml.Unmarshal(data, ^&cfg); fmt.Printf(\"Kafka Enabled: %%v\n\", cfg.Kafka.Enabled); fmt.Printf(\"Security Protocol: %%s\n\", cfg.Kafka.Security.Protocol); fmt.Printf(\"SASL Mechanism: %%s\n\", cfg.Kafka.Security.SASLMechanism) }"

if %ERRORLEVEL% NEQ 0 (
    echo [错误] 配置验证失败
    pause
    exit /b 1
)

echo.
echo ========================================
pause
