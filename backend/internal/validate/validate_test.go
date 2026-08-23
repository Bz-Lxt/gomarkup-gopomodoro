package validate

import "testing"

func TestEmailAndColumn(t *testing.T) {
	if _, err := Email("not-an-email"); err == nil {
		t.Fatal("email")
	}
	if _, err := Email("geek@gopomodoro.dev"); err != nil {
		t.Fatal(err)
	}
	if _, err := Column("doing"); err == nil {
		t.Fatal("bad column")
	}
	if _, err := Column("in_progress"); err != nil {
		t.Fatal(err)
	}
	if _, err := Password("short"); err == nil {
		t.Fatal("short password")
	}
}
