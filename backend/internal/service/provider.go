package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IProviderService 供应商管理（admin）
type IProviderService interface {
	List(ctx context.Context, page, size int, keyword string, status int) (*model.PageResult, error)
	Detail(ctx context.Context, id uint) (*model.Provider, error)
	Create(ctx context.Context, req *model.ProviderSaveReq) (uint, error)
	Update(ctx context.Context, id uint, req *model.ProviderSaveReq) error
	// Delete 删除前校验：名下存在模型或密钥时拒绝
	Delete(ctx context.Context, id uint) error
	SetStatus(ctx context.Context, id uint, status int) error
}

var Provider IProviderService
