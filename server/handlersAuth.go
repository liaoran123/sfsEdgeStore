package server

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"sfsEdgeStore/auth"
	"sfsEdgeStore/database"
)

func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		UserID    string `json:"user_id"`
		Role      string `json:"role"`
		ExpiresIn int    `json:"expires_in"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.UserID == "" || req.Role == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User ID and role are required"})
		return
	}

	authManager := auth.NewAuthManager()
	apiKey, err := authManager.CreateAPIKey(req.UserID, req.Role, time.Duration(req.ExpiresIn)*time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"api_key":    apiKey.Key,
		"user_id":    apiKey.UserID,
		"role":       apiKey.Role,
		"expires_at": apiKey.ExpiresAt,
	})
}

func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	authManager := auth.NewAuthManager()
	apiKeys, err := authManager.ListAPIKeys()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var response []map[string]interface{}
	for _, key := range apiKeys {
		response = append(response, map[string]interface{}{
			"id":         key.ID,
			"user_id":    key.UserID,
			"role":       key.Role,
			"created_at": key.CreatedAt,
			"expires_at": key.ExpiresAt,
			"active":     key.Active,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"api_keys": response,
	})
}

func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		APIKey string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "API key is required"})
		return
	}

	authManager := auth.NewAuthManager()
	if err := authManager.RevokeAPIKey(req.APIKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "API key revoked successfully",
	})
}

func (s *Server) handleRotateEncryptionKey(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		NewKey string `json:"new_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.NewKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "New encryption key is required"})
		return
	}

	if err := database.RotateEncryptionKey(req.NewKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Encryption key rotated successfully",
	})
}

func (s *Server) handleGetEncryptionStatus(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	enabled, algorithm, err := database.GetEncryptionStatus()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"enabled":   enabled,
		"algorithm": algorithm,
	})
}


func (s *Server) handleGetHardwareInfo(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	cpuSerial := getCPUSerialNumber()
	hostname, _ := os.Hostname()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"cpu_serial":    cpuSerial,
		"hostname":      hostname,
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"cpu_cores":     runtime.NumCPU(),
		"go_version":    runtime.Version(),
	})
}

func getCPUSerialNumber() string {
	if runtime.GOOS == "windows" {
		return getWindowsCPUSerial()
	}
	if runtime.GOOS == "linux" {
		return getLinuxCPUSerial()
	}
	if runtime.GOOS == "darwin" {
		return getDarwinCPUSerial()
	}
	return "UNKNOWN"
}

func getWindowsCPUSerial() string {
	cmd := exec.Command("wmic", "cpu", "get", "ProcessorId")
	output, err := cmd.Output()
	if err != nil {
		return "UNKNOWN"
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(strings.ToLower(line), "processorid") {
			return line
		}
	}
	return "UNKNOWN"
}

func getLinuxCPUSerial() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "UNKNOWN"
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Serial") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "UNKNOWN"
}

func getDarwinCPUSerial() string {
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err != nil {
		return "UNKNOWN"
	}
	return strings.TrimSpace(string(output))
}
