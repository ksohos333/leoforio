import AsyncStorage from '@react-native-async-storage/async-storage';

import {
  readFavourites,
  toFavourite,
  writeFavourites,
} from '../useFavourites';

const mockedStorage = AsyncStorage as unknown as {
  getItem: jest.Mock;
  setItem: jest.Mock;
};

beforeEach(async () => {
  mockedStorage.getItem.mockReset();
  mockedStorage.setItem.mockReset();
  mockedStorage.setItem.mockResolvedValue(undefined);
});

test('reads nothing when nothing was stored', async () => {
  mockedStorage.getItem.mockResolvedValue(null);

  await expect(readFavourites()).resolves.toEqual([]);
});

test('reads the stored favourites', async () => {
  mockedStorage.getItem.mockResolvedValue(
    JSON.stringify([{ id: '10175', name: 'ΠΛ. ΚΑΝΙΓΓΟΣ', lat: 37.98, lon: 23.73 }]),
  );

  const favourites = await readFavourites();

  expect(favourites).toHaveLength(1);
  expect(favourites[0].id).toBe('10175');
});

test('ignores stored rubbish', async () => {
  mockedStorage.getItem.mockResolvedValue(JSON.stringify({ not: 'a list' }));

  await expect(readFavourites()).resolves.toEqual([]);
});

test('survives a storage failure', async () => {
  mockedStorage.getItem.mockRejectedValue(new Error('no storage'));

  await expect(readFavourites()).resolves.toEqual([]);
});

test('writes the favourites as JSON', async () => {
  const favourites = [
    { id: '10175', name: 'ΠΛ. ΚΑΝΙΓΓΟΣ', lat: 37.98, lon: 23.73 },
  ];

  await writeFavourites(favourites);

  expect(mockedStorage.setItem).toHaveBeenCalledWith(
    'leoforio.favourites.v1',
    JSON.stringify(favourites),
  );
});

test('keeps the fields a stop needs', () => {
  const favourite = toFavourite({
    id: '10361',
    code: '000156',
    name: 'ΣΥΝΤΑΓΜΑ',
    lat: 37.9749,
    lon: 23.7348,
  });

  expect(favourite).toEqual({
    id: '10361',
    name: 'ΣΥΝΤΑΓΜΑ',
    lat: 37.9749,
    lon: 23.7348,
  });
});
