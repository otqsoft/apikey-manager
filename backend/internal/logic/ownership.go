// Package logic 数据归属辅助：providers/models/api_keys 均带 owner_id（创建者），
// admin 全量可见可操作；普通用户仅能操作自己名下资源。
package logic

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/pkg/errs"
)

// applyOwnerScope 列表查询的数据归属过滤：非 admin 限定 owner_id = userID。
// column 传带表别名的列名（如 "k.owner_id"），避免 join 场景下被错误限定表名。
func applyOwnerScope(m *gdb.Model, userID uint, isAdmin bool, column string) *gdb.Model {
	if !isAdmin {
		m = m.Where(column, userID)
	}
	return m
}

// assertResourceOwned 单资源归属校验：资源不存在返回 NotFound；
// 存在但非 admin 且 owner_id != userID 返回 Forbidden。返回的 bool 表示资源存在。
func assertResourceOwned(ctx context.Context, table string, id, userID uint, isAdmin bool, label string) (bool, error) {
	rec, err := g.DB().Model(table).Where("id", id).Fields("owner_id").One()
	if err != nil {
		return false, err
	}
	if rec == nil || rec.IsEmpty() {
		return false, nil
	}
	if !isAdmin && rec["owner_id"].Uint() != userID {
		return true, errs.Forbidden("无权操作该" + label)
	}
	return true, nil
}

// assertProviderOwned 供应商归属校验（非 admin 创建模型/密钥时，供应商必须是自己名下）
func assertProviderOwned(ctx context.Context, providerID, userID uint, isAdmin bool) error {
	if isAdmin {
		return nil
	}
	rec, err := g.DB().Model("providers").Where("id", providerID).Fields("owner_id").One()
	if err != nil {
		return err
	}
	if rec == nil || rec.IsEmpty() {
		return errs.BadRequest("所属供应商不存在")
	}
	if rec["owner_id"].Uint() != userID {
		return errs.Forbidden("无权使用该供应商，请先在供应商管理中创建自己的供应商")
	}
	return nil
}
