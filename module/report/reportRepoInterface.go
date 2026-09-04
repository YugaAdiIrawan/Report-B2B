package report

import (
	"context"

	"github.com/YugaAdiIrawan/model/report"
)

type ReportRepository interface {
	FetchReportData(ctx context.Context, filter report.ReportFilter) ([]report.AutoUWReport, error)
	FetchMonthlyReportData(ctx context.Context, filter report.MonthlyReportFilter) ([]report.MonthlyReport, error)
}

type AutomailReportLogRepository interface {
	Insert(ctx context.Context, log report.AutomailReportLog) error
}
