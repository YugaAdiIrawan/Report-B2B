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

const (
	inquiryTimeout        = 10 * time.Second
	maxInquiryRetries     = 5
	inquiryRetryBaseDelay = 2 * time.Second
)

func NewComoEmailService(cfg report.ComoConfig, logRepo report2.AutomailReportLogRepository) *comoEmailService {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.SendBaseURL == "" || cfg.SendApiKey == "" {
		log.Fatal().Msg("como_email: SendBaseURL/SendApiKey kosong — pastikan LoadComoSendEmailFromDB sudah dipanggil")
	}
	if cfg.InqBaseURL == "" || cfg.InqApiKey == "" {
		log.Fatal().Msg("como_email: InqBaseURL/InqApiKey kosong — pastikan LoadComoInqEmailFromDB sudah dipanggil")
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

	return s.confirmDelivery(ctx, requestID, activityID, &entry)
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
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("create activity: unexpected status %d: %s", resp.StatusCode, respBodyStr)
		entry.ErrorMessage = &errMsg
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	var comoResp report.ComoActivityResponse
	if errUnmarshal := json.Unmarshal(respBody, &comoResp); errUnmarshal != nil {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("create activity: response body is not valid envelope: %s (raw: %s)", errUnmarshal.Error(), respBodyStr)
		entry.ErrorMessage = &errMsg
		log.Error().
			Err(errUnmarshal).
			Str("request_id", requestID).
			Str("response_body", respBodyStr).
			Msg("como_email: failed to parse create activity response")
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	if !comoResp.IsRequestAccepted() {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("create activity: como rejected request (httpStatus=%d, statCode=%s, body=%s)",
			resp.StatusCode, comoResp.Data.StatCode, respBodyStr)
		entry.ErrorMessage = &errMsg
		log.Error().
			Str("request_id", requestID).
			Int("http_status", resp.StatusCode).
			Str("stat_code", comoResp.Data.StatCode).
			Msg("como_email: como returned 2xx but request was not accepted")
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	if comoResp.Data.ActivityID == "" {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("create activity: statCode 00 but activityId is empty (body=%s)", respBodyStr)
		entry.ErrorMessage = &errMsg
		log.Error().
			Str("request_id", requestID).
			Str("response_body", respBodyStr).
			Msg("como_email: missing activityId despite successful statCode")
		return "", fmt.Errorf("como_email: %s", errMsg)
	}

	log.Info().
		Str("request_id", requestID).
		Str("activity_id", comoResp.Data.ActivityID).
		Msg("como_email: activity created, proceeding to delivery confirmation")

	return comoResp.Data.ActivityID, nil
}

func (s *comoEmailService) confirmDelivery(ctx context.Context, requestID string, activityID string, entry *report.AutomailReportLog) error {
	var lastErr error

	for attempt := 1; attempt <= maxInquiryRetries; attempt++ {
		if attempt > 1 {
			delay := inquiryRetryBaseDelay * time.Duration(1<<(attempt-2)) // 2s, 4s, 8s, 16s
			log.Warn().
				Str("request_id", requestID).
				Str("activity_id", activityID).
				Int("attempt", attempt).
				Dur("delay", delay).
				Msg("como_email: activity not found yet, retrying inquiry after backoff")

			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				entry.Status = report.AutomailLogStatusHTTPError
				errMsg := fmt.Sprintf("inquiry activity: context cancelled while waiting to retry: %s", ctx.Err())
				entry.ErrorMessage = &errMsg
				return fmt.Errorf("como_email: %s", errMsg)
			case <-timer.C:
			}
		}

		comoResp, respBodyStr, httpStatus, err := s.inquireOnce(ctx, requestID, activityID, entry)
		if err != nil {
			lastErr = err
			return lastErr
		}

		if comoResp.IsNotFoundTransient() {
			lastErr = fmt.Errorf("como_email: inquiry activity: activity not found yet (attempt %d/%d)", attempt, maxInquiryRetries)
			entry.Status = report.AutomailLogStatusHTTPError
			errMsg := fmt.Sprintf("inquiry activity: activity not found after %d attempt(s) (httpStatus=%d, body=%s)", attempt, httpStatus, respBodyStr)
			entry.ErrorMessage = &errMsg
			continue
		}

		if !comoResp.IsDelivered() {
			entry.Status = report.AutomailLogStatusHTTPError
			errMsg := fmt.Sprintf("inquiry activity: email not delivered (statCode=%s, sendStatus=%s, body=%s)",
				comoResp.Data.StatCode, comoResp.Data.SendStatus, respBodyStr)
			entry.ErrorMessage = &errMsg
			log.Error().
				Str("request_id", requestID).
				Str("activity_id", activityID).
				Str("stat_code", comoResp.Data.StatCode).
				Str("send_status", comoResp.Data.SendStatus).
				Msg("como_email: inquiry confirmed delivery did not succeed (final state, not retrying)")
			return fmt.Errorf("como_email: %s", errMsg)
		}

		entry.Status = report.AutomailLogStatusSuccess
		log.Info().
			Str("request_id", requestID).
			Str("activity_id", activityID).
			Str("send_date", comoResp.Data.SendDate).
			Int("attempt", attempt).
			Msg("como_email: delivery confirmed via inquiry")
		return nil
	}

	log.Error().
		Str("request_id", requestID).
		Str("activity_id", activityID).
		Int("max_retries", maxInquiryRetries).
		Msg("como_email: activity still not found after exhausting all retries")
	return lastErr
}

func (s *comoEmailService) inquireOnce(ctx context.Context, requestID string, activityID string, entry *report.AutomailReportLog) (*report.ComoActivityResponse, string, int, error) {
	inqCtx, cancel := context.WithTimeout(ctx, inquiryTimeout)
	defer cancel()

	inquiryURL := fmt.Sprintf("%s=%s", s.inqBaseURL, activityID)

	log.Debug().
		Str("request_id", requestID).
		Str("inquiry_url", inquiryURL).
		Msg("como_email: constructed inquiry URL")

	httpReq, err := http.NewRequestWithContext(inqCtx, http.MethodGet, inquiryURL, nil)
	if err != nil {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("inquiry activity: create request: %s", err.Error())
		entry.ErrorMessage = &errMsg
		return nil, "", 0, fmt.Errorf("como_email: %s", errMsg)
	}
	httpReq.Header.Set("Authorization", s.inqApiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("inquiry activity: httpDo: %s", err.Error())
		entry.ErrorMessage = &errMsg
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("activity_id", activityID).
			Msg("como_email: inquiry request failed")
		return nil, "", 0, fmt.Errorf("como_email: %s", errMsg)
	}
	defer resp.Body.Close()

	respBody, errRespBody := io.ReadAll(resp.Body)
	if errRespBody != nil {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("inquiry activity: read response: %s", errRespBody.Error())
		entry.ErrorMessage = &errMsg
		return nil, "", 0, fmt.Errorf("como_email: %s", errMsg)
	}

	httpStatus := resp.StatusCode
	respBodyStr := string(respBody)
	entry.HTTPStatus = &httpStatus
	entry.ResponseBody = &respBodyStr

	log.Debug().
		Str("request_id", requestID).
		Str("activity_id", activityID).
		Int("status", resp.StatusCode).
		Str("response", respBodyStr).
		Msg("como_email: inquiry activity response received")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		entry.Status = report.AutomailLogStatusHTTPError
		errMsg := fmt.Sprintf("inquiry activity: unexpected status %d: %s", resp.StatusCode, respBodyStr)
		entry.ErrorMessage = &errMsg
		return nil, respBodyStr, httpStatus, fmt.Errorf("como_email: %s", errMsg)
	}

	var comoResp report.ComoActivityResponse
	if errUnmarshal := json.Unmarshal(respBody, &comoResp); errUnmarshal != nil {
		entry.Status = report.AutomailLogStatusHTTPError

		var comoErr report.ComoErrorResponse
		detail := respBodyStr
		if errSecondary := json.Unmarshal(respBody, &comoErr); errSecondary == nil && comoErr.Code != "" {
			detail = fmt.Sprintf("upstream transport error code=%s syscall=%s (raw: %s)", comoErr.Code, comoErr.Syscall, respBodyStr)
		}

		errMsg := fmt.Sprintf("inquiry activity: delivery confirmation failed: %s", detail)
		entry.ErrorMessage = &errMsg
		log.Error().
			Str("request_id", requestID).
			Str("activity_id", activityID).
			Str("response_body", respBodyStr).
			Msg("como_email: inquiry returned non-standard error envelope, treating as delivery failure")
		return nil, respBodyStr, httpStatus, fmt.Errorf("como_email: %s", errMsg)
	}

	return &comoResp, respBodyStr, httpStatus, nil
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
