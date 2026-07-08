package healthcheck

import "time"

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type ComponentCheck struct {
	Name    string       `json:"name"`
	Status  HealthStatus `json:"status"`
	Latency string       `json:"latency"`
	Error   string       `json:"error,omitempty"`
}

type LastLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type HealthCheckResponse struct {
	Status        HealthStatus     `json:"status"`
	Timrstamp     time.Time        `json:"timestamp"`
	UptimeSeconds int64            `json:"uptime_seconds"`
	Components    []ComponentCheck `json:"components"`
	LastLog       LastLogEntry     `json:"last_log"`
}
