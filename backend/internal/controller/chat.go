package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/middleware"
	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Chat 对话控制器：会话管理 + 双协议转发
var Chat = chatController{}

type chatController struct{}

// ListSessions admin 全部；user 仅本人；倒序分页
func (chatController) ListSessions(r *ghttp.Request) {
	page, size := pageParams(r)
	data, err := service.Chat.ListSessions(r.Context(), middleware.UserID(r), middleware.IsAdmin(r), page, size)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// CreateSession 创建会话
func (chatController) CreateSession(r *ghttp.Request) {
	var req model.CreateSessionReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	data, err := service.Chat.CreateSession(r.Context(), middleware.UserID(r), &req)
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Messages 会话消息正序；越权 403
func (chatController) Messages(r *ghttp.Request) {
	data, err := service.Chat.ListMessages(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r))
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// DeleteSession 仅本人或 admin
func (chatController) DeleteSession(r *ghttp.Request) {
	if err := service.Chat.DeleteSession(r.Context(), r.Get("id").Uint(), middleware.UserID(r), middleware.IsAdmin(r)); err != nil {
		fail(r, err)
		return
	}
	response.Success(r, nil)
}

// Completions 对话转发（stream=true 走 SSE；stream=false 一次性 JSON）
func (chatController) Completions(r *ghttp.Request) {
	var req model.ChatCompletionReq
	if err := r.Parse(&req); err != nil {
		response.BadRequest(r, err.Error())
		return
	}
	if err := service.Chat.Complete(r, middleware.UserID(r), &req); err != nil {
		fail(r, err)
		return
	}
}
