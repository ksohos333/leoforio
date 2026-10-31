package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ksohos333/leoforio/backend/internal/model"
)

type stubArrivals struct {
	calls int
	err   error
}

func (s *stubArrivals) Arrivals(_ context.Context, stopID string) ([]model.Arrival, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return []model.Arrival{{
		StopID:   stopID,
		LineID:   "815",
		RouteID:  "815-A",
		Expected: time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
	}}, nil
}

type stubPositions struct {
	calls int
}

func (s *stubPositions) Positions(context.Context) ([]model.VehiclePosition, error) {
	s.calls++
	return []model.VehiclePosition{{
		VehicleID: "v1",
		LineID:    "815",
		Position:  model.Position{Latitude: 37.98, Longitude: 23.72},
	}}, nil
}

func TestHealth(t *testing.T) {
	srv := New(nil, nil, "test")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestArrivalsWithoutSource(t *testing.T) {
	srv := New(nil, nil, "test")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/stop-1/arrivals", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no source is configured, got %d", rec.Code)
	}
}

func TestArrivalsAreCached(t *testing.T) {
	stub := &stubArrivals{}
	srv := New(stub, nil, "test")

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/stop-1/arrivals", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	if stub.calls != 1 {
		t.Errorf("expected 1 upstream call for 3 requests, got %d", stub.calls)
	}
}

func TestVehiclesWithoutSource(t *testing.T) {
	srv := New(nil, nil, "test")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no source is configured, got %d", rec.Code)
	}
}

func TestVehiclesAreCached(t *testing.T) {
	stub := &stubPositions{}
	srv := New(nil, stub, "test")

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	if stub.calls != 1 {
		t.Errorf("expected 1 upstream call for 3 requests, got %d", stub.calls)
	}
}

func TestUpstreamFailureIsBadGateway(t *testing.T) {
	stub := &stubArrivals{err: errors.New("upstream down")}
	srv := New(stub, nil, "test")

	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/stop-1/arrivals", nil))

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 for a failing upstream, got %d", rec.Code)
	}
}
