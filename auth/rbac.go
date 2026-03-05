package auth

// Role 角色类型
type Role string

// 预定义角色
const (
	RoleAdmin     Role = "admin"
	RoleUser      Role = "user"
	RoleReadOnly  Role = "readonly"
)

// Permission 权限类型
type Permission string

// 预定义权限
const (
	PermissionRead      Permission = "read"
	PermissionWrite     Permission = "write"
	PermissionAdmin     Permission = "admin"
	PermissionBackup    Permission = "backup"
	PermissionRestore   Permission = "restore"
)

// RolePermissions 角色权限映射
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionRead,
		PermissionWrite,
		PermissionAdmin,
		PermissionBackup,
		PermissionRestore,
	},
	RoleUser: {
		PermissionRead,
		PermissionWrite,
		PermissionBackup,
	},
	RoleReadOnly: {
		PermissionRead,
	},
}

// HasPermission 检查角色是否拥有指定权限
func HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// GetRolePermissions 获取角色的所有权限
func GetRolePermissions(role Role) []Permission {
	if permissions, ok := RolePermissions[role]; ok {
		return permissions
	}
	return []Permission{}
}

// ValidateRole 验证角色是否有效
func ValidateRole(role string) bool {
	_, ok := RolePermissions[Role(role)]
	return ok
}
