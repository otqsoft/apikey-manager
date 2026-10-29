package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/pkg/util"
	"apikey-manager/internal/service"
)

func init() {
	service.Key = &keyService{}
}

type keyService struct{}

// List admin 全量分页列表；api_key 掩码回显，附带 used_quota/max_quota
func (s *keyService) List(ctx context.Context, page, size int, providerID uint, keyword string) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	m := g.DB().Model("api_keys k").
		LeftJoin("providers p", "p.id = k.provider_id")
	if providerID > 0 {
		m = m.Where("k.provider_id", providerID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		// 显式括号包裹 OR 组：与 provider_id 的 AND 条件组合时不被优先级穿透
		m = m.Where("(k.name LIKE ? OR k.api_key LIKE ?)", like, like)
	}
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	records, err := m.Fields("k.*, p.name as provider_name, p.api_format").
		Page(page, size).OrderDesc("k.id").All()
	if err != nil {
		return nil, err
	}
	// 掩码处理：列表绝不返回完整 key
	for _, r := range records {
		raw := r["api_key"].String()
		masked := util.MaskAPIKey(raw)
		r["key_masked"] = gvar.New(masked)
		r["api_key"] = gvar.New(masked)
	}
	return &model.PageResult{List: safeList(records), Total: total, Page: page, PageSize: size}, nil
}

// Detail 详情返回完整 api_key；admin 全量，普通用户仅限被分配者
func (s *keyService) Detail(ctx context.Context, id, userID uint, isAdmin bool) (*model.ApiKey, error) {
	var k model.ApiKey
	if err := g.DB().Model("api_keys").Where("id", id).Scan(&k); err != nil {
		return nil, scanErr(err, "密钥不存在")
	}
	if k.ID == 0 {
		return nil, errs.NotFound("密钥不存在")
	}
	if !isAdmin {
		cnt, err := g.DB().Model("key_user").Where("key_id", id).Where("user_id", userID).Count()
		if err != nil {
			return nil, err
		}
		if cnt == 0 {
			return nil, errs.Forbidden("无权查看该密钥")
		}
	}
	return &k, nil
}

// Create 新建密钥：事务内先写 api_keys，再幂等写 key_user（创建者自动可见于 /me/keys）
func (s *keyService) Create(ctx context.Context, creatorID uint, req *model.KeySaveReq) (uint, error) {
	if req.APIKey == "" {
		return 0, errs.BadRequest("api_key 不能为空")
	}
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

	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	id, err := tx.Model("api_keys").InsertAndGetId(g.Map{
		"name":        req.Name,
		"provider_id": req.ProviderID,
		"api_key":     req.APIKey,
		"expire_date": req.ExpireDate,
		"max_quota":   req.MaxQuota,
		"used_quota":  0,
		"status":      status,
		"remark":      req.Remark,
	})
	if err != nil {
		return 0, err
	}
	newID := uint(id)

	// 创建者幂等分配：先 Count 存在性，避免唯一约束冲突/重复分配
	if creatorID > 0 {
		assigned, err := tx.Model("key_user").
			Where("key_id", newID).Where("user_id", creatorID).Count()
		if err != nil {
			return 0, err
		}
		if assigned == 0 {
			if _, err := tx.Model("key_user").Insert(g.Map{
				"key_id": newID,
				"user_id": creatorID,
			}); err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newID, nil
}

// Update 更新密钥；api_key 留空表示不修改明文
func (s *keyService) Update(ctx context.Context, id uint, req *model.KeySaveReq) error {
	cnt, err := g.DB().Model("api_keys").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("密钥不存在")
	}
	cnt, err = g.DB().Model("providers").Where("id", req.ProviderID).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.BadRequest("所属供应商不存在")
	}
	data := g.Map{
		"name":        req.Name,
		"provider_id": req.ProviderID,
		"expire_date": req.ExpireDate,
		"max_quota":   req.MaxQuota,
		"status":      req.Status,
		"remark":      req.Remark,
	}
	if req.APIKey != "" {
		data["api_key"] = req.APIKey
	}
	_, err = g.DB().Model("api_keys").Where("id", id).Update(data)
	return err
}

// Delete 删除密钥并级联清理分配关系
func (s *keyService) Delete(ctx context.Context, id uint) error {
	cnt, err := g.DB().Model("api_keys").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("密钥不存在")
	}
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Model("key_user").Where("key_id", id).Delete(); err != nil {
		return err
	}
	if _, err := tx.Model("api_keys").Where("id", id).Delete(); err != nil {
		return err
	}
	return tx.Commit()
}

// SetStatus 启停密钥
func (s *keyService) SetStatus(ctx context.Context, id uint, status int) error {
	cnt, err := g.DB().Model("api_keys").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("密钥不存在")
	}
	_, err = g.DB().Model("api_keys").Where("id", id).Update(g.Map{"status": status})
	return err
}

// ListUsers 已分配用户简表 + user_id 数组
func (s *keyService) ListUsers(ctx context.Context, id uint) ([]map[string]interface{}, error) {
	records, err := g.DB().Model("key_user ku").
		LeftJoin("users u", "u.id = ku.user_id").
		Fields("u.id as user_id, u.username, u.display_name, u.role, u.status").
		Where("ku.key_id", id).
		OrderAsc("u.id").
		All()
	if err != nil {
		return nil, err
	}
	// 无分配记录时也返回空数组（非 null），保证前端 v-for 可直接遍历
	users := make([]map[string]interface{}, 0)
	userIDs := make([]uint, 0)
	for _, r := range records {
		users = append(users, r.Map())
		userIDs = append(userIDs, r["user_id"].Uint())
	}
	// 契约要求：user_id 数组 + 用户简表，一次性返回
	return []map[string]interface{}{
		{"user_ids": userIDs, "users": users},
	}, nil
}

// AssignUsers 全量覆盖分配关系（去重后替换）
func (s *keyService) AssignUsers(ctx context.Context, id uint, userIDs []uint) error {
	cnt, err := g.DB().Model("api_keys").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("密钥不存在")
	}
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Model("key_user").Where("key_id", id).Delete(); err != nil {
		return err
	}
	seen := map[uint]bool{}
	for _, uid := range userIDs {
		if uid == 0 || seen[uid] {
			continue
		}
		seen[uid] = true
		exists, err := tx.Model("users").Where("id", uid).Count()
		if err != nil || exists == 0 {
			continue
		}
		if _, err := tx.Model("key_user").Insert(g.Map{"key_id": id, "user_id": uid}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// MeList 普通用户核心接口：全量返回已分配密钥，附带可用性标记（不可用项前端置灰并展示原因）
func (s *keyService) MeList(ctx context.Context, userID uint) ([]model.MeKeyItem, error) {
	records, err := g.DB().Model("key_user ku").
		LeftJoin("api_keys k", "k.id = ku.key_id").
		LeftJoin("providers p", "p.id = k.provider_id").
		Fields("k.*, p.name as provider_name, p.api_format, p.base_url").
		Where("ku.user_id", userID).
		All()
	if err != nil {
		return nil, err
	}
	list := make([]model.MeKeyItem, 0, len(records))
	for _, r := range records {
		k := model.ApiKey{
			ID:         r["id"].Uint(),
			Name:       r["name"].String(),
			ProviderID: r["provider_id"].Uint(),
			APIKey:     r["api_key"].String(),
			ExpireDate: r["expire_date"].String(),
			MaxQuota:   r["max_quota"].Float64(),
			UsedQuota:  r["used_quota"].Float64(),
			Status:     r["status"].Int(),
			Remark:     r["remark"].String(),
		}
		p := model.Provider{
			ID:        r["provider_id"].Uint(),
			Name:      r["provider_name"].String(),
			ApiFormat: r["api_format"].String(),
			BaseURL:   r["base_url"].String(),
			Status:    r["status"].Int(), // join 后同名字段歧义，下面单独覆盖
		}
		// 单独取供应商启用状态，避免同名字段冲突
		pStatus, err := g.DB().Model("providers").Where("id", p.ID).Value("status")
		if err != nil || pStatus == nil {
			p.Status = 0
		} else {
			p.Status = pStatus.Int()
		}
		item := model.MeKeyItem{
			ID:         k.ID,
			Name:       k.Name,
			ProviderID: k.ProviderID,
			KeyMasked:  util.MaskAPIKey(k.APIKey),
			ExpireDate: k.ExpireDate,
			MaxQuota:   k.MaxQuota,
			UsedQuota:  k.UsedQuota,
			Status:     k.Status,
			Remark:     k.Remark,
			Provider:   model.MeProvider{ID: p.ID, Name: p.Name, ApiFormat: p.ApiFormat, BaseURL: p.BaseURL},
			Models:     []model.MeModel{},
			Usable:     true,
		}
		// 可用性判定：不可用项不再静默过滤，返回标记与原因供前端置灰提示
		if p.ID == 0 {
			item.Usable = false
			item.UnusableReason = "所属供应商已删除"
		} else if uerr := validateUsable(k, p); uerr != nil {
			item.Usable = false
			item.UnusableReason = uerr.Error()
		}
		// 该供应商下启用模型列表
		modelRecords, err := g.DB().Model("models").
			Where("provider_id", p.ID).Where("status", 1).
			OrderAsc("id").All()
		if err == nil {
			for _, mr := range modelRecords {
				item.Models = append(item.Models, model.MeModel{
					ID:                mr["id"].Uint(),
					Name:              mr["name"].String(),
					ContextLength:     mr["context_length"].Int(),
					SupportThinking:   mr["support_thinking"].Bool(),
					SupportMultimodal: mr["support_multimodal"].Bool(),
				})
			}
		}
		list = append(list, item)
	}
	return list, nil
}

// CheckAvailable 对话前可用性校验，返回密钥与供应商（含超时、并发参数）
func (s *keyService) CheckAvailable(ctx context.Context, keyID uint) (*model.ApiKey, *model.Provider, error) {
	var k model.ApiKey
	if err := g.DB().Model("api_keys").Where("id", keyID).Scan(&k); err != nil {
		return nil, nil, err
	}
	if k.ID == 0 {
		return nil, nil, errs.NotFound("密钥不存在")
	}
	var p model.Provider
	if err := g.DB().Model("providers").Where("id", k.ProviderID).Scan(&p); err != nil {
		return nil, nil, err
	}
	if p.ID == 0 {
		return nil, nil, errs.BadRequest("密钥所属供应商不存在")
	}
	if err := validateUsable(k, p); err != nil {
		return nil, nil, err
	}
	return &k, &p, nil
}

// validateUsable 可用性判定（契约 1.3）
func validateUsable(k model.ApiKey, p model.Provider) error {
	if k.Status != 1 {
		return errs.BadRequest("密钥已停用，请联系管理员启用")
	}
	if p.Status != 1 {
		return errs.BadRequest("供应商「" + p.Name + "」已停用，暂不可用")
	}
	if k.ExpireDate != "" {
		today := time.Now().Format("2006-01-02")
		if k.ExpireDate < today {
			return errs.BadRequest("密钥已过期（有效期至 " + k.ExpireDate + "）")
		}
	}
	if k.MaxQuota > 0 && k.UsedQuota >= k.MaxQuota {
		return errs.BadRequest("密钥额度已用完（已用额度 " + formatFloat(k.UsedQuota) + " 元 / 限额 " + formatFloat(k.MaxQuota) + " 元）")
	}
	return nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
