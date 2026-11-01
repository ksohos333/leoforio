package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ksohos333/leoforio/backend/internal/model"
	"github.com/ksohos333/leoforio/backend/internal/store"
)

type stubArrivals struct {
	calls int
	err   error
}

func (s *stubArrivals) Arrivals(_ context.Context, stopCode string) ([]model.Arrival, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return []model.Arrival{{
		StopID:   stopCode,
		RouteID:  "2484",
		Expected: time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
	}}, nil
}

type stubPositions struct {
	calls int
}

func (s *stubPositions) Positions(_ context.Context, routeCode string) ([]model.VehiclePosition, error) {
	s.calls++
	return []model.VehiclePosition{{
		VehicleID: "v1",
		RouteID:   routeCode,
		Position:  model.Position{Latitude: 37.98, Longitude: 23.72},
	}}, nil
}

type stubNetwork struct {
	calls int
}

func (s *stubNetwork) Lines(context.Context) ([]model.Line, error) {
	s.calls++
	return []model.Line{{ID: "021", Code: "021", Name: "ΠΛΑΤΕΙΑ ΚΑΝΙΓΓΟΣ"}}, nil
}

func (s *stubNetwork) Routes(_ context.Context, lineCode string) ([]model.Route, error) {
	s.calls++
	return []model.Route{{ID: "2484", LineID: lineCode, Name: "PLATEIA KANIGKOS"}}, nil
}

func (s *stubNetwork) RoutesForStop(_ context.Context, stopCode string) ([]model.Route, error) {
	s.calls++
	return []model.Route{
		{
			ID:         "2484",
			LineID:     "1151",
			LineNumber: "021",
			Direction:  "ΠΛΑΤΕΙΑ ΚΑΝΙΓΓΟΣ - ΓΚΥΖΗ",
			LineName:   "ΠΛΑΤΕΙΑ ΚΑΝΙΓΓΟΣ - ΓΚΥΖH (ΚΥΚΛΙΚΗ)",
		},
	}, nil
}

func (s *stubNetwork) RouteStops(_ context.Context, routeCode string) ([]model.Stop, error) {
	s.calls++
	return []model.Stop{{ID: "10175", Code: "18", Name: "ΠΛ. ΚΑΝΙΓΓΟΣ"}}, nil
}

func (s *stubNetwork) RouteGeometry(_ context.Context, _ string) ([]model.Position, error) {
	s.calls++
	return []model.Position{
		{Latitude: 37.98545, Longitude: 23.73153},
		{Latitude: 37.98522, Longitude: 23.73147},
	}, nil
}

type stubStops struct {
	calls int
}

func (s *stubStops) SearchStops(_ context.Context, query string, _ int) ([]store.Stop, error) {
	s.calls++
	if query == "" {
		return nil, nil
	}
	return []store.Stop{{ID: "10175", Code: "18", Name: "ΠΛ. ΚΑΝΙΓΓΟΣ", Latitude: 37.98566, Longitude: 23.73113}}, nil
}

func (s *stubStops) NearbyStops(_ context.Context, _, _, _ float64, _ int) ([]store.Stop, error) {
	s.calls++
	return []store.Stop{{ID: "60718", Name: "ΣΥΝΤΑΓΜΑ", Distance: 0.023}}, nil
}

func (s *stubStops) StopByID(_ context.Context, id string) (store.Stop, error) {
	s.calls++
	if id != "10175" {
		return store.Stop{}, errors.New("no such stop")
	}
	return store.Stop{ID: "10175", Name: "ΠΛ. ΚΑΝΙΓΓΟΣ"}, nil
}

func TestStopSearchRequiresAQuery(t *testing.T) {
	srv := New(Dependencies{Stops: &stubStops{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/search", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without a query, got %d", rec.Code)
	}
}

func TestStopSearchIsServed(t *testing.T) {
	srv := New(Dependencies{Stops: &stubStops{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/search?q=kaniggos", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !contains(body, "10175") {
		t.Errorf("expected the stop in the response, got %s", body)
	}
}

func TestNearbyStopsAreServed(t *testing.T) {
	srv := New(Dependencies{Stops: &stubStops{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/nearby?lat=37.97&lon=23.73", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !contains(body, "60718") {
		t.Errorf("expected the nearby stop in the response, got %s", body)
	}
}

func TestNearbyStopsRejectBadCoordinates(t *testing.T) {
	srv := New(Dependencies{Stops: &stubStops{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/nearby?lat=north", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad coordinates, got %d", rec.Code)
	}
}

func TestStopDetailsAreServed(t *testing.T) {
	srv := New(Dependencies{Stops: &stubStops{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/10175", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/does-not-exist", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an unknown stop, got %d", rec.Code)
	}
}

func TestHealth(t *testing.T) {
	srv := New(Dependencies{Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMetaCarriesAttribution(t *testing.T) {
	srv := New(Dependencies{Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !contains(body, "OASA") {
		t.Errorf("expected the attribution to be present, got %s", body)
	}
}

func TestArrivalsWithoutSource(t *testing.T) {
	srv := New(Dependencies{Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/10175/arrivals", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no source is configured, got %d", rec.Code)
	}
}

func TestArrivalsAreCached(t *testing.T) {
	stub := &stubArrivals{}
	srv := New(Dependencies{Arrivals: stub, Version: "test"})

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/10175/arrivals", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	if stub.calls != 1 {
		t.Errorf("expected 1 upstream call for 3 requests, got %d", stub.calls)
	}
}

func TestVehiclesWithoutSource(t *testing.T) {
	srv := New(Dependencies{Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/routes/2484/vehicles", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no source is configured, got %d", rec.Code)
	}
}

func TestVehiclesAreCachedPerRoute(t *testing.T) {
	stub := &stubPositions{}
	srv := New(Dependencies{Positions: stub, Version: "test"})

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/routes/2484/vehicles", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/routes/4198/vehicles", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if stub.calls != 2 {
		t.Errorf("expected 2 upstream calls for two routes, got %d", stub.calls)
	}
}

func TestRoutesRequireALine(t *testing.T) {
	srv := New(Dependencies{Network: &stubNetwork{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/routes", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without a line parameter, got %d", rec.Code)
	}
}

func TestLinesAreServed(t *testing.T) {
	srv := New(Dependencies{Network: &stubNetwork{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/lines", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !contains(body, "ΠΛΑΤΕΙΑ") {
		t.Errorf("expected the line list in the response, got %s", body)
	}
}

func TestGeometryIsGeoJSON(t *testing.T) {
	srv := New(Dependencies{Network: &stubNetwork{}, Version: "test"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/routes/2484/geometry", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !contains(body, "LineString") {
		t.Errorf("expected a GeoJSON LineString, got %s", body)
	}
	if !contains(body, "23.73153") {
		t.Errorf("expected the coordinates in the response, got %s", body)
	}
}

func TestArrivalsAreLabelledWithTheLineNumber(t *testing.T) {
	arrivals := &stubArrivals{}
	srv := New(Dependencies{Arrivals: arrivals, Network: &stubNetwork{}, Version: "test"})

	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/10175/arrivals", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !contains(body, "021") {
		t.Errorf("expected the line number in the response, got %s", body)
	}
	if !contains(body, "ΚΑΝΙΓΓΟΣ") {
		t.Errorf("expected the line name in the response, got %s", body)
	}
}

func TestUpstreamFailureIsBadGateway(t *testing.T) {
	stub := &stubArrivals{err: errors.New("upstream down")}
	srv := New(Dependencies{Arrivals: stub, Version: "test"})

	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stops/10175/arrivals", nil))

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 for a failing upstream, got %d", rec.Code)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
