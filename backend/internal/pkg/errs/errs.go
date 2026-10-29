// Package errs 业务错误：携带契约响应码（400/401/403），控制器统一转换为响应体。
package errs

import (
	"errors"

	"apikey-manager/internal/pkg/response"
)

// BizError 带业务码的错误
type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

// New 构造业务错误
func New(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// BadRequest 400
func BadRequest(msg string) *BizError { return New(response.CodeBadReq, msg) }

// Forbidden 403
func Forbidden(msg string) *BizError { return New(response.CodeForbid, msg) }

// Unauthorized 401
func Unauthorized(msg string) *BizError { return New(response.CodeNoAuth, msg) }

// NotFound 资源不存在统一按 400 语义提示（契约未单列 404）
func NotFound(msg string) *BizError { return BadRequest(msg) }

// From 从 error 提取业务码与提示；非业务错误默认 500。
func From(err error) (code int, msg string) {
	if err == nil {
		return response.CodeOK, ""
	}
	var be *BizError
	if errors.As(err, &be) {
		return be.Code, be.Msg
	}
	return response.CodeServer, err.Error()
}
