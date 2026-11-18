// Package response 统一响应封装。
// 契约格式：{ "code": 0, "msg": "ok", "data": {} }
// code: 0 成功；400 参数错误；401 未登录；403 无权限/越权；500 服务端错误。
package response

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	CodeOK      = 0
	CodeBadReq  = 400
	CodeNoAuth  = 401
	CodeForbid  = 403
	CodeServer  = 500
)

// ctxBizCodeKey 请求上下文中的业务码键名，供操作日志中间件判定处理成功/失败。
const ctxBizCodeKey = "_biz_code"

// SetBizCode 在上下文记录本次请求最终业务码（JSON 写出时调用）
func SetBizCode(r *ghttp.Request, code int) {
	r.SetCtxVar(ctxBizCodeKey, code)
}

// BizCode 读取上下文业务码；未设置（如流式响应）默认按成功处理
func BizCode(r *ghttp.Request) int {
	return r.GetCtxVar(ctxBizCodeKey, CodeOK).Int()
}

// JSON 写出统一响应（HTTP 状态码与业务 code 保持一致，便于前端拦截器识别 401/403）。
func JSON(r *ghttp.Request, code int, msg string, data interface{}) {
	if msg == "" {
		msg = "ok"
	}
	SetBizCode(r, code)
	// 成功用 200；错误码沿用对应 HTTP 状态语义
	httpStatus := 200
	if code == CodeNoAuth || code == CodeForbid || code == CodeBadReq || code == CodeServer {
		httpStatus = code
	}
	r.Response.WriteStatus(httpStatus, g.Map{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

// Success 成功响应
func Success(r *ghttp.Request, data interface{}) {
	JSON(r, CodeOK, "ok", data)
}

// Fail 业务失败响应
func Fail(r *ghttp.Request, code int, msg string) {
	JSON(r, code, msg, nil)
}

// BadRequest 参数错误
func BadRequest(r *ghttp.Request, msg string) {
	Fail(r, CodeBadReq, msg)
}

// Unauthorized 未认证
func Unauthorized(r *ghttp.Request, msg string) {
	if msg == "" {
		msg = "登录已失效，请重新登录"
	}
	Fail(r, CodeNoAuth, msg)
}

// Forbidden 无权限/越权
func Forbidden(r *ghttp.Request, msg string) {
	if msg == "" {
		msg = "无权执行此操作"
	}
	Fail(r, CodeForbid, msg)
}

// ServerError 服务端错误（不向上游泄露敏感信息）
func ServerError(r *ghttp.Request, err error) {
	g.Log().Errorf(r.Context(), "服务端错误: %v", err)
	Fail(r, CodeServer, "服务内部错误，请稍后重试")
}
