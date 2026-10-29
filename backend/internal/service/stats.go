package service

import "context"

// IStatsService 统计服务（契约 2.8 概览 + 第 6 节三个新增接口）
type IStatsService interface {
	// Overview admin 全局维度；user 本人维度
	Overview(ctx context.Context, userID uint, isAdmin bool) (map[string]interface{}, error)
	// Rankings 排行榜：dim=model|provider，metric=calls|tokens|cost，limit 1~50
	Rankings(ctx context.Context, userID uint, isAdmin bool, dim, metric string, limit int) ([]map[string]interface{}, error)
	// Trend 近 days 个自然日每日聚合（days 1~90，无数据日期补 0，日期升序连续）
	Trend(ctx context.Context, userID uint, isAdmin bool, days int) ([]map[string]interface{}, error)
	// Breakdown Token 构成三项：缓存命中输入/未命中输入/输出；cost 固定 0（仅 Token 口径）
	Breakdown(ctx context.Context, userID uint, isAdmin bool) ([]map[string]interface{}, error)
}

var Stats IStatsService
