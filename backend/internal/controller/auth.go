package controller

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/middleware"
	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Auth 认证控制器
var Auth = authController{}

type authController struct{}

// Login 登录（公开接口）；无论成功失败均写入登录日志
func (authController) Login(r *ghttp.Request) {
	var req model.LoginReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	resp, err := service.Auth.Login(r.Context(), &req)
	if err != nil {
		_, reason := errs.From(err)
		recordLogin(r, 0, req.Username, "", 0, reason)
		fail(r, err)
		return
	}
	recordLogin(r, resp.User.ID, resp.User.Username, resp.User.DisplayName, 1, "")
	response.Success(r, resp)
}

// recordLogin 写一条登录日志；失败仅告警，不影响登录主流程
func recordLogin(r *ghttp.Request, userID uint, username, displayName string, status int, failReason string) {
	entry := &model.LoginLog{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		Status:      status,
		IP:          r.GetClientIp(),
		UserAgent:   r.Header.Get("User-Agent"),
		FailReason:  failReason,
	}
	if err := service.Log.RecordLogin(r.Context(), entry); err != nil {
		g.Log().Warningf(r.Context(), "写登录日志失败: %v", err)
	}
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
