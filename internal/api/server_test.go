package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/store"
)

type mockStore struct {
	jobs      []model.Job
	savedJobs []model.SavedJobWithJob
	sources   []model.Source
}

func (m *mockStore) GetActiveSources(ctx context.Context) ([]model.Source, error) {
	return m.sources, nil
}
func (m *mockStore) GetSourceByBoard(ctx context.Context, sourceType, boardIdentifier string) (*model.Source, error) {
	return nil, nil
}
func (m *mockStore) UpsertSource(ctx context.Context, src *model.Source) error {
	return nil
}
func (m *mockStore) UpsertJob(ctx context.Context, job *model.Job) (bool, bool, error) {
	return true, false, nil
}
func (m *mockStore) GetJobByID(ctx context.Context, id int64) (*model.Job, error) {
	for _, j := range m.jobs {
		if j.ID == id {
			return &j, nil
		}
	}
	return nil, nil
}
func (m *mockStore) ListJobs(ctx context.Context, filter store.JobFilter) ([]model.Job, int, error) {
	return m.jobs, len(m.jobs), nil
}
func (m *mockStore) ReconcileSourceJobs(ctx context.Context, sourceID int64, activeExternalIDs []string) (int64, error) {
	return 0, nil
}
func (m *mockStore) SaveJob(ctx context.Context, jobID int64, status, notes string) error {
	m.savedJobs = append(m.savedJobs, model.SavedJobWithJob{
		SavedJob: model.SavedJob{JobID: jobID, Status: status, Notes: notes},
	})
	return nil
}
func (m *mockStore) UnsaveJob(ctx context.Context, jobID int64) error {
	return nil
}
func (m *mockStore) GetSavedJobs(ctx context.Context) ([]model.SavedJobWithJob, error) {
	return m.savedJobs, nil
}
func (m *mockStore) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	return &model.DashboardStats{
		TotalActiveJobs:    len(m.jobs),
		NewJobsLast24h:     2,
		TotalSourcesActive: 5,
		TopSkills:          map[string]int{"Kubernetes": 10},
	}, nil
}
func (m *mockStore) StartCrawlRun(ctx context.Context, sourceID int64) (int64, error) {
	return 1, nil
}
func (m *mockStore) FinishCrawlRun(ctx context.Context, runID int64, status string, discovered, matched, newJobs, updatedJobs int, errMsg *string) error {
	return nil
}
func (m *mockStore) Migrate(ctx context.Context) error {
	return nil
}
func (m *mockStore) Close() {}

func TestServer_Endpoints(t *testing.T) {
	ms := &mockStore{
		jobs: []model.Job{
			{
				ID:            101,
				Title:         "Lead Kubernetes Architect",
				CompanyName:   "GitLab",
				Location:      "Bengaluru, India",
				MatchedSkills: []string{"Kubernetes", "Helm", "Go"},
				CanonicalURL:  "https://example.com/job/101",
				FirstSeenAt:   time.Now(),
			},
		},
		sources: []model.Source{
			{ID: 1, CompanyName: "GitLab", SourceType: "greenhouse", Enabled: true},
		},
	}

	server := NewServer(ms, nil)

	// 1. Healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want 200", w.Code)
	}

	// 2. List Jobs
	req = httptest.NewRequest(http.MethodGet, "/api/v1/jobs?skills=Kubernetes", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/jobs status = %d, want 200", w.Code)
	}

	var listResp struct {
		Jobs  []model.Job `json:"jobs"`
		Total int         `json:"total"`
	}
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode jobs list response: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Jobs) != 1 {
		t.Fatalf("expected 1 job, got total=%d len=%d", listResp.Total, len(listResp.Jobs))
	}
	if listResp.Jobs[0].Title != "Lead Kubernetes Architect" {
		t.Errorf("job title = %s, want Lead Kubernetes Architect", listResp.Jobs[0].Title)
	}

	// 3. Get Job by ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/jobs/101", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/jobs/101 status = %d, want 200", w.Code)
	}

	// 4. Save Job
	saveBody := `{"status": "applied", "notes": "Applied via referral"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/jobs/101/save", strings.NewReader(saveBody))
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/jobs/101/save status = %d, want 200", w.Code)
	}

	// 5. Get Saved Jobs
	req = httptest.NewRequest(http.MethodGet, "/api/v1/saved", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/saved status = %d, want 200", w.Code)
	}

	// 6. Get Stats
	req = httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/stats status = %d, want 200", w.Code)
	}
}
