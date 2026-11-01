import Geolocation from '@react-native-community/geolocation';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  PermissionsAndroid,
  Platform,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';

import { api } from '../api/client';
import type { Stop } from '../api/types';

interface SearchScreenProps {
  onSelect: (stop: Stop) => void;
}

export function SearchScreen({ onSelect }: SearchScreenProps) {
  const [query, setQuery] = useState('');
  const [nearby, setNearby] = useState<Stop[] | null>(null);
  const [locationError, setLocationError] = useState<string | null>(null);

  const results = useQuery({
    queryKey: ['stop-search', query],
    queryFn: () => api.searchStops(query),
    enabled: query.trim().length > 1,
  });

  const findNearby = async () => {
    setLocationError(null);

    if (Platform.OS === 'android') {
      const granted = await PermissionsAndroid.request(
        PermissionsAndroid.PERMISSIONS.ACCESS_COARSE_LOCATION,
      );
      if (granted !== PermissionsAndroid.RESULTS.GRANTED) {
        setLocationError('Location permission was not granted.');
        return;
      }
    }

    try {
      const position = await currentPosition();
      const stops = await api.nearbyStops(
        position.coords.latitude,
        position.coords.longitude,
      );
      setNearby(stops);
    } catch (error) {
      setLocationError(locationErrorMessage(error));
    }
  };

  const stops = nearby ?? results.data ?? [];

  return (
    <View style={styles.container}>
      <View style={styles.searchRow}>
        <TextInput
          style={styles.input}
          value={query}
          onChangeText={text => {
            setQuery(text);
            setNearby(null);
          }}
          placeholder="Where are you waiting?"
          placeholderTextColor="#7b8794"
          autoCorrect={false}
          accessibilityLabel="Search for a stop"
        />
        <Pressable
          style={styles.nearbyButton}
          onPress={findNearby}
          accessibilityRole="button"
          accessibilityLabel="Find stops near me">
          <Text style={styles.nearbyButtonText}>Near me</Text>
        </Pressable>
      </View>

      {locationError ? (
        <Text style={styles.error}>{locationError}</Text>
      ) : null}

      {results.isFetching && query.trim().length > 1 ? (
        <ActivityIndicator style={styles.loading} />
      ) : null}

      <FlatList
        data={stops}
        keyExtractor={stop => stop.id}
        contentContainerStyle={styles.list}
        keyboardShouldPersistTaps="handled"
        renderItem={({ item }) => (
          <Pressable
            style={styles.row}
            onPress={() => onSelect(item)}
            accessibilityRole="button"
            accessibilityLabel={`Stop ${item.name}`}>
            <View style={styles.rowText}>
              <Text style={styles.name}>{item.name}</Text>
              <Text style={styles.code}>{item.code}</Text>
            </View>
            {item.distanceKm !== undefined ? (
              <Text style={styles.distance}>
                {Math.round(item.distanceKm * 1000)} m
              </Text>
            ) : null}
          </Pressable>
        )}
        ListEmptyComponent={
          query.trim().length > 1 && !results.isFetching ? (
            <Text style={styles.empty}>No stop matches that name.</Text>
          ) : nearby === null ? (
            <Text style={styles.empty}>
              Search for a stop by name, or find the stops near you.
            </Text>
          ) : undefined
        }
      />
    </View>
  );
}

/**
 * Turns a location failure into something a person can act on.
 *
 * The raw messages are written for developers: "Location request timed out"
 * says nothing about what to do next.
 */
function locationErrorMessage(error: unknown): string {
  const message = error instanceof Error ? error.message : '';

  if (message.includes('timed out')) {
    return 'This device could not get a location in time. Search for your stop by name instead.';
  }
  if (message.includes('provider')) {
    return 'This device has no location provider available. Search for your stop by name instead.';
  }
  if (message.includes('denied') || message.includes('Permission')) {
    return 'The location permission was not granted.';
  }
  return 'The location is unavailable. Search for your stop by name instead.';
}

function currentPosition(): Promise<{ coords: { latitude: number; longitude: number } }> {
  return new Promise((resolve, reject) => {
    Geolocation.getCurrentPosition(
      resolve,
      (error: { message?: string } | undefined) =>
        reject(new Error(error?.message ?? 'The location is unavailable.')),
      { enableHighAccuracy: false, timeout: 15000, maximumAge: 60000 },
    );
  });
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  searchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 10,
  },
  input: {
    flex: 1,
    height: 44,
    borderRadius: 10,
    paddingHorizontal: 14,
    backgroundColor: '#f1f3f5',
    color: '#1f2933',
    fontSize: 16,
  },
  nearbyButton: {
    height: 44,
    paddingHorizontal: 16,
    borderRadius: 10,
    backgroundColor: '#2f6fed',
    alignItems: 'center',
    justifyContent: 'center',
  },
  nearbyButtonText: {
    color: '#ffffff',
    fontSize: 15,
    fontWeight: '600',
  },
  loading: {
    marginVertical: 8,
  },
  error: {
    paddingHorizontal: 16,
    paddingBottom: 8,
    fontSize: 13,
    color: '#c0392b',
  },
  list: {
    paddingBottom: 24,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#d8dce2',
  },
  rowText: {
    flex: 1,
  },
  name: {
    fontSize: 15,
    color: '#1f2933',
  },
  code: {
    fontSize: 12,
    color: '#7b8794',
    marginTop: 2,
  },
  distance: {
    fontSize: 13,
    color: '#3e4c59',
  },
  empty: {
    padding: 24,
    fontSize: 14,
    color: '#3e4c59',
    lineHeight: 20,
  },
});
