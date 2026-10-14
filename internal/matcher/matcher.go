package matcher

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bikrams/getfetch/internal/model"
)

type skillRule struct {
	Name    string
	Pattern *regexp.Regexp
}

type roleRule struct {
	Name    string
	Pattern *regexp.Regexp
}

var (
	// Primary target keywords (Kubernetes / K8s)
	k8sRegex = regexp.MustCompile(`(?i)\b(kubernetes|k8s)\b`)

	// Skill definitions with case-insensitive word-boundary patterns
	skillRules = []skillRule{
		{Name: "Kubernetes", Pattern: regexp.MustCompile(`(?i)\b(kubernetes|k8s)\b`)},
		{Name: "Helm", Pattern: regexp.MustCompile(`(?i)\bhelm\b`)},
		{Name: "Terraform", Pattern: regexp.MustCompile(`(?i)\b(terraform|opentofu)\b`)},
		{Name: "Docker", Pattern: regexp.MustCompile(`(?i)\b(docker|containerd|podman)\b`)},
		{Name: "ArgoCD", Pattern: regexp.MustCompile(`(?i)\bargo\s*cd\b`)},
		{Name: "Flux", Pattern: regexp.MustCompile(`(?i)\bflux\s*cd\b`)},
		{Name: "Istio", Pattern: regexp.MustCompile(`(?i)\bistio\b`)},
		{Name: "Envoy", Pattern: regexp.MustCompile(`(?i)\benvoy\b`)},
		{Name: "Prometheus", Pattern: regexp.MustCompile(`(?i)\bprometheus\b`)},
		{Name: "Grafana", Pattern: regexp.MustCompile(`(?i)\bgrafana\b`)},
		{Name: "OpenTelemetry", Pattern: regexp.MustCompile(`(?i)\b(opentelemetry|otel)\b`)},
		{Name: "Cilium", Pattern: regexp.MustCompile(`(?i)\bcilium\b`)},
		{Name: "eBPF", Pattern: regexp.MustCompile(`(?i)\be?bpf\b`)},
		{Name: "Ansible", Pattern: regexp.MustCompile(`(?i)\bansible\b`)},
		{Name: "Linux", Pattern: regexp.MustCompile(`(?i)\blinux\b`)},
		{Name: "GKE", Pattern: regexp.MustCompile(`(?i)\b(gke|google\s*kubernetes\s*engine)\b`)},
		{Name: "EKS", Pattern: regexp.MustCompile(`(?i)\b(eks|elastic\s*kubernetes\s*service)\b`)},
		{Name: "AKS", Pattern: regexp.MustCompile(`(?i)\b(aks|azure\s*kubernetes\s*service)\b`)},
		{Name: "GCP", Pattern: regexp.MustCompile(`(?i)\b(gcp|google\s*cloud(\s*platform)?)\b`)},
		{Name: "AWS", Pattern: regexp.MustCompile(`(?i)\b(aws|amazon\s*web\s*services)\b`)},
		{Name: "Azure", Pattern: regexp.MustCompile(`(?i)\b(azure|microsoft\s*azure)\b`)},
		{Name: "Go", Pattern: regexp.MustCompile(`(?i)(\bgolang\b|\bGo\s*(developer|engineer|programming|code)?\b)`)},
		{Name: "Python", Pattern: regexp.MustCompile(`(?i)\bpython\b`)},
		{Name: "Rust", Pattern: regexp.MustCompile(`(?i)\brust\b`)},
		{Name: "Bash", Pattern: regexp.MustCompile(`(?i)\b(bash|shell\s*scripting)\b`)},
	}

	// Role titles
	roleRules = []roleRule{
		{Name: "Site Reliability Engineer", Pattern: regexp.MustCompile(`(?i)\b(sre|site\s*reliability\s*engineer)\b`)},
		{Name: "Platform Engineer", Pattern: regexp.MustCompile(`(?i)\bplatform\s*engineer\b`)},
		{Name: "DevOps Engineer", Pattern: regexp.MustCompile(`(?i)\bdevops(\s*engineer)?\b`)},
		{Name: "Infrastructure Engineer", Pattern: regexp.MustCompile(`(?i)\binfrastructure\s*engineer\b`)},
		{Name: "Cloud Engineer", Pattern: regexp.MustCompile(`(?i)\bcloud(\s+infrastructure)?\s*engineer\b`)},
		{Name: "Kubernetes Engineer", Pattern: regexp.MustCompile(`(?i)\bkubernetes\s*(engineer|admin(istrator)?)\b`)},
		{Name: "Production Engineer", Pattern: regexp.MustCompile(`(?i)\bproduction\s*engineer\b`)},
		{Name: "Systems Engineer", Pattern: regexp.MustCompile(`(?i)\bsystems\s*engineer\b`)},
	}

	// Work mode patterns
	remoteRegex = regexp.MustCompile(`(?i)\b(remote|work\s*from\s*home|wfh|anywhere|distributed)\b`)
	hybridRegex = regexp.MustCompile(`(?i)\bhybrid\b`)
	onsiteRegex = regexp.MustCompile(`(?i)\b(on-?site|in-?office)\b`)

	// India regions
	indiaRegex = regexp.MustCompile(`(?i)\b(india|bengaluru|bangalore|hyderabad|pune|gurgaon|gurugram|noida|delhi|mumbai|chennai|in)\b`)
)

// Matcher analyzes raw job titles and descriptions to evaluate eligibility and extract metadata
type Matcher struct{}

// NewMatcher creates an initialized Matcher instance
func NewMatcher() *Matcher {
	return &Matcher{}
}

// Evaluate evaluates a job against target skills, roles, and location rules
func (m *Matcher) Evaluate(title, location, descText string, adapterRemote string) model.MatchResult {
	combined := title + " \n " + descText
	res := model.MatchResult{
		MatchedSkills: []string{},
		MatchedRoles:  []string{},
		RemoteType:    model.RemoteTypeUnknown,
	}

	// 1. Detect Skills
	hasK8s := false
	for _, rule := range skillRules {
		if rule.Pattern.MatchString(combined) {
			res.MatchedSkills = append(res.MatchedSkills, rule.Name)
			if rule.Name == "Kubernetes" {
				hasK8s = true
			}
		}
	}

	// 2. Detect Roles
	isInfraRole := false
	for _, rule := range roleRules {
		if rule.Pattern.MatchString(title) || rule.Pattern.MatchString(combined) {
			res.MatchedRoles = append(res.MatchedRoles, rule.Name)
			isInfraRole = true
		}
	}

	// 3. Determine if target match
	// Qualifies if:
	// - Contains Kubernetes / K8s anywhere
	// - OR title is an infra/SRE/Platform role AND mentions Terraform/Docker/Helm/GKE/EKS/AWS/GCP
	if hasK8s {
		res.IsTargetMatch = true
	} else if isInfraRole && len(res.MatchedSkills) >= 2 {
		res.IsTargetMatch = true
	}

	// 4. Extract snippet showing why it matched
	if res.IsTargetMatch {
		res.SnippetMatch = m.extractSnippet(descText, title)
	}

	// 5. Detect Remote / Work mode
	res.RemoteType = m.detectRemoteType(adapterRemote, title, location, descText)

	return res
}

// detectRemoteType infers the remote classification from location and text
func (m *Matcher) detectRemoteType(adapterRemote, title, location, desc string) string {
	lowerAdapter := strings.ToLower(adapterRemote)
	if lowerAdapter == "remote" || lowerAdapter == "true" {
		return model.RemoteTypeRemote
	}
	if lowerAdapter == "hybrid" {
		return model.RemoteTypeHybrid
	}
	if lowerAdapter == "onsite" || lowerAdapter == "on-site" {
		return model.RemoteTypeOnsite
	}

	combinedLoc := title + " " + location
	if remoteRegex.MatchString(combinedLoc) {
		return model.RemoteTypeRemote
	}
	if hybridRegex.MatchString(combinedLoc) {
		return model.RemoteTypeHybrid
	}
	if onsiteRegex.MatchString(combinedLoc) {
		return model.RemoteTypeOnsite
	}

	// Check description if not found in location
	if remoteRegex.MatchString(desc) {
		return model.RemoteTypeRemote
	}

	return model.RemoteTypeUnknown
}

// IsIndiaLocation checks if the location or title refers to India
func (m *Matcher) IsIndiaLocation(location, country string) bool {
	if strings.EqualFold(country, "India") || strings.EqualFold(country, "IN") {
		return true
	}
	return indiaRegex.MatchString(location)
}

// extractSnippet captures a concise window around the matching keyword
func (m *Matcher) extractSnippet(desc, title string) string {
	loc := k8sRegex.FindStringIndex(desc)
	if loc == nil {
		// Fallback to SRE or Platform
		re := regexp.MustCompile(`(?i)\b(sre|platform|reliability|terraform|infrastructure)\b`)
		loc = re.FindStringIndex(desc)
	}

	if loc == nil {
		// If not found in description, snippet from title
		return fmt.Sprintf("Matched role: %s", title)
	}

	matchStart := loc[0]
	matchEnd := loc[1]

	// Look backwards ~150 chars
	start := matchStart - 150
	if start < 0 {
		start = 0
	} else {
		// find nearest space
		if idx := strings.Index(desc[start:matchStart], " "); idx != -1 {
			start = start + idx + 1
		}
	}

	// Look forward ~200 chars
	end := matchEnd + 200
	if end > len(desc) {
		end = len(desc)
	} else {
		// find nearest space
		if idx := strings.LastIndex(desc[matchEnd:end], " "); idx != -1 {
			end = matchEnd + idx
		}
	}

	snippet := strings.TrimSpace(desc[start:end])
	// clean up line breaks
	snippet = regexp.MustCompile(`\s+`).ReplaceAllString(snippet, " ")

	prefix := "..."
	suffix := "..."
	if start == 0 {
		prefix = ""
	}
	if end == len(desc) {
		suffix = ""
	}

	return prefix + snippet + suffix
}
