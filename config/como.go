package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YugaAdiIrawan/model/report"
)

type Config struct {
	ComoEmail    report.ComoConfig
	AutoUWReport report.AutoUWSchedulerConfig
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ComoEmail: report.ComoConfig{
			BaseURL:   mustGetEnv("COMO_EMAIL_BASE_URL"),
			ApiKey:    mustGetEnv("COMO_EMAIL_API_KEY"),
			FromEmail: mustGetEnv("COMO_EMAIL_FROM"),
			Timeout:   getDurationEnv("COMO_EMAIL_TIMEOUT_SECONDS", 30*time.Second),
		},
		AutoUWReport: report.AutoUWSchedulerConfig{
			Recipients: getSliceEnv("AUTO_UW_REPORT_RECIPIENTS"),
			CCList:     getSliceEnv("AUTO_UW_REPORT_CC"),
			Timezone:   getEnvWithDefault("AUTO_UW_REPORT_TIMEZONE", "Asia/Jakarta"),
		},
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if len(c.AutoUWReport.Recipients) == 0 {
		return fmt.Errorf("AUTO_UW_REPORT_RECIPIENTS tidak boleh kosong")
	}
	return nil
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Errorf("environment variable %s not set", key)
	}
	return value
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getSliceEnv(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	seconds, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return time.Duration(seconds) * time.Second
}
