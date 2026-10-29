package logic

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/pkg/util"
	"apikey-manager/internal/service"
)

func init() {
	service.User = &userService{}
}

type userService struct{}

// List 用户分页列表（排除密码哈希字段）
func (s *userService) List(ctx context.Context, page, size int, keyword string) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	m := g.DB().Model("users")
	if keyword != "" {
		like := "%" + keyword + "%"
		// 显式括号包裹 OR 组，避免与其他 AND 条件组合时被优先级穿透
		m = m.Where("username LIKE ? OR display_name LIKE ?", like, like)
	}
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	records, err := m.Fields("id,username,role,display_name,status,created_at,updated_at").
		Page(page, size).OrderDesc("id").All()
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: safeList(records), Total: total, Page: page, PageSize: size}, nil
}

// Create 新建用户
func (s *userService) Create(ctx context.Context, req *model.UserSaveReq) (uint, error) {
	cnt, err := g.DB().Model("users").Where("username", req.Username).Count()
	if err != nil {
		return 0, err
	}
	if cnt > 0 {
		return 0, errs.BadRequest("账号已存在")
	}
	role := req.Role
	if role == "" {
		role = "user"
	}
	if role != "admin" && role != "user" {
		return 0, errs.BadRequest("角色非法，仅支持 admin/user")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	status := req.Status
	if status == 0 && req.Status == 0 {
		status = 1 // 默认启用
	}
	id, err := g.DB().Model("users").InsertAndGetId(g.Map{
		"username":      req.Username,
		"password_hash": string(hash),
		"role":          role,
		"display_name": req.DisplayName,
		"status":        status,
	})
	return uint(id), err
}

// Detail 用户详情（不返回哈希）
func (s *userService) Detail(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	if err := g.DB().Model("users").Where("id", id).Scan(&u); err != nil {
		return nil, scanErr(err, "用户不存在")
	}
	if u.ID == 0 {
		return nil, errs.NotFound("用户不存在")
	}
	u.PasswordHash = ""
	return &u, nil
}

// Update 更新用户基础信息（不含密码）
func (s *userService) Update(ctx context.Context, id uint, req *model.UserUpdateReq) error {
	cnt, err := g.DB().Model("users").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("用户不存在")
	}
	if req.Role != "" && req.Role != "admin" && req.Role != "user" {
		return errs.BadRequest("角色非法")
	}
	_, err = g.DB().Model("users").Where("id", id).Update(g.Map{
		"role":         req.Role,
		"display_name": req.DisplayName,
		"status":       req.Status,
	})
	return err
}

// SetStatus 启停用户
func (s *userService) SetStatus(ctx context.Context, id uint, status int) error {
	cnt, err := g.DB().Model("users").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("用户不存在")
	}
	_, err = g.DB().Model("users").Where("id", id).Update(g.Map{"status": status})
	return err
}

// ResetPassword 管理员重置密码
func (s *userService) ResetPassword(ctx context.Context, id uint, password string) error {
	cnt, err := g.DB().Model("users").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("用户不存在")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = g.DB().Model("users").Where("id", id).Update(g.Map{"password_hash": string(hash)})
	return err
}

// Delete 删除用户：禁止删除自己；事务内清理分配关系与会话消息
func (s *userService) Delete(ctx context.Context, id, currentUserID uint) error {
	if id == currentUserID {
		return errs.Forbidden("不能删除当前登录账号")
	}
	cnt, err := g.DB().Model("users").Where("id", id).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.NotFound("用户不存在")
	}
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Model("key_user").Where("user_id", id).Delete(); err != nil {
		return err
	}
	// 清理该用户的会话与消息，保持数据一致
	sessionIDs, err := tx.Model("chat_sessions").Fields("id").Where("user_id", id).Array("id")
	if err == nil && len(sessionIDs) > 0 {
		if _, err := tx.Model("chat_messages").WhereIn("session_id", sessionIDs).Delete(); err != nil {
			return err
		}
	}
	if _, err := tx.Model("chat_sessions").Where("user_id", id).Delete(); err != nil {
		return err
	}
	if _, err := tx.Model("users").Where("id", id).Delete(); err != nil {
		return err
	}
	return tx.Commit()
}

// AssignedKeys 该用户已分配密钥列表（掩码），用户管理弹窗使用
func (s *userService) AssignedKeys(ctx context.Context, userID uint) ([]map[string]interface{}, error) {
	records, err := g.DB().Model("key_user ku").
		LeftJoin("api_keys k", "k.id = ku.key_id").
		LeftJoin("providers p", "p.id = k.provider_id").
		Fields("k.id, k.name, k.provider_id, p.name as provider_name, k.api_key, k.expire_date, k.max_quota, k.used_quota, k.status, k.remark").
		Where("ku.user_id", userID).
		OrderDesc("k.id").
		All()
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(records))
	for _, r := range records {
		item := r.Map()
		if raw, ok := item["api_key"].(string); ok {
			item["key_masked"] = util.MaskAPIKey(raw)
		}
		delete(item, "api_key")
		list = append(list, item)
	}
	return list, nil
}
