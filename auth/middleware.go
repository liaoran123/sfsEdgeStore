package auth

import (
	"encoding/json"
	"net/http"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 第一层：IP白名单（网络层）
		clientIP := r.RemoteAddr
		// 检查是否是本地访问（包含端口号的情况）
		isLocalAccess := false
		
		// 检查 127.0.0.1 或 ::1
		if len(clientIP) >= 5 && (clientIP[:5] == "127.0." || clientIP[:3] == "[::1") {
			isLocalAccess = true
		}
		
		// 检查局域网IP
		if !isLocalAccess && len(clientIP) >= 7 {
			// 检查 192.168.x.x
			if clientIP[:8] == "192.168." {
				isLocalAccess = true
			}
			// 检查 10.x.x.x
			if !isLocalAccess && clientIP[:3] == "10." {
				isLocalAccess = true
			}
			// 检查 172.16.x.x - 172.31.x.x
			if !isLocalAccess && len(clientIP) >= 8 && clientIP[:4] == "172." {
				// 检查第二个 octet 是否在 16-31 之间
				if (clientIP[4] == '1' && clientIP[5] >= '6') || (clientIP[4] == '2') || (clientIP[4] == '3' && clientIP[5] <= '1') {
					isLocalAccess = true
				}
			}
		}
		
		// 对于开发环境，简化认证：允许所有访问
		// 这是为了方便开发和测试，生产环境应该移除这部分代码
		isLocalAccess = true
		
		// 如果不是本地访问，拒绝请求
		if !isLocalAccess {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Access denied: not from local network"})
			return
		}

		// 第二层：简单Token认证（应用层）
		// 从请求头获取API Key
		authHeader := r.Header.Get("X-API-Key")
		// 对于局域网环境，简化认证：如果没有API Key，直接通过
		// 这样可以方便本地开发和测试
		if authHeader == "" {
			// 本地访问，直接通过认证，设置默认管理员角色
			r = setUserInfo(r, "local-admin", "admin")
			next(w, r)
			return
		}

		// 如果有API Key，验证它
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
