package controller

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Model 模型控制器（admin）
var Model = modelController{}

type modelController struct{}

// List 分页 + provider_id + keyword；列表冗余 provider_name
func (modelController) List(r *ghttp.Request) {
	page, size := pageParams(r)
	providerID := r.GetQuery("provider_id").Uint()
	keyword := r.GetQuery("keyword").String()
	data, err := service.Model.List(r.Context(), page, size, providerID, keyword)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Create 新建模型
func (modelController) Create(r *ghttp.Request) {
	var req model.ModelSaveReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	id, err := service.Model.Create(r.Context(), &req)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, g.Map{"id": id})
}

// Detail 模型详情
func (modelController) Detail(r *ghttp.Request) {
	data, err := service.Model.Detail(r.Context(), r.Get("id").Uint())
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Update 更新模型
func (modelController) Update(r *ghttp.Request) {
	var req model.ModelSaveReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Model.Update(r.Context(), r.Get("id").Uint(), &req); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// Delete 删除模型
func (modelController) Delete(r *ghttp.Request) {
	if err := service.Model.Delete(r.Context(), r.Get("id").Uint()); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// SetStatus 启停模型
func (modelController) SetStatus(r *ghttp.Request) {
	var req model.StatusReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Model.SetStatus(r.Context(), r.Get("id").Uint(), req.Status); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}
