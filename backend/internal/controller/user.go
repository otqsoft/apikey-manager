package controller

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/middleware"
	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// User 用户管理控制器（仅 admin）
var User = userController{}

type userController struct{}

// List 分页 + keyword；不返回密码哈希
func (userController) List(r *ghttp.Request) {
	page, size := pageParams(r)
	keyword := r.GetQuery("keyword").String()
	data, err := service.User.List(r.Context(), page, size, keyword)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Create 新建用户
func (userController) Create(r *ghttp.Request) {
	var req model.UserSaveReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	id, err := service.User.Create(r.Context(), &req)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, g.Map{"id": id})
}

// Detail 用户详情
func (userController) Detail(r *ghttp.Request) {
	data, err := service.User.Detail(r.Context(), r.Get("id").Uint())
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Update 更新角色/显示名/状态（不改密码、不改账号）
func (userController) Update(r *ghttp.Request) {
	var req model.UserUpdateReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.User.Update(r.Context(), r.Get("id").Uint(), &req); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// SetStatus 启停用户
func (userController) SetStatus(r *ghttp.Request) {
	var req model.StatusReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.User.SetStatus(r.Context(), r.Get("id").Uint(), req.Status); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// ResetPassword 管理员重置密码
func (userController) ResetPassword(r *ghttp.Request) {
	var req model.ResetPasswordReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.User.ResetPassword(r.Context(), r.Get("id").Uint(), req.Password); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// Delete 删除用户（禁止删除自己；清理 key_user 与会话）
func (userController) Delete(r *ghttp.Request) {
	if err := service.User.Delete(r.Context(), r.Get("id").Uint(), middleware.UserID(r)); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// AssignedKeys 该用户已分配 key 列表（用户管理弹窗用）
func (userController) AssignedKeys(r *ghttp.Request) {
	data, err := service.User.AssignedKeys(r.Context(), r.Get("id").Uint())
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}
