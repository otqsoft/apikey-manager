// Package logic 数据归属辅助。
//
// 权限模型：
//   - providers / models：共享资源。所有登录用户可读（创建 API-KEY 时选用），
//     增删改仅管理员（路由级 RequireAdmin）。owner_id 仅用于记录创建者。
//   - api_keys：严格按归属隔离。每个用户（含管理员）只管理自己名下的密钥，
//     owner 可把自己的 Key 分配给其他用户使用（key_user）。
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

// assertProviderOwned 供应商归属校验（模型管理为管理员操作，admin 旁路）
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

// assertProviderExists 供应商存在性校验：供应商为共享资源，创建/编辑 API-KEY
// 时可选用任意供应商（不再要求归属）
func assertProviderExists(ctx context.Context, providerID uint) error {
	cnt, err := g.DB().Model("providers").Where("id", providerID).Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errs.BadRequest("所属供应商不存在")
	}
	return nil
}
