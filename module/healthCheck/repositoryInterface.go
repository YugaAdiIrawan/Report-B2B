package healthCheck

import (
	"context"
	"time"
)

type HealthRepo interface {
	PingDB(ctx context.Context) (time.Duration, error)
}
