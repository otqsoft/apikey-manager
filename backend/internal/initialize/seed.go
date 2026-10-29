package initialize

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

// Seed 首次启动种子数据：管理员账号 + 示例供应商/模型/密钥。
// 规则：users 非空则跳过管理员；providers 非空则跳过示例数据，绝不覆盖用户已有数据。
func Seed(ctx context.Context) error {
	if err := seedAdmin(ctx); err != nil {
		return err
	}
	if err := seedSamples(ctx); err != nil {
		return err
	}
	return nil
}

// seedAdmin 内置管理员（默认 admin/admin123，可在 config.yaml default 节修改）
func seedAdmin(ctx context.Context) error {
	cnt, err := g.DB().Model("users").Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	username := g.Cfg().MustGet(ctx, "default.username", "admin").String()
	password := g.Cfg().MustGet(ctx, "default.password", "admin123").String()
	display := g.Cfg().MustGet(ctx, "default.displayName", "系统管理员").String()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = g.DB().Model("users").Insert(g.Map{
		"username":      username,
		"password_hash": string(hash),
		"role":          "admin",
		"display_name":  display,
		"status":        1,
	})
	if err != nil {
		return err
	}
	g.Log().Infof(ctx, "已内置初始管理员账号: %s（请登录后立即修改默认密码）", username)
	return nil
}

// seedSamples 契约第 5 节示例数据，全表为空时插入，便于首次进入看到界面形态。
func seedSamples(ctx context.Context) error {
	cnt, err := g.DB().Model("providers").Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	remarkSample := "示例数据，请勿直接用于生产环境"

	// 示例供应商 1：OpenAI 兼容格式
	pid1, err := g.DB().Model("providers").InsertAndGetId(g.Map{
		"name":            "示例·OpenAI 兼容平台",
		"api_format":      "openai",
		"base_url":       "https://api.openai.com/v1",
		"platform_url":    "https://platform.openai.com",
		"balance":         0,
		"max_concurrency": 10,
		"timeout_seconds": 60,
		"status":          1,
		"remark":          remarkSample,
	})
	if err != nil {
		return err
	}
	// 示例供应商 2：Anthropic 格式
	pid2, err := g.DB().Model("providers").InsertAndGetId(g.Map{
		"name":            "示例·Anthropic Claude 平台",
		"api_format":      "anthropic",
		"base_url":       "https://api.anthropic.com",
		"platform_url":    "https://console.anthropic.com",
		"balance":         0,
		"max_concurrency": 10,
		"timeout_seconds": 60,
		"status":          1,
		"remark":          remarkSample,
	})
	if err != nil {
		return err
	}
	// 示例模型 1
	if _, err := g.DB().Model("models").Insert(g.Map{
		"name":               "gpt-4o",
		"provider_id":        pid1,
		"input_price_cached": 2.5,
		"input_price":        5.0,
		"output_price":       15.0,
		"context_length":    128000,
		"support_thinking":   false,
		"support_multimodal": true,
		"status":             1,
		"remark":             remarkSample,
	}); err != nil {
		return err
	}
	// 示例模型 2
	if _, err := g.DB().Model("models").Insert(g.Map{
		"name":               "claude-3-5-sonnet",
		"provider_id":        pid2,
		"input_price_cached": 1.5,
		"input_price":        3.0,
		"output_price":       15.0,
		"context_length":    200000,
		"support_thinking":   true,
		"support_multimodal": true,
		"status":             1,
		"remark":             remarkSample,
	}); err != nil {
		return err
	}
	// 示例密钥（掩码占位值，不可真实调用）
	if _, err := g.DB().Model("api_keys").Insert(g.Map{
		"name":        "示例密钥-1（掩码占位）",
		"provider_id": pid1,
		"api_key":     "sk-example-xxxx000011112222",
		"expire_date": "",
		"max_quota":   0,
		"used_quota":  0,
		"status":      1,
		"remark":      "示例数据，仅用于界面演示；接入真实平台后请替换为有效密钥",
	}); err != nil {
		return err
	}
	g.Log().Info(ctx, "已写入示例供应商/模型/密钥数据")
	return nil
}
