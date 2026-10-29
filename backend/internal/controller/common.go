// Package controller HTTP 控制器层：参数绑定校验 → 调用 service → 统一响应封装。
package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/pkg/response"
)

// fail 统一把 service 层业务错误转换为响应体
func fail(r *ghttp.Request, err error) {
	if err == nil {
		return
	}
	code, msg := errs.From(err)
	response.Fail(r, code, msg)
}

// pageParams 读取分页参数；page<=0 时返回默认值 1 / 10
func pageParams(r *ghttp.Request) (page, size int) {
	page = r.GetQuery("page").Int()
	size = r.GetQuery("page_size").Int()
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return
}

// optionalStatus 读取可选状态过滤参数；未提供返回 -1（不过滤）
func optionalStatus(r *ghttp.Request) int {
	v := r.GetQuery("status")
	if v.IsEmpty() {
		return -1
	}
	return v.Int()
}
