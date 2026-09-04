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
		var isAutoAccepted sql.NullBool

		err := rows.Scan(
			&item.LeadID,
			&item.NoRef,
			&item.OrderID,
			&item.Partner,
			&item.NameOfInsurance,
			&dateOfBirth,
			&item.Gender,
			&item.MCUPackage,
			&isAutoAccepted,
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

		item.IsAutoAccepted = isAutoAccepted.Valid && isAutoAccepted.Bool

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repoReport: rows.Err: %w", err)
	}
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
			condition = append(condition, fmt.Sprintf("order_id LIKE ?"))
			args = append(args, "018-%")

		case report.SubmissionSourceExternal:
			condition = append(condition, fmt.Sprintf("source = ?"))
			args = append(args, 1)
			condition = append(condition, fmt.Sprintf("order_id NOT LIKE ?"))
			args = append(args, "018-%")
		}
	}

	if filter.IsAutoAccepted != nil {
		condition = append(condition, fmt.Sprintf("is_auto_accepted = ?"))
		args = append(args, *filter.IsAutoAccepted)
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
			is_auto_accepted,
			tenor,
			up,
			tanggal_order,
			source,
			status_sent_cn
		FROM vw_report
		WHERE ` + strings.Join(condition, " AND ") + ` 
		ORDER BY created_at ASC`

	//log.Debug().Interface("args", args).Msg("repoReport: build query data")
	return query, args, nil
}

func (repo *reportRepo) FetchMonthlyReportData(ctx context.Context, filter report.MonthlyReportFilter) ([]report.MonthlyReport, error) {
	query, args := repo.buildMonthlyQuery(filter)

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repoReport: select monthly: %w", err)
	}
	defer rows.Close()

	var result []report.MonthlyReport
	for rows.Next() {
		var item report.MonthlyReport
		if err := rows.Scan(
			&item.NameOfInsurance,
			&item.Connectivity,
			&item.Status,
			&item.Partner,
			&item.NoRef,
			&item.MCUPackage,
			&item.TanggalCreate,
		); err != nil {
			return nil, fmt.Errorf("repoReport: scan monthly: %w", err)
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repoReport: rows.Err monthly: %w", err)
	}
	return result, nil
}

func (repo *reportRepo) buildMonthlyQuery(filter report.MonthlyReportFilter) (string, []interface{}) {
	var (
		conditions []string
		args       []interface{}
	)

	conditions = append(conditions, "(l.deleted_at IS NULL OR l.deleted_at = 0)")
	conditions = append(conditions, "FROM_UNIXTIME(l.created_at / 1000) >= ?")
	args = append(args, filter.StartDate.Format("2006-01-02"))
	conditions = append(conditions, "FROM_UNIXTIME(l.created_at / 1000) < ?")
	args = append(args, filter.EndDate.Format("2006-01-02"))
	conditions = append(conditions, "mpt.id <> 31")

	if filter.PartnerID != nil {
		conditions = append(conditions, "p.id = ?")
		args = append(args, *filter.PartnerID)
	}

	if filter.Status != nil {
		conditions = append(conditions, "l.status = ?")
		args = append(args, *filter.Status)
	}

	query := `
		SELECT
			lp.name_of_insurance AS name_of_insurance,
			CASE
				WHEN l.is_submission = true THEN 'esubmission'
				ELSE 'upload'
			END AS 'Connectivity',
			CASE
				WHEN l.status = 0 THEN 'Invalid'
				WHEN l.status = 1 THEN 'New'
				WHEN l.status = 2 THEN 'OnProcess'
				WHEN l.status = 3 THEN 'Accepted'
				WHEN l.status = 4 THEN 'Rejected'
				WHEN l.status = 5 THEN 'Canceled'
				WHEN l.status = 6 THEN 'Expired'
				WHEN l.status = 7 THEN 'Paid'
				WHEN l.status = 8 THEN 'Postpone'
				ELSE ''
			END AS 'Status',
			p.partner_name AS partner,
			li.ref_number AS no_ref,
			COALESCE(mpt.mcu_type_name, '-') AS mcu_package,
			FROM_UNIXTIME(l.created_at / 1000) AS tanggal_create
		FROM leads l
		JOIN lead_personals lp ON l.id = lp.lead_id AND lp.deleted_at IS NULL
		LEFT JOIN partners p ON l.partner_id = p.id
		JOIN lead_installments li ON l.id = li.lead_id AND li.deleted_at IS NULL
		LEFT JOIN mcu_package_types mpt ON l.mcu_package_type_id = mpt.id
		LEFT JOIN mcu_packages mp ON mp.mcu_package_type_id = mpt.id
		WHERE ` + strings.Join(conditions, " AND ") + `
		ORDER BY tanggal_create ASC`

	return query, args
}

func (repo *reportRepo) Insert(ctx context.Context, entry report.AutomailReportLog) error {
	query := `
        INSERT INTO automail_report_logs (
            request_id,
            activity_id,
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
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	data, err := repo.db.ExecContext(ctx, query,
		entry.RequestID,
		entry.ActivityID,
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

	log.Debug().Msgf("automail_log_repo: insert: %v", data)
	return nil
}
