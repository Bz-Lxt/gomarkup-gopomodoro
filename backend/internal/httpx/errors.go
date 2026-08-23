package httpx

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Details    map[string]any
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

func (e *AppError) WithDetails(d map[string]any) *AppError {
	cp := *e
	cp.Details = d
	return &cp
}

var (
	ErrValidation         = New("E_VALIDATION", "请求参数校验失败", 400)
	ErrUnauthorized       = New("E_UNAUTHORIZED", "未登录或令牌无效", 401)
	ErrForbidden          = New("E_FORBIDDEN", "无权访问该资源", 403)
	ErrNotFound           = New("E_NOT_FOUND", "资源不存在", 404)
	ErrConflict           = New("E_CONFLICT", "资源冲突", 409)
	ErrInvalidTransition  = New("E_INVALID_TRANSITION", "非法的番茄钟状态迁移", 409)
	ErrSessionBusy        = New("E_SESSION_BUSY", "已有进行中的番茄钟会话", 409)
	ErrOptimisticLock     = New("E_OPTIMISTIC_LOCK", "会话已被其他请求更新，请重试", 409)
	ErrInternal           = New("E_INTERNAL", "服务器内部错误", 500)
)

func IsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
