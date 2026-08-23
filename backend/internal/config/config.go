package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env                   string
	HTTPAddr              string
	DatabaseURL           string
	JWTSecret             string
	JWTExpiry             time.Duration
	FocusDuration         time.Duration
	GracePeriod           time.Duration
	WSPingInterval        time.Duration
	WSPongTimeout         time.Duration
	LogLevel              string
	AllowTestComplete     bool
	EventWorkers          int
}

func Load() (*Config, error) {
	c := &Config{
		Env:               getenv("APP_ENV", "development"),
		HTTPAddr:          getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:       getenv("DATABASE_URL", "postgres://gopomo:gopomo_dev@127.0.0.1:35174/gopomodoro?sslmode=disable"),
		JWTSecret:         getenv("JWT_SECRET", "gopomo-dev-jwt-secret-change-me"),
		JWTExpiry:         time.Duration(getenvInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
		FocusDuration:     time.Duration(getenvInt("FOCUS_DURATION_SECONDS", 1500)) * time.Second,
		GracePeriod:       time.Duration(getenvInt("GRACE_PERIOD_SECONDS", 120)) * time.Second,
		WSPingInterval:    15 * time.Second,
		WSPongTimeout:     45 * time.Second,
		LogLevel:          getenv("LOG_LEVEL", "info"),
		AllowTestComplete: getenvBool("ALLOW_TEST_COMPLETE", false),
		EventWorkers:      getenvInt("EVENT_WORKERS", 4),
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if strings.TrimSpace(c.JWTSecret) == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if c.FocusDuration < time.Second {
		return nil, fmt.Errorf("FOCUS_DURATION_SECONDS must be >= 1")
	}
	if c.GracePeriod < time.Second {
		return nil, fmt.Errorf("GRACE_PERIOD_SECONDS must be >= 1")
	}
	return c, nil
}

func (c *Config) Production() bool {
	return strings.EqualFold(c.Env, "production")
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	return raw == "1" || raw == "true" || raw == "yes"
}
