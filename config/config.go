package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	MaxIdle  int
	MaxOpen  int

	// Optional — gunakan default jika 0
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (c *MySqlConfig) applyDefaults() {
	if c.MaxIdle == 0 {
		c.MaxIdle = 5
	}
	if c.MaxOpen == 0 {
		c.MaxOpen = 25
	}
	// Harus lebih kecil dari MySQL wait_timeout (default 8 jam)
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 5 * time.Minute
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 2 * time.Minute
	}
}

func NewMySQLDB(cfg MySqlConfig) (*sql.DB, error) {
	cfg.applyDefaults()

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FJakarta",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("NewMySQLDB: open connection: %w", err)
	}

	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("NewMySQLDB: ping failed — check host/port/credentials: %w", err)
	}
	return db, nil
}
