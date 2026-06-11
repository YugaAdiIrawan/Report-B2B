package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/YugaAdiIrawan/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func LoadServiceAuthConfig() (*config.ServiceAuthConfig, error) {
	username := os.Getenv("REPORT_SERVICE_USERNAME")
	password := os.Getenv("REPORT_SERVICE_PASSWORD")
	// temporary debug
	log.Printf("[ServiceAuthConfig] username: %q", username)
	log.Printf("[ServiceAuthConfig] password: %q", password)
	if username == "" || password == "" {
		return nil, fmt.Errorf("LoadServiceAuthConfig: REPORT_SERVICE_USERNAME and REPORT_SERVICE_PASSWORD are required")
	}

	return &config.ServiceAuthConfig{
		ReportServiceUsername: username,
		ReportServicePassword: password,
	}, nil
}

type BasicAuthConfig struct {
	Username string
	Password string
}

func BasicAuth(cfg BasicAuthConfig) gin.HandlerFunc {
	if cfg.Username == "" || cfg.Password == "" {
		panic("BasicAuth middleware: username or password is empty")
	}

	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.Printf("[BasicAuth] Authorization header: %q", authHeader)
		log.Printf("[BasicAuth] Expected username: %q", cfg.Username)

		if !strings.HasPrefix(authHeader, "Basic ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		incomingUser := parts[0]
		incomingPassword := parts[1]

		log.Printf("[BasicAuth] incoming user: %q", incomingUser)
		log.Printf("[BasicAuth] incoming password: %q", incomingPassword)
		log.Printf("[BasicAuth] cfg username: %q", cfg.Username)
		log.Printf("[BasicAuth] cfg password: %q", cfg.Password)

		userMatch := subtle.ConstantTimeCompare([]byte(incomingUser), []byte(cfg.Username))
		passMatch := subtle.ConstantTimeCompare([]byte(incomingPassword), []byte(cfg.Password))
		log.Printf("[BasicAuth] userMatch: %d, passMatch: %d", userMatch, passMatch)

		if userMatch != 1 || passMatch != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("caller", incomingUser)
		c.Next()
	}
}
