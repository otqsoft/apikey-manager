package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

// 上下文中的登录态键名
const (
	CtxUserID   = "auth_user_id"
	CtxUsername = "auth_username"
	CtxRole     = "auth_role"
)

// UserID 从请求上下文取当前登录用户 ID
func UserID(r *ghttp.Request) uint {
	return r.GetCtxVar(CtxUserID).Uint()
}

// Username 当前登录用户名
func Username(r *ghttp.Request) string {
	return r.GetCtxVar(CtxUsername).String()
}

// Role 当前登录角色
func Role(r *ghttp.Request) string {
	return r.GetCtxVar(CtxRole).String()
}

// IsAdmin 是否管理员
func IsAdmin(r *ghttp.Request) bool {
	return Role(r) == "admin"
}
