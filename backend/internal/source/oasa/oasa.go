package oasa

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/ksohos333/leoforio/backend/internal/model"
)

// athens is the timezone the upstream timestamps are expressed in.
//
// The upstream sends local time without an offset, so the location has to
// carry the daylight saving rules: Athens is UTC+2 in winter and UTC+3 in
// summer.
var athens = loadAthens()

func loadAthens() *time.Location {
	location, err := time.LoadLocation("Europe/Athens")
	if err != nil {
		return time.FixedZone("Europe/Athens", 3*60*60)
	}
	return location
}

// Source implements the source interfaces against the OASA Telematics API.
type Source struct {
	client *Client
	now    func() time.Time
}

// New returns a Source backed by the given client.
func New(client *Client) *Source {
	return &Source{client: client, now: time.Now}
}

// Lines returns every line published by the operator.
func (s *Source) Lines(ctx context.Context) ([]model.Line, error) {
	payload, err := s.client.list(ctx, "webGetLines", nil)
	if err != nil {
		return nil, err
	}

	lines := make([]model.Line, 0, len(payload))
	for _, item := range payload {
		lines = append(lines, model.Line{
			ID:   item.str("LineCode"),
			Code: item.str("LineID"),
			Name: item.str("LineDescr"),
		})
	}
	return lines, nil
}

// Routes returns the routes that belong to a line.
func (s *Source) Routes(ctx context.Context, lineCode string) ([]model.Route, error) {
	payload, err := s.client.list(ctx, "webGetRoutes", []string{lineCode})
	if err != nil {
		return nil, err
	}

	routes := make([]model.Route, 0, len(payload))
	for _, item := range payload {
		routes = append(routes, model.Route{
			ID:         item.str("RouteCode"),
			LineID:     item.str("LineCode"),
			LineNumber: item.str("LineID"),
			Direction:  item.str("RouteDescr"),
			Name:       item.str("RouteDescrEng"),
			LineName:   item.str("LineDescr"),
		})
	}
	return routes, nil
}

// RoutesForStop returns the routes that serve a stop.
//
// This is also how a route code is resolved to the line number a rider sees,
// which is what the arrivals are labelled with.
func (s *Source) RoutesForStop(ctx context.Context, stopCode string) ([]model.Route, error) {
	payload, err := s.client.list(ctx, "webRoutesForStop", []string{stopCode})
	if err != nil {
		return nil, err
	}

	routes := make([]model.Route, 0, len(payload))
	for _, item := range payload {
		routes = append(routes, model.Route{
			ID:         item.str("RouteCode"),
			LineID:     item.str("LineCode"),
			LineNumber: item.str("LineID"),
			Direction:  item.str("RouteDescr"),
			Name:       item.str("RouteDescrEng"),
			LineName:   item.str("LineDescr"),
		})
	}
	return routes, nil
}

// RouteStops returns the stops of a route, in the order the operator lists
// them.
func (s *Source) RouteStops(ctx context.Context, routeCode string) ([]model.Stop, error) {
	payload, err := s.client.list(ctx, "webGetStops", []string{routeCode})
	if err != nil {
		return nil, err
	}
	return toStops(payload), nil
}

// RouteGeometry returns the path of a route as ordered points.
func (s *Source) RouteGeometry(ctx context.Context, routeCode string) ([]model.Position, error) {
	payload, err := s.client.object(ctx, "webGetRoutesDetailsAndStops", []string{routeCode})
	if err != nil {
		return nil, err
	}

	type point struct {
		order    int
		position model.Position
	}

	details := splitList(payload["details"])
	points := make([]point, 0, len(details))
	for _, item := range details {
		latitude, err := strconv.ParseFloat(trim(item.str("routed_y")), 64)
		if err != nil {
			continue
		}
		longitude, err := strconv.ParseFloat(trim(item.str("routed_x")), 64)
		if err != nil {
			continue
		}
		order, err := strconv.Atoi(trim(item.str("routed_order")))
		if err != nil {
			order = len(points)
		}
		points = append(points, point{
			order:    order,
			position: model.Position{Latitude: latitude, Longitude: longitude},
		})
	}

	sort.Slice(points, func(i, j int) bool { return points[i].order < points[j].order })

	geometry := make([]model.Position, 0, len(points))
	for _, item := range points {
		geometry = append(geometry, item.position)
	}
	return geometry, nil
}

// Arrivals returns the expected arrivals at a stop.
func (s *Source) Arrivals(ctx context.Context, stopCode string) ([]model.Arrival, error) {
	payload, err := s.client.list(ctx, "getStopArrivals", []string{stopCode})
	if err != nil {
		return nil, err
	}

	now := s.now()
	arrivals := make([]model.Arrival, 0, len(payload))
	for _, item := range payload {
		minutes, err := strconv.Atoi(trim(item.str("btime2")))
		if err != nil {
			continue
		}
		routeCode := item.str("route_code")
		arrivals = append(arrivals, model.Arrival{
			StopID:    stopCode,
			RouteID:   routeCode,
			VehicleID: item.str("veh_code"),
			Expected:  now.Add(time.Duration(minutes) * time.Minute),
		})
	}
	return arrivals, nil
}

// Positions returns the last known positions of the vehicles on a route.
func (s *Source) Positions(ctx context.Context, routeCode string) ([]model.VehiclePosition, error) {
	payload, err := s.client.list(ctx, "getBusLocation", []string{routeCode})
	if err != nil {
		return nil, err
	}

	positions := make([]model.VehiclePosition, 0, len(payload))
	for _, item := range payload {
		latitude, err := strconv.ParseFloat(trim(item.str("CS_LAT")), 64)
		if err != nil {
			continue
		}
		longitude, err := strconv.ParseFloat(trim(item.str("CS_LNG")), 64)
		if err != nil {
			continue
		}
		recordedAt, err := parseUpstreamTime(item.str("CS_DATE"))
		if err != nil {
			continue
		}
		heading, _ := strconv.Atoi(trim(item.str("VEH_HEADING")))

		positions = append(positions, model.VehiclePosition{
			VehicleID:  item.str("VEH_NO"),
			LineID:     item.str("ROUTE_CODE"),
			RouteID:    item.str("ROUTE_CODE"),
			Position:   model.Position{Latitude: latitude, Longitude: longitude},
			Heading:    heading,
			RecordedAt: recordedAt,
		})
	}
	return positions, nil
}

func toStops(payload []object) []model.Stop {
	stops := make([]model.Stop, 0, len(payload))
	for _, item := range payload {
		latitude, _ := strconv.ParseFloat(trim(item.str("StopLat")), 64)
		longitude, _ := strconv.ParseFloat(trim(item.str("StopLng")), 64)
		stops = append(stops, model.Stop{
			ID:        item.str("StopCode"),
			Code:      item.str("StopID"),
			Name:      item.str("StopDescr"),
			Latitude:  latitude,
			Longitude: longitude,
		})
	}
	return stops
}

// parseUpstreamTime reads the timestamp format the upstream uses, for
// example "Oct 31 2026 10:19:23:000PM": a millisecond field is glued to the
// AM/PM marker and no timezone is given.
func parseUpstreamTime(value string) (time.Time, error) {
	trimmed := trim(value)

	if index := lastIndex(trimmed, ":"); index >= 0 {
		tail := trimmed[index+1:]
		if len(tail) >= 4 && (hasSuffix(tail, "AM") || hasSuffix(tail, "PM")) {
			trimmed = trimmed[:index] + tail[3:]
		}
	}

	parsed, err := time.ParseInLocation("Jan _2 2006 03:04:05PM", trimmed, athens)
	if err != nil {
		return time.Time{}, fmt.Errorf("oasa: parsing timestamp %q: %w", value, err)
	}
	return parsed, nil
}

func lastIndex(value, separator string) int {
	for i := len(value) - len(separator); i >= 0; i-- {
		if value[i:i+len(separator)] == separator {
			return i
		}
	}
	return -1
}

func hasSuffix(value, suffix string) bool {
	return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
}
