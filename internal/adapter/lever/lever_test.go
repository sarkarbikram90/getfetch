package lever

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bikrams/getfetch/internal/model"
)

func TestLeverAdapter_FetchJobs(t *testing.T) {
	mockResponse := `[
		{
			"id": "lever-12345",
			"text": "Platform Engineer (Kubernetes)",
			"createdAt": 1728500000000,
			"hostedUrl": "https://jobs.lever.co/palantir/lever-12345?lever-source=linkedin",
			"applyUrl": "https://jobs.lever.co/palantir/lever-12345/apply",
			"categories": {
				"commitment": "Full-time",
				"department": "Infrastructure",
				"location": "Remote - India",
				"team": "Cloud"
			},
			"description": "<p>Operate multi-tenant Kubernetes clusters on AWS.</p>",
			"workplaceType": "remote",
			"country": "India"
		}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/palantir" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("mode") != "json" {
			t.Errorf("Expected mode=json query param")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	adapter.baseURL = server.URL

	src := model.Source{
		CompanyName:     "Palantir",
		BoardIdentifier: "palantir",
	}

	jobs, err := adapter.FetchJobs(context.Background(), src)
	if err != nil {
		t.Fatalf("FetchJobs returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	job := jobs[0]
	if job.ExternalID != "lever-12345" {
		t.Errorf("ExternalID = %s, want lever-12345", job.ExternalID)
	}
	if job.CanonicalURL != "https://jobs.lever.co/palantir/lever-12345" {
		t.Errorf("CanonicalURL not normalized: %s", job.CanonicalURL)
	}
	if job.Location != "Remote - India" {
		t.Errorf("Location = %s, want Remote - India", job.Location)
	}
	if job.RemoteType != "remote" {
		t.Errorf("RemoteType = %s, want remote", job.RemoteType)
	}
	if job.Department != "Infrastructure" {
		t.Errorf("Department = %s, want Infrastructure", job.Department)
	}
}
