package ashby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bikrams/getfetch/internal/model"
)

func TestAshbyAdapter_FetchJobs(t *testing.T) {
	mockResponse := `{
		"apiVersion": "0.1",
		"jobs": [
			{
				"id": "ashby-999",
				"title": "Lead Infrastructure Engineer",
				"department": "Platform",
				"location": "Bengaluru, India",
				"isRemote": true,
				"jobUrl": "https://jobs.ashbyhq.com/replit/ashby-999?utm_source=twitter",
				"applyUrl": "https://jobs.ashbyhq.com/replit/ashby-999/application",
				"publishedAt": "2026-10-11T12:00:00.000Z",
				"descriptionHtml": "<div>Manage large-scale Kubernetes clusters and Cilium service meshes.</div>"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/replit" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	adapter := NewAdapter(server.Client())
	adapter.baseURL = server.URL

	src := model.Source{
		CompanyName:     "Replit",
		BoardIdentifier: "replit",
	}

	jobs, err := adapter.FetchJobs(context.Background(), src)
	if err != nil {
		t.Fatalf("FetchJobs returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	job := jobs[0]
	if job.ExternalID != "ashby-999" {
		t.Errorf("ExternalID = %s, want ashby-999", job.ExternalID)
	}
	if job.CanonicalURL != "https://jobs.ashbyhq.com/replit/ashby-999" {
		t.Errorf("CanonicalURL not normalized: %s", job.CanonicalURL)
	}
	if job.RemoteType != "remote" {
		t.Errorf("RemoteType = %s, want remote", job.RemoteType)
	}
	if job.Department != "Platform" {
		t.Errorf("Department = %s, want Platform", job.Department)
	}
}
