package matcher

import (
	"strings"
	"testing"

	"github.com/bikrams/getfetch/internal/model"
)

func TestMatcher_Evaluate(t *testing.T) {
	m := NewMatcher()

	tests := []struct {
		name          string
		title         string
		location      string
		desc          string
		adapterRemote string
		wantMatch     bool
		wantSkills    []string
		wantRoles     []string
		wantRemote    string
	}{
		{
			name:          "Direct Kubernetes SRE role",
			title:         "Senior Site Reliability Engineer",
			location:      "Bengaluru, India",
			desc:          "We operate mission-critical clusters using Kubernetes, Helm, and Terraform on AWS.",
			adapterRemote: "hybrid",
			wantMatch:     true,
			wantSkills:    []string{"Kubernetes", "Helm", "Terraform", "AWS"},
			wantRoles:     []string{"Site Reliability Engineer"},
			wantRemote:    model.RemoteTypeHybrid,
		},
		{
			name:          "K8s abbreviation match",
			title:         "Platform Engineer",
			location:      "Remote - India",
			desc:          "Experience with k8s pod management, Prometheus monitoring, and Golang.",
			adapterRemote: "",
			wantMatch:     true,
			wantSkills:    []string{"Kubernetes", "Prometheus", "Go"},
			wantRoles:     []string{"Platform Engineer"},
			wantRemote:    model.RemoteTypeRemote,
		},
		{
			name:          "Non-matching role (Frontend React)",
			title:         "Frontend Software Engineer",
			location:      "San Francisco, CA",
			desc:          "Build UI components with React, TypeScript, and CSS. No infra mentioned.",
			adapterRemote: "onsite",
			wantMatch:     false,
			wantSkills:    []string{},
			wantRoles:     []string{},
			wantRemote:    model.RemoteTypeOnsite,
		},
		{
			name:          "Infra role with multiple skills (Docker + Terraform)",
			title:         "Cloud Infrastructure Engineer",
			location:      "Hyderabad",
			desc:          "Manage cloud infrastructure using Terraform, Docker, and GCP.",
			adapterRemote: "remote",
			wantMatch:     true,
			wantSkills:    []string{"Terraform", "Docker", "GCP"},
			wantRoles:     []string{"Infrastructure Engineer", "Cloud Engineer"},
			wantRemote:    model.RemoteTypeRemote,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := m.Evaluate(tt.title, tt.location, tt.desc, tt.adapterRemote)
			if res.IsTargetMatch != tt.wantMatch {
				t.Fatalf("Evaluate() IsTargetMatch = %v, want %v", res.IsTargetMatch, tt.wantMatch)
			}

			if tt.wantRemote != "" && res.RemoteType != tt.wantRemote {
				t.Errorf("RemoteType = %v, want %v", res.RemoteType, tt.wantRemote)
			}

			for _, skill := range tt.wantSkills {
				found := false
				for _, s := range res.MatchedSkills {
					if s == skill {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected skill %s in MatchedSkills: %v", skill, res.MatchedSkills)
				}
			}

			for _, role := range tt.wantRoles {
				found := false
				for _, r := range res.MatchedRoles {
					if r == role {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected role %s in MatchedRoles: %v", role, res.MatchedRoles)
				}
			}

			if tt.wantMatch && res.SnippetMatch == "" {
				t.Errorf("Expected non-empty snippet match for qualified job")
			}
		})
	}
}

func TestMatcher_IsIndiaLocation(t *testing.T) {
	m := NewMatcher()

	if !m.IsIndiaLocation("Bengaluru, Karnataka", "") {
		t.Errorf("Expected Bengaluru to match India")
	}
	if !m.IsIndiaLocation("Pune", "IN") {
		t.Errorf("Expected Pune to match India")
	}
	if m.IsIndiaLocation("London, UK", "GB") {
		t.Errorf("Expected London, UK NOT to match India")
	}
}

func TestMatcher_SnippetExtraction(t *testing.T) {
	m := NewMatcher()
	desc := "Our team manages 50 production clusters. You will deploy and scale Kubernetes applications using Helm and Flux in a multi-region environment. We also maintain observability pipelines."
	res := m.Evaluate("SRE", "Remote", desc, "remote")

	if !strings.Contains(strings.ToLower(res.SnippetMatch), "kubernetes") {
		t.Errorf("Expected snippet to contain kubernetes, got: %s", res.SnippetMatch)
	}
}
