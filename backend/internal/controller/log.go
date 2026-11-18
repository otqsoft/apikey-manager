package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Log 日志控制器（仅 admin，路由级 RequireAdmin 守卫）
var Log = logController{}

type logController struct{}

// OperationList 操作日志分页列表：page/page_size/keyword/module/status
func (logController) OperationList(r *ghttp.Request) {
	page, size := pageParams(r)
	keyword := r.GetQuery("keyword").String()
	module := r.GetQuery("module").String()
	status := optionalStatus(r)
	data, err := service.Log.OperationList(r.Context(), page, size, keyword, module, status)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// LoginList 登录日志分页列表：page/page_size/keyword/status
func (logController) LoginList(r *ghttp.Request) {
	page, size := pageParams(r)
	keyword := r.GetQuery("keyword").String()
	status := optionalStatus(r)
	data, err := service.Log.LoginList(r.Context(), page, size, keyword, status)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}
