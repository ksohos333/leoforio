// Package store persists the static network model.
//
// The schema is GTFS-shaped so that an official or community feed can be
// imported without a translation layer (see ADR 0001).
package store

// Schema is the SQL schema for the static network model.
const Schema = `
CREATE TABLE IF NOT EXISTS stops (
	id         TEXT PRIMARY KEY,
	code       TEXT NOT NULL,
	name       TEXT NOT NULL,
	latitude   REAL NOT NULL,
	longitude  REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS lines (
	id          TEXT PRIMARY KEY,
	code        TEXT NOT NULL,
	name        TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS routes (
	id         TEXT PRIMARY KEY,
	line_id    TEXT NOT NULL REFERENCES lines(id),
	direction  TEXT NOT NULL,
	name       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS route_stops (
	route_id   TEXT NOT NULL REFERENCES routes(id),
	stop_id    TEXT NOT NULL REFERENCES stops(id),
	position   INTEGER NOT NULL,
	PRIMARY KEY (route_id, stop_id)
);

CREATE TABLE IF NOT EXISTS trips (
	id         TEXT PRIMARY KEY,
	route_id   TEXT NOT NULL REFERENCES routes(id),
	service_id TEXT NOT NULL,
	headsign   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS stop_times (
	trip_id       TEXT NOT NULL REFERENCES trips(id),
	stop_id       TEXT NOT NULL REFERENCES stops(id),
	sequence      INTEGER NOT NULL,
	arrival_sec   INTEGER NOT NULL,
	departure_sec INTEGER NOT NULL,
	PRIMARY KEY (trip_id, stop_id)
);

CREATE INDEX IF NOT EXISTS idx_stop_times_stop ON stop_times(stop_id);
CREATE INDEX IF NOT EXISTS idx_routes_line ON routes(line_id);
`
