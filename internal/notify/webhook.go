package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/httputil"
)

// WebhookNotifier sends alerts to Slack/Discord incoming webhook endpoints
type WebhookNotifier struct {
	webhookURL string
	client     adapter.HTTPDoer
}

// NewWebhookNotifier creates an initialized webhook notifier
func NewWebhookNotifier(webhookURL string, client adapter.HTTPDoer) *WebhookNotifier {
	if client == nil {
		client = httputil.NewClient()
	}
	return &WebhookNotifier{
		webhookURL: webhookURL,
		client:     client,
	}
}

func (w *WebhookNotifier) Name() string {
	return "webhook"
}

// SendDigest posts a formatted summary to Slack or Discord webhook
func (w *WebhookNotifier) SendDigest(ctx context.Context, payload DigestPayload) error {
	if w.webhookURL == "" {
		return fmt.Errorf("webhook URL is missing")
	}
	if len(payload.NewJobs) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(":zap: *KubeJobs Digest: %d New Roles Discovered*\n\n", len(payload.NewJobs)))

	for i, j := range payload.NewJobs {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("_...and %d more roles._\n", len(payload.NewJobs)-10))
			break
		}
		sb.WriteString(fmt.Sprintf("*<%s|%s>* — `%s`\n", j.CanonicalURL, j.Title, j.CompanyName))
		sb.WriteString(fmt.Sprintf(":round_pushpin: %s | :desktop_computer: %s\n", j.Location, j.RemoteType))
		sb.WriteString(fmt.Sprintf(":hammer_and_wrench: %s\n", strings.Join(j.MatchedSkills, ", ")))
		if j.SnippetMatch != "" {
			sb.WriteString(fmt.Sprintf("> %s\n", j.SnippetMatch))
		}
		sb.WriteString("\n")
	}

	reqBody := map[string]string{
		"text": sb.String(),
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to build webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("webhook endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
