// Package api exposes the leoforio HTTP API.
//
// The API serves only cached, normalised data. It never calls an upstream
// directly: every request goes through a cache in front of a source.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ksohos333/leoforio/backend/internal/cache"
	"github.com/ksohos333/leoforio/backend/internal/model"
	"github.com/ksohos333/leoforio/backend/internal/source"
)

// ArrivalsTTL is how long an arrivals response is reused.
const ArrivalsTTL = 25 * time.Second

// PositionsTTL is how long a vehicle positions response is reused.
const PositionsTTL = 15 * time.Second

// Server serves the API.
type Server struct {
	arrivals       source.ArrivalsSource
	positions      source.PositionsSource
	arrivalsCache  *cache.Cache[[]model.Arrival]
	positionsCache *cache.Cache[[]model.VehiclePosition]
	version        string
	startedAt      time.Time
}

// New returns a Server. Sources may be nil, in which case the endpoints that
// depend on them report that the source is not configured.
func New(arrivals source.ArrivalsSource, positions source.PositionsSource, version string) *Server {
	return &Server{
		arrivals:       arrivals,
		positions:      positions,
		arrivalsCache:  cache.New[[]model.Arrival](ArrivalsTTL),
		positionsCache: cache.New[[]model.VehiclePosition](PositionsTTL),
		version:        version,
		startedAt:      time.Now(),
	}
}

// Routes returns the HTTP handler for the API.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/v1/meta", s.handleMeta)
	mux.HandleFunc("GET /api/v1/stops/{stopID}/arrivals", s.handleArrivals)
	mux.HandleFunc("GET /api/v1/vehicles", s.handleVehicles)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"uptime":  time.Since(s.startedAt).Round(time.Second).String(),
		"version": s.version,
	})
}

func (s *Server) handleMeta(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":       "leoforio",
		"apiVersion": "v1",
		"version":    s.version,
	})
}

func (s *Server) handleArrivals(w http.ResponseWriter, r *http.Request) {
	if s.arrivals == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	stopID := r.PathValue("stopID")
	arrivals, stale, err := s.arrivalsCache.Get(r.Context(), "arrivals:"+stopID, func(ctx context.Context) ([]model.Arrival, error) {
		return s.arrivals.Arrivals(ctx, stopID)
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if stale {
		w.Header().Set("X-Cache", "stale")
	}
	writeJSON(w, http.StatusOK, arrivals)
}

func (s *Server) handleVehicles(w http.ResponseWriter, r *http.Request) {
	if s.positions == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	positions, stale, err := s.positionsCache.Get(r.Context(), "positions", func(ctx context.Context) ([]model.VehiclePosition, error) {
		return s.positions.Positions(ctx)
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if stale {
		w.Header().Set("X-Cache", "stale")
	}
	writeJSON(w, http.StatusOK, positions)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}
