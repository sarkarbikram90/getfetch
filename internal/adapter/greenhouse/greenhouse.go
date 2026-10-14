package greenhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/normalizer"
)

type Adapter struct {
	client  *http.Client
	baseURL string
}

// NewAdapter creates an initialized Greenhouse Adapter
func NewAdapter(client *http.Client) *Adapter {
	if client == nil {
		client = adapter.DefaultHTTPClient
	}
	return &Adapter{
		client:  client,
		baseURL: "https://boards-api.greenhouse.io/v1/boards",
	}
}

func (a *Adapter) Name() string {
	return model.SourceTypeGreenhouse
}

type ghResponse struct {
	Jobs []ghJob `json:"jobs"`
}

type ghJob struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Location    ghLocation     `json:"location"`
	Content     string         `json:"content"`
	AbsoluteURL string         `json:"absolute_url"`
	UpdatedAt   string         `json:"updated_at"`
	Offices     []ghNamedItem  `json:"offices"`
	Departments []ghNamedItem  `json:"departments"`
	Metadata    []ghCustomMeta `json:"metadata"`
}

type ghLocation struct {
	Name string `json:"name"`
}

type ghNamedItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ghCustomMeta struct {
	Name      string `json:"name"`
	Value     any    `json:"value"`
	ValueType string `json:"value_type"`
}

// FetchJobs retrieves published jobs with descriptions for a given board identifier
func (a *Adapter) FetchJobs(ctx context.Context, source model.Source) ([]model.RawJob, error) {
	if source.BoardIdentifier == "" {
		return nil, fmt.Errorf("missing board identifier for source %s", source.CompanyName)
	}

	url := fmt.Sprintf("%s/%s/jobs?content=true", a.baseURL, source.BoardIdentifier)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", adapter.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to greenhouse failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("greenhouse board %q not found (404)", source.BoardIdentifier)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("greenhouse returned status %d: %s", resp.StatusCode, string(body))
	}

	var ghResp ghResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return nil, fmt.Errorf("failed to decode greenhouse response: %w", err)
	}

	rawJobs := make([]model.RawJob, 0, len(ghResp.Jobs))
	for _, j := range ghResp.Jobs {
		var postedAt *time.Time
		if j.UpdatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, j.UpdatedAt); err == nil {
				postedAt = &parsed
			}
		}

		dept := ""
		if len(j.Departments) > 0 {
			dept = j.Departments[0].Name
		}

		rawJobs = append(rawJobs, model.RawJob{
			ExternalID:     strconv.FormatInt(j.ID, 10),
			CanonicalURL:   normalizer.NormalizeURL(j.AbsoluteURL),
			CompanyName:    source.CompanyName,
			Title:          j.Title,
			Location:       j.Location.Name,
			Department:     dept,
			DescriptionRaw: j.Content,
			PostedAt:       postedAt,
			SourcePayload: map[string]any{
				"offices":     j.Offices,
				"departments": j.Departments,
			},
		})
	}

	return rawJobs, nil
}
