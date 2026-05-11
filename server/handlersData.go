package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"sfsEdgeStore/backup"
	"sfsEdgeStore/common"
	"sfsEdgeStore/database"
	"sfsEdgeStore/edgex"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// handleQueryReadings 处理查询数据的请求
func (s *Server) handleQueryReadings(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	deviceName := r.URL.Query().Get("deviceName")
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")
	limitStr := r.URL.Query().Get("limit")

	if deviceName != "" {
		deviceName = common.FormatDeviceName(deviceName)
	}

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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":    len(readings),
		"readings": readings,
	})
}

// handleBackup 处理备份数据库请求
func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	backupPath := r.URL.Query().Get("path")
	if backupPath == "" {
		backupPath = "./backups"
	}

	backupManager := backup.NewBackupManager(storage.GetDBManager().GetDB())
	backupFile, err := backupManager.Backup(backupPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"backupFile": backupFile,
	})
}

// handleRestore 处理恢复数据库请求
func (s *Server) handleRestore(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	backupFile := r.URL.Query().Get("file")
	if backupFile == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Backup file path is required"})
		return
	}

	backupManager := backup.NewBackupManager(storage.GetDBManager().GetDB())
	if err := backupManager.Restore(backupFile); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Database restored successfully",
	})
}

// ------------下面的都是测试用-----------------------------
// handleTestEdgeX 处理测试 EdgeX 消息的请求
func (s *Server) handleTestEdgeX(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var edgexMsg edgex.EdgeXMessage
	if err := json.NewDecoder(r.Body).Decode(&edgexMsg); err != nil {
		r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to read request body: %v", err)})
			return
		}

		var event edgex.EdgeXEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to process message: %v", err)})
			return
		}

		event.DeviceName = common.FormatDeviceName(event.DeviceName)
		handleEdgeXEvent(s, w, &event)
		return
	}

	msgBytes, err := json.Marshal(edgexMsg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	event, err := edgex.ProcessMessage(msgBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to process message: %v", err)})
		return
	}
	if event == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored_non_event_message"})
		return
	}
	handleEdgeXEvent(s, w, event)
}

// handleEdgeXEvent 处理 EdgeX 消息事件
func handleEdgeXEvent(s *Server, w http.ResponseWriter, event *edgex.EdgeXEvent) {
	var records []*map[string]any
	for _, reading := range event.Readings {
		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		value := common.ParseValue(reading.Value)

		data := map[string]any{
			"id":          reading.ID,
			"deviceName":  event.DeviceName,
			"profileName": event.ProfileName,
			"reading":     reading.ResourceName,
			"value":       value,
			"valueType":   reading.ValueType,
			"baseType":    reading.BaseType,
			"timestamp":   reading.Origin,
			"metadata":    metadataStr,
		}
		records = append(records, &data)
	}

	if len(records) > 0 {
		_, err := s.Table.BatchInsertNoIncIoT(records, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Batch stored %d readings from %s", len(records), event.DeviceName),
		})
	} else {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "No readings to store",
		})
	}
}
