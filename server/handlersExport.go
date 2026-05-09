package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"sfsEdgeStore/database"

	"github.com/liaoran123/sfsDb/record"
)

// handleExportCSV 处理导出 CSV 请求
func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	tempFile, err := os.CreateTemp("", "export-*.csv")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create temp file"})
		return
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	if err := database.ExportTableToCSV(database.Table, tempPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=export.csv")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	file, err := os.Open(tempPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open export file"})
		return
	}
	defer file.Close()

	io.Copy(w, file)
}

// handleExportJSON 处理导出 JSON 请求
func (s *Server) handleExportJSON(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	tempFile, err := os.CreateTemp("", "export-*.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create temp file"})
		return
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	if err := database.ExportTableToJSON(database.Table, tempPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=export.json")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	file, err := os.Open(tempPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open export file"})
		return
	}
	defer file.Close()

	io.Copy(w, file)
}

// handleExportSQL 处理导出 SQL 请求
func (s *Server) handleExportSQL(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		filePath = "./export.sql"
	}

	if err := database.ExportTableToSQL(database.Table, filePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"file":   filePath,
	})
}

// handleImportCSV 处理导入 CSV 请求
func (s *Server) handleImportCSV(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File path is required"})
		return
	}

	if err := database.ImportTableFromCSV(database.Table, filePath, 100); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data imported successfully from CSV",
	})
}

// handleImportJSON 处理导入 JSON 请求
func (s *Server) handleImportJSON(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File path is required"})
		return
	}

	if err := database.ImportTableFromJSON(database.Table, filePath, 100); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data imported successfully from JSON",
	})
}

// handleDataExport 处理导出数据请求
func (s *Server) handleDataExport(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	deviceName := r.URL.Query().Get("deviceName")
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")
	limitStr := r.URL.Query().Get("limit")

	var limit int
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var err error
	var readings record.Records
	if limit > 0 {
		readings, err = database.QueryRecords(database.Table, deviceName, startTime, endTime, false, limit)
	} else {
		readings, err = database.QueryRecords(database.Table, deviceName, startTime, endTime, false)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer readings.Release()

	readingsMap := make([]map[string]any, len(readings))
	for i, reading := range readings {
		readingsMap[i] = reading
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":    len(readings),
			"readings": readingsMap,
		})
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=data.csv")
		if len(readingsMap) > 0 {
			fields := make([]string, 0)
			for field := range readingsMap[0] {
				fields = append(fields, field)
			}
			writer := csv.NewWriter(w)
			defer writer.Flush()
			writer.Write(fields)
			for _, reading := range readingsMap {
				row := make([]string, len(fields))
				for i, field := range fields {
					value := reading[field]
					row[i] = fmt.Sprintf("%v", value)
				}
				writer.Write(row)
			}
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unsupported format"})
	}
}
