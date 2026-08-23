package handler

import (
	"gopomodoro/internal/burndown"
	"gopomodoro/internal/config"
	"gopomodoro/internal/eventbus"
	"gopomodoro/internal/pomodoro"
	"gopomodoro/internal/store"
	"gopomodoro/internal/ws"
)

type Deps struct {
	Cfg      *config.Config
	DB       *store.DB
	Registry *pomodoro.Registry
	Engine   *burndown.Engine
	Bus      *eventbus.Bus
	Hub      *ws.Hub
}
