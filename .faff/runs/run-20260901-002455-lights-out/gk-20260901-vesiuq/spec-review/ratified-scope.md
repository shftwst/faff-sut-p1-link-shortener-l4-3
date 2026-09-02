## Ratified scope

Assembled by `faff ratified-scope` from files committed to this repository. The spec under review is not a source and cannot write to any of these files.

### Ratified goals: PRD `link-shortener` (docs/prd/link-shortener.md)

- A running service that mints a code for an absolute URL and redirects that code to the URL.
- Codes survive an api container restart, proving the datastore is real rather than in-memory.
- An optional time-to-live is honoured: an expired code stops resolving.
- The packaged stack starts healthy and answers a liveness check quickly.
- An automated test suite covers every acceptance scenario below and passes.

### Non-goals: PRD `link-shortener` (docs/prd/link-shortener.md)

- Authentication and authorization.
- Custom aliases (caller-chosen codes).
- Analytics or click counting.
- A user interface.
- Deduplication of identical URLs (two shortenings of one URL may differ).
