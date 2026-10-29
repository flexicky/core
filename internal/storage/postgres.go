package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     int
	SSLMode  string
}

// DSN renders the connection string. Credentials are escaped, so passwords
// containing '@', ':' or '/' are handled correctly.
func (c DBConfig) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   "/" + c.Name,
	}

	query := u.Query()
	query.Set("sslmode", c.SSLMode)
	u.RawQuery = query.Encode()

	return u.String()
}

type Storage struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func New(ctx context.Context, cfg DBConfig, log *slog.Logger) (*Storage, error) {
	const op = "storage.New"

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("%s: parse dsn: %w", op, err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: create pool: %w", op, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: ping database: %w", op, err)
	}

	log.Info("database connection pool initialized", slog.String("host", cfg.Host), slog.String("database", cfg.Name))

	return &Storage{pool: pool, log: log}, nil
}

func (s *Storage) Close() {
	if s.pool == nil {
		return
	}

	s.pool.Close()
	s.log.Info("database connection pool closed")
}

func (s *Storage) Pool() *pgxpool.Pool {
	return s.pool
}
