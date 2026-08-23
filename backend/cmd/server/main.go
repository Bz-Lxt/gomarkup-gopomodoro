package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"gopomodoro/internal/burndown"
	"gopomodoro/internal/config"
	"gopomodoro/internal/eventbus"
	"gopomodoro/internal/handler"
	"gopomodoro/internal/logger"
	"gopomodoro/internal/model"
	"gopomodoro/internal/pomodoro"
	"gopomodoro/internal/seed"
	"gopomodoro/internal/store"
	"gopomodoro/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger.Init(cfg.LogLevel, cfg.Production())
	log := logger.L()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	db, err := store.Open(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	migDir := os.Getenv("MIGRATIONS_DIR")
	if migDir == "" {
		migDir = "migrations"
		if _, err := os.Stat(migDir); err != nil {
			migDir = filepath.Join("..", "migrations")
		}
	}
	if err := db.Migrate(context.Background(), migDir); err != nil {
		log.Error("migrate failed", "err", err)
		os.Exit(1)
	}
	if err := seed.Run(context.Background(), db); err != nil {
		log.Error("seed failed", "err", err)
		os.Exit(1)
	}

	hub := ws.NewHub()
	engine := burndown.NewEngine(db, hub)
	bus := eventbus.New(128, cfg.EventWorkers, func(ev model.DomainEvent) error {
		return engine.Handle(context.Background(), ev)
	})
	reg := pomodoro.NewRegistry(db, pomodoro.RealClock{}, cfg.FocusDuration, cfg.GracePeriod, bus)
	reg.Hub = hub
	if err := reg.Rebuild(context.Background()); err != nil {
		log.Error("registry rebuild failed", "err", err)
		os.Exit(1)
	}
	reg.StartSweep()
	snap := burndown.NewSnapshotter(db, engine)
	snap.Start()
	defer snap.Stop()

	deps := &handler.Deps{Cfg: cfg, DB: db, Registry: reg, Engine: engine, Bus: bus, Hub: hub}
	wsh := &ws.Handler{Hub: hub, Registry: reg, Ping: cfg.WSPingInterval, PongWait: cfg.WSPongTimeout}
	engine.Hub = hub

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.NewRouter(deps, wsh),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("http listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down")
	reg.StopSweep()
	shCtx, shCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shCancel()
	_ = srv.Shutdown(shCtx)
	bus.Close(shCtx)
}
