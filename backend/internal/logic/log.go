package logic

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/model"
	"apikey-manager/internal/service"
)

func init() {
	service.Log = &logService{}
}

type logService struct{}

// OperationList 操作日志分页查询
func (s *logService) OperationList(ctx context.Context, page, size int, keyword, module string, status int) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	m := g.DB().Model("operation_logs")
	if keyword != "" {
		like := "%" + keyword + "%"
		m = m.Where("username LIKE ? OR display_name LIKE ? OR path LIKE ? OR action LIKE ?", like, like, like, like)
	}
	if module != "" {
		m = m.Where("module", module)
	}
	if status == 0 || status == 1 {
		m = m.Where("status", status)
	}
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	records, err := m.Page(page, size).OrderDesc("id").All()
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: safeList(records), Total: total, Page: page, PageSize: size}, nil
}

// LoginList 登录日志分页查询
func (s *logService) LoginList(ctx context.Context, page, size int, keyword string, status int) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	m := g.DB().Model("login_logs")
	if keyword != "" {
		like := "%" + keyword + "%"
		m = m.Where("username LIKE ? OR display_name LIKE ?", like, like)
	}
	if status == 0 || status == 1 {
		m = m.Where("status", status)
	}
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	records, err := m.Page(page, size).OrderDesc("id").All()
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: safeList(records), Total: total, Page: page, PageSize: size}, nil
}

// RecordOperation 写入一条操作日志
func (s *logService) RecordOperation(ctx context.Context, entry *model.OperationLog) error {
	if entry == nil {
		return nil
	}
	_, err := g.DB().Model("operation_logs").Insert(g.Map{
		"user_id":      entry.UserID,
		"username":     entry.Username,
		"display_name": entry.DisplayName,
		"module":       entry.Module,
		"action":       entry.Action,
		"method":       entry.Method,
		"path":         entry.Path,
		"detail":       entry.Detail,
		"ip":           entry.IP,
		"user_agent":   entry.UserAgent,
		"status":       entry.Status,
	})
	return err
}

// RecordLogin 写入一条登录日志
func (s *logService) RecordLogin(ctx context.Context, entry *model.LoginLog) error {
	if entry == nil {
		return nil
	}
	_, err := g.DB().Model("login_logs").Insert(g.Map{
		"user_id":      entry.UserID,
		"username":     entry.Username,
		"display_name": entry.DisplayName,
		"status":       entry.Status,
		"ip":           entry.IP,
		"user_agent":   entry.UserAgent,
		"fail_reason":  entry.FailReason,
	})
	return err
}
