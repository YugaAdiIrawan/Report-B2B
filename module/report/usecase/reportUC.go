package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/YugaAdiIrawan/client"
	"github.com/YugaAdiIrawan/helpers"
	report2 "github.com/YugaAdiIrawan/model/report"
	"github.com/YugaAdiIrawan/module/report"
	"github.com/rs/zerolog/log"
)

type reportUsecase struct {
	reportRepo   report.ReportRepository
	emailService client.EmailService
}

func NewReportUsecase(repo report.ReportRepository, email client.EmailService) *reportUsecase {
	return &reportUsecase{
		reportRepo:   repo,
		emailService: email,
	}
}

func (uc *reportUsecase) GenerateReport(ctx context.Context, filter report2.ReportFilter) ([]byte, string, error) {
	data, err := uc.reportRepo.FetchReportData(ctx, filter)
	if err != nil {
		return nil, "", fmt.Errorf("report_usecase: fetch report data: %w", err)
	}
	log.Debug().Msgf("report data: %v", data)

	//excelBytes, err := helpers.GenerateAutoUWExel(data, filter.StartDate)
	excelBytes, err := helpers.GenerateAutoUWExcelHTML(data, filter.StartDate)
	if err != nil {
		return nil, "", fmt.Errorf("report_usecase: generate excel: %w", err)
	}
	fileName := fmt.Sprintf("AutoUW_Report_%s_%s.xls", filter.StartDate.Format("2006-01-02"), filter.EndDate.Format("2006-01-02"))
	return excelBytes, fileName, nil
}

func (uc *reportUsecase) GenerateMonthlyReport(ctx context.Context, filter report2.MonthlyReportFilter) ([]byte, string, error) {
	data, err := uc.reportRepo.FetchMonthlyReportData(ctx, filter)
	if err != nil {
		return nil, "", fmt.Errorf("report_usecase: fetch monthly data: %w", err)
	}
	log.Debug().Msgf("monthly report data: %v", data)

	excelBytes, err := helpers.GenerateMonthlyReportExcelHTML(data, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, "", fmt.Errorf("report_usecase: generate monthly excel: %w", err)
	}

	fileName := fmt.Sprintf("Monthly_Report_%s_to_%s.xls",
		filter.StartDate.Format("2006-01-02"),
		filter.EndDate.Format("2006-01-02"),
	)
	return excelBytes, fileName, nil
}

func (uc *reportUsecase) SendDailyReport(ctx context.Context, targetDate time.Time, recipients []string) error {
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, targetDate.Location())

	filter := report2.ReportFilter{
		StartDate: startOfDay,
		EndDate:   endOfDay,
	}

	excelBytes, filename, err := uc.GenerateReport(ctx, filter)
	if err != nil {
		return fmt.Errorf("report_usecase: send daily: generate report: %w", err)
	}

	subject := fmt.Sprintf(
		"[Auto UW] Daily Report - %s",
		targetDate.Format("02 January 2006"),
	)

	htmlBody := buildEmailHTML(targetDate, len(excelBytes))

	emailReq := report2.SendReportEmailRequest{
		To:       recipients,
		Subject:  subject,
		HTMLBody: htmlBody,
		Attachment: report2.EmailAttachment{
			Filename:    filename,
			ContentType: "application/vnd.ms-excel",
			Data:        excelBytes,
		},
	}

	if err := uc.emailService.SendReportEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("report_usecase: send daily: send email: %w", err)
	}

	return nil
}

func buildEmailHTML(reportDate time.Time, fileSizeBytes int) string {
	return fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; color: #333;">
  <h3 style="color: #1F4E79;">Auto UW Daily Report</h3>
  <p>Terlampir adalah <strong>Report Auto UW</strong> untuk tanggal <strong>%s</strong>.</p>
  <p>File Excel terlampir berisi data seluruh pengajuan Auto UW pada hari tersebut.</p>
  <br/>
  <p style="color: #888; font-size: 10px;">Email ini dikirim otomatis oleh sistem. Mohon tidak membalas email ini.</p>
</body>
</html>`,
		reportDate.Format("02 January 2006"),
	)
}
