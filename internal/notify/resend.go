package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/httputil"
)

// ResendNotifier sends digests via the Resend transactional email API
type ResendNotifier struct {
	apiKey      string
	fromAddress string
	client      adapter.HTTPDoer
	apiURL      string
}

// NewResendNotifier creates an initialized Resend email notifier
func NewResendNotifier(apiKey, fromAddress string, client adapter.HTTPDoer) *ResendNotifier {
	if client == nil {
		client = httputil.NewClient()
	}
	if fromAddress == "" {
		fromAddress = "KubeJobs <onboarding@resend.dev>"
	}
	return &ResendNotifier{
		apiKey:      apiKey,
		fromAddress: fromAddress,
		client:      client,
		apiURL:      "https://api.resend.com/emails",
	}
}

func (r *ResendNotifier) Name() string {
	return "resend"
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

// SendDigest sends an email digest to the specified recipient
func (r *ResendNotifier) SendDigest(ctx context.Context, payload DigestPayload) error {
	if r.apiKey == "" {
		return fmt.Errorf("resend api key is missing")
	}
	if payload.Recipient == "" {
		return fmt.Errorf("recipient email address is missing")
	}
	if len(payload.NewJobs) == 0 && len(payload.UpdatedJobs) == 0 {
		return nil
	}

	htmlBody, err := RenderHTML(payload)
	if err != nil {
		return fmt.Errorf("failed to render html email: %w", err)
	}

	textBody := RenderPlainText(payload)
	subject := fmt.Sprintf("⚡ KubeJobs Digest: %d New Roles Found (%s)",
		len(payload.NewJobs), time.Now().Format("Jan 02"))

	reqBody := resendRequest{
		From:    r.fromAddress,
		To:      []string{payload.Recipient},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "getfetch/1.0")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email via resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("resend api returned error HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}
