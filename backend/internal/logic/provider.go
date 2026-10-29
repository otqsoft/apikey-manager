package logic

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/service"
)

func init() {
	service.Provider = &providerService{}
}

type providerService struct{}

// List 供应商分页列表；status=-1 表示不按状态过滤
func (s *providerService) List(ctx context.Context, page, size int, keyword string, status int) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	m := g.DB().Model("providers")
	if keyword != "" {
		m = m.WhereLike("name", "%"+keyword+"%")
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

// Detail 供应商详情
func (s *providerService) Detail(ctx context.Context, id uint) (*model.Provider, error) {
	var p model.Provider
	if err := g.DB().Model("providers").Where("id", id).Scan(&p); err != nil {
		return nil, scanErr(err, "供应商不存在")
	}
	if p.ID == 0 {
		return nil, errs.NotFound("供应商不存在")
	}
	return &p, nil
}

// Create 新建供应商，补默认值并校验名称唯一
func (s *providerService) Create(ctx context.Context, req *model.ProviderSaveReq) (uint, error) {
	if req.ApiFormat != "openai" && req.ApiFormat != "anthropic" {
		return 0, errs.BadRequest("api_format 仅支持 openai / anthropic")
	}
	cnt, err := g.DB().Model("providers").Where("name", req.Name).Count()
	if err != nil {
		return 0, err
	}
	if cnt > 0 {
		return 0, errs.BadRequest("供应商名称已存在")
	}
	data := s.applyDefaults(req)
	id, err := g.DB().Model("providers").InsertAndGetId(data)
	return uint(id), err
}

// Update 更新供应商
func (s *providerService) Update(ctx context.Context, id uint, req *model.ProviderSaveReq) error {
	cnt, err := g.DB().Model("providers").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("供应商不存在")
	}
	if req.ApiFormat != "openai" && req.ApiFormat != "anthropic" {
		return errs.BadRequest("api_format 仅支持 openai / anthropic")
	}
	cnt, err = g.DB().Model("providers").Where("name", req.Name).WhereNot("id", id).Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errs.BadRequest("供应商名称已存在")
	}
	_, err = g.DB().Model("providers").Where("id", id).Update(s.applyDefaults(req))
	return err
}

// Delete 删除前校验：名下存在模型或密钥则拒绝
func (s *providerService) Delete(ctx context.Context, id uint) error {
	cnt, err := g.DB().Model("providers").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("供应商不存在")
	}
	modelCnt, err := g.DB().Model("models").Where("provider_id", id).Count()
	if err != nil {
		return err
	}
	if modelCnt > 0 {
		return errs.BadRequest("该供应商下仍绑定模型，请先删除或迁移模型后再删除供应商")
	}
	keyCnt, err := g.DB().Model("api_keys").Where("provider_id", id).Count()
	if err != nil {
		return err
	}
	if keyCnt > 0 {
		return errs.BadRequest("该供应商下仍绑定 API-KEY，请先删除或迁移密钥后再删除供应商")
	}
	_, err = g.DB().Model("providers").Where("id", id).Delete()
	return err
}

// SetStatus 启停供应商
func (s *providerService) SetStatus(ctx context.Context, id uint, status int) error {
	cnt, err := g.DB().Model("providers").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("供应商不存在")
	}
	_, err = g.DB().Model("providers").Where("id", id).Update(g.Map{"status": status})
	return err
}

// applyDefaults 补全默认值（契约默认：并发 10、超时 60、状态启用）
func (s *providerService) applyDefaults(req *model.ProviderSaveReq) g.Map {
	maxConcurrency := req.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}
	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 60
	}
	status := req.Status
	if status != 0 && status != 1 {
		status = 1
	}
	return g.Map{
		"name":            req.Name,
		"api_format":      req.ApiFormat,
		"base_url":       req.BaseURL,
		"platform_url":    req.PlatformURL,
		"balance":         req.Balance,
		"max_concurrency": maxConcurrency,
		"timeout_seconds": timeout,
		"status":          status,
		"remark":          req.Remark,
	}
}
