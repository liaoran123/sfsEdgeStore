package server

import (
	"encoding/json"
	"net/http"
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
