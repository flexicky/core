package storage

import (
	"net/url"
	"testing"
)

func TestDSNEscapesCredentials(t *testing.T) {
	t.Parallel()

	cfg := DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "p@ss:w/rd",
		Name:     "core",
		SSLMode:  "disable",
	}

	parsed, err := url.Parse(cfg.DSN())
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}

	password, _ := parsed.User.Password()
	if password != cfg.Password {
		t.Fatalf("password = %q, want %q", password, cfg.Password)
	}

	if parsed.User.Username() != cfg.User {
		t.Fatalf("user = %q, want %q", parsed.User.Username(), cfg.User)
	}

	if parsed.Host != "localhost:5432" {
		t.Fatalf("host = %q, want %q", parsed.Host, "localhost:5432")
	}

	if parsed.Path != "/core" {
		t.Fatalf("path = %q, want %q", parsed.Path, "/core")
	}

	if got := parsed.Query().Get("sslmode"); got != "disable" {
		t.Fatalf("sslmode = %q, want %q", got, "disable")
	}
}
