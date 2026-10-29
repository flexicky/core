package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type GRPC struct {
	Port    int           `yaml:"port" env:"GRPC_PORT" env-default:"9090"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"10s"`
}

type DB struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Name     string `yaml:"name" env:"DB_NAME" env-required:"true"`
	User     string `yaml:"user" env:"DB_USER" env-required:"true"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
}

// Redis holds the connection parameters. Prefix optionally namespaces the
// cache keys; leave it empty to keep the historical "session-<id>" layout.
type Redis struct {
	Host     string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	Prefix   string `yaml:"prefix" env:"REDIS_PREFIX"`
}

type Config struct {
	Env string `yaml:"env" env:"ENV" env-default:"dev"`

	// AccessTokenTTL is the lifetime of issued access tokens.
	AccessTokenTTL time.Duration `yaml:"access_token_ttl" env:"ACCESS_TOKEN_TTL" env-default:"1h"`
	// SessionTTL is the lifetime of an authentication session (refresh token).
	SessionTTL time.Duration `yaml:"session_ttl" env:"SESSION_TTL" env-default:"720h"`
	// SessionCacheTTL is the lifetime of a session entry in the Redis cache.
	SessionCacheTTL time.Duration `yaml:"session_cache_ttl" env:"SESSION_CACHE_TTL" env-default:"1h"`
	// ShutdownTimeout bounds the graceful shutdown of the application.
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"10s"`

	// JWTSeed is an optional base64 encoded 32 byte ed25519 seed used to sign
	// access tokens. When empty, a random key pair is generated on start-up,
	// which invalidates every token issued before the restart.
	JWTSeed string `yaml:"jwt_seed" env:"JWT_SEED"`

	GRPC  GRPC  `yaml:"grpc"`
	DB    DB    `yaml:"db"`
	Redis Redis `yaml:"redis"`
}

// Load reads the configuration from path and validates it.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is empty: pass -config or set CONFIG_PATH")
	}

	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("config file %q: %w", path, err)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// MustLoad is Load that panics on failure. It is meant for main packages.
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}

	return cfg
}

func (c *Config) validate() error {
	if c.GRPC.Port <= 0 || c.GRPC.Port > 65535 {
		return fmt.Errorf("grpc.port must be in 1..65535, got %d", c.GRPC.Port)
	}

	if c.AccessTokenTTL <= 0 {
		return errors.New("access_token_ttl must be positive")
	}

	if c.SessionTTL <= 0 {
		return errors.New("session_ttl must be positive")
	}

	if c.SessionCacheTTL <= 0 {
		return errors.New("session_cache_ttl must be positive")
	}

	if c.DB.Name == "" || c.DB.User == "" {
		return errors.New("db.name and db.user are required")
	}

	return nil
}
