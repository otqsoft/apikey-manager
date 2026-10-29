package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IKeyService 密钥管理：admin 全量；普通用户管理自己名下密钥（owner_id 过滤）
type IKeyService interface {
	// List 分页列表（掩码回显）；admin 全量，普通用户仅自己名下
	List(ctx context.Context, userID uint, isAdmin bool, page, size int, providerID uint, keyword string) (*model.PageResult, error)
	// Detail 详情返回完整 api_key；admin 全量，普通用户限 owner 或被分配者
	Detail(ctx context.Context, id, userID uint, isAdmin bool) (*model.ApiKey, error)
	// Create 新建密钥；creatorID 为创建者（当前登录用户）且为 owner，
	// 同一事务内幂等写入 key_user 实现“创建即分配”
	Create(ctx context.Context, creatorID uint, isAdmin bool, req *model.KeySaveReq) (uint, error)
	Update(ctx context.Context, id, userID uint, isAdmin bool, req *model.KeySaveReq) error
	// Delete 级联清理 key_user 分配关系（归属校验）
	Delete(ctx context.Context, id, userID uint, isAdmin bool) error
	SetStatus(ctx context.Context, id, userID uint, isAdmin bool, status int) error
	// ListUsers 已分配用户简表（仅 owner 或 admin）
	ListUsers(ctx context.Context, id, userID uint, isAdmin bool) ([]map[string]interface{}, error)
	// AssignUsers 全量覆盖分配关系（仅 owner 或 admin）
	AssignUsers(ctx context.Context, id, userID uint, isAdmin bool, userIDs []uint) error
	// MeList 我可用的密钥：名下（owner）+ 被分配（key_user），附带供应商与启用模型
	MeList(ctx context.Context, userID uint) ([]model.MeKeyItem, error)
	// CheckAvailable 对话前可用性校验：key 启用、供应商启用、未过期、未超限额
	CheckAvailable(ctx context.Context, keyID uint) (*model.ApiKey, *model.Provider, error)
}

var Key IKeyService
