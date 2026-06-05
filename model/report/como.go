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
	BaseURL   string
	ApiKey    string
	FromEmail string
	Timeout   time.Duration
}

type AutoUWSchedulerConfig struct {
	Recipients  []string
	CCList      []string
	Timezone    string
	RunHour     int
	RunMinute   int
	IntervalMin int
}

type ComoEmailResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    *ComoEmailResponseData `json:"data,omitempty"`
}

type ComoEmailResponseData struct {
	MessageID string `json:"messageId,omitempty"`
}
