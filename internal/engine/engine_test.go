package engine

import (
	"context"
	"testing"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/matcher"
	"github.com/bikrams/getfetch/internal/model"
)

type mockAdapter struct {
	name string
	jobs []model.RawJob
	err  error
}

func (m *mockAdapter) Name() string {
	return m.name
}

func (m *mockAdapter) FetchJobs(ctx context.Context, source model.Source) ([]model.RawJob, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.jobs, nil
}

func TestEngine_CrawlSource_DryRun(t *testing.T) {
	reg := adapter.NewRegistry()
	mockAd := &mockAdapter{
		name: model.SourceTypeGreenhouse,
		jobs: []model.RawJob{
			{
				ExternalID:     "101",
				Title:          "Senior Site Reliability Engineer",
				Location:       "Bengaluru, India",
				CanonicalURL:   "https://example.com/jobs/101",
				DescriptionRaw: "<p>We run high-scale services using Kubernetes, Helm, and Terraform.</p>",
			},
			{
				ExternalID:     "102",
				Title:          "Frontend Engineer",
				Location:       "Remote",
				CanonicalURL:   "https://example.com/jobs/102",
				DescriptionRaw: "<p>React, CSS, and Figma mockups.</p>",
			},
		},
	}
	reg.Register(mockAd)

	eng := NewEngine(nil, reg, matcher.NewMatcher(), nil)

	src := model.Source{
		ID:              1,
		SourceType:      model.SourceTypeGreenhouse,
		CompanyName:     "TestCloud",
		BoardIdentifier: "testcloud",
	}

	summary := eng.CrawlSource(context.Background(), src, true)

	if summary.Error != nil {
		t.Fatalf("CrawlSource returned unexpected error: %v", errToString(summary.Error))
	}
	if summary.Discovered != 2 {
		t.Errorf("Discovered = %d, want 2", summary.Discovered)
	}
	if summary.Matched != 1 {
		t.Errorf("Matched = %d, want 1 (only SRE/K8s job)", summary.Matched)
	}
	if len(summary.MatchedJobs) != 1 {
		t.Fatalf("MatchedJobs len = %d, want 1", len(summary.MatchedJobs))
	}

	matched := summary.MatchedJobs[0]
	if matched.ExternalID != "101" {
		t.Errorf("Matched job ExternalID = %s, want 101", matched.ExternalID)
	}
	if matched.Country != "India" {
		t.Errorf("Matched job Country = %s, want India", matched.Country)
	}
}

func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
