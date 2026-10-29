package logic

import (
	"database/sql"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"

	"apikey-manager/internal/pkg/errs"
)

// safeList 空结果兜底为空数组，避免 JSON 序列化为 null 导致前端遍历异常。
func safeList(records gdb.Result) gdb.Result {
	if records == nil {
		return gdb.Result{}
	}
	return records
}

// scanErr 把 Scan 无记录场景（sql.ErrNoRows）统一转为业务 400 NotFound，其余错误原样返回。
func scanErr(err error, notFoundMsg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errs.NotFound(notFoundMsg)
	}
	return err
}
