package controller

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/middleware"
	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Key 密钥控制器：管理端接口 admin 路由守卫；MeList 登录用户可用
var Key = keyController{}

type keyController struct{}

// List admin 全量分页列表（掩码回显）
func (keyController) List(r *ghttp.Request) {
	page, size := pageParams(r)
	providerID := r.GetQuery("provider_id").Uint()
	keyword := r.GetQuery("keyword").String()
	data, err := service.Key.List(r.Context(), middleware.UserID(r), middleware.IsAdmin(r), page, size, providerID, keyword)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Create 新建密钥（admin）
func (keyController) Create(r *ghttp.Request) {
	var req model.KeySaveReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	id, err := service.Key.Create(r.Context(), middleware.UserID(r), middleware.IsAdmin(r), &req)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, g.Map{"id": id})
}

// Detail 详情返回完整 api_key；admin 全量，普通用户仅限被分配者
func (keyController) Detail(r *ghttp.Request) {
	data, err := service.Key.Detail(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r))
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Update 更新密钥（admin）
func (keyController) Update(r *ghttp.Request) {
	var req model.KeySaveReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Key.Update(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r), &req); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// Delete 删除密钥并级联清理分配
func (keyController) Delete(r *ghttp.Request) {
	if err := service.Key.Delete(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r)); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// SetStatus 启停密钥
func (keyController) SetStatus(r *ghttp.Request) {
	var req model.StatusReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Key.SetStatus(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r), req.Status); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// ListUsers 已分配用户（admin）
func (keyController) ListUsers(r *ghttp.Request) {
	data, err := service.Key.ListUsers(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r))
	if err != nil {
		fail(r, err)
		return
	}
	// service 返回 [ {user_ids, users} ]，取首项直接输出
	if len(data) > 0 {
		response.Success(r, data[0])
		return
	}
	response.Success(r, g.Map{"user_ids": []uint{}, "users": []interface{}{}})
}

// AssignUsers 全量覆盖分配（admin）
func (keyController) AssignUsers(r *ghttp.Request) {
	var req model.KeyUsersReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Key.AssignUsers(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r), req.UserIDs); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// MeList 普通用户核心接口：已分配且全部可用条件满足的 key + 供应商启用模型
func (keyController) MeList(r *ghttp.Request) {
	data, err := service.Key.MeList(r.Context(), middleware.UserID(r))
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}
