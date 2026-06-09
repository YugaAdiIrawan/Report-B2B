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
			&item.OrderID,
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

	condition = append(condition, fmt.Sprintf("created_at >= ?"))
	args = append(args, filter.StartDate)

	endOfDay := time.Date(filter.EndDate.Year(), filter.EndDate.Month(), filter.EndDate.Day(), 23, 59, 59, 999999999, filter.EndDate.Location())
	condition = append(condition, fmt.Sprintf("created_at <= ?"))
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
		switch *filter.SubmissionSource {
		case report.SubmissionSourceUpload:
			condition = append(condition, fmt.Sprintf("source = ?"))
			args = append(args, 0)

		case report.SubmissionSourceESubmission:
			condition = append(condition, fmt.Sprintf("source = ?"))
			args = append(args, 1)
			condition = append(condition, fmt.Sprintf("order_id LIKE = ?"))
			args = append(args, "018-%")

		case report.SubmissionSourceExternal:
			condition = append(condition, fmt.Sprintf("source = ?"))
			args = append(args, 1)
			condition = append(condition, fmt.Sprintf("order_id NOT LIKE = ?"))
			args = append(args, "018-%")
		}
	}
	query := `
		SELECT
			lead_id,
			no_ref,
			order_id,
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
		ORDER BY created_at ASC`

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
