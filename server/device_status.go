package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"sfsEdgeStore/database"
)

type deviceStatusEntry struct {
	DeviceName string `json:"deviceName"`
	LastActive int64  `json:"lastActive"`
	DataCount  int    `json:"dataCount"`
	IsOnline   bool   `json:"isOnline"`
}

type deviceStatusCache struct {
	mu      sync.RWMutex
	devices map[string]*deviceStatusEntry
}

func newDeviceStatusCache() *deviceStatusCache {
	return &deviceStatusCache{
		devices: make(map[string]*deviceStatusEntry),
	}
}

func (c *deviceStatusCache) recordStored(deviceName string, timestamp int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.devices[deviceName]
	if !ok {
		c.devices[deviceName] = &deviceStatusEntry{
			DeviceName: deviceName,
			LastActive: timestamp,
			DataCount:  1,
		}
	} else {
		entry.DataCount++
		if timestamp > entry.LastActive {
			entry.LastActive = timestamp
		}
	}
}

func (c *deviceStatusCache) getAll() []deviceStatusEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now().UnixNano()
	offlineThreshold := int64(5 * time.Minute)

	entries := make([]deviceStatusEntry, 0, len(c.devices))
	for _, d := range c.devices {
		entry := *d
		entry.IsOnline = now-entry.LastActive < offlineThreshold
		entries = append(entries, entry)
	}
	return entries
}

func (s *Server) handleDeviceStatus(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()
	w.Header().Set("Content-Type", "application/json")

	entries := s.deviceStatusCache.getAll()

	var devices []deviceStatusEntry
	for _, entry := range entries {
		devices = append(devices, entry)
	}
	if devices == nil {
		devices = []deviceStatusEntry{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"devices": devices,
	})
}

func (s *Server) SeedDeviceStatusCache() {
	startTime := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
	records, err := database.QueryRecords(database.Table, "", startTime, "", false)
	if err != nil {
		return
	}
	defer records.Release()

	for _, rec := range records {
		deviceName, _ := rec["deviceName"].(string)
		timestamp, _ := rec["timestamp"].(int64)
		if deviceName != "" && timestamp > 0 {
			s.deviceStatusCache.recordStored(deviceName, timestamp)
		}
	}
}

func (s *Server) OnRecordStored(deviceName string, timestamp int64) {
	s.deviceStatusCache.recordStored(deviceName, timestamp)
}