package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bikrams/getfetch/internal/model"
)

func samplePayload() DigestPayload {
	return DigestPayload{
		Recipient:   "user@example.com",
		GeneratedAt: time.Now(),
		NewJobs: []model.Job{
			{
				Title:         "Staff Site Reliability Engineer",
				CompanyName:   "GitLab",
				Location:      "Bengaluru, India",
				RemoteType:    "remote",
				CanonicalURL:  "https://example.com/jobs/1",
				MatchedSkills: []string{"Kubernetes", "Helm", "Go"},
				SnippetMatch:  "...manage high-scale Kubernetes clusters and Helm charts in production...",
			},
		},
	}
}

func TestRenderHTML(t *testing.T) {
	payload := samplePayload()
	html, err := RenderHTML(payload)
	if err != nil {
		t.Fatalf("RenderHTML returned error: %v", err)
	}

	if !strings.Contains(html, "Staff Site Reliability Engineer") {
		t.Errorf("Expected job title in HTML")
	}
	if !strings.Contains(html, "GitLab") {
		t.Errorf("Expected company in HTML")
	}
	if !strings.Contains(html, "Kubernetes") {
		t.Errorf("Expected matched skills in HTML")
	}
}

func TestRenderPlainText(t *testing.T) {
	payload := samplePayload()
	text := RenderPlainText(payload)

	if !strings.Contains(text, "Staff Site Reliability Engineer") {
		t.Errorf("Expected job title in plain text")
	}
	if !strings.Contains(text, "https://example.com/jobs/1") {
		t.Errorf("Expected link in plain text")
	}
}

func TestResendNotifier_SendDigest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Missing or invalid authorization header")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "msg_123"}`))
	}))
	defer server.Close()

	n := NewResendNotifier("test-key", "alerts@kubejobs.dev", server.Client())
	n.apiURL = server.URL

	err := n.SendDigest(context.Background(), samplePayload())
	if err != nil {
		t.Fatalf("SendDigest failed: %v", err)
	}
}

func TestWebhookNotifier_SendDigest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	w := NewWebhookNotifier(server.URL, server.Client())
	err := w.SendDigest(context.Background(), samplePayload())
	if err != nil {
		t.Fatalf("Webhook SendDigest failed: %v", err)
	}
}
