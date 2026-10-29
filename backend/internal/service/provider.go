package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IProviderService 供应商管理：admin 全量；普通用户仅自己名下（owner_id 过滤）
type IProviderService interface {
	List(ctx context.Context, userID uint, isAdmin bool, page, size int, keyword string, status int) (*model.PageResult, error)
	Detail(ctx context.Context, id, userID uint, isAdmin bool) (*model.Provider, error)
	// Create 创建供应商，归属 ownerID（当前登录用户）
	Create(ctx context.Context, ownerID uint, req *model.ProviderSaveReq) (uint, error)
	Update(ctx context.Context, id, userID uint, isAdmin bool, req *model.ProviderSaveReq) error
	// Delete 删除前校验：名下存在模型或密钥时拒绝
	Delete(ctx context.Context, id, userID uint, isAdmin bool) error
	SetStatus(ctx context.Context, id, userID uint, isAdmin bool, status int) error
}

var Provider IProviderService
