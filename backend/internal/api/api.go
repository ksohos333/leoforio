// Package api exposes the leoforio HTTP API.
//
// The API serves only cached, normalised data. It never calls an upstream
// directly: every request goes through a cache in front of a source.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ksohos333/leoforio/backend/internal/cache"
	"github.com/ksohos333/leoforio/backend/internal/model"
	"github.com/ksohos333/leoforio/backend/internal/source"
	"github.com/ksohos333/leoforio/backend/internal/store"
)

// StopFinder searches the locally imported stop index.
type StopFinder interface {
	SearchStops(ctx context.Context, query string, limit int) ([]store.Stop, error)
	NearbyStops(ctx context.Context, latitude, longitude, radiusKm float64, limit int) ([]store.Stop, error)
	StopByID(ctx context.Context, id string) (store.Stop, error)
}

// Cache lifetimes per resource.
const (
	ArrivalsTTL      = 25 * time.Second
	PositionsTTL     = 15 * time.Second
	NetworkStaticTTL = 6 * time.Hour
)

// Dependencies are the sources and stores the API serves from.
type Dependencies struct {
	Arrivals  source.ArrivalsSource
	Positions source.PositionsSource
	Network   source.NetworkSource
	Stops     StopFinder
	Version   string
}

// Server serves the API.
type Server struct {
	arrivals  source.ArrivalsSource
	positions source.PositionsSource
	network   source.NetworkSource
	stops     StopFinder
	version   string
	startedAt time.Time

	arrivalsCache  *cache.Cache[[]model.Arrival]
	positionsCache *cache.Cache[[]model.VehiclePosition]
	linesCache     *cache.Cache[[]model.Line]
	routesCache    *cache.Cache[[]model.Route]
	stopsCache     *cache.Cache[[]model.Stop]
	geometryCache  *cache.Cache[[]model.Position]
}

// New returns a Server. Any dependency may be nil, in which case the
// endpoints that depend on it report that the source is not configured.
func New(dependencies Dependencies) *Server {
	return &Server{
		arrivals:       dependencies.Arrivals,
		positions:      dependencies.Positions,
		network:        dependencies.Network,
		stops:          dependencies.Stops,
		version:        dependencies.Version,
		startedAt:      time.Now(),
		arrivalsCache:  cache.New[[]model.Arrival](ArrivalsTTL),
		positionsCache: cache.New[[]model.VehiclePosition](PositionsTTL),
		linesCache:     cache.New[[]model.Line](NetworkStaticTTL),
		routesCache:    cache.New[[]model.Route](NetworkStaticTTL),
		stopsCache:     cache.New[[]model.Stop](NetworkStaticTTL),
		geometryCache:  cache.New[[]model.Position](NetworkStaticTTL),
	}
}

// Routes returns the HTTP handler for the API.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/v1/meta", s.handleMeta)
	mux.HandleFunc("GET /api/v1/lines", s.handleLines)
	mux.HandleFunc("GET /api/v1/routes", s.handleRoutes)
	mux.HandleFunc("GET /api/v1/routes/{routeCode}/vehicles", s.handleVehicles)
	mux.HandleFunc("GET /api/v1/routes/{routeCode}/stops", s.handleRouteStops)
	mux.HandleFunc("GET /api/v1/routes/{routeCode}/geometry", s.handleRouteGeometry)
	mux.HandleFunc("GET /api/v1/stops/search", s.handleStopSearch)
	mux.HandleFunc("GET /api/v1/stops/nearby", s.handleStopsNearby)
	mux.HandleFunc("GET /api/v1/stops/{stopCode}", s.handleStop)
	mux.HandleFunc("GET /api/v1/stops/{stopCode}/arrivals", s.handleArrivals)
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
		"name":            "leoforio",
		"apiVersion":      "v1",
		"version":         s.version,
		"dataAttribution": "OASA",
		"unofficial":      true,
	})
}

func (s *Server) handleLines(w http.ResponseWriter, r *http.Request) {
	if s.network == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	lines, stale, err := s.linesCache.Get(r.Context(), "lines", func(ctx context.Context) ([]model.Line, error) {
		return s.network.Lines(ctx)
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if stale {
		w.Header().Set("X-Cache", "stale")
	}
	writeJSON(w, http.StatusOK, lines)
}

func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	if s.network == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	lineCode := r.URL.Query().Get("line")
	if lineCode == "" {
		writeError(w, http.StatusBadRequest, "the line query parameter is required")
		return
	}

	routes, stale, err := s.routesCache.Get(r.Context(), "routes:"+lineCode, func(ctx context.Context) ([]model.Route, error) {
		return s.network.Routes(ctx, lineCode)
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if stale {
		w.Header().Set("X-Cache", "stale")
	}
	writeJSON(w, http.StatusOK, routes)
}

func (s *Server) handleVehicles(w http.ResponseWriter, r *http.Request) {
	if s.positions == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	routeCode := r.PathValue("routeCode")
	positions, stale, err := s.positionsCache.Get(r.Context(), "positions:"+routeCode, func(ctx context.Context) ([]model.VehiclePosition, error) {
		return s.positions.Positions(ctx, routeCode)
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

func (s *Server) handleRouteStops(w http.ResponseWriter, r *http.Request) {
	if s.network == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	routeCode := r.PathValue("routeCode")
	stops, stale, err := s.stopsCache.Get(r.Context(), "stops:"+routeCode, func(ctx context.Context) ([]model.Stop, error) {
		return s.network.RouteStops(ctx, routeCode)
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if stale {
		w.Header().Set("X-Cache", "stale")
	}
	writeJSON(w, http.StatusOK, stops)
}

func (s *Server) handleRouteGeometry(w http.ResponseWriter, r *http.Request) {
	if s.network == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	routeCode := r.PathValue("routeCode")
	geometry, stale, err := s.geometryCache.Get(r.Context(), "geometry:"+routeCode, func(ctx context.Context) ([]model.Position, error) {
		return s.network.RouteGeometry(ctx, routeCode)
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if stale {
		w.Header().Set("X-Cache", "stale")
	}
	writeJSON(w, http.StatusOK, geoJSON(geometry))
}

func (s *Server) handleStopSearch(w http.ResponseWriter, r *http.Request) {
	if s.stops == nil {
		writeError(w, http.StatusServiceUnavailable, "the stop index is not available")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "the q query parameter is required")
		return
	}

	limit := queryLimit(r, 20)
	stops, err := s.stops.SearchStops(r.Context(), query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stops)
}

func (s *Server) handleStopsNearby(w http.ResponseWriter, r *http.Request) {
	if s.stops == nil {
		writeError(w, http.StatusServiceUnavailable, "the stop index is not available")
		return
	}

	latitude, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "the lat query parameter must be a number")
		return
	}
	longitude, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "the lon query parameter must be a number")
		return
	}

	radius := 1.0
	if value, err := strconv.ParseFloat(r.URL.Query().Get("radius"), 64); err == nil && value > 0 {
		radius = value
	}

	stops, err := s.stops.NearbyStops(r.Context(), latitude, longitude, radius, queryLimit(r, 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stops)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if s.stops == nil {
		writeError(w, http.StatusServiceUnavailable, "the stop index is not available")
		return
	}

	stop, err := s.stops.StopByID(r.Context(), r.PathValue("stopCode"))
	if err != nil {
		writeError(w, http.StatusNotFound, "no such stop")
		return
	}
	writeJSON(w, http.StatusOK, stop)
}

// queryLimit reads a result limit from the query string.
func queryLimit(r *http.Request, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || value <= 0 {
		return fallback
	}
	if value > 100 {
		return 100
	}
	return value
}

func (s *Server) handleArrivals(w http.ResponseWriter, r *http.Request) {
	if s.arrivals == nil {
		writeError(w, http.StatusServiceUnavailable, source.ErrNotConfigured.Error())
		return
	}

	stopCode := r.PathValue("stopCode")
	arrivals, stale, err := s.arrivalsCache.Get(r.Context(), "arrivals:"+stopCode, func(ctx context.Context) ([]model.Arrival, error) {
		return s.arrivals.Arrivals(ctx, stopCode)
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

// geoJSON renders a route path as a GeoJSON LineString, which is what the
// map renders.
func geoJSON(geometry []model.Position) map[string]any {
	coordinates := make([][]float64, 0, len(geometry))
	for _, point := range geometry {
		coordinates = append(coordinates, []float64{point.Longitude, point.Latitude})
	}

	return map[string]any{
		"type": "Feature",
		"geometry": map[string]any{
			"type":        "LineString",
			"coordinates": coordinates,
		},
		"properties": map[string]any{},
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}
