// Package main 多平台大模型 API-KEY 集中管理工具 —— GoFrame v2 后端入口。
// 启动方式：
//   - 默认（推荐）：go run . 或 ./bin/api，框架自动搜索 manifest/config/config.yaml
//   - 指定配置：./bin/api -c manifest/config/config.yaml（兼容写法）
//     或框架原生：./bin/api --gf.gcfg.file=manifest/config/config.yaml
package main

import (
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2" // SQLite 纯 Go 驱动（glebarez，无 CGO）
	// 切换 MySQL 时启用下面一行（同时在 config.yaml 中把 database.default.type 改为 mysql）：
	// _ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"apikey-manager/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.New())
}
