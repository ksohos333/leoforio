import { useQuery } from '@tanstack/react-query';
import { FlatList, Pressable, StyleSheet, Text, View } from 'react-native';

import { api } from '../api/client';
import type { Line } from '../api/types';

interface LinesScreenProps {
  onSelect: (line: Line) => void;
}

export function LinesScreen({ onSelect }: LinesScreenProps) {
  const { data, isPending, isError } = useQuery({
    queryKey: ['lines'],
    queryFn: api.lines,
  });

  if (isError) {
    return (
      <View style={styles.message}>
        <Text style={styles.messageText}>
          The line list is unavailable right now. Check that the backend is
          reachable.
        </Text>
      </View>
    );
  }

  return (
    <FlatList
      data={data ?? []}
      keyExtractor={line => line.id}
      contentContainerStyle={styles.list}
      renderItem={({ item }) => (
        <Pressable
          style={styles.row}
          onPress={() => onSelect(item)}
          accessibilityRole="button"
          accessibilityLabel={`Line ${item.code}`}>
          <Text style={styles.code}>{item.code}</Text>
          <Text style={styles.name}>{item.name}</Text>
        </Pressable>
      )}
      ListEmptyComponent={
        isPending ? undefined : (
          <Text style={styles.messageText}>No lines found.</Text>
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
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#d8dce2',
  },
  code: {
    width: 64,
    fontSize: 16,
    fontWeight: '700',
    color: '#1f2933',
  },
  name: {
    flex: 1,
    fontSize: 14,
    color: '#3e4c59',
  },
  message: {
    padding: 24,
  },
  messageText: {
    fontSize: 14,
    color: '#3e4c59',
  },
});
