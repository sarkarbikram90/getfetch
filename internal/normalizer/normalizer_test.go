package normalizer

import (
	"strings"
	"testing"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // substrings that must be in output
		excluded []string // substrings that must NOT be in output
	}{
		{
			name:  "Basic HTML with paragraphs and lists",
			input: "<h1>Senior SRE</h1><p>We are hiring a Kubernetes engineer.</p><ul><li>Go</li><li>Terraform</li></ul>",
			expected: []string{
				"Senior SRE",
				"We are hiring a Kubernetes engineer.",
				"Go",
				"Terraform",
			},
			excluded: []string{"<h1>", "</h1>", "<p>", "</p>", "<li>", "</li>", "<ul>"},
		},
		{
			name:  "HTML with script and style tags",
			input: "<style>.title { color: red; }</style><p>Real Content</p><script>alert('bad');</script>",
			expected: []string{
				"Real Content",
			},
			excluded: []string{"style", "color: red", "alert('bad')", "script"},
		},
		{
			name:  "HTML entities unescaped",
			input: "<p>K8s &amp; Cloud &gt; Legacy &lt;systems&gt; &quot;High Scale&quot;</p>",
			expected: []string{
				"K8s & Cloud > Legacy <systems> \"High Scale\"",
			},
			excluded: []string{"&amp;", "&gt;", "&lt;", "&quot;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := StripHTML(tt.input)
			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("expected output to contain %q, but got:\n%s", exp, output)
				}
			}
			for _, exc := range tt.excluded {
				if strings.Contains(output, exc) {
					t.Errorf("expected output NOT to contain %q, but got:\n%s", exc, output)
				}
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Strips tracking params and trailing slash",
			input:    "https://boards.greenhouse.io/gitlab/jobs/12345/?utm_source=linkedin&utm_medium=jobboard&gh_src=abc123",
			expected: "https://boards.greenhouse.io/gitlab/jobs/12345",
		},
		{
			name:     "Preserves valid query params while sorting",
			input:    "https://jobs.lever.co/company/abc?page=2&filter=eng&utm_campaign=winter",
			expected: "https://jobs.lever.co/company/abc?filter=eng&page=2",
		},
		{
			name:     "Lowercases domain and removes fragment",
			input:    "HTTPS://JOBS.LEVER.CO/Company/123#apply-now",
			expected: "https://jobs.lever.co/Company/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeURL(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestComputeContentHash(t *testing.T) {
	hash1 := ComputeContentHash("Senior SRE", "GitLab", "Requires Kubernetes and Go.")
	hash2 := ComputeContentHash("Senior SRE", "GitLab", "Requires Kubernetes and Go.")
	hash3 := ComputeContentHash("Senior SRE", "GitLab", "Requires Kubernetes, Go, and Python.")

	if hash1 != hash2 {
		t.Errorf("expected hash1 == hash2, got %s and %s", hash1, hash2)
	}
	if hash1 == hash3 {
		t.Errorf("expected hash1 != hash3 when content changes")
	}
}
