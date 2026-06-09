package report

import (
	"strings"
	"time"
)

type CNReleaseStatus string

const (
	CNStatusNull                   CNReleaseStatus = "null"
	CNStatusSuccessSentSFTP        CNReleaseStatus = "success sent sftp"
	CNStatusSuccessCopied          CNReleaseStatus = "success copied"
	CNStatusProgress               CNReleaseStatus = "progress"
	CNStatusFailedDirNotFound      CNReleaseStatus = "Failed,Dir Not Found"
	CNStatusFailedAfterRetry       CNReleaseStatus = "failed connect to Server, after 3 retrying"
	CNStatusFailedConnectToPartner CNReleaseStatus = "Failed Connect to Partner"
)

func (s CNReleaseStatus) IsValid() bool {
	switch s {
	case CNStatusNull,
		CNStatusSuccessSentSFTP,
		CNStatusSuccessCopied,
		CNStatusProgress,
		CNStatusFailedDirNotFound,
		CNStatusFailedAfterRetry,
		CNStatusFailedConnectToPartner:
		return true
	}
	return false
}

// SubmissionSource: false = Upload, true = API
type SubmissionSourceType string

const (
	SubmissionSourceUpload      SubmissionSourceType = "UPLOAD"
	SubmissionSourceESubmission SubmissionSourceType = "ESUBMISSION"
	SubmissionSourceExternal    SubmissionSourceType = "EXTERNAL"
)

func (s SubmissionSourceType) IsValid() bool {
	switch s {
	case SubmissionSourceUpload,
		SubmissionSourceESubmission,
		SubmissionSourceExternal:
		return true
	}
	return false
}

func ResolveSubmissionSource(source int, noRef string) SubmissionSourceType {
	if source == 0 {
		return SubmissionSourceUpload
	}
	if strings.HasPrefix(noRef, "018-") {
		return SubmissionSourceESubmission
	}
	return SubmissionSourceExternal
}

type ReportFilter struct {
	StartDate        time.Time
	EndDate          time.Time
	CNReleaseStatus  *CNReleaseStatus      // nil = all
	SubmissionSource *SubmissionSourceType // nil = all, true = API, false = Upload atau External
}

type AutoUWReport struct {
	LeadID          int64
	NoRef           string
	OrderID         string
	Partner         string
	NameOfInsurance string
	DateOfBirth     *time.Time
	Gender          string
	MCUPackage      string
	Tenor           float64
	UP              float64
	TanggalOrder    time.Time
	Source          int
	StatusSentCN    *string
	PicEmail        *string
}

type SendReportEmailRequest struct {
	To         []string
	Subject    string
	HTMLBody   string
	Attachment EmailAttachment
}

type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type ReportRequest struct {
	StartDate        string  `form:"start_date" binding:"required"` // format: 2006-01-02
	EndDate          string  `form:"end_date" binding:"required"`   // format: 2006-01-02
	CNReleaseStatus  *string `form:"cn_release_status"`             // optional
	SubmissionSource *string `form:"submission_source"`             // optional: 0=API, 1=Upload
}

type AutomailReportLog struct {
	ID             int64
	RequestID      string
	FromEmail      string
	ToEmails       string
	CcEmails       string
	Subject        string
	HasAttachment  bool
	AttachmentName string
	AttachmentSize int
	HTTPStatus     *int
	ResponseBody   *string
	ErrorMessage   *string
	Status         AutomailLogStatus
	SentAt         time.Time
	CreatedAt      time.Time
}

type AutomailLogStatus string

const (
	AutomailLogStatusSuccess   AutomailLogStatus = "SUCCESS"
	AutomailLogStatusFailed    AutomailLogStatus = "FAILED"     // error sebelum HTTP (marshal, create req)
	AutomailLogStatusHTTPError AutomailLogStatus = "HTTP_ERROR" // dapat response tapi non-2xx
)
