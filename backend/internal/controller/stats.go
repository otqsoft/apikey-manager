package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/middleware"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// Stats 统计控制器
var Stats = statsController{}

type statsController struct{}

// Overview admin 全局统计；user 本人统计
func (statsController) Overview(r *ghttp.Request) {
	data, err := service.Stats.Overview(r.Context(), middleware.UserID(r), middleware.IsAdmin(r))
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Rankings 排行榜：dim=model|provider，metric=calls|tokens|cost，limit 1~50
func (statsController) Rankings(r *ghttp.Request) {
	data, err := service.Stats.Rankings(
		r.Context(), middleware.UserID(r), middleware.IsAdmin(r),
		r.GetQuery("dim").String(), r.GetQuery("metric").String(), r.GetQuery("limit").Int())
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Trend 近 N 个自然日聚合（days 1~90，默认 30）
func (statsController) Trend(r *ghttp.Request) {
	data, err := service.Stats.Trend(r.Context(), middleware.UserID(r), middleware.IsAdmin(r), r.GetQuery("days").Int())
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}

// Breakdown Token 构成三项（缓存命中输入/未命中输入/输出）
func (statsController) Breakdown(r *ghttp.Request) {
	data, err := service.Stats.Breakdown(r.Context(), middleware.UserID(r), middleware.IsAdmin(r))
	if err != nil {
		fail(r, err)
		return
	}
	response.Success(r, data)
}
