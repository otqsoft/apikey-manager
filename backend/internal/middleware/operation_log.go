package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/service"
)

// 请求参数中需要脱敏的字段（不写入操作日志 detail）
var sensitiveKeys = map[string]bool{
	"password":     true,
	"old_password": true,
	"new_password": true,
	"api_key":      true,
	"token":        true,
	"confirm":      true,
}

// OperationLog 操作日志中间件：记录已登录用户的写操作（POST/PUT/DELETE）。
// 需挂载在 JwtAuth 之后，以便从上下文读取操作人信息。记录失败不影响主流程。
func OperationLog(r *ghttp.Request) {
	method := r.Method
	// 仅记录写操作；GET/HEAD/OPTIONS 不记录，避免查询与预检请求刷屏
	isWrite := method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
	if !isWrite {
		r.Middleware.Next()
		return
	}

	path := r.URL.Path
	userID := UserID(r)
	username := Username(r)
	module, action := describeOperation(method, path)
	detail := buildDetail(r)

	r.Middleware.Next()

	// Next 之后读取业务码判定成功/失败（流式等未设置场景默认成功）
	status := 1
	if response.BizCode(r) != response.CodeOK {
		status = 0
	}

	// 昵称
	displayName := username
	if userID > 0 {
		if name, err := g.DB().Model("users").Where("id", userID).Value("display_name"); err == nil && name.String() != "" {
			displayName = name.String()
		}
	}

	entry := &model.OperationLog{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		Module:      module,
		Action:      action,
		Method:      method,
		Path:        path,
		Detail:      detail,
		IP:          r.GetClientIp(),
		UserAgent:   r.Header.Get("User-Agent"),
		Status:      status,
	}
	if err := service.Log.RecordOperation(r.Context(), entry); err != nil {
		g.Log().Warningf(r.Context(), "写操作日志失败: %v", err)
	}
}

// describeOperation 根据 HTTP 方法与路径推断「模块」与「动作」的中文描述。
func describeOperation(method, path string) (module, action string) {
	p := strings.TrimPrefix(path, "/api")
	switch {
	case strings.HasPrefix(p, "/providers"):
		module = "供应商管理"
	case strings.HasPrefix(p, "/models"):
		module = "模型管理"
	case strings.HasPrefix(p, "/keys"), strings.HasPrefix(p, "/me/keys"):
		module = "API-KEY 管理"
	case strings.HasPrefix(p, "/users"):
		module = "用户管理"
	case strings.HasPrefix(p, "/chat"):
		module = "对话测试"
	case strings.HasPrefix(p, "/auth"):
		module = "认证"
	default:
		module = "其他"
	}

	switch method {
	case http.MethodPost:
		switch {
		case strings.HasSuffix(p, "/reset-password"):
			action = "重置密码"
		case strings.Contains(p, "/chat/completions"):
			action = "发起对话"
		case strings.HasSuffix(p, "/sessions"):
			action = "创建会话"
		case strings.HasSuffix(p, "/register"):
			action = "注册"
		case strings.HasSuffix(p, "/logout"):
			action = "退出登录"
		default:
			action = "新增"
		}
	case http.MethodPut:
		switch {
		case strings.HasSuffix(p, "/status"):
			action = "状态变更"
		case strings.HasSuffix(p, "/users"):
			action = "分配用户"
		case strings.Contains(p, "/password"):
			action = "修改密码"
		default:
			action = "更新"
		}
	case http.MethodDelete:
		action = "删除"
	default:
		action = method
	}
	return
}

// buildDetail 收集请求参数并脱敏，序列化为 JSON 字符串（截断到 1000 字符）。
func buildDetail(r *ghttp.Request) string {
	params := r.GetMap()
	for k := range params {
		if sensitiveKeys[strings.ToLower(k)] {
			params[k] = "******"
		}
	}
	if len(params) == 0 {
		return ""
	}
	b, err := json.Marshal(params)
	if err != nil {
		return ""
	}
	const maxLen = 1000
	if len(b) > maxLen {
		return string(b[:maxLen]) + "..."
	}
	return string(b)
}
