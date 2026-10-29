package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IKeyService 密钥管理（admin 全量；普通用户仅 /me/keys 与可访问详情）
type IKeyService interface {
	// List admin 全量分页列表（掩码回显）
	List(ctx context.Context, page, size int, providerID uint, keyword string) (*model.PageResult, error)
	// Detail 详情返回完整 api_key；admin 全量，普通用户仅限被分配者
	Detail(ctx context.Context, id, userID uint, isAdmin bool) (*model.ApiKey, error)
	// Create 新建密钥；creatorID 为创建者（当前登录用户），同一事务内幂等写入 key_user 实现“创建即分配”
	Create(ctx context.Context, creatorID uint, req *model.KeySaveReq) (uint, error)
	Update(ctx context.Context, id uint, req *model.KeySaveReq) error
	// Delete 级联清理 key_user 分配关系
	Delete(ctx context.Context, id uint) error
	SetStatus(ctx context.Context, id uint, status int) error
	// ListUsers 已分配用户简表
	ListUsers(ctx context.Context, id uint) ([]map[string]interface{}, error)
	// AssignUsers 全量覆盖分配关系
	AssignUsers(ctx context.Context, id uint, userIDs []uint) error
	// MeList 普通用户核心接口：已分配 + 全部可用条件满足的 key，附带供应商与启用模型
	MeList(ctx context.Context, userID uint) ([]model.MeKeyItem, error)
	// CheckAvailable 对话前可用性校验：key 启用、供应商启用、未过期、未超限额
	CheckAvailable(ctx context.Context, keyID uint) (*model.ApiKey, *model.Provider, error)
}

var Key IKeyService
