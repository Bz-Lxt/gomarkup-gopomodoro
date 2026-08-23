package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopomodoro/internal/httpx"
)

const ctxUserID = "user_id"
const ctxEmail = "user_email"

func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			raw = c.Query("token")
		}
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
		if raw == "" {
			httpx.AbortJSON(c, httpx.ErrUnauthorized)
			return
		}
		claims, err := Parse(secret, raw)
		if err != nil {
			httpx.AbortJSON(c, httpx.ErrUnauthorized)
			return
		}
		id, err := uuid.Parse(claims.UserID)
		if err != nil {
			httpx.AbortJSON(c, httpx.ErrUnauthorized)
			return
		}
		c.Set(ctxUserID, id)
		c.Set(ctxEmail, claims.Email)
		c.Next()
	}
}

func UserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(ctxUserID)
	id, _ := v.(uuid.UUID)
	return id
}
