jest.mock('@maplibre/maplibre-react-native', () => ({
  Map: 'Map',
  Camera: 'Camera',
  GeoJSONSource: 'GeoJSONSource',
  Layer: 'Layer',
}));

global.fetch = jest.fn().mockResolvedValue({
  ok: true,
  json: async () => [],
});
