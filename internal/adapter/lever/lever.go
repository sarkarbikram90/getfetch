package lever

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/httputil"
	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/normalizer"
)

type Adapter struct {
	client  adapter.HTTPDoer
	baseURL string
}

// NewAdapter creates an initialized Lever Adapter
func NewAdapter(client adapter.HTTPDoer) *Adapter {
	if client == nil {
		client = httputil.NewClient()
	}
	return &Adapter{
		client:  client,
		baseURL: "https://api.lever.co/v0/postings",
	}
}

func (a *Adapter) Name() string {
	return model.SourceTypeLever
}

type leverPosting struct {
	ID               string          `json:"id"`
	Text             string          `json:"text"`
	CreatedAt        int64           `json:"createdAt"`
	HostedURL        string          `json:"hostedUrl"`
	ApplyURL         string          `json:"applyUrl"`
	Categories       leverCategories `json:"categories"`
	Description      string          `json:"description"`
	DescriptionPlain string          `json:"descriptionPlain"`
	WorkplaceType    string          `json:"workplaceType"` // "remote", "hybrid", "on-site"
	Country          string          `json:"country"`
}

type leverCategories struct {
	Commitment string `json:"commitment"`
	Department string `json:"department"`
	Location   string `json:"location"`
	Team       string `json:"team"`
}

// FetchJobs retrieves published postings for a given Lever board
func (a *Adapter) FetchJobs(ctx context.Context, source model.Source) ([]model.RawJob, error) {
	if source.BoardIdentifier == "" {
		return nil, fmt.Errorf("missing board identifier for source %s", source.CompanyName)
	}

	url := fmt.Sprintf("%s/%s?mode=json", a.baseURL, source.BoardIdentifier)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", adapter.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to lever failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lever board %q not found (404)", source.BoardIdentifier)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("lever returned status %d: %s", resp.StatusCode, string(body))
	}

	var postings []leverPosting
	if err := json.NewDecoder(resp.Body).Decode(&postings); err != nil {
		return nil, fmt.Errorf("failed to decode lever postings: %w", err)
	}

	rawJobs := make([]model.RawJob, 0, len(postings))
	for _, p := range postings {
		var postedAt *time.Time
		if p.CreatedAt > 0 {
			t := time.UnixMilli(p.CreatedAt)
			postedAt = &t
		}

		dept := p.Categories.Department
		if dept == "" {
			dept = p.Categories.Team
		}

		desc := p.Description
		if desc == "" {
			desc = p.DescriptionPlain
		}

		rawJobs = append(rawJobs, model.RawJob{
			ExternalID:     p.ID,
			CanonicalURL:   normalizer.NormalizeURL(p.HostedURL),
			CompanyName:    source.CompanyName,
			Title:          p.Text,
			Location:       p.Categories.Location,
			RemoteType:     p.WorkplaceType,
			Country:        p.Country,
			Department:     dept,
			DescriptionRaw: desc,
			PostedAt:       postedAt,
			SourcePayload: map[string]any{
				"commitment": p.Categories.Commitment,
				"team":       p.Categories.Team,
				"apply_url":  p.ApplyURL,
			},
		})
	}

	return rawJobs, nil
}
