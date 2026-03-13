package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"data-cleaning-service/internal/config"
	"data-cleaning-service/internal/service"
	"data-cleaning-service/pkg/logger"

	"github.com/judwhite/go-svc"
	"go.uber.org/zap"
)

var (
	configPath string
)

func init() {
	// 默认配置文件路径
	defaultConfigPath := "./config.toml"

	flag.StringVar(&configPath, "config", defaultConfigPath, "配置文件路径")
	flag.StringVar(&configPath, "c", defaultConfigPath, "配置文件路径（短选项）")
}

func main() {
	flag.Parse()

	// 获取可执行文件目录
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法获取可执行文件路径: %v\n", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)

	// 如果配置路径是相对路径，转换为绝对路径
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(execDir, configPath)
	}

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	if err := logger.InitLogger(
		cfg.Logging.Level,
		cfg.Logging.FilePath,
		cfg.Logging.MaxSizeMB,
		cfg.Logging.MaxBackups,
		cfg.Logging.MaxAgeDays,
	); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("========================================")
	logger.Info("数据清洗服务启动",
		zap.String("name", cfg.Service.Name),
		zap.String("version", "1.0.0"),
		zap.String("config", configPath),
	)
	logger.Info("========================================")

	// 创建服务实例
	svcInstance := service.NewCleaningService(cfg)

	// 根据平台运行服务
	if err := svc.Run(svcInstance); err != nil {
		logger.Fatal("服务运行失败", zap.String("error", err.Error()))
		os.Exit(1)
	}
}

// program 实现svc.Service接口
type program struct {
	service *service.CleaningService
}

func (p *program) Init(env svc.Environment) error {
	return p.service.Init(env)
}

func (p *program) Start() error {
	return p.service.Start()
}

func (p *program) Stop() error {
	return p.service.Stop()
}

// handleSignals 处理系统信号
func handleSignals(stopChan chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Infof("接收到信号: %v", sig)
		close(stopChan)
	case <-stopChan:
	}
}
