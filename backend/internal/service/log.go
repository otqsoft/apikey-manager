package service

import (
	"context"

	"apikey-manager/internal/model"
)

// ILogService 日志服务：操作日志 + 登录日志（查询仅 admin，写入供中间件/登录埋点调用）
type ILogService interface {
	// OperationList 操作日志分页查询：keyword 匹配账号/昵称/路径；module 模块过滤（空=全部）；status 过滤（-1=全部）
	OperationList(ctx context.Context, page, size int, keyword, module string, status int) (*model.PageResult, error)
	// LoginList 登录日志分页查询：keyword 匹配账号/昵称；status 过滤（-1=全部）
	LoginList(ctx context.Context, page, size int, keyword string, status int) (*model.PageResult, error)
	// RecordOperation 写入一条操作日志（失败不影响主流程，由调用方忽略错误）
	RecordOperation(ctx context.Context, entry *model.OperationLog) error
	// RecordLogin 写入一条登录日志（失败不影响登录主流程，由调用方忽略错误）
	RecordLogin(ctx context.Context, entry *model.LoginLog) error
}

var Log ILogService
