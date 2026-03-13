package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// logger 全局logger实例
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// InitLogger 初始化日志系统
func InitLogger(level string, filePath string, maxSizeMB, maxBackups, maxAgeDays int) error {
	// 解析日志级别
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件输出
	fileWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   true,
	}

	// 控制台输出
	consoleWriter := zapcore.Lock(os.Stdout)

	// 创建Core
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		logLevel,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		consoleWriter,
		logLevel,
	)

	// 合并Core
	core := zapcore.NewTee(fileCore, consoleCore)

	// 创建logger
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = logger.Sugar()

	return nil
}

// Debug 调试级别日志
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Debugf 调试级别格式化日志
func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

// Info 信息级别日志
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Infof 信息级别格式化日志
func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

// Warn 警告级别日志
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Warnf 警告级别格式化日志
func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

// Error 错误级别日志
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Errorf 错误级别格式化日志
func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

// Fatal 致命错误级别日志
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// Fatalf 致命错误级别格式化日志
func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}

// With 创建带有字段的logger
func With(fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}

// Sync 同步日志缓冲区
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// String 字符串类型
type String string

// String 转换为字符串
func (s String) String() string {
	return string(s)
}

// Int 整数类型
type Int int

// Int 转换为整数
func (i Int) Int() int {
	return int(i)
}

// Any 任意类型
type Any interface{}
