# SUT P1 — Link-shortener (greenfield micro-service)

A small, production-shaped URL shortener with a datastore.

## Stack preference (the architecture proposer reads this — there's no mined infra profile on a fresh repo)
- Language/runtime: **TypeScript on Node 20** (or Go 1.22 — pick the build-biased best fit).
- Datastore: **Postgres** (a real persistent store, exercised by env-compose + seed).
- Packaged to run under **docker-compose** (api + db), health-checked at `GET /healthz` → 200.
- Production-shaped: env-config, a migration for the schema, structured errors. Not a toy single-file.

## What to build
A REST service:
- `POST /shorten` `{ "url": "<absolute http(s) url>", "ttl_seconds"?: <int> }`
- `GET /{code}` → 302 redirect to the stored URL
- `GET /healthz` → 200 (liveness; env-compose health-checks this)

## Scenarios (born-verifiable — Given/When/Then)
- Given a valid absolute URL, When `POST /shorten`, Then status 201 and body has a `code` of exactly 7 base62 chars.
- Given a code returned by /shorten, When `GET /{code}`, Then status 302 and `Location` equals the original URL.
- Given an unknown code, When `GET /{code}`, Then status 404.
- Given a code created with `ttl_seconds: 1`, When `GET /{code}` after 2 seconds, Then status 404 (expired).
- Given the same URL shortened twice, When both POSTs return, Then the two codes are different (no dedup required).
- Given the api restarts, When `GET /{code}` for a pre-restart code, Then status 302 (persistence holds — proves the datastore is real, not in-memory).

## Out of scope
Auth, custom aliases, analytics, a UI. Keep it to the spine above.
