package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/YugaAdiIrawan/config"
	"github.com/YugaAdiIrawan/middleware"
	report2 "github.com/YugaAdiIrawan/model/report"
	"github.com/YugaAdiIrawan/module/report"
	role "github.com/YugaAdiIrawan/module/role/middlewareUsecase"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ReportHandler struct {
	reportUsecase report.ReportUsecase
	cfg           *config.ServiceAuthConfig
	role          role.Middleware
}

func NewReportHandler(r *gin.Engine, reportUsecase report.ReportUsecase, cfg *config.ServiceAuthConfig, roleMiddleware role.Middleware) {
	handler := ReportHandler{
		reportUsecase: reportUsecase,
		cfg:           cfg,
		role:          roleMiddleware,
	}
	authMiddleware := middleware.BasicAuth(middleware.BasicAuthConfig{
		Username: cfg.ReportServiceUsername,
		Password: cfg.ReportServicePassword,
	})
	reportUW := r.Group("/auto-uw")
	reportUW.Use(authMiddleware)
	reportUW.Use(handler.role.MiddlewareRouteRoles)
	{
		reportUW.GET("/export", handler.ExportReportAutoUW)
	}
}

func (h *ReportHandler) ExportReportAutoUW(c *gin.Context) {
	var req report2.ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"message": err.Error(),
		})
		return
	}

	jakartaLoc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "timezone_error",
			"message": "gagal load timezone",
		})
		return
	}
	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, jakartaLoc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_start_date",
			"message": "format harus: YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.ParseInLocation("2006-01-02", req.EndDate, jakartaLoc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_end_date",
			"message": "format harus: YYYY-MM-DD",
		})
		return
	}
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_date_range",
			"message": "end_date tidak boleh sebelum start_date",
		})
		return
	}
	if endDate.Sub(startDate).Hours() > 31*24 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "date_range_too_large",
			"message": "rentang tanggal maksimal 31 hari",
		})
		return
	}

	filter := report2.ReportFilter{
		StartDate:      startDate,
		EndDate:        endDate,
		IsAutoAccepted: req.IsAutoAccepted,
	}
	if req.CNReleaseStatus != nil {
		status := report2.CNReleaseStatus(*req.CNReleaseStatus)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_cn_release_status",
				"message": "nilai cn_release_status tidak valid",
			})
			return
		}
		filter.CNReleaseStatus = &status
	}

	if req.SubmissionSource != nil {
		sourceType := report2.SubmissionSourceType(*req.SubmissionSource)
		if !sourceType.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_submission_source",
				"message": "nilai submission_source harus: UPLOAD, ESUBMISSION, atau EXTERNAL",
			})
			return
		}
		filter.SubmissionSource = &sourceType
	}
	excelBytes, filename, err := h.reportUsecase.GenerateReport(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "generate_report_failed",
			"message": err.Error(), //"Gagal generate report, silakan coba kembali",
		})
		return
	}
	logEvent := log.Debug().Interface("req", req)
	if req.IsAutoAccepted != nil {
		logEvent = logEvent.Bool("is_auto_accepted", *req.IsAutoAccepted)
	}
	logEvent.Msg("request")

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/vnd.ms-excel; charset=UTF-8")
	c.Header("Content-Length", fmt.Sprintf("%d", len(excelBytes)))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, "application/vnd.ms-excel; charset=UTF-8", excelBytes)
}
