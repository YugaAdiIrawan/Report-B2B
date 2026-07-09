package healthCheck

import (
	"context"

	health "github.com/YugaAdiIrawan/model/healthcheck"
)

type HealthUseCase interface {
	CheckHealth(ctx context.Context) health.HealthCheckResponse
}
