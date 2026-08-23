package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopomodoro/internal/logger"
)

type Envelope struct {
	OK      bool           `json:"ok"`
	Data    any            `json:"data,omitempty"`
	Error   *ErrorPayload  `json:"error,omitempty"`
}

type ErrorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{OK: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{OK: true, Data: data})
}

func Fail(c *gin.Context, err error) {
	if ae, ok := IsAppError(err); ok {
		if ae.HTTPStatus >= 500 {
			logger.L().Error("request failed", "code", ae.Code, "err", err)
		}
		c.JSON(ae.HTTPStatus, Envelope{OK: false, Error: &ErrorPayload{
			Code: ae.Code, Message: ae.Message, Details: ae.Details,
		}})
		return
	}
	logger.L().Error("unhandled error", "err", err)
	c.JSON(http.StatusInternalServerError, Envelope{OK: false, Error: &ErrorPayload{
		Code: ErrInternal.Code, Message: ErrInternal.Message,
	}})
}

func AbortJSON(c *gin.Context, err *AppError) {
	c.AbortWithStatusJSON(err.HTTPStatus, Envelope{OK: false, Error: &ErrorPayload{
		Code: err.Code, Message: err.Message, Details: err.Details,
	}})
}
