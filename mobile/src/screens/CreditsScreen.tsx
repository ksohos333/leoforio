import { ScrollView, StyleSheet, Text } from 'react-native';

/**
 * Credits and disclaimer.
 *
 * The disclaimer is not decoration: the transit data may only be used for
 * personal, non-commercial purposes, and the operator has to be credited.
 * See docs/legal.md.
 */
export function CreditsScreen() {
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.title}>Credits</Text>

      <Text style={styles.body}>
        leoforio is an unofficial app. It is not affiliated with, endorsed by
        or supported by OASA (Οργανισμός Αστικών Συγκοινωνιών Αθηνών) or by
        any other transport operator.
      </Text>

      <Text style={styles.body}>
        Transit data © OASA. The static network data comes from the OASA feed
        published on data.gov.gr, and the live arrivals, vehicle positions and
        timetables come from the OASA Telematics service.
      </Text>

      <Text style={styles.body}>
        The live feed is used for personal, non-commercial purposes only, as
        required by the OASA Telematics terms of use. It is queried when you
        open the app, held in a short-lived cache, and never stored or
        redistributed.
      </Text>

      <Text style={styles.body}>
        The Telematics service is undocumented and may change or disappear
        without notice, so no availability is promised for it. When data is
        missing, this app says so instead of guessing.
      </Text>

      <Text style={styles.body}>
        Map data © OpenStreetMap contributors. Tiles are served by a
        configured provider.
      </Text>

      <Text style={styles.body}>
        The software is licensed under the AGPL-3.0. The data is not: the
        terms that govern it restrict it to personal, non-commercial use.
      </Text>

      <Text style={styles.small}>
        https://github.com/ksohos333/leoforio
      </Text>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    padding: 20,
    paddingBottom: 48,
  },
  title: {
    fontSize: 20,
    fontWeight: '700',
    color: '#1f2933',
    marginBottom: 16,
  },
  body: {
    fontSize: 14,
    lineHeight: 21,
    color: '#3e4c59',
    marginBottom: 14,
  },
  small: {
    fontSize: 13,
    color: '#7b8794',
  },
});
