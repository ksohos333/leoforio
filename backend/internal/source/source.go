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

// ArrivalsSource provides expected arrivals for a stop.
type ArrivalsSource interface {
	Arrivals(ctx context.Context, stopID string) ([]model.Arrival, error)
}

// PositionsSource provides the last known positions of vehicles.
type PositionsSource interface {
	Positions(ctx context.Context) ([]model.VehiclePosition, error)
}

// StaticSource provides the static network model.
type StaticSource interface {
	Lines(ctx context.Context) ([]model.Line, error)
	Stops(ctx context.Context) ([]model.Stop, error)
	Routes(ctx context.Context) ([]model.Route, error)
}
