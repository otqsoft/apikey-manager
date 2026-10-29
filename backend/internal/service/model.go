package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IModelService 模型管理：admin 全量；普通用户仅自己名下（owner_id 过滤）
type IModelService interface {
	// List 分页；可按 provider_id 过滤；列表冗余 provider_name
	List(ctx context.Context, userID uint, isAdmin bool, page, size int, providerID uint, keyword string) (*model.PageResult, error)
	Detail(ctx context.Context, id, userID uint, isAdmin bool) (*model.Model, error)
	// Create 创建模型，归属 ownerID；非 admin 校验供应商归属
	Create(ctx context.Context, ownerID uint, isAdmin bool, req *model.ModelSaveReq) (uint, error)
	Update(ctx context.Context, id, userID uint, isAdmin bool, req *model.ModelSaveReq) error
	Delete(ctx context.Context, id, userID uint, isAdmin bool) error
	SetStatus(ctx context.Context, id, userID uint, isAdmin bool, status int) error
}

var Model IModelService
