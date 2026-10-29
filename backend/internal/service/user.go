package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IUserService 用户管理（仅 admin）
type IUserService interface {
	// List 分页列表（不返回密码哈希）
	List(ctx context.Context, page, size int, keyword string) (*model.PageResult, error)
	// Create 新建用户（bcrypt 哈希落库）
	Create(ctx context.Context, req *model.UserSaveReq) (uint, error)
	// Detail 用户详情（不含哈希）
	Detail(ctx context.Context, id uint) (*model.User, error)
	// Update 更新角色/显示名/状态（不改密码、不改账号）
	Update(ctx context.Context, id uint, req *model.UserUpdateReq) error
	// SetStatus 启停
	SetStatus(ctx context.Context, id uint, status int) error
	// ResetPassword 管理员重置密码
	ResetPassword(ctx context.Context, id uint, password string) error
	// Delete 删除用户；禁止删除自己；清理 key_user 分配
	Delete(ctx context.Context, id, currentUserID uint) error
	// AssignedKeys 该用户已分配的 key（用户管理弹窗用）
	AssignedKeys(ctx context.Context, userID uint) ([]map[string]interface{}, error)
}

var User IUserService
