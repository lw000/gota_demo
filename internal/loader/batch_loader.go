package loader

import (
	"fmt"

	"github.com/go-gota/gota/dataframe"
	"data-cleaning-service/internal/config"
)

// BatchLoader 批量加载器（根据配置选择数据源）
type BatchLoader struct {
	loader DataLoader
}

// NewBatchLoader 创建批量加载器
func NewBatchLoader(dataConfig *config.DataConfig) (*BatchLoader, error) {
	var loader DataLoader

	// 根据配置选择数据源
	if dataConfig.Kafka.Enabled {
		var err error
		loader, err = NewKafkaLoader(&dataConfig.Kafka)
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka loader: %w", err)
		}
	} else if dataConfig.CSV.Enabled {
		loader = NewCSVLoader(dataConfig.CSV.InputPath)
	} else {
		return nil, fmt.Errorf("no data source enabled")
	}

	return &BatchLoader{
		loader: loader,
	}, nil
}

// Load 加载数据
func (l *BatchLoader) Load() (dataframe.DataFrame, error) {
	return l.loader.Load()
}

// Close 关闭加载器
func (l *BatchLoader) Close() error {
	if l.loader != nil {
		return l.loader.Close()
	}
	return nil
}

// GetSourceInfo 获取数据源信息
func (l *BatchLoader) GetSourceInfo() map[string]interface{} {
	if l.loader != nil {
		return l.loader.GetSourceInfo()
	}
	return map[string]interface{}{}
}

// GetLoader 获取底层加载器
func (l *BatchLoader) GetLoader() DataLoader {
	return l.loader
}
