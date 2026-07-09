package schaduler

import (
	"net/http"

	"github.com/YugaAdiIrawan/config"
	"github.com/YugaAdiIrawan/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AdminSchedulerHandler struct {
	autoUWScheduler *AutoUWScheduler
	cfg             *config.ServiceAuthConfig
}

func NewAdminSchedulerHandler(r *gin.Engine, s *AutoUWScheduler, config *config.ServiceAuthConfig) {
	handlerSchaduler := &AdminSchedulerHandler{
		autoUWScheduler: s,
		cfg:             config,
	}
	authMiddleware := middleware.BasicAuth(middleware.BasicAuthConfig{
		Username: config.ReportServiceUsername,
		Password: config.ReportServicePassword,
	})
	schaduler := r.Group("/scheduler")
	schaduler.Use(authMiddleware)
	{
		schaduler.POST("/stop", handlerSchaduler.PauseAutoUWScheduler)
		schaduler.POST("/start", handlerSchaduler.ResumeAutoUWScheduler)
		schaduler.GET("/status", handlerSchaduler.GetAutoUWSchedulerStatus)
	}
}

func (h *AdminSchedulerHandler) PauseAutoUWScheduler(c *gin.Context) {
	if h.autoUWScheduler.IsPaused() {
		c.JSON(http.StatusOK, gin.H{
			"status":  "already_paused",
			"message": "auto_uw_scheduler is already paused",
		})
		return
	}

	h.autoUWScheduler.Pause()
	log.Warn().
		Str("actor", c.GetString("caller")).
		Msg("auto_uw_scheduler: paused via admin API")

	c.JSON(http.StatusOK, gin.H{
		"status":  "paused",
		"message": "auto_uw_scheduler paused successfully",
	})
}

func (h *AdminSchedulerHandler) ResumeAutoUWScheduler(c *gin.Context) {
	if !h.autoUWScheduler.IsPaused() {
		c.JSON(http.StatusOK, gin.H{
			"status":  "already_running",
			"message": "auto_uw_scheduler is already running",
		})
		return
	}

	h.autoUWScheduler.Resume()
	log.Warn().
		Str("actor", c.GetString("admin_username")).
		Msg("auto_uw_scheduler: resumed via admin API")

	c.JSON(http.StatusOK, gin.H{
		"status":  "resumed",
		"message": "auto_uw_scheduler resumed successfully",
	})
}

func (h *AdminSchedulerHandler) GetAutoUWSchedulerStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"paused": h.autoUWScheduler.IsPaused(),
	})
}
