package database

import "github.com/liaoran123/sfsDb/engine"

// ExportTableToCSV 导出表数据为CSV格式
func ExportTableToCSV(tbl *engine.Table, filePath string) error {
	return tbl.ExportToCSV(filePath)
}

// ImportTableFromCSV 从CSV文件导入数据到表
func ImportTableFromCSV(tbl *engine.Table, filePath string, batchSize int) error {
	return tbl.ImportFromCSV(filePath, batchSize)
}

// ExportTableToJSON 导出表数据为JSON格式
func ExportTableToJSON(tbl *engine.Table, filePath string) error {
	return tbl.ExportToJSON(filePath)
}

// ImportTableFromJSON 从JSON文件导入数据到表
func ImportTableFromJSON(tbl *engine.Table, filePath string, batchSize int) error {
	return tbl.ImportFromJSON(filePath, batchSize)
}

// ExportTableToSQL 导出表数据为SQL格式
func ExportTableToSQL(tbl *engine.Table, filePath string) error {
	return tbl.ExportToSQL(filePath)
}
