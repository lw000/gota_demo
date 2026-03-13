package cleaner

import (
	"strings"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"data-cleaning-service/internal/config"
	"data-cleaning-service/pkg/logger"
	"go.uber.org/zap"
)

// OperationRecord 清洗操作记录
type OperationRecord struct {
	Operation     string    `json:"operation"`
	AffectedRows  int       `json:"affected_rows"`
	Description   string    `json:"description"`
	Timestamp     time.Time `json:"timestamp"`
}

// DataCleaner 数据清洗器
type DataCleaner struct {
	rules          *config.CleaningRules
	operations     []OperationRecord
	originalRows   int
	cleanedRows    int
}

// NewDataCleaner 创建数据清洗器
func NewDataCleaner(rules *config.CleaningRules) *DataCleaner {
	return &DataCleaner{
		rules:      rules,
		operations: make([]OperationRecord, 0),
	}
}

// Clean 执行数据清洗
func (c *DataCleaner) Clean(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	startTime := time.Now()
	
	c.originalRows = df.Nrow()
	logger.Info("开始数据清洗",
		zap.Int("original_rows", c.originalRows),
		zap.Int("columns", df.Ncol()),
	)

	cleanedDF := df

	// 执行各种清洗操作
	var err error

	// 1. 去重
	if c.rules.RemoveDuplicates {
		cleanedDF, err = c.removeDuplicates(cleanedDF)
		if err != nil {
			return dataframe.DataFrame{}, err
		}
	}

	// 2. 处理缺失值
	if c.rules.HandleMissingValues != "" && c.rules.HandleMissingValues != "none" {
		cleanedDF, err = c.handleMissingValues(cleanedDF)
		if err != nil {
			return dataframe.DataFrame{}, err
		}
	}

	// 3. 数值范围校验
	if len(c.rules.NumericRanges) > 0 {
		cleanedDF, err = c.validateNumericRanges(cleanedDF)
		if err != nil {
			return dataframe.DataFrame{}, err
		}
	}

	// 4. 字符串规范化
	if c.rules.StringNormalization {
		cleanedDF, err = c.normalizeStrings(cleanedDF)
		if err != nil {
			return dataframe.DataFrame{}, err
		}
	}

	c.cleanedRows = cleanedDF.Nrow()
	duration := time.Since(startTime)

	logger.Info("数据清洗完成",
		zap.Int("cleaned_rows", c.cleanedRows),
		zap.Int("removed_rows", c.originalRows-c.cleanedRows),
		zap.Duration("duration", duration),
	)

	return cleanedDF, nil
}

// removeDuplicates 去除重复行
func (c *DataCleaner) removeDuplicates(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	startTime := time.Now()
	originalRows := df.Nrow()

	logger.Info("执行去重操作")

	// 使用map去重
	seen := make(map[string]bool)
	var validRows []int

	for i := 0; i < df.Nrow(); i++ {
		// 生成行唯一标识
		rowKey := c.generateRowKey(df, i)
		if !seen[rowKey] {
			seen[rowKey] = true
			validRows = append(validRows, i)
		}
	}

	// 过滤重复行
	cleanedDF, err := c.filterRows(df, validRows)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	removedRows := originalRows - cleanedDF.Nrow()
	duration := time.Since(startTime)

	// 记录操作
	c.operations = append(c.operations, OperationRecord{
		Operation:    "remove_duplicates",
		AffectedRows: removedRows,
		Description:  "去除重复行",
		Timestamp:    startTime,
	})

	logger.Info("去重完成",
		zap.Int("removed_rows", removedRows),
		zap.Duration("duration", duration),
	)

	return cleanedDF, nil
}

// generateRowKey 生成行的唯一标识
func (c *DataCleaner) generateRowKey(df dataframe.DataFrame, rowIndex int) string {
	var keyParts []string
	colNames := df.Names()
	for i := 0; i < df.Ncol(); i++ {
		s := df.Col(colNames[i])
		val := s.Elem(rowIndex).String()
		keyParts = append(keyParts, val)
	}
	return strings.Join(keyParts, "|")
}

// handleMissingValues 处理缺失值
func (c *DataCleaner) handleMissingValues(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	startTime := time.Now()

	switch c.rules.HandleMissingValues {
	case "delete":
		return c.deleteMissingValues(df, startTime)
	case "fill_mean":
		return c.fillMissingValuesWithMean(df, startTime)
	case "fill_zero":
		return c.fillMissingValuesWithZero(df, startTime)
	case "fill_custom":
		return c.fillMissingValuesWithCustom(df, startTime)
	default:
		logger.Warn("未知的缺失值处理方式",
			zap.String("method", c.rules.HandleMissingValues),
		)
		return df, nil
	}
}

// deleteMissingValues 删除包含缺失值的行
func (c *DataCleaner) deleteMissingValues(df dataframe.DataFrame, startTime time.Time) (dataframe.DataFrame, error) {
	logger.Info("删除包含缺失值的行")
	originalRows := df.Nrow()

	var validRows []int
	colNames := df.Names()
	for i := 0; i < df.Nrow(); i++ {
		hasMissing := false
		for j := 0; j < df.Ncol(); j++ {
			s := df.Col(colNames[j])
			if s.Elem(i).IsNA() {
				hasMissing = true
				break
			}
		}
		if !hasMissing {
			validRows = append(validRows, i)
		}
	}

	cleanedDF, err := c.filterRows(df, validRows)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	removedRows := originalRows - cleanedDF.Nrow()
	duration := time.Since(startTime)

	c.operations = append(c.operations, OperationRecord{
		Operation:    "delete_missing_values",
		AffectedRows: removedRows,
		Description:  "删除包含缺失值的行",
		Timestamp:    startTime,
	})

	logger.Info("删除缺失值完成",
		zap.Int("removed_rows", removedRows),
		zap.Duration("duration", duration),
	)

	return cleanedDF, nil
}

// fillMissingValuesWithMean 使用均值填充缺失值
func (c *DataCleaner) fillMissingValuesWithMean(df dataframe.DataFrame, startTime time.Time) (dataframe.DataFrame, error) {
	logger.Info("使用均值填充缺失值")

	// TODO: 实现均值填充逻辑
	// 目前先返回原始数据
	logger.Warn("均值填充功能待实现")

	return df, nil
}

// fillMissingValuesWithZero 使用0填充缺失值
func (c *DataCleaner) fillMissingValuesWithZero(df dataframe.DataFrame, startTime time.Time) (dataframe.DataFrame, error) {
	logger.Info("使用0填充缺失值")

	// TODO: 实现零值填充逻辑
	// 目前先返回原始数据
	logger.Warn("零值填充功能待实现")

	return df, nil
}

// fillMissingValuesWithCustom 使用自定义值填充缺失值
func (c *DataCleaner) fillMissingValuesWithCustom(df dataframe.DataFrame, startTime time.Time) (dataframe.DataFrame, error) {
	logger.Info("使用自定义值填充缺失值")

	// TODO: 实现自定义值填充逻辑
	// 目前先返回原始数据
	logger.Warn("自定义值填充功能待实现")

	return df, nil
}

// validateNumericRanges 验证数值范围
func (c *DataCleaner) validateNumericRanges(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	startTime := time.Now()

	logger.Info("验证数值范围",
		zap.Int("ranges_count", len(c.rules.NumericRanges)),
	)

	originalRows := df.Nrow()
	var validRows []int
	colNames := df.Names()

	for i := 0; i < df.Nrow(); i++ {
		valid := true
		for _, rangeRule := range c.rules.NumericRanges {
			// 检查列是否存在
			colIndex := -1
			for j := 0; j < len(colNames); j++ {
				if colNames[j] == rangeRule.Column {
					colIndex = j
					break
				}
			}

			if colIndex == -1 {
				logger.Warn("列不存在，跳过验证",
					zap.String("column", rangeRule.Column),
				)
				continue
			}

			// 获取列值
			col := df.Col(colNames[colIndex])
			elem := col.Elem(i)

			if elem.IsNA() {
				continue
			}

			// 检查数值范围
			val := elem.Float()
			if val < rangeRule.Min || val > rangeRule.Max {
				valid = false
				logger.Debug("值超出范围",
					zap.Int("row", i),
					zap.String("column", rangeRule.Column),
					zap.Float64("value", val),
					zap.Float64("min", rangeRule.Min),
					zap.Float64("max", rangeRule.Max),
				)
				break
			}
		}

		if valid {
			validRows = append(validRows, i)
		}
	}

	cleanedDF, err := c.filterRows(df, validRows)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	removedRows := originalRows - cleanedDF.Nrow()
	duration := time.Since(startTime)

	c.operations = append(c.operations, OperationRecord{
		Operation:    "validate_numeric_ranges",
		AffectedRows: removedRows,
		Description:  "验证数值范围",
		Timestamp:    startTime,
	})

	logger.Info("数值范围验证完成",
		zap.Int("removed_rows", removedRows),
		zap.Duration("duration", duration),
	)

	return cleanedDF, nil
}

// normalizeStrings 字符串规范化
func (c *DataCleaner) normalizeStrings(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	startTime := time.Time{} // 使用零值，因为我们会更新时间

	logger.Info("规范化字符串")
	startTime = time.Now()

	// 遍历所有列
	var seriesArray []series.Series
	colNames := df.Names()
	for i := 0; i < df.Ncol(); i++ {
		s := df.Col(colNames[i])

		// 只处理字符串类型的列
		if s.Type() != series.String {
			seriesArray = append(seriesArray, s)
			continue
		}

		// 规范化字符串：去除前后空格，统一转小写
		var normalizedValues []interface{}
		for j := 0; j < s.Len(); j++ {
			elem := s.Elem(j)
			if elem.IsNA() {
				normalizedValues = append(normalizedValues, nil)
				continue
			}

			strVal := elem.String()
			normalized := strings.TrimSpace(strings.ToLower(strVal))
			normalizedValues = append(normalizedValues, normalized)
		}

		normalizedSeries := series.New(normalizedValues, series.String, s.Name)
		seriesArray = append(seriesArray, normalizedSeries)
	}

	cleanedDF := dataframe.New(seriesArray...)
	duration := time.Since(startTime)

	c.operations = append(c.operations, OperationRecord{
		Operation:    "normalize_strings",
		AffectedRows: cleanedDF.Nrow(),
		Description:  "字符串规范化（去空格、转小写）",
		Timestamp:    startTime,
	})

	logger.Info("字符串规范化完成",
		zap.Duration("duration", duration),
	)

	return cleanedDF, nil
}

// filterRows 过滤行
func (c *DataCleaner) filterRows(df dataframe.DataFrame, validRows []int) (dataframe.DataFrame, error) {
	if len(validRows) == 0 {
		return dataframe.New(), nil
	}

	// 为每列创建过滤后的Series
	var seriesArray []series.Series
	colNames := df.Names()
	for i := 0; i < df.Ncol(); i++ {
		s := df.Col(colNames[i])
		var filteredValues []interface{}

		for _, rowIndex := range validRows {
			if rowIndex >= s.Len() {
				continue
			}
			elem := s.Elem(rowIndex)
			if elem.IsNA() {
				filteredValues = append(filteredValues, nil)
			} else {
				// 根据类型获取值
				switch s.Type() {
				case series.String:
					filteredValues = append(filteredValues, elem.String())
				case series.Int:
					if val, err := elem.Int(); err == nil {
						filteredValues = append(filteredValues, val)
					} else {
						filteredValues = append(filteredValues, nil)
					}
				case series.Float:
					filteredValues = append(filteredValues, elem.Float())
				case series.Bool:
					if val, err := elem.Bool(); err == nil {
						filteredValues = append(filteredValues, val)
					} else {
						filteredValues = append(filteredValues, nil)
					}
				default:
					filteredValues = append(filteredValues, nil)
				}
			}
		}

		filteredSeries := series.New(filteredValues, s.Type(), s.Name)
		seriesArray = append(seriesArray, filteredSeries)
	}

	return dataframe.New(seriesArray...), nil
}

// GetOperations 获取操作记录
func (c *DataCleaner) GetOperations() []OperationRecord {
	return c.operations
}

// GetOriginalRows 获取原始行数
func (c *DataCleaner) GetOriginalRows() int {
	return c.originalRows
}

// GetCleanedRows 获取清洗后行数
func (c *DataCleaner) GetCleanedRows() int {
	return c.cleanedRows
}

// GetDataQuality 获取数据质量指标
func (c *DataCleaner) GetDataQuality() map[string]float64 {
	if c.originalRows == 0 {
		return map[string]float64{}
	}

	removedRows := c.originalRows - c.cleanedRows
	completeness := float64(c.cleanedRows) / float64(c.originalRows) * 100

	return map[string]float64{
		"completeness": completeness,
		"data_loss":    float64(removedRows) / float64(c.originalRows) * 100,
	}
}
