package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "s")
	os.Unsetenv("FOCUS_DURATION_SECONDS")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.FocusDuration != 1500*time.Second {
		t.Fatal(c.FocusDuration)
	}
	if c.GracePeriod != 120*time.Second {
		t.Fatal(c.GracePeriod)
	}
}

func TestLoadRejectsTinyFocus(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("FOCUS_DURATION_SECONDS", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}
