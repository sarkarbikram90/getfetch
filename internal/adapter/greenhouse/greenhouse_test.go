package greenhouse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bikrams/getfetch/internal/model"
)

func TestGreenhouseAdapter_FetchJobs(t *testing.T) {
	mockResponse := `{
		"jobs": [
			{
				"id": 9876543,
				"title": "Senior Site Reliability Engineer - Kubernetes",
				"location": { "name": "Bengaluru, India" },
				"content": "<p>Looking for a Kubernetes and Terraform specialist in India.</p>",
				"absolute_url": "https://boards.greenhouse.io/example/jobs/9876543?gh_src=test",
				"updated_at": "2026-10-12T10:00:00Z",
				"departments": [{"id": 1, "name": "Infrastructure"}]
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/testco/jobs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("content") != "true" {
			t.Errorf("Expected content=true query param")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	adapter.baseURL = server.URL

	src := model.Source{
		CompanyName:     "TestCo",
		BoardIdentifier: "testco",
	}

	jobs, err := adapter.FetchJobs(context.Background(), src)
	if err != nil {
		t.Fatalf("FetchJobs returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	job := jobs[0]
	if job.ExternalID != "9876543" {
		t.Errorf("ExternalID = %s, want 9876543", job.ExternalID)
	}
	if job.CanonicalURL != "https://boards.greenhouse.io/example/jobs/9876543" {
		t.Errorf("CanonicalURL not normalized: %s", job.CanonicalURL)
	}
	if job.Location != "Bengaluru, India" {
		t.Errorf("Location = %s, want Bengaluru, India", job.Location)
	}
	if job.Department != "Infrastructure" {
		t.Errorf("Department = %s, want Infrastructure", job.Department)
	}
}
