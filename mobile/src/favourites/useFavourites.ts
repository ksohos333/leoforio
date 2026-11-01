import AsyncStorage from '@react-native-async-storage/async-storage';
import { useCallback, useEffect, useState } from 'react';

import type { Stop } from '../api/types';

const storageKey = 'leoforio.favourites.v1';

/**
 * A stop the user wants to keep an eye on.
 *
 * Favourites live on the device only. Nothing about them is sent anywhere,
 * and the backend never learns which stops a person waits at.
 */
export interface Favourite {
  id: string;
  name: string;
  lat: number;
  lon: number;
}

export function toFavourite(stop: Stop): Favourite {
  return {
    id: stop.id,
    name: stop.name,
    lat: stop.lat,
    lon: stop.lon,
  };
}

export async function readFavourites(): Promise<Favourite[]> {
  try {
    const stored = await AsyncStorage.getItem(storageKey);
    if (!stored) {
      return [];
    }
    const parsed = JSON.parse(stored);
    if (!Array.isArray(parsed)) {
      return [];
    }
    return parsed.filter(
      item =>
        typeof item?.id === 'string' && typeof item?.name === 'string',
    ) as Favourite[];
  } catch {
    return [];
  }
}

export async function writeFavourites(favourites: Favourite[]): Promise<void> {
  try {
    await AsyncStorage.setItem(storageKey, JSON.stringify(favourites));
  } catch {
    // A failure to store a favourite is not worth interrupting the user
    // for; the screen keeps working without it.
  }
}

/**
 * useFavourites keeps the list of favourite stops.
 */
export function useFavourites() {
  const [favourites, setFavourites] = useState<Favourite[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    let active = true;
    readFavourites().then(stored => {
      if (active) {
        setFavourites(stored);
        setLoaded(true);
      }
    });
    return () => {
      active = false;
    };
  }, []);

  const add = useCallback((stop: Stop) => {
    const favourite = toFavourite(stop);
    setFavourites(current => {
      const next = [
        favourite,
        ...current.filter(item => item.id !== favourite.id),
      ];
      writeFavourites(next);
      return next;
    });
  }, []);

  const remove = useCallback((id: string) => {
    setFavourites(current => {
      const next = current.filter(item => item.id !== id);
      writeFavourites(next);
      return next;
    });
  }, []);

  const has = useCallback(
    (id: string) => favourites.some(item => item.id === id),
    [favourites],
  );

  return { favourites, loaded, add, remove, has };
}
