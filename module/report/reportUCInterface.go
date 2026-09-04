package report

import (
	"context"
	"time"

	"github.com/YugaAdiIrawan/model/report"
)

type ReportUsecase interface {
	GenerateReport(ctx context.Context, filter report.ReportFilter) ([]byte, string, error)
	GenerateMonthlyReport(ctx context.Context, filter report.MonthlyReportFilter) ([]byte, string, error)
	SendDailyReport(ctx context.Context, targetDate time.Time, recipients []string) error
}
