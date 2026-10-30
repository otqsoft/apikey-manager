// Package cmd 命令入口：初始化数据库、注册路由、启动 HTTP 服务。
package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gcmd"

	"apikey-manager/internal/controller"
	"apikey-manager/internal/initialize"
	"apikey-manager/internal/middleware"

	// 触发 logic 包 init()，完成 service 接口绑定
	_ "apikey-manager/internal/logic"
)

// Main 主命令
var Main = gcmd.Command{
	Name:  "apikey-manager",
	Usage: "apikey-manager（默认自动搜索 manifest/config/config.yaml；可用 -c 指定配置文件）",
	Brief: "多平台大模型 API-KEY 集中管理工具（GoFrame v2 后端）",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		// 配置文件加载（必须在 g.DB() / g.Server() 首次使用前完成，
		// 保证数据库连接与监听地址均来自指定配置）：
		//   1) 框架原生：--gf.gcfg.file=path 或环境变量 GF_GCFG_FILE=path
		//   2) 习惯写法兼容：-c path / --c=path
		cfgFile := parser.GetOpt("c").String()
		if cfgFile == "" {
			cfgFile = parser.GetOpt("gf.gcfg.file").String()
		}
		if cfgFile != "" {
			if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
				adapter.SetFileName(cfgFile)
				adapter.Clear() // 清掉默认配置缓存，强制按指定文件重读
			}
		}

		// 1. 自动建表 + 首次启动种子数据
		if err := initialize.CreateTables(ctx); err != nil {
			return err
		}
		if err := initialize.Seed(ctx); err != nil {
			return err
		}

		// 2. 注册路由并启动服务（监听地址来自 config.yaml server.address，默认 :8090）
		s := g.Server()
		s.Group("/api", func(group *ghttp.RouterGroup) {
			// 公开接口：登录、注册（无需 JWT；注册受 app.allow_register 开关控制）
			group.POST("/auth/login", controller.Auth.Login)
			group.POST("/auth/register", controller.Auth.Register)

			// 需登录接口组
			group.Group("/", func(auth *ghttp.RouterGroup) {
				auth.Middleware(middleware.JwtAuth)

				// 认证
				auth.POST("/auth/logout", controller.Auth.Logout)
				auth.GET("/auth/profile", controller.Auth.Profile)
				auth.PUT("/auth/password", controller.Auth.ChangePassword)

				// 普通用户：我的可用密钥
				auth.GET("/me/keys", controller.Key.MeList)

				// 供应商 / 模型 / 密钥管理：admin 全量；普通用户仅自己名下数据
				// （数据归属过滤与归属校验统一在 logic 层按 owner_id 执行）
				auth.GET("/providers", controller.Provider.List)
				auth.POST("/providers", controller.Provider.Create)
				auth.GET("/providers/:id", controller.Provider.Detail)
				auth.PUT("/providers/:id", controller.Provider.Update)
				auth.DELETE("/providers/:id", controller.Provider.Delete)
				auth.PUT("/providers/:id/status", controller.Provider.SetStatus)

				auth.GET("/models", controller.Model.List)
				auth.POST("/models", controller.Model.Create)
				auth.GET("/models/:id", controller.Model.Detail)
				auth.PUT("/models/:id", controller.Model.Update)
				auth.DELETE("/models/:id", controller.Model.Delete)
				auth.PUT("/models/:id/status", controller.Model.SetStatus)

				auth.GET("/keys", controller.Key.List)
				// 详情（返回完整明文 api_key，供列表页「复制完整 Key」使用）：
				// 归属校验在 logic 层，admin 全量，普通用户限 owner 或被分配者
				auth.GET("/keys/:id", controller.Key.Detail)
				auth.POST("/keys", controller.Key.Create)
				auth.PUT("/keys/:id", controller.Key.Update)
				auth.DELETE("/keys/:id", controller.Key.Delete)
				auth.PUT("/keys/:id/status", controller.Key.SetStatus)
				auth.GET("/keys/:id/users", controller.Key.ListUsers)
				auth.PUT("/keys/:id/users", controller.Key.AssignUsers)

				// 对话（admin / user 均可，数据归属在逻辑层过滤）
				auth.GET("/chat/sessions", controller.Chat.ListSessions)
				auth.POST("/chat/sessions", controller.Chat.CreateSession)
				auth.GET("/chat/sessions/:id/messages", controller.Chat.Messages)
				auth.DELETE("/chat/sessions/:id", controller.Chat.DeleteSession)
				auth.POST("/chat/completions", controller.Chat.Completions)

				// 统计（角色在逻辑层区分全局/本人）
				auth.GET("/stats/overview", controller.Stats.Overview)
				auth.GET("/stats/rankings", controller.Stats.Rankings)
				auth.GET("/stats/trend", controller.Stats.Trend)
				auth.GET("/stats/breakdown", controller.Stats.Breakdown)

				// 用户列表：登录即可读（「分配用户」弹窗需要选择用户）
				auth.GET("/users", controller.User.List)

				// 仅管理员：用户管理（增删改、启停、重置密码、查看分配）
				auth.Group("/", func(admin *ghttp.RouterGroup) {
					admin.Middleware(middleware.RequireAdmin)
					admin.POST("/users", controller.User.Create)
					admin.GET("/users/:id", controller.User.Detail)
					admin.PUT("/users/:id", controller.User.Update)
					admin.PUT("/users/:id/status", controller.User.SetStatus)
					admin.POST("/users/:id/reset-password", controller.User.ResetPassword)
					admin.DELETE("/users/:id", controller.User.Delete)
					admin.GET("/users/:id/keys", controller.User.AssignedKeys)
				})
			})
		})

		s.Run()
		return nil
	},
}
