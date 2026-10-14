package ashby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/normalizer"
)

type Adapter struct {
	client  *http.Client
	baseURL string
}

// NewAdapter creates an initialized Ashby Adapter
func NewAdapter(client *http.Client) *Adapter {
	if client == nil {
		client = adapter.DefaultHTTPClient
	}
	return &Adapter{
		client:  client,
		baseURL: "https://api.ashbyhq.com/posting-api/job-board",
	}
}

func (a *Adapter) Name() string {
	return model.SourceTypeAshby
}

type ashbyResponse struct {
	APIVersion string     `json:"apiVersion"`
	Jobs       []ashbyJob `json:"jobs"`
}

type ashbyJob struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Department       string `json:"department"`
	Location         string `json:"location"`
	IsRemote         bool   `json:"isRemote"`
	JobURL           string `json:"jobUrl"`
	ApplyURL         string `json:"applyUrl"`
	PublishedAt      string `json:"publishedAt"`
	DescriptionHTML  string `json:"descriptionHtml"`
	DescriptionPlain string `json:"descriptionPlain"`
}

// FetchJobs retrieves published jobs from an Ashby job board
func (a *Adapter) FetchJobs(ctx context.Context, source model.Source) ([]model.RawJob, error) {
	if source.BoardIdentifier == "" {
		return nil, fmt.Errorf("missing board identifier for source %s", source.CompanyName)
	}

	url := fmt.Sprintf("%s/%s", a.baseURL, source.BoardIdentifier)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", adapter.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to ashby failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("ashby board %q not found (404)", source.BoardIdentifier)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("ashby returned status %d: %s", resp.StatusCode, string(body))
	}

	var ashbyResp ashbyResponse
	if err := json.NewDecoder(resp.Body).Decode(&ashbyResp); err != nil {
		return nil, fmt.Errorf("failed to decode ashby response: %w", err)
	}

	rawJobs := make([]model.RawJob, 0, len(ashbyResp.Jobs))
	for _, j := range ashbyResp.Jobs {
		var postedAt *time.Time
		if j.PublishedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, j.PublishedAt); err == nil {
				postedAt = &parsed
			}
		}

		remoteType := model.RemoteTypeUnknown
		if j.IsRemote {
			remoteType = model.RemoteTypeRemote
		}

		desc := j.DescriptionHTML
		if desc == "" {
			desc = j.DescriptionPlain
		}

		rawJobs = append(rawJobs, model.RawJob{
			ExternalID:     j.ID,
			CanonicalURL:   normalizer.NormalizeURL(j.JobURL),
			CompanyName:    source.CompanyName,
			Title:          j.Title,
			Location:       j.Location,
			RemoteType:     remoteType,
			Department:     j.Department,
			DescriptionRaw: desc,
			PostedAt:       postedAt,
			SourcePayload: map[string]any{
				"is_remote":  j.IsRemote,
				"apply_url":  j.ApplyURL,
				"department": j.Department,
			},
		})
	}

	return rawJobs, nil
}
