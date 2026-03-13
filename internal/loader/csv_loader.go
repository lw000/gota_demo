package loader

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-gota/gota/dataframe"
	"data-cleaning-service/pkg/logger"
	"go.uber.org/zap"
)

// DataLoader 数据加载器接口
type DataLoader interface {
	Load() (dataframe.DataFrame, error)
	Close() error
	GetSourceInfo() map[string]interface{}
}

// CSVLoader CSV文件加载器
type CSVLoader struct {
	inputPath string
}

// NewCSVLoader 创建CSV加载器
func NewCSVLoader(inputPath string) *CSVLoader {
	return &CSVLoader{
		inputPath: inputPath,
	}
}

// Load 加载CSV文件
func (l *CSVLoader) Load() (dataframe.DataFrame, error) {
	startTime := time.Now()
	logger.Info("开始加载CSV文件",
		zap.String("file", l.inputPath),
	)

	// 打开文件
	file, err := os.Open(l.inputPath)
	if err != nil {
		return dataframe.DataFrame{}, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// 读取CSV数据到DataFrame
	df := dataframe.ReadCSV(file)
	if df.Err != nil {
		return dataframe.DataFrame{}, fmt.Errorf("failed to read CSV data: %w", df.Err)
	}

	duration := time.Since(startTime)
	logger.Info("CSV文件加载完成",
		zap.Int("rows", df.Nrow()),
		zap.Int("columns", df.Ncol()),
		zap.Duration("duration", duration),
	)

	return df, nil
}

// Close 关闭加载器
func (l *CSVLoader) Close() error {
	// CSVLoader不需要关闭资源
	return nil
}

// GetSourceInfo 获取数据源信息
func (l *CSVLoader) GetSourceInfo() map[string]interface{} {
	info := map[string]interface{}{
		"type":      "csv",
		"file_path": l.inputPath,
	}
	
	// 获取文件大小
	if stat, err := os.Stat(l.inputPath); err == nil {
		info["file_size"] = stat.Size()
		info["modified_time"] = stat.ModTime()
	}

	return info
}

// NewCSVLoaderFromReader 从io.Reader创建CSV加载器（用于测试）
func NewCSVLoaderFromReader(reader io.Reader) (*CSVLoaderFromReader, error) {
	return &CSVLoaderFromReader{
		reader: reader,
	}, nil
}

// CSVLoaderFromReader 从io.Reader加载CSV（用于测试）
type CSVLoaderFromReader struct {
	reader io.Reader
}

// Load 加载CSV数据
func (l *CSVLoaderFromReader) Load() (dataframe.DataFrame, error) {
	startTime := time.Now()
	logger.Info("开始从Reader加载CSV数据")

	df := dataframe.ReadCSV(l.reader)
	if df.Err != nil {
		return dataframe.DataFrame{}, fmt.Errorf("failed to read CSV data: %w", df.Err)
	}

	duration := time.Since(startTime)
	logger.Info("CSV数据加载完成",
		zap.Int("rows", df.Nrow()),
		zap.Int("columns", df.Ncol()),
		zap.Duration("duration", duration),
	)

	return df, nil
}

// Close 关闭加载器
func (l *CSVLoaderFromReader) Close() error {
	// 不需要关闭reader
	return nil
}

// GetSourceInfo 获取数据源信息
func (l *CSVLoaderFromReader) GetSourceInfo() map[string]interface{} {
	return map[string]interface{}{
		"type": "csv_reader",
	}
}
