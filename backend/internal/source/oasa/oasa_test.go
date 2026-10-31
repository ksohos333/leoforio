package oasa

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fixtureBody loads a captured response from the repository fixtures.
//
// The fixtures are trimmed samples, so a list fixture is served as the
// "sample" array and an object fixture is served as it is.
func fixtureBody(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("..", "..", "..", "..", "docs", "data-sources", "fixtures", "oasa", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}

	var wrapper struct {
		Sample json.RawMessage `json:"sample"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Sample) > 0 {
		return wrapper.Sample
	}
	return data
}

// stubServer serves the given body and records the last request it saw.
func stubServer(t *testing.T, body []byte, recorded *http.Request) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if recorded != nil {
			*recorded = *r
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestRequestUsesPositionalParameters(t *testing.T) {
	var recorded http.Request
	server := stubServer(t, []byte(`[]`), &recorded)

	source := New(NewClient(server.URL))
	if _, err := source.Arrivals(context.Background(), "10175"); err != nil {
		t.Fatalf("arrivals failed: %v", err)
	}

	query := recorded.URL.Query()
	if got := query.Get("act"); got != "getStopArrivals" {
		t.Errorf("expected act=getStopArrivals, got %q", got)
	}
	if got := query.Get("p1"); got != "10175" {
		t.Errorf("expected p1=10175, got %q", got)
	}
	if got := recorded.Header.Get("User-Agent"); got != DefaultUserAgent {
		t.Errorf("expected a descriptive user agent, got %q", got)
	}
}

func TestArrivalsFromFixture(t *testing.T) {
	server := stubServer(t, fixtureBody(t, "getStopArrivals"), nil)

	source := New(NewClient(server.URL))
	fixed := time.Date(2026, 10, 31, 21, 0, 0, 0, time.UTC)
	source.now = func() time.Time { return fixed }

	arrivals, err := source.Arrivals(context.Background(), "10175")
	if err != nil {
		t.Fatalf("arrivals failed: %v", err)
	}
	if len(arrivals) != 1 {
		t.Fatalf("expected 1 arrival, got %d", len(arrivals))
	}

	arrival := arrivals[0]
	if arrival.StopID != "10175" {
		t.Errorf("expected stop 10175, got %q", arrival.StopID)
	}
	if arrival.RouteID != "2484" {
		t.Errorf("expected route 2484, got %q", arrival.RouteID)
	}
	if arrival.VehicleID != "12237" {
		t.Errorf("expected vehicle 12237, got %q", arrival.VehicleID)
	}
	// The fixture carries btime2 = 11 minutes.
	want := fixed.Add(11 * time.Minute)
	if !arrival.Expected.Equal(want) {
		t.Errorf("expected arrival at %s, got %s", want, arrival.Expected)
	}
}

func TestPositionsFromFixture(t *testing.T) {
	server := stubServer(t, fixtureBody(t, "getBusLocation"), nil)

	source := New(NewClient(server.URL))
	positions, err := source.Positions(context.Background(), "2484")
	if err != nil {
		t.Fatalf("positions failed: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}

	position := positions[0]
	if position.VehicleID != "12237" {
		t.Errorf("expected vehicle 12237, got %q", position.VehicleID)
	}
	if position.RouteID != "2484" {
		t.Errorf("expected route 2484, got %q", position.RouteID)
	}
	if position.Position.Latitude < 37.98 || position.Position.Latitude > 37.99 {
		t.Errorf("unexpected latitude %v", position.Position.Latitude)
	}
	if position.Position.Longitude < 23.74 || position.Position.Longitude > 23.75 {
		t.Errorf("unexpected longitude %v", position.Position.Longitude)
	}
	if position.Heading != 247 {
		t.Errorf("expected heading 247, got %d", position.Heading)
	}

	// The fixture carries the timestamp "Oct 31 2026 10:19:23:000PM", which is
	// 20:19:23 UTC because Athens is on UTC+2 in late October.
	want := time.Date(2026, 10, 31, 20, 19, 23, 0, time.UTC)
	if !position.RecordedAt.Equal(want) {
		t.Errorf("expected recorded at %s, got %s", want, position.RecordedAt)
	}
}

func TestRouteStopsFromFixture(t *testing.T) {
	server := stubServer(t, fixtureBody(t, "webGetStops"), nil)

	source := New(NewClient(server.URL))
	stops, err := source.RouteStops(context.Background(), "2484")
	if err != nil {
		t.Fatalf("route stops failed: %v", err)
	}
	if len(stops) != 5 {
		t.Fatalf("expected 5 stops in the trimmed fixture, got %d", len(stops))
	}

	first := stops[0]
	if first.ID != "10175" {
		t.Errorf("expected stop 10175, got %q", first.ID)
	}
	if first.Latitude == 0 || first.Longitude == 0 {
		t.Errorf("expected coordinates for stop 10175, got %v", first)
	}
}

func TestRouteGeometryFromFixture(t *testing.T) {
	server := stubServer(t, fixtureBody(t, "webGetRoutesDetailsAndStops"), nil)

	source := New(NewClient(server.URL))
	geometry, err := source.RouteGeometry(context.Background(), "2484")
	if err != nil {
		t.Fatalf("route geometry failed: %v", err)
	}
	if len(geometry) != 5 {
		t.Fatalf("expected 5 points in the trimmed fixture, got %d", len(geometry))
	}

	first := geometry[0]
	if first.Longitude < 23.73 || first.Longitude > 23.74 {
		t.Errorf("unexpected first longitude %v", first.Longitude)
	}
	if first.Latitude < 37.98 || first.Latitude > 37.99 {
		t.Errorf("unexpected first latitude %v", first.Latitude)
	}
}

func TestParseUpstreamTime(t *testing.T) {
	// The upstream sends local Athens time without an offset, so the
	// expectations are given as UTC instants: Athens is UTC+2 in winter and
	// UTC+3 in summer.
	cases := []struct {
		input string
		want  time.Time
	}{
		{
			input: "Oct 31 2026 10:19:23:000PM",
			want:  time.Date(2026, 10, 31, 20, 19, 23, 0, time.UTC),
		},
		{
			input: "Jan 03 2026 08:05:00:000AM",
			want:  time.Date(2026, 1, 3, 6, 5, 0, 0, time.UTC),
		},
		{
			input: "Jul 13 2016 11:05:32:000PM",
			want:  time.Date(2016, 7, 13, 20, 5, 32, 0, time.UTC),
		},
	}

	for _, testCase := range cases {
		got, err := parseUpstreamTime(testCase.input)
		if err != nil {
			t.Fatalf("parsing %q: %v", testCase.input, err)
		}
		if !got.Equal(testCase.want) {
			t.Errorf("parsing %q: expected %s, got %s", testCase.input, testCase.want, got)
		}
	}
}

func TestUpstreamErrorIsReported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	source := New(NewClient(server.URL))
	if _, err := source.Arrivals(context.Background(), "10175"); err == nil {
		t.Fatal("expected an error for a failing upstream")
	}
}

// TestNumericFieldsAreAccepted covers the upstream changing a field from a
// string to a number without notice.
func TestNumericFieldsAreAccepted(t *testing.T) {
	server := stubServer(t, []byte(`[{"route_code": 2484, "veh_code": 12237, "btime2": 16}]`), nil)

	source := New(NewClient(server.URL))
	fixed := time.Date(2026, 10, 31, 21, 0, 0, 0, time.UTC)
	source.now = func() time.Time { return fixed }

	arrivals, err := source.Arrivals(context.Background(), "10175")
	if err != nil {
		t.Fatalf("arrivals failed: %v", err)
	}
	if len(arrivals) != 1 {
		t.Fatalf("expected 1 arrival, got %d", len(arrivals))
	}
	if arrivals[0].RouteID != "2484" {
		t.Errorf("expected route 2484, got %q", arrivals[0].RouteID)
	}
	if arrivals[0].VehicleID != "12237" {
		t.Errorf("expected vehicle 12237, got %q", arrivals[0].VehicleID)
	}
	if !arrivals[0].Expected.Equal(fixed.Add(16 * time.Minute)) {
		t.Errorf("expected arrival at %s, got %s", fixed.Add(16*time.Minute), arrivals[0].Expected)
	}
}

func TestNumericPositionsAreAccepted(t *testing.T) {
	server := stubServer(t, []byte(`[{
		"VEH_NO": 12237,
		"CS_DATE": "Oct 31 2026 10:19:23:000PM",
		"CS_LAT": 37.988269,
		"CS_LNG": 23.746624,
		"ROUTE_CODE": 2484,
		"VEH_HEADING": 247
	}]`), nil)

	source := New(NewClient(server.URL))
	positions, err := source.Positions(context.Background(), "2484")
	if err != nil {
		t.Fatalf("positions failed: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	if positions[0].VehicleID != "12237" {
		t.Errorf("expected vehicle 12237, got %q", positions[0].VehicleID)
	}
	if positions[0].Position.Latitude != 37.988269 {
		t.Errorf("unexpected latitude %v", positions[0].Position.Latitude)
	}
	if positions[0].Heading != 247 {
		t.Errorf("expected heading 247, got %d", positions[0].Heading)
	}
}

// TestUpstreamErrorObjectIsReported covers the upstream answering with an
// error object where a list was expected.
func TestUpstreamErrorObjectIsReported(t *testing.T) {
	server := stubServer(t, []byte(`{"error":"nothing requested"}`), nil)

	source := New(NewClient(server.URL))
	_, err := source.Arrivals(context.Background(), "10175")
	if err == nil {
		t.Fatal("expected an error for an upstream error object")
	}
	if !errors.Is(err, ErrUpstreamError) {
		t.Errorf("expected an upstream error, got %v", err)
	}
	if !strings.Contains(err.Error(), "nothing requested") {
		t.Errorf("expected the upstream message in %v", err)
	}
}

// TestRejectionStopsFurtherRequests covers the operator rate limiting the
// client: the next calls must fail fast instead of hammering the upstream.
func TestRejectionStopsFurtherRequests(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	source := New(NewClient(server.URL))

	if _, err := source.Positions(context.Background(), "2484"); !errors.Is(err, ErrRejected) {
		t.Fatalf("expected a rejection, got %v", err)
	}
	if _, err := source.Positions(context.Background(), "2484"); !errors.Is(err, ErrRejected) {
		t.Fatalf("expected a rejection, got %v", err)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("expected the client to stop after one rejection, got %d calls", got)
	}
}

// TestLive exercises the upstream service itself. It is skipped unless
// LEOFORIO_LIVE=1 is set, so the normal test run stays hermetic.
func TestLive(t *testing.T) {
	if os.Getenv("LEOFORIO_LIVE") != "1" {
		t.Skip("set LEOFORIO_LIVE=1 to run against the live service")
	}

	source := New(NewClient(""))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	lines, err := source.Lines(ctx)
	if err != nil {
		t.Fatalf("lines: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected lines from the live service")
	}
	t.Logf("live: %d lines", len(lines))

	routes, err := source.Routes(ctx, lines[0].ID)
	if err != nil {
		t.Fatalf("routes: %v", err)
	}
	if len(routes) == 0 {
		t.Fatal("expected routes from the live service")
	}
	t.Logf("live: %d routes for line %s", len(routes), lines[0].ID)

	stops, err := source.RouteStops(ctx, routes[0].ID)
	if err != nil {
		t.Fatalf("route stops: %v", err)
	}
	if len(stops) == 0 {
		t.Fatal("expected stops from the live service")
	}
	t.Logf("live: %d stops on route %s", len(stops), routes[0].ID)

	arrivals, err := source.Arrivals(ctx, stops[0].ID)
	if err != nil {
		t.Fatalf("arrivals: %v", err)
	}
	t.Logf("live: %d arrivals at stop %s", len(arrivals), stops[0].ID)

	positions, err := source.Positions(ctx, routes[0].ID)
	if err != nil {
		t.Fatalf("positions: %v", err)
	}
	t.Logf("live: %d vehicles on route %s", len(positions), routes[0].ID)
}
