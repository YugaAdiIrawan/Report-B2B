package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YugaAdiIrawan/globals"
	"github.com/YugaAdiIrawan/model/report"
	"github.com/rs/zerolog/log"
)

type Config struct {
	ComoEmail    report.ComoConfig
	AutoUWReport report.AutoUWSchedulerConfig
}

const (
	H2HConfigIDComoSendEmail = 5
)

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ComoEmail: report.ComoConfig{
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

func (c *Config) validateComoEmailConfig() error {
	var missing []string

	if c.ComoEmail.SendBaseURL == "" {
		missing = append(missing, "SendBaseURL (h2h_configs id=5, base_url+path_url)")
	}
	if c.ComoEmail.SendApiKey == "" {
		missing = append(missing, "SendApiKey (h2h_configs id=5, notes)")
	}
	if c.ComoEmail.FromEmail == "" {
		missing = append(missing, "FromEmail (env COMO_EMAIL_FROM)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("como email config tidak lengkap, field kosong: %s", strings.Join(missing, "; "))
	}
	return nil
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal().Msgf("config: environment variable %s wajib di-set tapi kosong", key)
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

func LoadComoSendEmailFromDB(cfg *Config, h2hRepo globals.GlobalRepository) error {
	data, err := h2hRepo.RetrieveH2hConfigAPIEmail(H2HConfigIDComoSendEmail)
	if err != nil {
		return fmt.Errorf("LoadComoSendEmailFromDB: query h2h_configs id=%d gagal: %w", H2HConfigIDComoSendEmail, err)
	}
	if data == nil {
		return fmt.Errorf("LoadComoSendEmailFromDB: h2h_configs id=%d tidak ditemukan (deleted/missing)", H2HConfigIDComoSendEmail)
	}

	baseURL := strings.TrimSpace(data.BaseURL)
	pathURL := strings.TrimSpace(data.PathURL)
	apiKey := strings.TrimSpace(data.Notes)

	if baseURL == "" {
		return fmt.Errorf("LoadComoSendEmailFromDB: base_url kosong pada h2h_configs id=%d", H2HConfigIDComoSendEmail)
	}
	if apiKey == "" {
		return fmt.Errorf("LoadComoSendEmailFromDB: notes (ApiKey) kosong pada h2h_configs id=%d", H2HConfigIDComoSendEmail)
	}

	cfg.ComoEmail.SendBaseURL = baseURL + pathURL
	cfg.ComoEmail.SendApiKey = apiKey
	
	return nil
}
