package usecase

import (
	"context"
	"time"

	"github.com/YugaAdiIrawan/helpers"
	health "github.com/YugaAdiIrawan/model/healthcheck"
	"github.com/YugaAdiIrawan/module/healthCheck"
)

type healthUC struct {
	healthrepo healthCheck.HealthRepo
	logHook    *helpers.LastLog
	startedAt  time.Time
}

func NewHealthUsecase(repo healthCheck.HealthRepo, logHook *helpers.LastLog, startedAt time.Time) healthCheck.HealthUseCase {
	return &healthUC{
		healthrepo: repo,
		logHook:    logHook,
		startedAt:  startedAt,
	}
}

func (uc *healthUC) CheckHealth(ctx context.Context) health.HealthCheckResponse {
	ctx, cencel := context.WithTimeout(ctx, 5*time.Second)
	defer cencel()

	overallStatus := health.StatusHealthy

	latency, errLatency := uc.healthrepo.PingDB(ctx)
	dbCheck := health.ComponentCheck{
		Name:    "ciputra_phase_2",
		Status:  health.StatusHealthy,
		Latency: latency.String(),
	}
	if errLatency != nil {
		dbCheck.Status = health.StatusUnhealthy
		dbCheck.Error = errLatency.Error()
		overallStatus = health.StatusUnhealthy
	}

	response := health.HealthCheckResponse{
		Status:        overallStatus,
		Timrstamp:     time.Now(),
		UptimeSeconds: int64(time.Since(uc.startedAt).Seconds()),
		Components: []health.ComponentCheck{
			dbCheck,
		},
	}
	if entry, ok := uc.logHook.Last(); ok {
		response.LastLog = health.LastLogEntry{
			Timestamp: entry.Timestamp,
			Level:     entry.Level.String(),
			Message:   entry.Message,
		}
	}
	return response
}
