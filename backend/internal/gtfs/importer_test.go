package gtfs

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksohos333/leoforio/backend/internal/store"
)

// buildFeed creates a zip archive with the given files.
func buildFeed(t *testing.T, files map[string]string) []byte {
	t.Helper()

	buffer := &bytes.Buffer{}
	archive := zip.NewWriter(buffer)

	for name, content := range files {
		writer, err := archive.Create(name)
		if err != nil {
			t.Fatalf("creating %s in the archive: %v", name, err)
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	if err := archive.Close(); err != nil {
		t.Fatalf("closing the archive: %v", err)
	}
	return buffer.Bytes()
}

func TestParseStopRows(t *testing.T) {
	rows := "stop_id,stop_code,stop_name,stop_lat,stop_lon,stop_url\n" +
		"10175,18,ΠΛ. ΚΑΝΙΓΓΟΣ,37.9856600,23.7311300,http://telematics.oasa.gr/#stationInfo_10175\n" +
		"60010,060010,ΝΑΥΑΡΙΝΟΥ,37.9837048,23.7349122,\n" +
		",,broken,not-a-number,23.7,\n"

	stops, err := parseStopRows(bytes.NewReader([]byte(rows)))
	if err != nil {
		t.Fatalf("parsing rows: %v", err)
	}
	if len(stops) != 2 {
		t.Fatalf("expected 2 usable stops, got %d", len(stops))
	}

	first := stops[0]
	if first.ID != "10175" {
		t.Errorf("expected id 10175, got %q", first.ID)
	}
	if first.Code != "18" {
		t.Errorf("expected code 18, got %q", first.Code)
	}
	if first.Name != "ΠΛ. ΚΑΝΙΓΓΟΣ" {
		t.Errorf("unexpected name %q", first.Name)
	}
	if first.Latitude != 37.98566 {
		t.Errorf("unexpected latitude %v", first.Latitude)
	}

	second := stops[1]
	if second.Code != "060010" {
		t.Errorf("expected the leading zero to be kept, got %q", second.Code)
	}
}

func TestParseStopRowsRejectsMissingColumns(t *testing.T) {
	rows := "stop_id,stop_name\n10175,ΠΛ. ΚΑΝΙΓΓΟΣ\n"

	if _, err := parseStopRows(bytes.NewReader([]byte(rows))); err == nil {
		t.Error("expected an error when a required column is missing")
	}
}

func TestImportLoadsStopsAndSkipsUnchangedFeeds(t *testing.T) {
	feed := buildFeed(t, map[string]string{
		"stops.txt": "stop_id,stop_code,stop_name,stop_lat,stop_lon\n" +
			"10175,18,ΠΛ. ΚΑΝΙΓΓΟΣ,37.9856600,23.7311300\n" +
			"60134,060134,ΠΛ. ΟΜΟΝΟΙΑΣ,37.9833682,23.7279031\n",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n",
	})

	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("If-None-Match") == "feed-v1" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", "feed-v1")
		_, _ = w.Write(feed)
	}))
	t.Cleanup(server.Close)

	target, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { target.Close() })

	importer := NewImporter(target, server.URL)

	result, err := importer.Import(context.Background())
	if err != nil {
		t.Fatalf("importing: %v", err)
	}
	if result.Unchanged {
		t.Error("expected a first import to change the index")
	}
	if result.Stops != 2 {
		t.Errorf("expected 2 stops, got %d", result.Stops)
	}

	count, err := target.StopCount(context.Background())
	if err != nil {
		t.Fatalf("counting stops: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 stops in the store, got %d", count)
	}

	version, err := target.ImportVersion(context.Background(), SourceName)
	if err != nil {
		t.Fatalf("reading version: %v", err)
	}
	if version != "feed-v1" {
		t.Errorf("expected version feed-v1, got %q", version)
	}

	second, err := importer.Import(context.Background())
	if err != nil {
		t.Fatalf("importing again: %v", err)
	}
	if !second.Unchanged {
		t.Error("expected the second import to be skipped as unchanged")
	}
	if requests != 2 {
		t.Errorf("expected 2 requests, got %d", requests)
	}
}

func TestImportReportsAMissingFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	target, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { target.Close() })

	importer := NewImporter(target, server.URL)

	if _, err := importer.Import(context.Background()); err == nil {
		t.Error("expected an error for a missing feed")
	}
}

func TestImportLeavesTheIndexAloneOnFailure(t *testing.T) {
	feed := buildFeed(t, map[string]string{
		"stops.txt": "stop_id,stop_code,stop_name,stop_lat,stop_lon\n10175,18,ΠΛ. ΚΑΝΙΓΓΟΣ,37.98,23.73\n",
	})

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			_, _ = w.Write(feed)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	target, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { target.Close() })

	importer := NewImporter(target, server.URL)

	if _, err := importer.Import(context.Background()); err != nil {
		t.Fatalf("first import: %v", err)
	}
	if _, err := importer.Import(context.Background()); err == nil {
		t.Fatal("expected the second import to fail")
	}

	count, err := target.StopCount(context.Background())
	if err != nil {
		t.Fatalf("counting stops: %v", err)
	}
	if count != 1 {
		t.Errorf("expected the index to survive the failed import, got %d stops", count)
	}
}

func TestTemporaryFeedFilesAreRemoved(t *testing.T) {
	feed := buildFeed(t, map[string]string{
		"stops.txt": "stop_id,stop_code,stop_name,stop_lat,stop_lon\n10175,18,ΠΛ. ΚΑΝΙΓΓΟΣ,37.98,23.73\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(feed)
	}))
	t.Cleanup(server.Close)

	target, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { target.Close() })

	before, err := filepath.Glob(filepath.Join(os.TempDir(), "leoforio-gtfs-*.zip"))
	if err != nil {
		t.Fatalf("listing temporary files: %v", err)
	}

	importer := NewImporter(target, server.URL)
	if _, err := importer.Import(context.Background()); err != nil {
		t.Fatalf("importing: %v", err)
	}

	after, err := filepath.Glob(filepath.Join(os.TempDir(), "leoforio-gtfs-*.zip"))
	if err != nil {
		t.Fatalf("listing temporary files: %v", err)
	}
	if len(after) != len(before) {
		t.Errorf("expected the temporary feed file to be removed, found %d", len(after)-len(before))
	}
}
