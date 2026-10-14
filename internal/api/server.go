package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server handles HTTP API requests for job searches, bookmarks, and sources
type Server struct {
	store  store.Store
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewServer initializes the API routing table with CORS and health endpoints
func NewServer(s store.Store, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	srv := &Server{
		store:  s,
		logger: logger,
		mux:    http.NewServeMux(),
	}

	srv.routes()
	return srv
}

func (s *Server) routes() {
	// Health & Metrics
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.Handle("GET /metrics", promhttp.Handler())

	// Core API v1 endpoints
	s.mux.HandleFunc("GET /api/v1/jobs", s.handleListJobs)
	s.mux.HandleFunc("GET /api/v1/jobs/{id}", s.handleGetJobByID)
	s.mux.HandleFunc("POST /api/v1/jobs/{id}/save", s.handleSaveJob)
	s.mux.HandleFunc("DELETE /api/v1/jobs/{id}/save", s.handleUnsaveJob)
	s.mux.HandleFunc("GET /api/v1/saved", s.handleListSavedJobs)
	s.mux.HandleFunc("GET /api/v1/sources", s.handleListSources)
	s.mux.HandleFunc("GET /api/v1/stats", s.handleGetStats)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for frontend clients
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]any{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	var skills []string
	if sParam := q.Get("skills"); sParam != "" {
		for _, sk := range strings.Split(sParam, ",") {
			if trimmed := strings.TrimSpace(sk); trimmed != "" {
				skills = append(skills, trimmed)
			}
		}
	}

	var roles []string
	if rParam := q.Get("roles"); rParam != "" {
		for _, ro := range strings.Split(rParam, ",") {
			if trimmed := strings.TrimSpace(ro); trimmed != "" {
				roles = append(roles, trimmed)
			}
		}
	}

	filter := store.JobFilter{
		Query:      strings.TrimSpace(q.Get("q")),
		Skills:     skills,
		Roles:      roles,
		RemoteType: strings.TrimSpace(q.Get("remote")),
		Country:    strings.TrimSpace(q.Get("country")),
		Status:     "active",
		Limit:      limit,
		Offset:     offset,
	}

	jobs, total, err := s.store.ListJobs(r.Context(), filter)
	if err != nil {
		s.logger.Error("Failed to query jobs", "err", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to fetch jobs")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]any{
		"jobs":   jobs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) handleGetJobByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := s.store.GetJobByID(r.Context(), id)
	if err != nil {
		s.logger.Error("Failed to get job", "id", id, "err", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to retrieve job")
		return
	}
	if job == nil {
		s.respondError(w, http.StatusNotFound, "Job not found")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]any{"job": job})
}

type saveJobRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

func (s *Server) handleSaveJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	var req saveJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.store.SaveJob(r.Context(), id, req.Status, req.Notes); err != nil {
		s.logger.Error("Failed to save job", "id", id, "err", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to save job")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) handleUnsaveJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	if err := s.store.UnsaveJob(r.Context(), id); err != nil {
		s.logger.Error("Failed to unsave job", "id", id, "err", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to remove job from saved")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) handleListSavedJobs(w http.ResponseWriter, r *http.Request) {
	saved, err := s.store.GetSavedJobs(r.Context())
	if err != nil {
		s.logger.Error("Failed to list saved jobs", "err", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to retrieve saved jobs")
		return
	}

	if saved == nil {
		saved = []model.SavedJobWithJob{}
	}

	s.respondJSON(w, http.StatusOK, map[string]any{
		"saved_jobs": saved,
		"count":      len(saved),
	})
}

func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := s.store.GetActiveSources(r.Context())
	if err != nil {
		s.logger.Error("Failed to list sources", "err", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to retrieve sources")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]any{
		"sources": sources,
		"count":   len(sources),
	})
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.GetDashboardStats(r.Context())
	if err != nil {
		s.logger.Error("Failed to get dashboard stats", "err", err)
		s.respondError(w, http.StatusInternalServerError, "Failed to load dashboard metrics")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]any{"stats": stats})
}

func (s *Server) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, map[string]string{"error": message})
}
