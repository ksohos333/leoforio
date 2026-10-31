import type { StyleSpecification } from '@maplibre/maplibre-react-native';

/**
 * Builds a raster style from a tile URL.
 *
 * The style is built in the app rather than fetched, so that a release can
 * point it at a provider that permits app-scale use without shipping a
 * third-party style URL (ADR 0005).
 */
export function rasterStyle(
  tileUrl: string,
  attribution: string,
): StyleSpecification {
  return {
    version: 8,
    sources: {
      base: {
        type: 'raster',
        tiles: [tileUrl],
        tileSize: 256,
        attribution,
      },
    },
    layers: [{ id: 'base', type: 'raster', source: 'base' }],
  };
}
