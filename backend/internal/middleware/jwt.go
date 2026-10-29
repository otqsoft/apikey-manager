package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/pkg/jwtx"
	"apikey-manager/internal/pkg/response"
)

// JwtAuth 登录校验中间件：解析 Bearer Token，校验用户存在且未停用，写入上下文。
func JwtAuth(r *ghttp.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		response.Unauthorized(r, "缺少登录凭证")
		return
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		response.Unauthorized(r, "凭证格式非法")
		return
	}
	claims, err := jwtx.Parse(strings.TrimSpace(parts[1]))
	if err != nil {
		response.Unauthorized(r, "登录已失效，请重新登录")
		return
	}
	// 二次校验用户状态（被停用/删除的 token 立即失效）
	count, err := g.DB().Model("users").Where("id", claims.UserID).Where("status", 1).Count(r.Context())
	if err != nil || count == 0 {
		response.Unauthorized(r, "账号已被停用或不存在")
		return
	}
	r.SetCtxVar(CtxUserID, claims.UserID)
	r.SetCtxVar(CtxUsername, claims.Username)
	r.SetCtxVar(CtxRole, claims.Role)
	r.Middleware.Next()
}

// RequireAdmin 角色守卫：仅管理员可访问；普通用户返回 403。
func RequireAdmin(r *ghttp.Request) {
	if Role(r) != "admin" {
		response.Forbidden(r, "仅管理员可执行该操作")
		return
	}
	r.Middleware.Next()
}
