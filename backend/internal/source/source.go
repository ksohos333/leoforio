// Package source defines the interfaces every upstream data source must
// implement.
//
// Upstream access is isolated behind these interfaces so that the API layer
// and the app never depend on an upstream response shape, and so that a
// source can be swapped or disabled without touching its callers.
package source

import (
	"context"
	"errors"

	"github.com/ksohos333/leoforio/backend/internal/model"
)

// ErrNotConfigured is returned when a source is not wired to an upstream yet.
var ErrNotConfigured = errors.New("source: not configured")

// ErrUnavailable is returned when an upstream cannot be reached or is
// currently refusing to serve data.
var ErrUnavailable = errors.New("source: upstream unavailable")

// ErrNotFound is returned when the upstream has no data for the request.
var ErrNotFound = errors.New("source: not found")

// ArrivalsSource provides expected arrivals for a stop.
type ArrivalsSource interface {
	Arrivals(ctx context.Context, stopCode string) ([]model.Arrival, error)
}

// PositionsSource provides the last known positions of the vehicles on a
// route.
//
// Positions are published per route by the upstream, never for the whole
// fleet, so a caller asks for the routes it is actually showing.
type PositionsSource interface {
	Positions(ctx context.Context, routeCode string) ([]model.VehiclePosition, error)
}

// NetworkSource provides the static network model: lines, and the routes and
// stops that belong to them.
type NetworkSource interface {
	Lines(ctx context.Context) ([]model.Line, error)
	Routes(ctx context.Context, lineCode string) ([]model.Route, error)
	RouteStops(ctx context.Context, routeCode string) ([]model.Stop, error)
	RouteGeometry(ctx context.Context, routeCode string) ([]model.Position, error)
}
