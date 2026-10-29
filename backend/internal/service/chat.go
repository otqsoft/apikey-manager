package service

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/model"
)

// IChatService 会话与对话转发
type IChatService interface {
	// CreateSession 创建会话（校验 key 可用性）
	CreateSession(ctx context.Context, userID uint, req *model.CreateSessionReq) (*model.ChatSession, error)
	// ListSessions admin 全部；user 仅本人；按创建时间倒序分页
	ListSessions(ctx context.Context, userID uint, isAdmin bool, page, size int) (*model.PageResult, error)
	// ListMessages 消息正序；越权 403
	ListMessages(ctx context.Context, sessionID, userID uint, isAdmin bool) ([]*model.ChatMessage, error)
	// DeleteSession 仅本人或 admin
	DeleteSession(ctx context.Context, sessionID, userID uint, isAdmin bool) error
	// Complete 对话转发：按 stream 分别写 SSE 流或一次性 JSON；负责落库、计费、额度累计
	Complete(r *ghttp.Request, userID uint, req *model.ChatCompletionReq) error
}

var Chat IChatService
