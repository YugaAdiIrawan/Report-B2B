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
		if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
			unauthorized(c, "missing or malformed Authorization header")
			return
		}

		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			unauthorized(c, "invalid base64 encoding")
			return
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			unauthorized(c, "malformed credentials format")
			return
		}

		incomingUser := parts[0]
		incomingPassword := parts[1]

		userMatch := subtle.ConstantTimeCompare([]byte(incomingUser), []byte(cfg.Username))
		passMatch := subtle.ConstantTimeCompare([]byte(incomingPassword), []byte(cfg.Password))

		if userMatch != 1 || passMatch != 1 {
			log.Warn().
				Str("attempted_user", incomingUser).
				Str("remote_addr", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Msg("basic_auth: unauthorized access attempt")

			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("caller", incomingUser)
		c.Next()
	}
}

func unauthorized(c *gin.Context, reason string) {
	log.Debug().
		Str("reason", reason).
		Str("remote_addr", c.ClientIP()).
		Str("path", c.Request.URL.Path).
		Msg("basic_auth: request rejected")

	c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
	c.AbortWithStatus(http.StatusUnauthorized)
}
