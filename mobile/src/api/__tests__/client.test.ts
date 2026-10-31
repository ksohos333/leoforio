import { ApiError, api } from '../client';

const fetchMock = jest.fn();

beforeEach(() => {
  fetchMock.mockReset();
  globalThis.fetch = fetchMock as unknown as typeof fetch;
});

function respondWith(payload: unknown, ok = true, status = 200) {
  fetchMock.mockResolvedValue({
    ok,
    status,
    statusText: ok ? 'OK' : 'Error',
    json: async () => payload,
    text: async () => JSON.stringify(payload),
  });
}

test('returns parsed vehicles on success', async () => {
  respondWith([
    {
      vehicleId: 'v1',
      lineId: '2484',
      routeId: '2484',
      position: { lat: 37.98, lon: 23.72 },
      heading: 233,
      recordedAt: '2026-10-31T22:30:49+02:00',
    },
  ]);

  const vehicles = await api.vehicles('2484');

  expect(vehicles).toHaveLength(1);
  expect(vehicles[0].vehicleId).toBe('v1');
  expect(vehicles[0].heading).toBe(233);
  expect(fetchMock).toHaveBeenCalledWith(
    expect.stringContaining('/api/v1/routes/2484/vehicles'),
    expect.objectContaining({ signal: expect.anything() }),
  );
});

test('raises ApiError with the response status', async () => {
  respondWith({ error: 'source: not configured' }, false, 503);

  await expect(api.vehicles('2484')).rejects.toBeInstanceOf(ApiError);
});

test('escapes the stop id in the arrivals path', async () => {
  respondWith([]);

  await api.arrivals('stop/with slash');

  expect(fetchMock).toHaveBeenCalledWith(
    expect.stringContaining('/api/v1/stops/stop%2Fwith%20slash/arrivals'),
    expect.objectContaining({ signal: expect.anything() }),
  );
});

test('passes the line to the routes endpoint', async () => {
  respondWith([]);

  await api.routes('1151');

  expect(fetchMock).toHaveBeenCalledWith(
    expect.stringContaining('/api/v1/routes?line=1151'),
    expect.objectContaining({ signal: expect.anything() }),
  );
});

test('requests the route geometry as GeoJSON', async () => {
  respondWith({
    type: 'Feature',
    geometry: { type: 'LineString', coordinates: [[23.7, 37.9]] },
    properties: {},
  });

  const geometry = await api.routeGeometry('2484');

  expect(geometry.geometry.type).toBe('LineString');
  expect(geometry.geometry.coordinates).toHaveLength(1);
});

test('requests the lines list', async () => {
  respondWith([{ id: '1151', code: '021', name: 'ΠΛΑΤΕΙΑ ΚΑΝΙΓΓΟΣ' }]);

  const lines = await api.lines();

  expect(lines).toHaveLength(1);
  expect(fetchMock).toHaveBeenCalledWith(
    expect.stringContaining('/api/v1/lines'),
    expect.objectContaining({ signal: expect.anything() }),
  );
});
