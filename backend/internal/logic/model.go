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

// List 模型分页列表，冗余供应商名称；支持 provider_id / keyword 过滤
func (s *modelService) List(ctx context.Context, page, size int, providerID uint, keyword string) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	m := g.DB().Model("models m").
		LeftJoin("providers p", "p.id = m.provider_id")
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

// Detail 模型详情
func (s *modelService) Detail(ctx context.Context, id uint) (*model.Model, error) {
	var m model.Model
	if err := g.DB().Model("models").Where("id", id).Scan(&m); err != nil {
		return nil, scanErr(err, "模型不存在")
	}
	if m.ID == 0 {
		return nil, errs.NotFound("模型不存在")
	}
	return &m, nil
}

// Create 新建模型，校验供应商存在
func (s *modelService) Create(ctx context.Context, req *model.ModelSaveReq) (uint, error) {
	cnt, err := g.DB().Model("providers").Where("id", req.ProviderID).Count()
	if err != nil {
		return 0, err
	}
	if cnt == 0 {
		return 0, errs.BadRequest("所属供应商不存在")
	}
	status := req.Status
	if status != 0 && status != 1 {
		status = 1
	}
	id, err := g.DB().Model("models").InsertAndGetId(g.Map{
		"name":                req.Name,
		"provider_id":         req.ProviderID,
		"input_price_cached": req.InputPriceCached,
		"input_price":         req.InputPrice,
		"output_price":        req.OutputPrice,
		"context_length":      req.ContextLength,
		"support_thinking":    req.SupportThinking,
		"support_multimodal":  req.SupportMultimodal,
		"status":              status,
		"remark":              req.Remark,
	})
	return uint(id), err
}

// Update 更新模型
func (s *modelService) Update(ctx context.Context, id uint, req *model.ModelSaveReq) error {
	cnt, err := g.DB().Model("models").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("模型不存在")
	}
	cnt, err = g.DB().Model("providers").Where("id", req.ProviderID).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.BadRequest("所属供应商不存在")
	}
	_, err = g.DB().Model("models").Where("id", id).Update(g.Map{
		"name":                req.Name,
		"provider_id":         req.ProviderID,
		"input_price_cached": req.InputPriceCached,
		"input_price":         req.InputPrice,
		"output_price":        req.OutputPrice,
		"context_length":      req.ContextLength,
		"support_thinking":    req.SupportThinking,
		"support_multimodal":  req.SupportMultimodal,
		"status":              req.Status,
		"remark":              req.Remark,
	})
	return err
}

// Delete 删除模型
func (s *modelService) Delete(ctx context.Context, id uint) error {
	cnt, err := g.DB().Model("models").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("模型不存在")
	}
	_, err = g.DB().Model("models").Where("id", id).Delete()
	return err
}

// SetStatus 启停模型
func (s *modelService) SetStatus(ctx context.Context, id uint, status int) error {
	cnt, err := g.DB().Model("models").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("模型不存在")
	}
	_, err = g.DB().Model("models").Where("id", id).Update(g.Map{"status": status})
	return err
}
