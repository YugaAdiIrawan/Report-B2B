package report

import "time"

type Ext struct {
	ID   string `json:"id"`
	Desc string `json:"desc"`
}

type PayloadComoEmail struct {
	From        string           `json:"from"`
	To          string           `json:"to"`
	Cc          string           `json:"cc"`
	Bcc         string           `json:"bcc"`
	ReplyTo     string           `json:"replyTo"`
	Subject     string           `json:"subject"`
	Text        string           `json:"text"`
	HTML        string           `json:"html"`
	Attachments []AttachmentItem `json:"attachments,omitempty"`
}

type RequestPayloadComo struct {
	Ext     Ext              `json:"ext"`
	Payload PayloadComoEmail `json:"payload"`
}

type AttachmentItem struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"` // "base64"
	ContentType string `json:"contentType,omitempty"`
}

type ComoConfig struct {
	SendBaseURL string
	SendApiKey  string
	InqBaseURL  string
	InqApiKey   string
	FromEmail   string
	Timeout     time.Duration
}

type AutoUWSchedulerConfig struct {
	Recipients  []string
	CCList      []string
	Timezone    string
	RunHour     int
	RunMinute   int
	IntervalMin int
}

type ComoActivityData struct {
	StatCode   string `json:"statCode"`
	StatRemark string `json:"statRemark"`
	ActivityID string `json:"activityId"`
	SeqNo      int    `json:"seqNo,omitempty"`
	SendStatus string `json:"sendStatus,omitempty"`
	SendDate   string `json:"sendDate,omitempty"`
}

type ComoActivityResponse struct {
	Data ComoActivityData `json:"data"`
}

const (
	StatCodeSuccess  = "00"
	SendStatusSent   = "sent"
	StatCodeNotFound = "01"
)

func (r *ComoActivityResponse) IsRequestAccepted() bool {
	return r != nil && r.Data.StatCode == StatCodeSuccess
}

func (r *ComoActivityResponse) IsDelivered() bool {
	return r != nil && r.Data.StatCode == StatCodeSuccess && r.Data.SendStatus == SendStatusSent
}

func (r *ComoActivityResponse) IsNotFoundTransient() bool {
	return r != nil && r.Data.StatCode == StatCodeNotFound
}

type ComoErrorResponse struct {
	Errno   int    `json:"errno"`
	Code    string `json:"code"`
	Syscall string `json:"syscall"`
	Command string `json:"command"`
}

type ComoEmailResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    *ComoEmailResponseData `json:"data,omitempty"`
}

type ComoEmailResponseData struct {
	MessageID string `json:"messageId,omitempty"`
}
