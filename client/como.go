package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/YugaAdiIrawan/model/report"
	report2 "github.com/YugaAdiIrawan/module/report"
	"github.com/rs/zerolog/log"
)

type EmailService interface {
	SendReportEmail(ctx context.Context, req report.SendReportEmailRequest) error
}

type comoEmailService struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	fromEmail  string
	logRepo    report2.AutomailReportLogRepository
}

func NewComoEmailService(cfg report.ComoConfig, logRepo report2.AutomailReportLogRepository) *comoEmailService {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &comoEmailService{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.ApiKey,
		fromEmail:  cfg.FromEmail,
		logRepo:    logRepo,
	}
}

func (s *comoEmailService) SendReportEmail(ctx context.Context, req report.SendReportEmailRequest) error {
	requestID := fmt.Sprintf("auto-uw-report-%d", time.Now().UnixMilli())
	sentAt := time.Now()
	toStr := joinsEmails(req.To)

	entry := report.AutomailReportLog{
		RequestID:      requestID,
		FromEmail:      s.fromEmail,
		ToEmails:       toStr,
		CcEmails:       "",
		Subject:        req.Subject,
		HasAttachment:  len(req.Attachment.Data) > 0,
		AttachmentName: req.Attachment.Filename,
		AttachmentSize: len(req.Attachment.Data),
		Status:         report.AutomailLogStatusFailed,
		SentAt:         sentAt,
	}

	defer func() {
		logCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if insertErr := s.logRepo.Insert(logCtx, entry); insertErr != nil {
			log.Error().
				Err(insertErr).
				Str("request_id", requestID).
				Msg("como_email: failed to insert automail report log")
		}
		log.Debug().Msgf("Data : %v", entry)
	}()

	payload := report.RequestPayloadComo{
		Ext: report.Ext{
			ID:   requestID,
			Desc: "Auto UW Dealy Report",
		}, Payload: report.PayloadComoEmail{
			From:    s.fromEmail,
			To:      toStr,
			Cc:      "",
			Bcc:     "",
			ReplyTo: s.fromEmail,
			Subject: req.Subject,
			Text:    "",
			HTML:    req.HTMLBody,
		},
	}
	if len(req.Attachment.Data) > 0 {
		payload.Payload.Attachments = []report.AttachmentItem{
			{
				Filename:    req.Attachment.Filename,
				Content:     base64.StdEncoding.EncodeToString(req.Attachment.Data),
				Encoding:    "base64",
				ContentType: req.Attachment.ContentType,
			},
		}

		log.Debug().
			Str("filename", req.Attachment.Filename).
			Str("content_type", req.Attachment.ContentType).
			Int("original_bytes", len(req.Attachment.Data)).
			Msg("como_email: attachment prepared")
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		errMsg := err.Error()
		entry.ErrorMessage = &errMsg
		return fmt.Errorf("send report email: marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.baseURL,
		bytes.NewReader(bodyBytes),
	)

	if err != nil {
		errMsg := err.Error()
		entry.ErrorMessage = &errMsg
		return fmt.Errorf("send report email: create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		errMsg := err.Error()
		entry.ErrorMessage = &errMsg
		return fmt.Errorf("send report email: httpDo: %w", err)
	}
	defer resp.Body.Close()

	respBody, errRespBody := io.ReadAll(resp.Body)
	if errRespBody != nil {
		errMsg := err.Error()
		entry.ErrorMessage = &errMsg
		return fmt.Errorf("send report email: read response: %w", errRespBody)
	}

	httpStatus := resp.StatusCode
	respBodyStr := string(respBody)
	entry.HTTPStatus = &httpStatus
	entry.ResponseBody = &respBodyStr

	log.Debug().
		Int("status", resp.StatusCode).
		Str("response", string(respBody)).
		Msg("como_email: response received")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, respBodyStr)
		entry.ErrorMessage = &errMsg
		return fmt.Errorf("como_email: %s", errMsg)
	}

	entry.Status = report.AutomailLogStatusSuccess
	return nil
}

func joinsEmails(emails []string) string {
	result := ""
	for i, e := range emails {
		if i > 0 {
			result += ","
		}
		result += e
	}
	return result
}
