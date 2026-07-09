package handler

import (
	"net/http"

	health "github.com/YugaAdiIrawan/model/healthcheck"
	"github.com/YugaAdiIrawan/module/healthCheck"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	uc healthCheck.HealthUseCase
}

func NewHealthHandler(r *gin.Engine, uc healthCheck.HealthUseCase) {
	handler := HealthHandler{
		uc: uc,
	}

	r.GET("/healthCheck", handler.Check)
}

func (h *HealthHandler) Check(c *gin.Context) {
	response := h.uc.CheckHealth(c.Request.Context())

	httpStatus := http.StatusOK
	if response.Status == health.StatusUnhealthy {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response)
}
