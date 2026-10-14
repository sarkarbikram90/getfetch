package notify

import (
	"context"
	"time"

	"github.com/bikrams/getfetch/internal/model"
)

// DigestPayload contains the batch of discovered and updated jobs to alert on
type DigestPayload struct {
	Recipient   string      `json:"recipient"`
	GeneratedAt time.Time   `json:"generated_at"`
	NewJobs     []model.Job `json:"new_jobs"`
	UpdatedJobs []model.Job `json:"updated_jobs"`
}

// Notifier defines the interface for alert dispatchers
type Notifier interface {
	Name() string
	SendDigest(ctx context.Context, payload DigestPayload) error
}
