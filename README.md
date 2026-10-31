# leoforio

*λεωφορείο — bus*

Open app for Greek public transport, starting with Athens (OASA network). **Status: idea stage** — deliberately on the back burner while I build my main project. This repo exists to timestamp the idea, capture the plan, and gather signal from others who share the pain.

## Repository

| Path | Purpose |
|---|---|
| [`backend/`](backend/README.md) | Go service: upstream adapters, polite caching, versioned JSON API |
| [`mobile/`](mobile/README.md) | React Native app (MapLibre Native over OpenStreetMap tiles) |
| [`docs/adr/`](docs/adr/) | architecture decision records |
| [`docs/data-sources/`](docs/data-sources/README.md) | every upstream data source, its terms and its captured fixtures |

The app never calls an upstream operator directly: all upstream access lives
behind adapters in the backend, so that a broken or changed upstream degrades
one service instead of every user's phone. See
[ADR 0004](docs/adr/0004-app-never-calls-upstream.md).

## The problem

Most people in Greece have no idea when the bus is coming. On a good day you might know half the time. The official tools exist but feel like a punishment: cluttered interfaces, routes that are hard to read, no real sense of "where is my bus right now."

Knowing when the bus comes should not be a privilege.

## Vision

- **Clean, calm UI** — see your route and your stop in seconds, not menus
- **Live bus tracking** on a map — where the bus actually is, not where the schedule wishes it was
- **Timetables that respect you** — merged from official open-data feeds and public sources (scraped where necessary; respect each source's terms, prefer official feeds when available)
- **Open map technologies first** — OpenStreetMap / Leaflet as the base; Google Maps as an optional layer where it genuinely helps
- Free, open source, no ads, no tracking

## Non-goals (for v1)

- Ride-hailing, ticketing, payments
- Coverage outside Greece

## Want this to exist?

Open an issue with your wishlist — what would make *you* actually know when your bus is coming. Signal now shapes the build later.

## License

AGPLv3 — see [LICENSE](LICENSE).
