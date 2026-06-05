package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/YugaAdiIrawan/model/report"
	"github.com/rs/zerolog/log"
)

type reportRepo struct {
	db *sql.DB
}

func NewReportRepo(db *sql.DB) *reportRepo {
	return &reportRepo{db: db}
}

func (repo *reportRepo) FetchReportData(ctx context.Context, filter report.ReportFilter) ([]report.AutoUWReport, error) {
	query, args, err := repo.buildQueryData(filter)
	if err != nil {
		return nil, fmt.Errorf("repoReport: build query: %w", err)
	}
	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repoReport: select: %w", err)
	}
	defer rows.Close()

	var result []report.AutoUWReport
	for rows.Next() {
		var item report.AutoUWReport
		var dateOfBirth sql.NullTime
		var statusSentCN sql.NullString

		err := rows.Scan(
			&item.LeadID,
			&item.NoRef,
			&item.Partner,
			&item.NameOfInsurance,
			&dateOfBirth,
			&item.Gender,
			&item.MCUPackage,
			&item.Tenor,
			&item.UP,
			&item.TanggalOrder,
			&item.Source,
			&statusSentCN,
		)
		if err != nil {
			return nil, fmt.Errorf("repoReport: scan: %w", err)
		}

		if dateOfBirth.Valid {
			item.DateOfBirth = &dateOfBirth.Time
		}
		if statusSentCN.Valid {
			item.StatusSentCN = &statusSentCN.String
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repoReport: rows.Err: %w", err)
	}
	log.Debug().Msgf("result : %v", result)
	return result, nil
}

func (repo *reportRepo) buildQueryData(filter report.ReportFilter) (string, []interface{}, error) {
	var (
		condition []string
		args      []interface{}
	)

	condition = append(condition, fmt.Sprintf("tanggal_order >= ?"))
	args = append(args, filter.StartDate)

	endOfDay := time.Date(filter.EndDate.Year(), filter.EndDate.Month(), filter.EndDate.Day(), 23, 59, 59, 999999999, filter.EndDate.Location())
	condition = append(condition, fmt.Sprintf("tanggal_order <= ?"))
	args = append(args, endOfDay)

	if filter.CNReleaseStatus != nil {
		if *filter.CNReleaseStatus == report.CNStatusNull {
			condition = append(condition, "status_sent_cn IS NULL")
		} else {
			condition = append(condition, fmt.Sprintf("status_sent_cn = ?"))
			args = append(args, string(*filter.CNReleaseStatus))
		}
	}

	if filter.SubmissionSource != nil {
		sourceVal := 0
		if *filter.SubmissionSource {
			sourceVal = 1
		}
		condition = append(condition, fmt.Sprintf("source = ?"))
		args = append(args, sourceVal)
	}
	query := `
		SELECT
			lead_id,
			no_ref,
			partner,
			name_of_insurance,
			date_of_birth,
			gender,
			mcu_package,
			tenor,
			up,
			tanggal_order,
			source,
			status_sent_cn
		FROM vw_report
		WHERE ` + strings.Join(condition, " AND ") + ` 
		ORDER BY tanggal_order ASC`

	// TEMPORARY DEBUG
	fmt.Printf("[DEBUG QUERY]\n%s\n", query)
	fmt.Printf("[DEBUG ARGS] %+v\n", args)
	return query, args, nil
}

func (repo *reportRepo) Insert(ctx context.Context, entry report.AutomailReportLog) error {
	query := `
        INSERT INTO automail_report_logs (
            request_id,
            from_email,
            to_emails,
            cc_emails,
            subject,
            has_attachment,
            attachment_name,
            attachment_size,
            http_status,
            response_body,
            error_message,
            status,
            sent_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := repo.db.ExecContext(ctx, query,
		entry.RequestID,
		entry.FromEmail,
		entry.ToEmails,
		entry.CcEmails,
		entry.Subject,
		entry.HasAttachment,
		entry.AttachmentName,
		entry.AttachmentSize,
		entry.HTTPStatus,
		entry.ResponseBody,
		entry.ErrorMessage,
		string(entry.Status),
		entry.SentAt,
	)
	if err != nil {
		return fmt.Errorf("automail_log_repo: insert: %w", err)
	}

	return nil
}
