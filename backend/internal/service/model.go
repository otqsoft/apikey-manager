package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IModelService 模型管理（admin）
type IModelService interface {
	// List 分页；可按 provider_id 过滤；列表冗余 provider_name
	List(ctx context.Context, page, size int, providerID uint, keyword string) (*model.PageResult, error)
	Detail(ctx context.Context, id uint) (*model.Model, error)
	Create(ctx context.Context, req *model.ModelSaveReq) (uint, error)
	Update(ctx context.Context, id uint, req *model.ModelSaveReq) error
	Delete(ctx context.Context, id uint) error
	SetStatus(ctx context.Context, id uint, status int) error
}

var Model IModelService
