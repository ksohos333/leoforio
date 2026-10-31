package store

import (
	"context"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()

	target, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { target.Close() })
	return target
}

func seed(t *testing.T, target *Store) {
	t.Helper()

	stops := []Stop{
		{ID: "10175", Code: "18", Name: "ΠΛ. ΚΑΝΙΓΓΟΣ", Latitude: 37.98566, Longitude: 23.73113},
		{ID: "60134", Code: "060134", Name: "ΠΛ. ΟΜΟΝΟΙΑΣ", Latitude: 37.9833682, Longitude: 23.7279031},
		{ID: "100001", Code: "100001", Name: "14η ΝΤΑΜΑΡΙΑ", Latitude: 38.0659257, Longitude: 23.7130708},
		{ID: "400058", Code: "400058", Name: "ΒΕΝΙΖΕΛΟΥ", Latitude: 37.9432677, Longitude: 23.6520113},
	}
	if err := target.ReplaceStops(context.Background(), stops, "test", "v1", "2026-10-31T00:00:00Z"); err != nil {
		t.Fatalf("seeding store: %v", err)
	}
}

func TestSearchIsCaseAndAccentInsensitive(t *testing.T) {
	target := openTestStore(t)
	seed(t, target)

	cases := []struct {
		query string
		want  string
	}{
		{query: "κανιγγος", want: "10175"},
		{query: "ΚΑΝΙΓΓΟΣ", want: "10175"},
		{query: "ομονοιας", want: "60134"},
		{query: "βενιζελου", want: "400058"},
	}

	for _, testCase := range cases {
		stops, err := target.SearchStops(context.Background(), testCase.query, 10)
		if err != nil {
			t.Fatalf("searching %q: %v", testCase.query, err)
		}
		if len(stops) == 0 {
			t.Errorf("searching %q found nothing", testCase.query)
			continue
		}
		if stops[0].ID != testCase.want {
			t.Errorf("searching %q: expected %s first, got %s", testCase.query, testCase.want, stops[0].ID)
		}
	}
}

func TestSearchFindsStopByIdentifier(t *testing.T) {
	target := openTestStore(t)
	seed(t, target)

	stops, err := target.SearchStops(context.Background(), "10175", 10)
	if err != nil {
		t.Fatalf("searching: %v", err)
	}
	if len(stops) != 1 || stops[0].ID != "10175" {
		t.Fatalf("expected the stop with id 10175, got %v", stops)
	}
}

func TestSearchTreatsWildcardsLiterally(t *testing.T) {
	target := openTestStore(t)
	seed(t, target)

	stops, err := target.SearchStops(context.Background(), "%", 10)
	if err != nil {
		t.Fatalf("searching: %v", err)
	}
	if len(stops) != 0 {
		t.Errorf("expected no results for a literal percent sign, got %d", len(stops))
	}
}

func TestNearbyStopsAreSortedByDistance(t *testing.T) {
	target := openTestStore(t)
	seed(t, target)

	// Omonia square is the closest of the seeded stops to this point.
	stops, err := target.NearbyStops(context.Background(), 37.98337, 23.72790, 1.0, 10)
	if err != nil {
		t.Fatalf("finding nearby stops: %v", err)
	}
	if len(stops) < 2 {
		t.Fatalf("expected several nearby stops, got %d", len(stops))
	}
	if stops[0].ID != "60134" {
		t.Errorf("expected Omonoia square first, got %s", stops[0].ID)
	}
	for i := 1; i < len(stops); i++ {
		if stops[i].Distance < stops[i-1].Distance {
			t.Errorf("stops are not sorted by distance: %v", stops)
			break
		}
	}
}

func TestNearbyStopsRespectsRadius(t *testing.T) {
	target := openTestStore(t)
	seed(t, target)

	stops, err := target.NearbyStops(context.Background(), 37.98337, 23.72790, 0.1, 10)
	if err != nil {
		t.Fatalf("finding nearby stops: %v", err)
	}
	for _, stop := range stops {
		if stop.Distance > 0.1 {
			t.Errorf("stop %s is %f km away, outside the radius", stop.ID, stop.Distance)
		}
	}
}

func TestStopByID(t *testing.T) {
	target := openTestStore(t)
	seed(t, target)

	stop, err := target.StopByID(context.Background(), "10175")
	if err != nil {
		t.Fatalf("reading stop: %v", err)
	}
	if stop.Name != "ΠΛ. ΚΑΝΙΓΓΟΣ" {
		t.Errorf("unexpected stop name %q", stop.Name)
	}
	if stop.Code != "18" {
		t.Errorf("unexpected stop code %q", stop.Code)
	}

	if _, err := target.StopByID(context.Background(), "does-not-exist"); err == nil {
		t.Error("expected an error for an unknown stop")
	}
}

func TestImportVersionIsRemembered(t *testing.T) {
	target := openTestStore(t)

	version, err := target.ImportVersion(context.Background(), "test")
	if err != nil {
		t.Fatalf("reading version: %v", err)
	}
	if version != "" {
		t.Errorf("expected no version before an import, got %q", version)
	}

	seed(t, target)

	version, err = target.ImportVersion(context.Background(), "test")
	if err != nil {
		t.Fatalf("reading version: %v", err)
	}
	if version != "v1" {
		t.Errorf("expected version v1, got %q", version)
	}
}
