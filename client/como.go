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
	sendURL    string
	sendApiKey string
	inqBaseURL string
	inqApiKey  string
	fromEmail  string
	logRepo    report2.AutomailReportLogRepository
}

func NewComoEmailService(cfg report.ComoConfig, logRepo report2.AutomailReportLogRepository) *comoEmailService {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.SendBaseURL == "" || cfg.SendApiKey == "" {
		log.Fatal().Msg("como_email: SendBaseURL/SendApiKey kosong — pastikan LoadComoSendEmailFromDB sudah dipanggil")
	}

	return &comoEmailService{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		sendURL:    cfg.SendBaseURL,
		sendApiKey: cfg.SendApiKey,
		inqBaseURL: cfg.InqBaseURL,
		inqApiKey:  cfg.InqApiKey,
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

	activityID, err := s.createActivity(ctx, requestID, toStr, req, &entry)
	if err != nil {
		return err
	}

	entry.ActivityID = activityID
	entry.Status = report.AutomailLogStatusPendingConfirmation

	return nil
}

func (s *comoEmailService) createActivity(ctx context.Context, requestID string, toStr string, req report.SendReportEmailRequest, entry *report.AutomailReportLog) (string, error) {
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
		return "", fmt.Errorf("send report email: marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.sendURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		errMsg := err.Error()
		entry.ErrorMessage = &errMsg
		return "", fmt.Errorf("send report email: create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", s.sendApiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		errMsg := err.Error()
		entry.ErrorMessage = &errMsg
		return "", fmt.Errorf("send report email: httpDo: %w", err)
	}
	defer resp.Body.Close()

	respBody, errRespBody := io.ReadAll(resp.Body)
	if errRespBody != nil {
		errMsg := errRespBody.Error()
		entry.ErrorMessage = &errMsg
		return "", fmt.Errorf("send report email: read response: %w", errRespBody)
	}

	httpStatus := resp.StatusCode
	respBodyStr := string(respBody)
	entry.HTTPStatus = &httpStatus
	entry.ResponseBody = &respBodyStr

	log.Debug().
		Str("request_id", requestID).
		Int("status", resp.StatusCode).
		Str("response", respBodyStr).
		Msg("como_email: create activity response received")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("create activity: unexpected http status %d: %s", resp.StatusCode, respBodyStr)
		entry.ErrorMessage = &errMsg
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	var comoResp report.ComoActivityResponse
	if err := json.Unmarshal(respBody, &comoResp); err != nil {
		errMsg := fmt.Sprintf("create activity: invalid response envelope: %s (raw: %s)", err.Error(), respBodyStr)
		entry.ErrorMessage = &errMsg
		entry.Status = report.AutomailLogStatusHTTPError
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("response_body", respBodyStr).
			Msg("como_email: failed to unmarshal create activity response")
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	if !comoResp.IsRequestAccepted() {
		errMsg := fmt.Sprintf("create activity: request rejected by Como (stat_code=%s, body=%s)",
			comoResp.Data.StatCode, respBodyStr)
		entry.ErrorMessage = &errMsg
		entry.Status = report.AutomailLogStatusHTTPError
		log.Error().
			Str("request_id", requestID).
			Str("stat_code", comoResp.Data.StatCode).
			Msg("como_email: Como returned 2xx but request was not accepted")
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	if comoResp.Data.ActivityID == "" {
		errMsg := fmt.Sprintf("create activity: stat_code 00 but activityId is empty (body=%s)", respBodyStr)
		entry.ErrorMessage = &errMsg
		entry.Status = report.AutomailLogStatusHTTPError
		log.Error().
			Str("request_id", requestID).
			Msg("como_email: missing activityId despite accepted stat_code")
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	return comoResp.Data.ActivityID, nil
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
