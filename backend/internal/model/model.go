// Package model holds the internal representation of the transit network.
//
// Every upstream source is normalised into these types, so that a change in
// an upstream response shape does not reach the API or the app.
package model

import "time"

// Position is a geographic point.
type Position struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// Line is a transport line as published upstream.
type Line struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// Route is an ordered sequence of stops belonging to a line direction.
type Route struct {
	ID        string     `json:"id"`
	LineID    string     `json:"lineId"`
	Direction string     `json:"direction"`
	Name      string     `json:"name"`
	StopIDs   []string   `json:"stopIds"`
	Geometry  []Position `json:"geometry,omitempty"`
}

// Stop is a boarding location.
type Stop struct {
	ID        string  `json:"id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// Arrival is an expected arrival of a vehicle at a stop.
type Arrival struct {
	StopID    string    `json:"stopId"`
	LineID    string    `json:"lineId"`
	RouteID   string    `json:"routeId"`
	VehicleID string    `json:"vehicleId,omitempty"`
	Scheduled time.Time `json:"scheduled"`
	Expected  time.Time `json:"expected"`
}

// VehiclePosition is the last known position of a vehicle.
type VehiclePosition struct {
	VehicleID  string    `json:"vehicleId"`
	LineID     string    `json:"lineId"`
	RouteID    string    `json:"routeId"`
	Position   Position  `json:"position"`
	RecordedAt time.Time `json:"recordedAt"`
}
