package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}

func TestLoadAppliesDefaults(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
env: local
db:
  name: core
  user: postgres
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "local" {
		t.Fatalf("Env = %q, want local", cfg.Env)
	}

	if cfg.GRPC.Port != 9090 {
		t.Fatalf("GRPC.Port = %d, want 9090", cfg.GRPC.Port)
	}

	if cfg.Redis.Port != 6379 {
		t.Fatalf("Redis.Port = %d, want 6379", cfg.Redis.Port)
	}

	if cfg.AccessTokenTTL != time.Hour {
		t.Fatalf("AccessTokenTTL = %v, want 1h", cfg.AccessTokenTTL)
	}

	if cfg.DB.SSLMode != "disable" {
		t.Fatalf("DB.SSLMode = %q, want disable", cfg.DB.SSLMode)
	}
}

func TestLoadExampleConfig(t *testing.T) {
	t.Parallel()

	cfg, err := Load(filepath.Join("..", "..", "config", "config.example.yaml"))
	if err != nil {
		t.Fatalf("Load(example) error = %v", err)
	}

	if cfg.DB.Name != "core" || cfg.GRPC.Port != 9090 {
		t.Fatalf("unexpected example config: %+v", cfg)
	}
}

func TestLoadRejectsBrokenConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"missing db name":   "db:\n  user: postgres\n",
		"port out of range": "db:\n  name: core\n  user: postgres\ngrpc:\n  port: 70000\n",
		"negative ttl":      "db:\n  name: core\n  user: postgres\naccess_token_ttl: -1h\n",
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := Load(writeConfig(t, body)); err == nil {
				t.Fatal("Load() error = nil, want error")
			}
		})
	}
}

func TestLoadRejectsMissingPath(t *testing.T) {
	t.Parallel()

	if _, err := Load(""); err == nil {
		t.Fatal("Load(\"\") error = nil, want error")
	}

	if _, err := Load(filepath.Join(t.TempDir(), "absent.yaml")); err == nil {
		t.Fatal("Load(absent) error = nil, want error")
	}
}
