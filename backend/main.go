// Package main 多平台大模型 API-KEY 集中管理工具 —— GoFrame v2 后端入口。
// 启动方式：
//   - 默认（推荐）：go run . 或 ./bin/api，框架自动搜索 manifest/config/config.yaml
//   - 指定配置：./bin/api -c manifest/config/config.yaml（兼容写法）
//     或框架原生：./bin/api --gf.gcfg.file=manifest/config/config.yaml
package main

import (
	// MySQL 驱动（当前默认数据库，连接参数见 manifest/config/config.yaml database.default）
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	// 切换回 SQLite 时启用下面一行（纯 Go 无 CGO），并把 config.yaml 的 type 改回 sqlite：
	// _ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"apikey-manager/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.New())
}
