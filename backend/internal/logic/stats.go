package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/service"
)

func init() {
	service.Stats = &statsService{}
}

type statsService struct{}

// Overview admin 全局维度；user 本人维度（契约 2.8）
func (s *statsService) Overview(ctx context.Context, userID uint, isAdmin bool) (map[string]interface{}, error) {
	if isAdmin {
		return s.adminOverview(ctx)
	}
	return s.userOverview(ctx, userID)
}

func (s *statsService) adminOverview(ctx context.Context) (map[string]interface{}, error) {
	providerCount, _ := g.DB().Model("providers").Count()
	modelCount, _ := g.DB().Model("models").Count()
	keyCount, _ := g.DB().Model("api_keys").Count()
	userCount, _ := g.DB().Model("users").Count()

	todayTokens, _ := g.DB().Model("chat_messages").
		Where("role", "assistant").Where(todayCondition("")).
		Sum("(prompt_tokens + completion_tokens)")
	todayCost, _ := g.DB().Model("chat_messages").
		Where("role", "assistant").Where(todayCondition("")).
		Sum("cost")
	totalCost, _ := g.DB().Model("chat_messages").
		Where("role", "assistant").Sum("cost")

	return map[string]interface{}{
		"provider_count": providerCount,
		"model_count":   modelCount,
		"key_count":     keyCount,
		"user_count":    userCount,
		"today_tokens":  int64(todayTokens),
		"today_cost":    todayCost,
		"total_cost":    totalCost,
	}, nil
}

func (s *statsService) userOverview(ctx context.Context, userID uint) (map[string]interface{}, error) {
	myKeyCount, _ := g.DB().Model("key_user").Where("user_id", userID).Count()
	mySessionCount, _ := g.DB().Model("chat_sessions").Where("user_id", userID).Count()

	msgModel := func() *gdb.Model {
		return g.DB().Model("chat_messages m").
			LeftJoin("chat_sessions s", "s.id = m.session_id").
			Where("s.user_id", userID).Where("m.role", "assistant")
	}
	todayTokens, _ := msgModel().Where(todayCondition("m")).Sum("(m.prompt_tokens + m.completion_tokens)")
	todayCost, _ := msgModel().Where(todayCondition("m")).Sum("m.cost")
	totalCost, _ := msgModel().Sum("m.cost")

	return map[string]interface{}{
		"my_key_count":     myKeyCount,
		"my_session_count": mySessionCount,
		"today_tokens":     int64(todayTokens),
		"today_cost":       todayCost,
		"total_cost":       totalCost,
	}, nil
}

// todayCondition 今日时间条件（双方言）；alias 为空表示不带表别名
func todayCondition(alias string) string {
	col := "created_at"
	if alias != "" {
		col = alias + "." + col
	}
	driver := g.Cfg().MustGet(nil, "database.default.type", "sqlite").String()
	if driver == "mysql" {
		return "DATE(" + col + ") = CURDATE()"
	}
	return "date(" + col + ") = date('now','localtime')"
}

// driverKind 当前数据库类型：mysql / sqlite（其余按 sqlite 兼容处理）
func driverKind() string {
	if g.Cfg().MustGet(nil, "database.default.type", "sqlite").String() == "mysql" {
		return "mysql"
	}
	return "sqlite"
}

// assistantModel 第 6 节统计统一基础模型：
// 仅统计 role='assistant' 消息；user 口径联 chat_sessions 并按 s.user_id 收敛，admin 不加用户过滤。
func (s *statsService) assistantModel(userID uint, isAdmin bool) *gdb.Model {
	m := g.DB().Model("chat_messages m").
		LeftJoin("chat_sessions s", "s.id = m.session_id").
		Where("m.role", "assistant")
	if !isAdmin {
		m = m.Where("s.user_id", userID)
	}
	return m
}

// dateGroupExpr 日期分组表达式（双方言）：SQLite date(col,'localtime')；MySQL DATE(col)。
// alias 为表别名（如 m），为空时不带别名。
func dateGroupExpr(alias string) string {
	col := "created_at"
	if alias != "" {
		col = alias + "." + col
	}
	if driverKind() == "mysql" {
		return "DATE(" + col + ")"
	}
	return "date(" + col + ",'localtime')"
}

// nonnegativeExpr 非负标量表达式：SQLite 用 MAX(a,b)；MySQL 用 GREATEST(a,b,0)。
func nonnegativeExpr(expr string) string {
	if driverKind() == "mysql" {
		return fmt.Sprintf("GREATEST(%s,0)", expr)
	}
	return fmt.Sprintf("MAX(%s,0)", expr)
}

// Rankings 排行榜（契约 6.1）：dim=model|provider；metric=calls|tokens|cost；按 metric 降序。
func (s *statsService) Rankings(ctx context.Context, userID uint, isAdmin bool, dim, metric string, limit int) ([]map[string]interface{}, error) {
	if dim != "provider" {
		dim = "model"
	}
	if metric != "calls" && metric != "tokens" && metric != "cost" {
		metric = "calls"
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	m := s.assistantModel(userID, isAdmin)
	var nameExpr, groupExpr string
	if dim == "provider" {
		// provider 维度：联 api_keys、providers；供应商缺失归为「未知供应商」
		m = m.
			LeftJoin("api_keys k", "k.id = s.key_id").
			LeftJoin("providers p", "p.id = k.provider_id")
		nameExpr = "IFNULL(p.name,'未知供应商')"
		groupExpr = nameExpr
	} else {
		// model 维度：按归并后的模型名分组，空串 / NULL 统一归入「未知模型」，避免重复成组
		nameExpr = "IFNULL(NULLIF(s.model,''),'未知模型')"
		groupExpr = nameExpr
	}

	rows, err := m.Fields(fmt.Sprintf(
		"%s AS name, COUNT(*) AS calls, SUM(m.prompt_tokens + m.completion_tokens) AS tokens, SUM(m.cost) AS cost",
		nameExpr)).
		Group(groupExpr).
		Order(metric + " DESC").
		Limit(0, limit).
		All()
	if err != nil {
		return nil, err
	}

	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]interface{}{
			"name":   r["name"].String(),
			"calls":  r["calls"].Int64(),
			"tokens": r["tokens"].Int64(),
			"cost":   r["cost"].Float64(),
		})
	}
	return out, nil
}

// Trend 近 N 个自然日每日聚合（契约 6.2）；无数据日期 Go 侧补 0，数组按日期升序连续。
func (s *statsService) Trend(ctx context.Context, userID uint, isAdmin bool, days int) ([]map[string]interface{}, error) {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	dateExpr := dateGroupExpr("m")
	rows, err := s.assistantModel(userID, isAdmin).
		Fields(fmt.Sprintf(
			"%s AS date, COUNT(*) AS calls, SUM(m.prompt_tokens + m.completion_tokens) AS tokens, SUM(m.cost) AS cost",
			dateExpr)).
		Group(dateExpr).
		All()
	if err != nil {
		return nil, err
	}

	// 数据库聚合结果按日期建立索引，缺失日期在 Go 侧补零
	agg := make(map[string]map[string]interface{}, len(rows))
	for _, r := range rows {
		agg[r["date"].String()] = map[string]interface{}{
			"calls":  r["calls"].Int64(),
			"tokens": r["tokens"].Int64(),
			"cost":   r["cost"].Float64(),
		}
	}

	out := make([]map[string]interface{}, 0, days)
	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		item := map[string]interface{}{"date": key, "calls": int64(0), "tokens": int64(0), "cost": float64(0)}
		if v, ok := agg[key]; ok {
			item["calls"] = v["calls"]
			item["tokens"] = v["tokens"]
			item["cost"] = v["cost"]
		}
		out = append(out, item)
	}
	return out, nil
}

// Breakdown Token 构成三项（契约 6.3），固定顺序与中文名。
// 说明：cost 统一返回 0 —— 费用精确拆分需按模型单价对各类 token 二次折算，本接口以 Token 构成为准。
func (s *statsService) Breakdown(ctx context.Context, userID uint, isAdmin bool) ([]map[string]interface{}, error) {
	// ①缓存命中输入：SUM(cached_tokens)
	cached, err := s.assistantModel(userID, isAdmin).Sum("m.cached_tokens")
	if err != nil {
		return nil, err
	}
	// ②未命中输入：MAX/GREATEST(prompt_tokens - cached_tokens, 0)；gdb Sum 会自动包外层 SUM
	unhitExpr := nonnegativeExpr("m.prompt_tokens - m.cached_tokens")
	unhit, err := s.assistantModel(userID, isAdmin).Sum(unhitExpr)
	if err != nil {
		return nil, err
	}
	// ③输出：SUM(completion_tokens)
	completion, err := s.assistantModel(userID, isAdmin).Sum("m.completion_tokens")
	if err != nil {
		return nil, err
	}

	return []map[string]interface{}{
		{"name": "缓存命中输入", "tokens": int64(cached), "cost": float64(0)},
		{"name": "未命中输入", "tokens": int64(unhit), "cost": float64(0)},
		{"name": "输出", "tokens": int64(completion), "cost": float64(0)},
	}, nil
}
