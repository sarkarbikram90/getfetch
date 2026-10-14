package adapter

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bikrams/getfetch/internal/model"
)

// DefaultHTTPClient provides a preconfigured client with safe timeouts and connection reuse
var DefaultHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

// UserAgent sent with crawler requests to identify the bot politely
const UserAgent = "getfetch-bot/1.0 (+https://github.com/bikrams/getfetch)"

// Adapter defines the common contract implemented by all ATS connectors
type Adapter interface {
	Name() string
	FetchJobs(ctx context.Context, source model.Source) ([]model.RawJob, error)
}

// Registry stores and resolves adapters by source type
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRegistry instantiates an empty Adapter Registry
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds an adapter to the registry
func (r *Registry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// Get retrieves an adapter by its source_type identifier
func (r *Registry) Get(name string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ad, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("unsupported adapter type: %q", name)
	}
	return ad, nil
}

// List returns all registered adapter names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}
