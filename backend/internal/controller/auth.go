package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/middleware"
	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Auth 认证控制器
var Auth = authController{}

type authController struct{}

// Login 登录（公开接口）
func (authController) Login(r *ghttp.Request) {
	var req model.LoginReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	resp, err := service.Auth.Login(r.Context(), &req)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, resp)
}

// Register 公开注册（无需 JWT；成功不签发 token，前端切回登录表单）
func (authController) Register(r *ghttp.Request) {
	var req model.RegisterReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	data, err := service.Auth.Register(r.Context(), &req)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Logout 登出：前端清除 token 即可；后端预留黑名单扩展点
func (authController) Logout(r *ghttp.Request) {
	response.Success(r, nil)
}

// Profile 当前用户信息 + 权限标识
func (authController) Profile(r *ghttp.Request) {
	data, err := service.Auth.Profile(r.Context(), middleware.UserID(r))
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// ChangePassword 本人修改密码
func (authController) ChangePassword(r *ghttp.Request) {
	var req model.ChangePasswordReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Auth.ChangePassword(r.Context(), middleware.UserID(r), &req); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}
