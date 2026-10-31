# ADR 0006: F-Droid and direct APK first, Google Play later

- Status: Accepted
- Date: 2026-10-31
- Deciders: @ksohos333

## Context

The project is AGPL-3.0, source-available by definition, and run by one person
on a back burner. Google Play imposes policy constraints that conflict with
the product: background location is restricted, and every release requires a
privacy policy, a data-safety form and a review cycle measured in weeks.
F-Droid and direct APK distribution suit an open-source, no-tracking project
and let releases ship on the project's own schedule.

## Decision

- v1 is distributed through F-Droid and a signed direct APK.
- Google Play is added after v1 is stable, with a Play-specific flavor where
  it is needed.
- Background location is optional, user-initiated, and never a prerequisite
  for using the app.
- Push notifications arrive in v1.1, where the Play flavor can use FCM and
  the F-Droid flavor can use a F-Droid friendly alternative.

## Consequences

- Fewer potential users early, but releases are not gated on store review.
- The build must be reproducible enough for F-Droid's requirements, which
  means avoiding proprietary SDKs in the default flavor.
- Publishing to Play later is additional work (policy forms, privacy policy
  hosting, Play release signing), budgeted separately from v1.
