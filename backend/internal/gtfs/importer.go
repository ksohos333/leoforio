// Package gtfs imports the official OASA GTFS feed.
//
// The feed is a 30 MB zip that contains a 225 MB stop_times file whose
// service calendar has expired, so the importer reads only the files that
// stay useful: the stops, which carry the identifiers the live API uses, and
// which are the index the app searches.
package gtfs

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ksohos333/leoforio/backend/internal/store"
)

// DefaultURL is the official OASA feed on the Greek open data portal.
const DefaultURL = "https://data.gov.gr/dataset/fb049bb1-aea6-4443-95fa-8b941dd6a057/resource/119db488-16ea-4c76-b560-41c472872390/download/osy_gtfs.zip"

// UserAgent identifies this project to the operator of the portal.
const UserAgent = "leoforio/0.1 (+https://github.com/ksohos333/leoforio)"

// SourceName is the name the import is recorded under.
const SourceName = "oasa-gtfs"

// Importer downloads the feed and loads it into the store.
type Importer struct {
	url    string
	client *http.Client
	store  *store.Store
}

// NewImporter returns an importer for the given store. An empty URL uses
// DefaultURL.
func NewImporter(target *store.Store, url string) *Importer {
	if url == "" {
		url = DefaultURL
	}
	return &Importer{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Minute},
		store:  target,
	}
}

// Result describes what an import did.
type Result struct {
	Version   string
	Stops     int
	Unchanged bool
}

// Import downloads the feed and replaces the stop index.
//
// When the feed has not changed since the last import, the store is left
// untouched.
func (i *Importer) Import(ctx context.Context) (Result, error) {
	previous, err := i.store.ImportVersion(ctx, SourceName)
	if err != nil {
		return Result{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, i.url, nil)
	if err != nil {
		return Result{}, fmt.Errorf("gtfs: building request: %w", err)
	}
	request.Header.Set("User-Agent", UserAgent)
	if previous != "" {
		request.Header.Set("If-None-Match", previous)
	}

	response, err := i.client.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf("gtfs: downloading feed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotModified {
		return Result{Version: previous, Unchanged: true}, nil
	}
	if response.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("gtfs: feed returned status %d", response.StatusCode)
	}

	etag := response.Header.Get("ETag")
	if etag == "" {
		etag = response.Header.Get("Last-Modified")
	}
	if etag == "" {
		etag = fmt.Sprintf("imported-%s", time.Now().UTC().Format(time.RFC3339))
	}

	stops, err := i.parseFeed(ctx, response.Body)
	if err != nil {
		return Result{}, err
	}
	if len(stops) == 0 {
		return Result{}, fmt.Errorf("gtfs: feed contained no stops")
	}

	if err := i.store.ReplaceStops(ctx, stops, SourceName, etag, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return Result{}, err
	}

	return Result{Version: etag, Stops: len(stops)}, nil
}

// parseFeed downloads the archive to a temporary file and reads the stops
// out of it.
//
// The archive is a zip, which needs a size, so it is spooled to disk rather
// than held in memory.
func (i *Importer) parseFeed(ctx context.Context, body io.Reader) ([]store.Stop, error) {
	temporary, err := os.CreateTemp("", "leoforio-gtfs-*.zip")
	if err != nil {
		return nil, fmt.Errorf("gtfs: creating temporary file: %w", err)
	}
	defer os.Remove(temporary.Name())
	defer temporary.Close()

	if _, err := io.Copy(temporary, body); err != nil {
		return nil, fmt.Errorf("gtfs: saving feed: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return nil, fmt.Errorf("gtfs: closing feed file: %w", err)
	}

	return parseStops(temporary.Name())
}

// parseStops reads stops.txt out of the feed archive.
//
// The archive is read from disk, so the large stop_times file inside it is
// never unpacked.
func parseStops(path string) ([]store.Stop, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("gtfs: reading feed archive: %w", err)
	}
	defer archive.Close()

	for _, file := range archive.File {
		if file.Name != "stops.txt" {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("gtfs: opening stops.txt: %w", err)
		}
		defer reader.Close()

		return parseStopRows(reader)
	}

	return nil, fmt.Errorf("gtfs: feed contains no stops.txt")
}

// parseStopRows reads the CSV rows of stops.txt.
//
// The columns are stop_id, stop_code, stop_name, stop_lat, stop_lon and
// stop_url. The stop_id is the identifier the live API uses, so a stop found
// here can be asked about directly.
func parseStopRows(reader io.Reader) ([]store.Stop, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1

	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("gtfs: reading stops.txt: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}

	header := rows[0]
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimSpace(name)] = index
	}

	required := []string{"stop_id", "stop_name", "stop_lat", "stop_lon"}
	for _, name := range required {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("gtfs: stops.txt has no %s column", name)
		}
	}

	stops := make([]store.Stop, 0, len(rows)-1)
	for _, row := range rows[1:] {
		get := func(name string) string {
			index, ok := columns[name]
			if !ok || index >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[index])
		}

		id := get("stop_id")
		if id == "" {
			continue
		}

		latitude, err := strconv.ParseFloat(get("stop_lat"), 64)
		if err != nil {
			continue
		}
		longitude, err := strconv.ParseFloat(get("stop_lon"), 64)
		if err != nil {
			continue
		}

		stops = append(stops, store.Stop{
			ID:        id,
			Code:      get("stop_code"),
			Name:      get("stop_name"),
			Latitude:  latitude,
			Longitude: longitude,
		})
	}

	return stops, nil
}
