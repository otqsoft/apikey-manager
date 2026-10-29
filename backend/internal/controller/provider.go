package controller

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Provider 供应商控制器（admin 路由守卫在路由层挂载）
var Provider = providerController{}

type providerController struct{}

// List 分页 + keyword + status 筛选
func (providerController) List(r *ghttp.Request) {
	page, size := pageParams(r)
	keyword := r.GetQuery("keyword").String()
	data, err := service.Provider.List(r.Context(), page, size, keyword, optionalStatus(r))
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Create 新建供应商
func (providerController) Create(r *ghttp.Request) {
	var req model.ProviderSaveReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	id, err := service.Provider.Create(r.Context(), &req)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, g.Map{"id": id})
}

// Detail 供应商详情
func (providerController) Detail(r *ghttp.Request) {
	data, err := service.Provider.Detail(r.Context(), r.Get("id").Uint())
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Update 更新供应商
func (providerController) Update(r *ghttp.Request) {
	var req model.ProviderSaveReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Provider.Update(r.Context(), r.Get("id").Uint(), &req); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// Delete 删除前校验名下模型/密钥
func (providerController) Delete(r *ghttp.Request) {
	if err := service.Provider.Delete(r.Context(), r.Get("id").Uint()); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// SetStatus 启停供应商
func (providerController) SetStatus(r *ghttp.Request) {
	var req model.StatusReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Provider.SetStatus(r.Context(), r.Get("id").Uint(), req.Status); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}
