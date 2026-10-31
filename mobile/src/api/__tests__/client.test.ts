import { ApiError, api } from '../client';

const fetchMock = jest.fn();

beforeEach(() => {
  fetchMock.mockReset();
  globalThis.fetch = fetchMock as unknown as typeof fetch;
});

test('returns parsed vehicles on success', async () => {
  fetchMock.mockResolvedValue({
    ok: true,
    json: async () => [
      {
        vehicleId: 'v1',
        lineId: '815',
        routeId: '815-A',
        position: { lat: 37.98, lon: 23.72 },
        recordedAt: '2026-10-31T12:00:00Z',
      },
    ],
  });

  const vehicles = await api.vehicles();

  expect(vehicles).toHaveLength(1);
  expect(vehicles[0].vehicleId).toBe('v1');
  expect(fetchMock).toHaveBeenCalledWith(
    expect.stringContaining('/api/v1/vehicles'),
    expect.objectContaining({ signal: expect.anything() }),
  );
});

test('raises ApiError with the response status', async () => {
  fetchMock.mockResolvedValue({
    ok: false,
    status: 503,
    statusText: 'Service Unavailable',
    text: async () => 'source: not configured',
  });

  await expect(api.vehicles()).rejects.toBeInstanceOf(ApiError);
});

test('escapes the stop id in the arrivals path', async () => {
  fetchMock.mockResolvedValue({
    ok: true,
    json: async () => [],
  });

  await api.arrivals('stop/with slash');

  expect(fetchMock).toHaveBeenCalledWith(
    expect.stringContaining('/api/v1/stops/stop%2Fwith%20slash/arrivals'),
    expect.objectContaining({ signal: expect.anything() }),
  );
});
