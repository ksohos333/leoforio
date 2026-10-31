import { useQuery } from '@tanstack/react-query';
import { FlatList, Pressable, StyleSheet, Text, View } from 'react-native';

import { api } from '../api/client';
import type { Line, Route } from '../api/types';

interface RoutesScreenProps {
  line: Line;
  onSelect: (route: Route) => void;
}

export function RoutesScreen({ line, onSelect }: RoutesScreenProps) {
  const { data, isPending, isError } = useQuery({
    queryKey: ['routes', line.id],
    queryFn: () => api.routes(line.id),
  });

  if (isError) {
    return (
      <View style={styles.message}>
        <Text style={styles.messageText}>
          The routes of this line are unavailable right now.
        </Text>
      </View>
    );
  }

  return (
    <FlatList
      data={data ?? []}
      keyExtractor={route => route.id}
      contentContainerStyle={styles.list}
      renderItem={({ item }) => (
        <Pressable
          style={styles.row}
          onPress={() => onSelect(item)}
          accessibilityRole="button"
          accessibilityLabel={`Route ${item.name || item.direction}`}>
          <Text style={styles.name}>{item.name || item.direction}</Text>
          <Text style={styles.id}>{item.id}</Text>
        </Pressable>
      )}
      ListEmptyComponent={
        isPending ? undefined : (
          <Text style={styles.messageText}>No routes for this line.</Text>
        )
      }
    />
  );
}

const styles = StyleSheet.create({
  list: {
    paddingVertical: 8,
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
  name: {
    flex: 1,
    fontSize: 15,
    color: '#1f2933',
  },
  id: {
    fontSize: 12,
    color: '#7b8794',
  },
  message: {
    padding: 24,
  },
  messageText: {
    fontSize: 14,
    color: '#3e4c59',
  },
});
