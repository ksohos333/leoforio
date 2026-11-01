import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';
import { Pressable, StatusBar, StyleSheet, Text, View } from 'react-native';
import {
  SafeAreaProvider,
  useSafeAreaInsets,
} from 'react-native-safe-area-context';

import type { Line, Route, Stop } from './src/api/types';
import { useFavourites } from './src/favourites/useFavourites';
import { CreditsScreen } from './src/screens/CreditsScreen';
import { LinesScreen } from './src/screens/LinesScreen';
import { LiveMapScreen } from './src/screens/LiveMapScreen';
import { RoutesScreen } from './src/screens/RoutesScreen';
import { SearchScreen } from './src/screens/SearchScreen';
import { StopScreen } from './src/screens/StopScreen';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

type Screen =
  | { name: 'search' }
  | { name: 'stop'; stop: Stop }
  | { name: 'lines' }
  | { name: 'routes'; line: Line }
  | { name: 'map'; route: Route }
  | { name: 'credits' };

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <SafeAreaProvider>
        <AppContent />
      </SafeAreaProvider>
    </QueryClientProvider>
  );
}

function AppContent() {
  const insets = useSafeAreaInsets();
  const [stack, setStack] = useState<Screen[]>([{ name: 'search' }]);
  const screen = stack[stack.length - 1];
  const favourites = useFavourites();

  const push = (next: Screen) => setStack(current => [...current, next]);
  const back = () =>
    setStack(current => (current.length > 1 ? current.slice(0, -1) : current));

  const title =
    screen.name === 'search'
      ? 'leoforio'
      : screen.name === 'stop'
      ? screen.stop.name
      : screen.name === 'lines'
      ? 'Lines'
      : screen.name === 'routes'
      ? `Line ${screen.line.code}`
      : screen.name === 'map'
      ? screen.route.name || screen.route.direction
      : 'Credits';

  return (
    <>
      <StatusBar barStyle="dark-content" />
      <View style={styles.container}>
        <View
          style={[
            styles.header,
            {
              paddingLeft: insets.left + 16,
              paddingRight: insets.right + 16,
              paddingTop: insets.top + 12,
            },
          ]}
        >
          {stack.length > 1 ? (
            <Pressable
              onPress={back}
              accessibilityRole="button"
              accessibilityLabel="Go back"
            >
              <Text style={styles.headerButton}>‹ Back</Text>
            </Pressable>
          ) : (
            <View />
          )}
          <Text style={styles.title} numberOfLines={1}>
            {title}
          </Text>
          {screen.name === 'credits' ? (
            <View />
          ) : (
            <Pressable
              onPress={() => push({ name: 'credits' })}
              accessibilityRole="button"
              accessibilityLabel="Credits"
            >
              <Text style={styles.headerButton}>Info</Text>
            </Pressable>
          )}
        </View>

        <View style={styles.body}>
          {screen.name === 'search' ? (
            <SearchScreen
              favourites={favourites.favourites}
              onSelect={stop => push({ name: 'stop', stop })}
              onSelectFavourite={favourite =>
                push({
                  name: 'stop',
                  stop: {
                    id: favourite.id,
                    code: '',
                    name: favourite.name,
                    lat: favourite.lat,
                    lon: favourite.lon,
                  },
                })
              }
              onRemoveFavourite={favourites.remove}
            />
          ) : null}
          {screen.name === 'stop' ? (
            <StopScreen
              stop={screen.stop}
              isFavourite={favourites.has(screen.stop.id)}
              onToggleFavourite={() => {
                if (favourites.has(screen.stop.id)) {
                  favourites.remove(screen.stop.id);
                } else {
                  favourites.add(screen.stop);
                }
              }}
              onSelectRoute={route => push({ name: 'map', route })}
            />
          ) : null}
          {screen.name === 'lines' ? (
            <LinesScreen onSelect={line => push({ name: 'routes', line })} />
          ) : null}
          {screen.name === 'routes' ? (
            <RoutesScreen
              line={screen.line}
              onSelect={route => push({ name: 'map', route })}
            />
          ) : null}
          {screen.name === 'map' ? (
            <LiveMapScreen route={screen.route} />
          ) : null}
          {screen.name === 'credits' ? <CreditsScreen /> : null}
        </View>

        {screen.name === 'search' ? (
          <View style={styles.tabs}>
            <Pressable
              style={styles.tab}
              onPress={() => push({ name: 'lines' })}
              accessibilityRole="button"
              accessibilityLabel="Browse lines"
            >
              <Text style={styles.tabText}>Browse lines</Text>
            </Pressable>
          </View>
        ) : null}
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#ffffff',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingBottom: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#d8dce2',
  },
  title: {
    flex: 1,
    textAlign: 'center',
    fontSize: 16,
    fontWeight: '600',
    color: '#1f2933',
  },
  headerButton: {
    fontSize: 15,
    color: '#2f6fed',
  },
  body: {
    flex: 1,
  },
  tabs: {
    paddingHorizontal: 16,
    paddingTop: 10,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: '#d8dce2',
  },
  tab: {
    alignSelf: 'flex-start',
    paddingVertical: 6,
  },
  tabText: {
    fontSize: 14,
    color: '#2f6fed',
  },
});

export default App;
