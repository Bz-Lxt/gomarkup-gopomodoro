package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/model"
)

func TestIssueAndParse(t *testing.T) {
	u := &model.User{ID: uuid.New(), Email: "geek@gopomodoro.dev"}
	tok, err := Issue("secret", u, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Parse("secret", tok.Token)
	if err != nil {
		t.Fatal(err)
	}
	if c.Email != u.Email {
		t.Fatal(c.Email)
	}
	if _, err := Parse("other", tok.Token); err == nil {
		t.Fatal("must reject wrong secret")
	}
}

func TestPasswordRoundtrip(t *testing.T) {
	h, err := HashPassword("pomodoro123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(h, "pomodoro123") {
		t.Fatal("match")
	}
	if CheckPassword(h, "wrong") {
		t.Fatal("mismatch")
	}
}
