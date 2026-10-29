package logic

import (
	"context"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/service"
)

func init() {
	service.Model = &modelService{}
}

type modelService struct{}

// List 模型分页列表，冗余供应商名称；admin 全量，普通用户仅自己名下
func (s *modelService) List(ctx context.Context, userID uint, isAdmin bool, page, size int, providerID uint, keyword string) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	m := applyOwnerScope(g.DB().Model("models m").LeftJoin("providers p", "p.id = m.provider_id"), userID, isAdmin, "m.owner_id")
	if providerID > 0 {
		m = m.Where("m.provider_id", providerID)
	}
	if keyword != "" {
		m = m.WhereLike("m.name", "%"+keyword+"%")
	}
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	records, err := m.Fields("m.*, p.name as provider_name").
		Page(page, size).OrderDesc("m.id").All()
	if err != nil {
		return nil, err
	}
	// SQLite 以 0/1 存储布尔，列表响应统一归一化为 JSON 布尔 true/false（契约 0 节）
	for _, r := range records {
		for _, field := range []string{"support_thinking", "support_multimodal"} {
			r[field] = gvar.New(r[field].Bool())
		}
	}
	return &model.PageResult{List: safeList(records), Total: total, Page: page, PageSize: size}, nil
}

// Detail 模型详情（归属校验）
func (s *modelService) Detail(ctx context.Context, id, userID uint, isAdmin bool) (*model.Model, error) {
	var m model.Model
	if err := g.DB().Model("models").Where("id", id).Scan(&m); err != nil {
		return nil, scanErr(err, "模型不存在")
	}
	if m.ID == 0 {
		return nil, errs.NotFound("模型不存在")
	}
	if !isAdmin && m.OwnerID != userID {
		return nil, errs.Forbidden("无权查看该模型")
	}
	return &m, nil
}

// Create 新建模型：归属当前用户；非 admin 校验供应商归属
func (s *modelService) Create(ctx context.Context, ownerID uint, isAdmin bool, req *model.ModelSaveReq) (uint, error) {
	if err := assertProviderOwned(ctx, req.ProviderID, ownerID, isAdmin); err != nil {
		return 0, err
	}
	status := req.Status
	if status != 0 && status != 1 {
		status = 1
	}
	id, err := g.DB().Model("models").InsertAndGetId(g.Map{
		"name":               req.Name,
		"owner_id":           ownerID,
		"provider_id":        req.ProviderID,
		"input_price_cached": req.InputPriceCached,
		"input_price":        req.InputPrice,
		"output_price":       req.OutputPrice,
		"context_length":     req.ContextLength,
		"support_thinking":   req.SupportThinking,
		"support_multimodal": req.SupportMultimodal,
		"status":             status,
		"remark":             req.Remark,
	})
	return uint(id), err
}

// Update 更新模型（归属校验 + 新供应商归属校验）
func (s *modelService) Update(ctx context.Context, id, userID uint, isAdmin bool, req *model.ModelSaveReq) error {
	exists, err := assertResourceOwned(ctx, "models", id, userID, isAdmin, "模型")
	if err != nil {
		return err
	}
	if !exists {
		return errs.NotFound("模型不存在")
	}
	if err := assertProviderOwned(ctx, req.ProviderID, userID, isAdmin); err != nil {
		return err
	}
	_, err = g.DB().Model("models").Where("id", id).Update(g.Map{
		"name":               req.Name,
		"provider_id":        req.ProviderID,
		"input_price_cached": req.InputPriceCached,
		"input_price":        req.InputPrice,
		"output_price":       req.OutputPrice,
		"context_length":     req.ContextLength,
		"support_thinking":   req.SupportThinking,
		"support_multimodal": req.SupportMultimodal,
		"status":             req.Status,
		"remark":             req.Remark,
	})
	return err
}

// Delete 删除模型（归属校验）
func (s *modelService) Delete(ctx context.Context, id, userID uint, isAdmin bool) error {
	exists, err := assertResourceOwned(ctx, "models", id, userID, isAdmin, "模型")
	if err != nil {
		return err
	}
	if !exists {
		return errs.NotFound("模型不存在")
	}
	_, err = g.DB().Model("models").Where("id", id).Delete()
	return err
}

// SetStatus 启停模型（归属校验）
func (s *modelService) SetStatus(ctx context.Context, id, userID uint, isAdmin bool, status int) error {
	exists, err := assertResourceOwned(ctx, "models", id, userID, isAdmin, "模型")
	if err != nil {
		return err
	}
	if !exists {
		return errs.NotFound("模型不存在")
	}
	_, err = g.DB().Model("models").Where("id", id).Update(g.Map{"status": status})
	return err
}
