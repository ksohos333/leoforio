// Package store persists the static network model in SQLite.
//
// The stop index comes from the official GTFS feed. The identifiers in that
// feed are the same identifiers the live API uses, so a stop found through
// search can be queried for arrivals without any translation.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// Schema is the SQL schema for the static network model.
const Schema = `
CREATE TABLE IF NOT EXISTS stops (
	id         TEXT PRIMARY KEY,
	code       TEXT NOT NULL,
	name       TEXT NOT NULL,
	search_key TEXT NOT NULL,
	latitude   REAL NOT NULL,
	longitude  REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stops_search_key ON stops(search_key);
CREATE INDEX IF NOT EXISTS idx_stops_latitude ON stops(latitude);
CREATE INDEX IF NOT EXISTS idx_stops_longitude ON stops(longitude);

CREATE TABLE IF NOT EXISTS import_state (
	source     TEXT PRIMARY KEY,
	version    TEXT NOT NULL,
	imported_at TEXT NOT NULL
);
`

// Stop is a boarding location as stored locally.
type Stop struct {
	ID        string  `json:"id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`

	// Distance is the distance from the queried point in kilometres. It is
	// only set for nearby lookups.
	Distance float64 `json:"distanceKm,omitempty"`
}

// Store is the local database.
type Store struct {
	db *sql.DB
}

// Open opens the store at the given path and applies the schema. The path
// may be ":memory:" for a temporary store.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("store: opening %s: %w", path, err)
	}

	// SQLite handles one writer at a time, and the import job is the only
	// writer.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: applying schema: %w", err)
	}

	return &Store{db: db}, nil
}

// Close releases the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// ImportVersion returns the version of the last successful import for a
// source, or an empty string when there was none.
func (s *Store) ImportVersion(ctx context.Context, source string) (string, error) {
	var version string
	err := s.db.QueryRowContext(ctx,
		`SELECT version FROM import_state WHERE source = ?`, source).Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("store: reading import version: %w", err)
	}
	return version, nil
}

// ReplaceStops replaces the stop index with the given stops and records the
// import version.
func (s *Store) ReplaceStops(ctx context.Context, stops []Stop, source, version string, importedAt string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: starting import: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM stops`); err != nil {
		return fmt.Errorf("store: clearing stops: %w", err)
	}

	statement, err := tx.PrepareContext(ctx,
		`INSERT INTO stops (id, code, name, search_key, latitude, longitude) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("store: preparing import: %w", err)
	}
	defer statement.Close()

	for _, stop := range stops {
		if _, err := statement.ExecContext(ctx,
			stop.ID, stop.Code, stop.Name, SearchKey(stop.Name), stop.Latitude, stop.Longitude); err != nil {
			return fmt.Errorf("store: importing stop %s: %w", stop.ID, err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO import_state (source, version, imported_at) VALUES (?, ?, ?)
		 ON CONFLICT(source) DO UPDATE SET version = excluded.version, imported_at = excluded.imported_at`,
		source, version, importedAt); err != nil {
		return fmt.Errorf("store: recording import: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: committing import: %w", err)
	}
	return nil
}

// StopCount returns the number of stops in the index.
func (s *Store) StopCount(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM stops`).Scan(&count); err != nil {
		return 0, fmt.Errorf("store: counting stops: %w", err)
	}
	return count, nil
}

// SearchStops finds stops whose name contains the query, or whose identifier
// matches it exactly.
//
// Matching runs against a normalised key, so a query in Greek finds a stop
// regardless of case or accents.
func (s *Store) SearchStops(ctx context.Context, query string, limit int) ([]Stop, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}

	key := SearchKey(trimmed)
	pattern := "%" + escapeLike(key) + "%"

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, code, name, latitude, longitude FROM stops
		 WHERE id = ? OR search_key LIKE ? ESCAPE '\'
		 ORDER BY LENGTH(name), name
		 LIMIT ?`,
		trimmed, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("store: searching stops: %w", err)
	}
	defer rows.Close()

	return scanStops(rows)
}

// NearbyStops returns the stops closest to a point, within a radius in
// kilometres.
func (s *Store) NearbyStops(ctx context.Context, latitude, longitude, radiusKm float64, limit int) ([]Stop, error) {
	if limit <= 0 {
		limit = 20
	}

	// A bounding box narrows the candidates before the distance is worked
	// out, so the sort stays cheap.
	latitudeDelta := radiusKm / 111.0
	longitudeDelta := radiusKm / (111.0 * math.Cos(latitude*math.Pi/180))

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, code, name, latitude, longitude FROM stops
		 WHERE latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?`,
		latitude-latitudeDelta, latitude+latitudeDelta,
		longitude-longitudeDelta, longitude+longitudeDelta)
	if err != nil {
		return nil, fmt.Errorf("store: finding nearby stops: %w", err)
	}
	defer rows.Close()

	stops, err := scanStops(rows)
	if err != nil {
		return nil, err
	}

	for i := range stops {
		stops[i].Distance = haversineKm(latitude, longitude, stops[i].Latitude, stops[i].Longitude)
	}

	sort.Slice(stops, func(i, j int) bool { return stops[i].Distance < stops[j].Distance })

	if len(stops) > limit {
		stops = stops[:limit]
	}
	return stops, nil
}

// StopByID returns a single stop.
func (s *Store) StopByID(ctx context.Context, id string) (Stop, error) {
	var stop Stop
	err := s.db.QueryRowContext(ctx,
		`SELECT id, code, name, latitude, longitude FROM stops WHERE id = ?`, id).
		Scan(&stop.ID, &stop.Code, &stop.Name, &stop.Latitude, &stop.Longitude)
	if err == sql.ErrNoRows {
		return Stop{}, fmt.Errorf("store: no stop with id %s", id)
	}
	if err != nil {
		return Stop{}, fmt.Errorf("store: reading stop %s: %w", id, err)
	}
	return stop, nil
}

func scanStops(rows *sql.Rows) ([]Stop, error) {
	stops := make([]Stop, 0)
	for rows.Next() {
		var stop Stop
		if err := rows.Scan(&stop.ID, &stop.Code, &stop.Name, &stop.Latitude, &stop.Longitude); err != nil {
			return nil, fmt.Errorf("store: reading stop: %w", err)
		}
		stops = append(stops, stop)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: reading stops: %w", err)
	}
	return stops, nil
}

// haversineKm returns the distance between two points in kilometres.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
