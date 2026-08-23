package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
	"gopomodoro/internal/validate"
)

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		email, err := validate.Email(req.Email)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		pass, err := validate.Password(req.Password)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		name, err := validate.RequiredString("display_name", req.DisplayName, 1, 40)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		hash, err := auth.HashPassword(pass)
		if err != nil {
			httpx.Fail(c, httpx.ErrInternal)
			return
		}
		u := &model.User{
			ID: uuid.New(), Email: email, PasswordHash: hash,
			DisplayName: name, Timezone: "Asia/Shanghai", CreatedAt: timeutil.Now(),
		}
		if err := d.DB.CreateUser(c.Request.Context(), u); err != nil {
			httpx.Fail(c, err)
			return
		}
		tok, err := auth.Issue(d.Cfg.JWTSecret, u, d.Cfg.JWTExpiry)
		if err != nil {
			httpx.Fail(c, httpx.ErrInternal)
			return
		}
		httpx.Created(c, gin.H{"user": u, "auth": tok})
	}
}

func Login(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		email, err := validate.Email(req.Email)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		u, err := d.DB.UserByEmail(c.Request.Context(), email)
		if err != nil {
			httpx.Fail(c, httpx.ErrUnauthorized)
			return
		}
		if !auth.CheckPassword(u.PasswordHash, req.Password) {
			httpx.Fail(c, httpx.ErrUnauthorized)
			return
		}
		tok, err := auth.Issue(d.Cfg.JWTSecret, u, d.Cfg.JWTExpiry)
		if err != nil {
			httpx.Fail(c, httpx.ErrInternal)
			return
		}
		httpx.OK(c, gin.H{"user": u, "auth": tok})
	}
}

func Me(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := d.DB.UserByID(c.Request.Context(), auth.UserID(c))
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, u)
	}
}
