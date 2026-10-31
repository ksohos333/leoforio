package oasa

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	// The timezone database is embedded so that the binary behaves the same
	// on a host without the system zoneinfo files.
	_ "time/tzdata"

	"github.com/ksohos333/leoforio/backend/internal/model"
)

// ErrRequestFailed means the upstream call did not succeed.
var ErrRequestFailed = errors.New("oasa: request failed")

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

type lineDTO struct {
	LineCode    string `json:"LineCode"`
	LineID      string `json:"LineID"`
	LineDescr   string `json:"LineDescr"`
	LineDescrEn string `json:"LineDescrEng"`
}

type routeDTO struct {
	RouteCode     string `json:"RouteCode"`
	LineCode      string `json:"LineCode"`
	RouteDescr    string `json:"RouteDescr"`
	RouteDescrEn  string `json:"RouteDescrEng"`
	RouteType     string `json:"RouteType"`
	RouteDistance string `json:"RouteDistance"`
}

type stopDTO struct {
	StopCode       string `json:"StopCode"`
	StopID         string `json:"StopID"`
	StopDescr      string `json:"StopDescr"`
	StopDescrEn    string `json:"StopDescrEng"`
	StopStreet     string `json:"StopStreet"`
	StopStreetEn   string `json:"StopStreetEng"`
	StopHeading    string `json:"StopHeading"`
	StopLat        string `json:"StopLat"`
	StopLng        string `json:"StopLng"`
	RouteStopOrder string `json:"RouteStopOrder"`
}

type arrivalDTO struct {
	RouteCode string `json:"route_code"`
	VehicleID string `json:"veh_code"`
	Minutes   string `json:"btime2"`
}

type positionDTO struct {
	VehicleID  string `json:"VEH_NO"`
	RecordedAt string `json:"CS_DATE"`
	Latitude   string `json:"CS_LAT"`
	Longitude  string `json:"CS_LNG"`
	RouteCode  string `json:"ROUTE_CODE"`
	Heading    string `json:"VEH_HEADING"`
}

type routeDetailsDTO struct {
	Details []struct {
		Longitude string `json:"routed_x"`
		Latitude  string `json:"routed_y"`
		Order     string `json:"routed_order"`
	} `json:"details"`
	Stops []stopDTO `json:"stops"`
}

// Lines returns every line published by the operator.
func (s *Source) Lines(ctx context.Context) ([]model.Line, error) {
	var payload []lineDTO
	if err := s.client.get(ctx, "webGetLines", nil, &payload); err != nil {
		return nil, err
	}

	lines := make([]model.Line, 0, len(payload))
	for _, item := range payload {
		lines = append(lines, model.Line{
			ID:   item.LineCode,
			Code: item.LineID,
			Name: item.LineDescr,
		})
	}
	return lines, nil
}

// Routes returns the routes that belong to a line.
func (s *Source) Routes(ctx context.Context, lineCode string) ([]model.Route, error) {
	var payload []routeDTO
	if err := s.client.get(ctx, "webGetRoutes", []string{lineCode}, &payload); err != nil {
		return nil, err
	}

	routes := make([]model.Route, 0, len(payload))
	for _, item := range payload {
		routes = append(routes, model.Route{
			ID:        item.RouteCode,
			LineID:    item.LineCode,
			Direction: item.RouteDescr,
			Name:      item.RouteDescrEn,
		})
	}
	return routes, nil
}

// RouteStops returns the stops of a route, in the order the operator lists
// them.
func (s *Source) RouteStops(ctx context.Context, routeCode string) ([]model.Stop, error) {
	var payload []stopDTO
	if err := s.client.get(ctx, "webGetStops", []string{routeCode}, &payload); err != nil {
		return nil, err
	}
	return toStops(payload), nil
}

// RouteGeometry returns the path of a route as ordered points.
func (s *Source) RouteGeometry(ctx context.Context, routeCode string) ([]model.Position, error) {
	var payload routeDetailsDTO
	if err := s.client.get(ctx, "webGetRoutesDetailsAndStops", []string{routeCode}, &payload); err != nil {
		return nil, err
	}

	type point struct {
		order    int
		position model.Position
	}
	points := make([]point, 0, len(payload.Details))
	for _, item := range payload.Details {
		latitude, err := strconv.ParseFloat(item.Latitude, 64)
		if err != nil {
			continue
		}
		longitude, err := strconv.ParseFloat(item.Longitude, 64)
		if err != nil {
			continue
		}
		order, err := strconv.Atoi(item.Order)
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
	var payload []arrivalDTO
	if err := s.client.get(ctx, "getStopArrivals", []string{stopCode}, &payload); err != nil {
		return nil, err
	}

	now := s.now()
	arrivals := make([]model.Arrival, 0, len(payload))
	for _, item := range payload {
		minutes, err := strconv.Atoi(strings.TrimSpace(item.Minutes))
		if err != nil {
			continue
		}
		arrivals = append(arrivals, model.Arrival{
			StopID:    stopCode,
			LineID:    item.RouteCode,
			RouteID:   item.RouteCode,
			VehicleID: item.VehicleID,
			Expected:  now.Add(time.Duration(minutes) * time.Minute),
		})
	}
	return arrivals, nil
}

// Positions returns the last known positions of the vehicles on a route.
func (s *Source) Positions(ctx context.Context, routeCode string) ([]model.VehiclePosition, error) {
	var payload []positionDTO
	if err := s.client.get(ctx, "getBusLocation", []string{routeCode}, &payload); err != nil {
		return nil, err
	}

	positions := make([]model.VehiclePosition, 0, len(payload))
	for _, item := range payload {
		latitude, err := strconv.ParseFloat(item.Latitude, 64)
		if err != nil {
			continue
		}
		longitude, err := strconv.ParseFloat(item.Longitude, 64)
		if err != nil {
			continue
		}
		recordedAt, err := parseUpstreamTime(item.RecordedAt)
		if err != nil {
			continue
		}
		heading, _ := strconv.Atoi(strings.TrimSpace(item.Heading))

		positions = append(positions, model.VehiclePosition{
			VehicleID:  item.VehicleID,
			LineID:     item.RouteCode,
			RouteID:    item.RouteCode,
			Position:   model.Position{Latitude: latitude, Longitude: longitude},
			Heading:    heading,
			RecordedAt: recordedAt,
		})
	}
	return positions, nil
}

func toStops(payload []stopDTO) []model.Stop {
	stops := make([]model.Stop, 0, len(payload))
	for _, item := range payload {
		latitude, _ := strconv.ParseFloat(item.StopLat, 64)
		longitude, _ := strconv.ParseFloat(item.StopLng, 64)
		stops = append(stops, model.Stop{
			ID:        item.StopCode,
			Code:      item.StopID,
			Name:      item.StopDescr,
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
	trimmed := strings.TrimSpace(value)

	if index := strings.LastIndex(trimmed, ":"); index >= 0 {
		tail := trimmed[index+1:]
		if len(tail) >= 4 && (strings.HasSuffix(tail, "AM") || strings.HasSuffix(tail, "PM")) {
			trimmed = trimmed[:index] + tail[3:]
		}
	}

	parsed, err := time.ParseInLocation("Jan _2 2006 03:04:05PM", trimmed, athens)
	if err != nil {
		return time.Time{}, fmt.Errorf("oasa: parsing timestamp %q: %w", value, err)
	}
	return parsed, nil
}
