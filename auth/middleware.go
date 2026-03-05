package auth

import (
	"encoding/json"
	"net/http"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取API Key
		authHeader := r.Header.Get("X-API-Key")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "API key required"})
			return
		}

		// 验证API Key
		apiKey, err := GetAPIKey(authHeader)
		if err != nil || !apiKey.IsValid() {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired API key"})
			return
		}

		// 将用户信息存储到请求上下文
		r = setUserInfo(r, apiKey.UserID, string(apiKey.Role))

		next(w, r)
	}
}

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(permission Permission, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求上下文获取用户角色
		role, ok := getUserRole(r)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "User not authenticated"})
			return
		}

		// 检查权限
		if !HasPermission(Role(role), permission) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Insufficient permissions"})
			return
		}

		next(w, r)
	}
}

// 简单的上下文管理（实际应用中应该使用context包）
func setUserInfo(r *http.Request, userID, role string) *http.Request {
	// 简化处理，实际应用中应该使用context
	r.Header.Set("X-User-ID", userID)
	r.Header.Set("X-User-Role", role)
	return r
}

func getUserRole(r *http.Request) (string, bool) {
	role := r.Header.Get("X-User-Role")
	return role, role != ""
}

// GetAPIKey 从存储中获取API Key
func GetAPIKey(key string) (*APIKey, error) {
	// 使用AuthManager从数据库获取API Key
	authManager := NewAuthManager()
	return authManager.GetAPIKeyByKey(key)
}
