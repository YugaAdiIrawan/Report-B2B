package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/YugaAdiIrawan/module/healthCheck"
)

type healthRepository struct {
	db *sql.DB
}

func NewHealthCHeckRepository(db *sql.DB) healthCheck.HealthRepo {
	return &healthRepository{db: db}
}

func (r *healthRepository) PingDB(ctx context.Context) (time.Duration, error) {
	if r.db == nil {
		return 0, fmt.Errorf("db connection is nil")
	}

	start := time.Now()
	if err := r.db.PingContext(ctx); err != nil {
		return time.Since(start), fmt.Errorf("db ping failed: %w", err)
	}
	return time.Since(start), nil
}
