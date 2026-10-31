# Fixtures

Captured upstream responses, one file per endpoint, named
`<source>/<endpoint>.json`.

Rules:

- A fixture is captured in the same session as the source document, so the
  document and the fixture cannot disagree.
- Fixtures are trimmed of anything that is not needed to understand the
  response: replace long identifier lists with a representative sample, keep
  the shape intact.
- Fixtures contain no personal data. Vehicle identifiers are truncated if
  they appear, and driver identifiers are removed entirely.
- Fixtures are the contract for adapter tests. When an upstream response
  shape changes, the test fails until both the adapter and the fixture are
  updated in the same commit.
- Fixtures are committed, not generated, so that contract tests are
  deterministic in CI.
