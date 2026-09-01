# E1 — Packaged service skeleton: docker-compose (api + Postgres), GET /healthz, env config

> Spec: faffter-dark-nlspec · 2026-09-01 · autonomous · claude-code/unknown · confidence: high · spec-review: approve.
> build-tier: complex
> Revised 2026-09-01 (spec-review round 1 `revise`): closed infosec objections (no-secret structured fatal log; non-root runtime user) and QA objections (born-verifiable graceful-shutdown, 405, and migration-runner scenarios; defined "structured fatal log").
> Revised 2026-09-01 (spec-review round 2, batch 1): rescoped migration verification — E1 asserts the golang-migrate runner RAN (empty `migrations/` → `ErrNoChange`), not a `schema_migrations` table (that is an E2 artifact); added a present-but-invalid-DSN no-secret-leak scenario, a non-root container-user scenario, and a connection-refused oracle for the "port never bound" claim.
> Revised 2026-09-01 (spec-review round 2, batch 2): required ALL fatal boot logs (config/migrate/db-connect) to scrub the DSN from wrapped pgx/golang-migrate driver errors; added an unreachable-DB (`db_connect_failed`) scenario verifying the failed-Ping path + no secret leak; strengthened the migration oracle to assert the runner log precedes the port-bound log (ordering, not just presence).

This is the build spec for epic E1 of the link-shortener service (gitkey `gk-20260901-r99do5`), written for the build agent and human reviewers. E1 is the first slice and the one that establishes the stack: a packaged, production-shaped service skeleton that runs under docker-compose with a real Postgres datastore, answers a `GET /healthz` liveness check, reads its configuration from the environment, wires in a schema-migration runner, and returns structured JSON errors. It deliberately stops short of the shorten, redirect, and TTL behaviour, which land in later epics (E2..E6). Everything the later epics need to exist as a runnable, health-checked surface is stood up here; nothing of the product behaviour itself is.

## 1. WHY — problem and principles

The load-bearing idea: this slice delivers a *runnable, health-checked container topology*, not a feature. Its whole value is that `docker compose up` brings an api container and a Postgres container to a healthy state, the api reaches the database, applies migrations, and answers `GET /healthz` with 200 within 60 seconds. Every later epic builds inside this skeleton, so the skeleton has to be real (real Postgres, real migration runner, real env config) rather than a stub.

Problem statement: the repository has no service to build on. Later epics need a packaged runtime, a real datastore, a migration harness, and a health surface before any endpoint can be written. E1 stands up exactly that skeleton and proves it healthy.

Design principles:

- **Scope is the skeleton, not the product.** Include only what the runnable, health-checked topology needs. The `POST /shorten`, `GET /{code}` redirect, and TTL behaviour are out of scope; adding them here would pull E2..E6 forward and break the slice boundary.
- **The definition of done is born-verifiable.** "api + db healthy and `GET /healthz` returns 200 within 60s of `docker compose up`" must be observable with `docker compose ps` and one HTTP GET, not asserted in prose.
- **A healthy api implies a reachable database.** The api fails fast at boot if it cannot read its config, connect to Postgres, or run migrations. Because it only starts after the db is healthy and crash-loops otherwise, a running, healthy api container is itself evidence the datastore is real and reachable.
- **Production-shaped, not a single file.** Configuration from the environment, a versioned migration, layered packages, and structured errors are in scope because the brief calls the skeleton production-shaped; a toy single-file server is explicitly rejected.

Reference context: none. This is a greenfield repository (only the SUT brief, PRD, and roadmap skeleton exist); there is no prior code, architecture, or datastore to mirror or extend.

Scope statement: E1 is the foundation layer of Project P1 (core shorten-to-redirect on a real datastore); it owns the packaging, config, health, and migration-runner surfaces that E2 (schema + persistence), E3 (`POST /shorten`), and E4 (`GET /{code}`) then build on.

## 2. OUT OF SCOPE

- **`POST /shorten` mint endpoint** — validating an absolute URL, minting a 7-char base62 code, persisting it. Why excluded: it is epic E3, and it depends on the persistence layer (E2) that is itself later. Extension point: a new handler registered on the `internal/httpapi` router, backed by `internal/store`.
- **`GET /{code}` redirect and 404 for unknown codes** — resolving a code to a URL and issuing a 302. Why excluded: epic E4, depends on the write path (E3). Extension point: a wildcard route (`GET /{code}`) on the same router.
- **The `links` table and its persistence layer** — the `code -> url` table with created/expiry columns and its query/insert methods. Why excluded: it is epic E2. E1 ships the migration *runner* and an empty `migrations/` directory, not the table. Extension point: the first real migration file under `migrations/`, plus a `Links` repository in `internal/store`.
- **TTL expiry semantics** — `ttl_seconds`, expiry-on-read, expired-code 404. Why excluded: epic E5. Extension point: an `expires_at` column added by an E2/E5 migration and an expiry check in the read path.
- **Deep readiness / dependency health in `/healthz`** — a `/healthz` that actively pings Postgres on every call. Why excluded: for this slice a shallow liveness check plus fail-fast-at-boot already makes "api + db healthy" born-verifiable (see the HOW section). A deep readiness probe is operational hardening, aligned with epic E6. Extension point: a separate `GET /readyz` handler, or a DB ping inside the existing `/healthz` handler.

## 3. WHAT — vocabulary, types, and interfaces

Vocabulary:

| Term | Definition |
|---|---|
| api | The Go service container: the HTTP server that answers `/healthz` and, in later epics, the product endpoints. |
| db | The Postgres 16 container that backs the service. |
| liveness | A shallow check that the api process is up and serving; it does not probe dependencies. |
| migration runner | The golang-migrate invocation the api performs at startup to bring the schema to the latest version before serving. |

### Configuration (read from the environment at startup)

```
RECORD Config:
  database_url: string      # required; Postgres DSN, e.g. postgres://user:pass@db:5432/linkshortener?sslmode=disable
  http_addr:    string      # optional; TCP listen address, default ":8080"
  # Validation: database_url MUST be non-empty. A missing/blank database_url is a fatal
  # config error — the process logs a structured fatal message and exits non-zero WITHOUT
  # binding a port. "Structured fatal message" has a concrete shape: a single-line JSON object
  # with at least {"level":"fatal","event":"config_invalid","reason":<which variable and why>}.
  # The log MUST NOT contain the value of DATABASE_URL (or any secret): a present-but-invalid
  # DSN is reported by NAMING the offending variable and the validation failure, never by
  # echoing the value — the DSN carries the database password. This rule covers ALL fatal
  # boot logs (config, migrate, db-connect), not just config validation: raw pgx / golang-migrate
  # driver errors frequently embed the connection string, so the fatal log records a stable
  # failure-class event (config_invalid / migrate_failed / db_connect_failed) and MUST scrub any
  # DSN substring (host/user/password) from a wrapped driver error — never pass driver error text
  # through verbatim if it could contain the DSN.
```

- **Chosen:** configuration is a single `DATABASE_URL` DSN plus an optional `HTTP_ADDR` (default `:8080`), read once at startup into an immutable `Config`, with fail-fast validation — rationale: one DSN is the 12-factor-friendly shape both pgx and golang-migrate accept directly, and fail-fast on a missing DSN turns misconfiguration into an immediate, observable boot failure rather than a half-up server. Discrete `PG*` variables are rejected as an alternative in the Design Decision Rationale.
- **Chosen:** every fatal boot log is a single-line JSON object (`level`, `event`, `reason`) and NEVER contains a secret value — an invalid `DATABASE_URL` is reported by naming the variable and the validation failure, never by echoing the DSN — rationale: echoing an "invalid" DSN into container logs (stdout/stderr) would leak the database password into log aggregation, and the structured-error envelope already forbids secrets in HTTP responses, so the boot log needs the same rule; fixing the JSON shape also turns the previously-vague "structured fatal log" into a parseable, testable line.

### HTTP surface (this slice)

| Method + path | Response | Notes |
|---|---|---|
| `GET /healthz` | `200` with JSON body `{"status":"ok"}` | Liveness. Only reachable once config loaded, DB pool established, migrations applied. |
| any other route | `404` with a structured JSON error body | Proves the structured-error foundation; later epics register real routes. |
| a known path with a wrong method | `405` with the same structured JSON error body | Optional-but-recommended; same envelope. |

### Structured error envelope

```
RECORD ErrorResponse:
  error: {
    code:    string   # stable machine-readable token, e.g. "not_found", "method_not_allowed"
    message: string   # human-readable, safe to surface; no internal detail or secrets
  }
```

- **Chosen:** every non-2xx response the service originates uses the `{"error":{"code","message"}}` envelope with `Content-Type: application/json` — rationale: the brief requires structured errors as a production-shaping property; defining the envelope now (even though only 404/405 use it in E1) gives E3..E6 a single error shape to reuse instead of each endpoint inventing its own.

### Filesystem / package layout

```
.
├── cmd/api/main.go              # entrypoint: load config -> connect DB -> migrate -> serve
├── internal/config/config.go    # env parsing + validation -> Config
├── internal/httpapi/router.go   # ServeMux wiring, /healthz handler, structured 404/405 fallback
├── internal/store/store.go      # pgxpool construction + Ping (connection foundation only)
├── internal/migrate/migrate.go  # golang-migrate runner invoked at startup
├── migrations/                  # golang-migrate SQL dir; EMPTY in E1 (E2 adds the first table)
├── Dockerfile                   # multi-stage: build static Go binary -> minimal runtime image, runs as a non-root user
├── docker-compose.yml           # services: api + db, with healthchecks
├── go.mod / go.sum
└── .env.example                 # documents DATABASE_URL, HTTP_ADDR (never committed real secrets)
```

- **Chosen:** the service is laid out across `cmd/` and `internal/{config,httpapi,store,migrate}` packages rather than one file — rationale: the brief explicitly rejects a toy single-file service; separate packages give E2..E6 clear seams (a store package to add repositories to, a router to register handlers on).
- **Chosen:** the runtime image runs the api as a dedicated non-root user (a uid created in the final stage, `USER` set to it) rather than the default root — rationale: a Go static binary in a minimal image runs as root unless a user is explicitly set, so a container escape or RCE in a later epic would hold root inside the container; dropping to a non-root uid now is a one-line hardening that shrinks blast radius and costs nothing here (the static binary needs no privileged port — `HTTP_ADDR` defaults to `:8080`).

## 4. HOW — behaviour

Architecture and approach: the api is a Go 1.22 program using the standard library `net/http` with `ServeMux` method-and-path routing (no third-party web framework). Startup is a strict, fail-fast sequence; only after it completes does the server bind and accept traffic. Postgres access is via `pgxpool`. Schema is owned by golang-migrate and applied at startup. The whole thing is packaged as two docker-compose services with healthchecks.

Startup sequence (the api entrypoint):

```
PROCEDURE main():
  1. cfg := config.Load(environment)
     - IF cfg invalid (e.g. DATABASE_URL empty):
         log a structured fatal message (single-line JSON, naming the offending variable, NEVER its value);
         EXIT non-zero. Do NOT bind a port.
  2. runner := migrate.New(cfg.database_url, "migrations/")
     - Apply all pending migrations (golang-migrate `up`).
     - Treat "no change" (no pending migrations) as SUCCESS — E1's migrations dir is empty.
     - IF the runner errors for any other reason: log structured fatal (event "migrate_failed",
       DSN scrubbed — do NOT wrap the raw golang-migrate error if it embeds the connection string);
       EXIT non-zero.
  3. pool := store.Open(cfg.database_url)
     - Establish a pgxpool; Ping once to confirm connectivity.
     - IF Ping fails: log structured fatal (event "db_connect_failed", DSN scrubbed — do NOT wrap
       the raw pgx error verbatim if it embeds the connection string); EXIT non-zero. Do NOT bind a port.
  4. mux := httpapi.NewRouter(pool)   # /healthz + structured 404/405 fallback
  5. Start the HTTP server on cfg.http_addr.
     - Install a signal handler (SIGINT/SIGTERM) for graceful shutdown.
     - On SIGINT/SIGTERM: stop accepting new connections, call http.Server.Shutdown with a bounded
       timeout (<= 10s) to drain in-flight requests, then EXIT 0. A shutdown that completes draining
       within the timeout is "graceful"; the observable is exit code 0 within the timeout window.
```

Behaviour summary: the process either reaches step 5 with a validated config, a migrated schema, and a live database connection, or it exits non-zero before binding a port. There is no partially-up state in which `/healthz` answers while the datastore is unreachable.

`GET /healthz` behaviour:

```
PROCEDURE healthz(request):
  1. Return 200 with Content-Type application/json and body {"status":"ok"}.
  # Shallow by design: reaching this handler already implies config + DB + migrations succeeded at boot.
```

Structured not-found / method-not-allowed:

```
PROCEDURE fallback(request):
  1. IF no route matches the path:
     Return 404, Content-Type application/json,
       body {"error":{"code":"not_found","message":"resource not found"}}.
  2. IF a path matches but the method does not:
     Return 405 with {"error":{"code":"method_not_allowed","message":"method not allowed"}}.
```

Edge cases and error handling:

- **Missing/blank `DATABASE_URL`** — terminal at boot; structured fatal log, non-zero exit, no port bound. Not retryable.
- **Postgres not yet accepting connections when the api starts** — mitigated by compose ordering (`api depends_on db: condition: service_healthy`). If a connection still fails, the api exits non-zero and compose restarts it (`restart: on-failure`), so a slow db start self-heals within the 60s window rather than leaving the api half-up.
- **golang-migrate returns `ErrNoChange`** — not an error; treated as success (E1 has no pending migrations).
- **Any other migration error** — terminal at boot; non-zero exit. Not retryable without a fix.

Failure modes:

- **The failure:** `/healthz` returns 200 while the database is not actually reachable, making "api + db healthy" a false positive. **How you'd know:** the api container would be `healthy` in `docker compose ps` while the db container is not, or while a `psql` into the db fails. **What it means:** proceed — the fail-fast boot sequence (connect + Ping + migrate before bind) plus `api depends_on db: service_healthy` makes this state unreachable; a healthy api is contingent on a successful DB connection at boot, so the shallow `/healthz` is sound for this slice.
- **The failure:** the migration runner is wired but never actually executes, so E2's migrations would silently not run. **How you'd know:** the api startup logs would not show the golang-migrate step running (no "applying migrations" / "no change" line) before the server binds. **What it means:** proceed — E1 asserts the runner *executes* at boot (fail-fast: any migration error is a non-zero exit before bind, so a healthy api is contingent on the runner completing), which is the observable the harness E2 depends on is live. E1 does NOT assert the `schema_migrations` table, because golang-migrate creates no version table for an empty `migrations/` dir (`ErrNoChange`); the table first appears with E2's initial migration.

Anti-patterns:

- **Anti-pattern:** making `/healthz` ping Postgres on every request. Why: it turns a liveness probe into a readiness probe, couples container health to transient DB blips, and is out of scope for this slice (see OUT OF SCOPE).
- **Anti-pattern:** adding a `links` table (or any product table) migration in E1. Why: that is epic E2; E1 ships only the runner and an empty `migrations/` dir.
- **Anti-pattern:** binding the HTTP port before config/DB/migrations succeed. Why: it creates a half-up server that answers `/healthz` 200 while the datastore is unreachable, breaking the "healthy api implies reachable db" invariant.

### docker-compose topology

```
services:
  db:
    image: postgres:16
    environment: POSTGRES_USER / POSTGRES_PASSWORD / POSTGRES_DB
    healthcheck: pg_isready -U <user> -d <db>   # interval/timeout/retries tuned to be healthy well within 60s
  api:
    build: .                                    # the multi-stage Dockerfile
    environment: DATABASE_URL (points at db:5432), HTTP_ADDR
    depends_on: db: { condition: service_healthy }
    ports: "8080:8080"
    restart: on-failure
    healthcheck: GET /healthz returns 200        # e.g. wget/curl -f http://localhost:8080/healthz
```

- **Chosen:** the api's compose healthcheck is `GET /healthz` and the db's is `pg_isready`, with `api depends_on db: service_healthy` — rationale: this makes the DoD's "api + db healthy" a first-class, observable compose state (`docker compose ps` shows both `healthy`), and hands the downstream env-compose provisioner a health-checked surface to point at.

## Scenarios

> 2 holdout scenario(s) withheld from this view — evaluated code-blind against the running feature; full spec on the tracker.

Given the repository at E1, When `docker compose up -d --build` is run, Then within 60 seconds both the `db` and `api` containers report `healthy` (observable via `docker compose ps`) and `GET http://localhost:8080/healthz` returns HTTP 200 with body `{"status":"ok"}`.

Given the running api, When a client sends `GET /healthz`, Then the status is 200, the `Content-Type` is `application/json`, and the body is `{"status":"ok"}`.

Given the running api, When a client sends `POST /healthz` (a defined path with an unsupported method), Then the status is 405, the `Content-Type` is `application/json`, and the body is a structured error of the form `{"error":{"code":"method_not_allowed","message":"..."}}`.

Given the running stack, When the api has reached `healthy` after `docker compose up`, Then the api's startup logs show the golang-migrate runner ran to completion at boot (applying all pending migrations, or logging `no change` for E1's empty `migrations/` dir) AND that runner log line appears strictly BEFORE the "listening"/port-bound log line — the observable that the runner is wired, executes, and runs before the server binds (fail-fast ordering, not an async side-task). (E1 does NOT assert the `schema_migrations` bookkeeping table: golang-migrate returns `ErrNoChange` and creates no version table when `migrations/` is empty; that table first appears with E2's initial migration.)

Given the running api, When it is sent SIGTERM (e.g. `docker compose stop api`), Then it stops accepting new connections, drains in-flight requests, and exits with code 0 within the 10-second shutdown timeout (observable as a clean, non-timeout container stop / exit status 0), rather than being force-killed on timeout.

Given the api started with a present-but-invalid `DATABASE_URL` (a syntactically malformed DSN carrying a recognisable password token, e.g. `DATABASE_URL="not://a valid dsn:sup3rSecret@"`), When the process starts, Then it logs a structured fatal message and exits non-zero without binding the HTTP port, AND the entire captured log output does NOT contain the secret token (`sup3rSecret`) or the raw DSN value — the offending variable is named, its value is never echoed.

Given the built api image, When the running container's effective user is inspected (e.g. `docker compose exec api id -u`, or reading the image config `User`), Then it is a non-root uid (not `0`) — the `Dockerfile` sets `USER` to a dedicated non-root account.

Given the api started with a well-formed `DATABASE_URL` that points at an UNREACHABLE database (e.g. a dead host/port, or the `db` service stopped, with a recognisable password token in the DSN), When the process starts, Then the Ping fails, the api logs a structured fatal message (event `db_connect_failed`) and exits non-zero WITHOUT binding the HTTP port (verified by the connection-refused oracle on `HTTP_ADDR`), AND the captured log output does NOT contain the password token or the raw DSN — the wrapped pgx driver error is scrubbed of the connection string. This exercises the failed-Ping path and the "healthy api implies a reachable db" invariant.

- The api process MUST NOT bind its HTTP port unless config loaded, the migration runner ran, and the Postgres connection succeeded at boot (a healthy api implies a reachable db).

## 6. Design decision rationale

**Which language/runtime for the service?**
- TypeScript on Node 20 — pros: brief's first-listed option, huge ecosystem, familiar REST+Postgres path. Cons: larger runtime image, slower cold start, more moving parts to reach a minimal healthy container.
- Go 1.22 — pros: single static binary → smallest api image and fastest cold start (the direct lever on the 60s `/healthz` DoD); Go 1.22's `net/http.ServeMux` gives method+path routing with no framework dependency; mature Postgres driver (pgx) and migration tool (golang-migrate). Cons: slightly more boilerplate for JSON handling than a TS framework.
- **Chosen:** Go 1.22 on stdlib `net/http` — the build-biased best fit: minimal dependencies, fastest path to a healthy container, and everything built and run locally under docker-compose with no managed service. (See the architecture-proposal block in section 9.)

**How is the schema owned?**
- Ad-hoc DDL at startup — pros: no tool. Cons: not versioned, not reviewable, no bookkeeping.
- golang-migrate SQL migrations applied on start — pros: deterministic, versioned, reviewable; gives E2 a first-class place to add the `links` table and the `schema_migrations` bookkeeping table it will create on first real migration. Cons: one dependency. (In E1 the `migrations/` dir is empty, so the runner returns `ErrNoChange` and creates no version table yet — E1 verifies the runner *ran*, not that a table exists.)
- **Chosen:** golang-migrate, run to latest at startup, `ErrNoChange` treated as success — rationale: the migration is a first-class artifact per the brief, and E1 needs the *runner* live even though the first real migration is E2's.

**Configuration shape?**
- Discrete `PGHOST/PGUSER/...` variables — pros: granular. Cons: more surface, more validation, diverges from what pgx/migrate accept directly.
- Single `DATABASE_URL` DSN (+ `HTTP_ADDR`) — pros: one value both pgx and golang-migrate consume directly; 12-factor-friendly; trivial fail-fast check. Cons: less granular.
- **Chosen:** `DATABASE_URL` + optional `HTTP_ADDR`, validated fail-fast at boot.

**Is `/healthz` shallow liveness or deep readiness?**
- Deep (ping DB each call) — pros: reflects live DB state. Cons: readiness, not liveness; couples health to transient blips; out of scope for the skeleton.
- Shallow (200 once serving) + fail-fast boot — pros: simple, and the boot sequence + `depends_on: service_healthy` already make "api + db healthy" born-verifiable. Cons: does not catch a DB that drops after boot (acceptable for this slice; a `/readyz` is the named extension point).
- **Chosen:** shallow liveness plus fail-fast boot; deep readiness is an extension point for later operational hardening (E6).

## 7. Open questions and assumptions

Open questions: none. Every decision in this slice is closed above; the slice is deliberately small and its choices follow directly from the brief and the architecture proposal.

Assumptions:

- **Assumes:** a Docker engine with the Compose v2 plugin is available in the environment that runs the DoD check (`docker compose up`, `docker compose ps`). Validation: `docker compose version` succeeds before the DoD is exercised.
- **Assumes:** `postgres:16` is pullable (or already cached) in that environment. Validation: `docker compose pull db` (or the first `up`) succeeds.
- **Assumes:** outbound module fetching (Go module proxy) is available at image build time, or modules are vendored. Validation: `docker compose build api` succeeds.

## 8. DONE — definition of done

### From WHY
- [ ] `docker compose up -d --build` brings both `api` and `db` containers to `healthy` within 60 seconds, observable via `docker compose ps`.
- [ ] `GET /healthz` returns 200 within 60 seconds of `docker compose up`.

### From WHAT (config)
- [ ] Config is read from the environment at startup: `DATABASE_URL` (required) and `HTTP_ADDR` (default `:8080`).
- [ ] A missing/blank `DATABASE_URL` causes a structured fatal log (single-line JSON with `level`/`event`/`reason`) and a non-zero exit with no port bound.
- [ ] No fatal (or any) boot log contains the value of `DATABASE_URL` or any secret; an invalid DSN is reported by naming the variable, not by echoing its value — verified for an unset DSN, a present-but-invalid (malformed) DSN, AND an unreachable-DB (`db_connect_failed`) case, asserting the captured logs contain no secret token from the DSN. Wrapped pgx / golang-migrate driver errors are scrubbed of the connection string (never passed through verbatim).

### From WHAT (HTTP surface + errors)
- [ ] `GET /healthz` returns 200, `Content-Type: application/json`, body `{"status":"ok"}`.
- [ ] An undefined route returns 404 with a structured JSON body `{"error":{"code":"not_found","message":...}}`.
- [ ] A defined path with an unsupported method returns 405 with a structured JSON body `{"error":{"code":"method_not_allowed","message":...}}`.
- [ ] A structured error envelope `{"error":{"code","message"}}` exists and is used by the 404 (and 405) fallback.

### From WHAT (layout)
- [ ] The service is organised across `cmd/api` and `internal/{config,httpapi,store,migrate}` packages (not a single file).
- [ ] A `migrations/` directory exists and is wired into the startup runner (empty in E1).

### From HOW (behaviour)
- [ ] The api runs the golang-migrate runner at startup (before binding the HTTP port), treating `ErrNoChange` for E1's empty `migrations/` dir as success; the runner's execution is observable in the startup logs, and its log line appears strictly BEFORE the port-bound log line (fail-fast ordering); any migration error is a non-zero exit before bind. (E1 does NOT assert a `schema_migrations` table — golang-migrate creates none for an empty migrations dir; that is an E2 observable.)
- [ ] The api establishes a pgxpool and Pings Postgres successfully at boot; a failed Ping is a non-zero exit — verified by a scenario that points `DATABASE_URL` at an unreachable DB and asserts a `db_connect_failed` fatal exit with no port bound.
- [ ] The api does not bind its HTTP port unless config, the migration runner, and the DB connection all succeed — verified by its oracle: during any boot-failure window a TCP connect to `HTTP_ADDR` is refused (connection refused), so the port is never bound.
- [ ] SIGINT/SIGTERM triggers a graceful HTTP shutdown: new connections stop, in-flight requests drain, and the process exits 0 within a bounded (<= 10s) timeout — verified by a clean stop / exit status 0 rather than a force-kill on timeout.

### From HOW (compose topology)
- [ ] `docker-compose.yml` defines `api` and `db` services; `db` has a `pg_isready` healthcheck; `api` has a `GET /healthz` healthcheck and `depends_on db: condition: service_healthy`.
- [ ] A multi-stage `Dockerfile` builds a static Go binary into a minimal runtime image, and the final image runs the api as a dedicated non-root user (`USER` is set, not root) — verified by inspecting the running container's effective uid (`id -u` != 0).
- [ ] `.env.example` documents `DATABASE_URL` and `HTTP_ADDR`; no real secrets are committed.

Integration smoke test:

```
1. docker compose up -d --build
2. Poll GET http://localhost:8080/healthz until 200 or 60s elapse.
3. Assert docker compose ps shows api healthy AND db healthy.
4. Assert GET /does-not-exist returns 404 with {"error":{"code":"not_found",...}}.
5. Assert POST /healthz returns 405 with {"error":{"code":"method_not_allowed",...}}.
6. Assert the api startup logs show the golang-migrate runner ran at boot AND before the port-bound line (e.g. docker compose logs api shows a "migrat"/"no change" line ordered before the "listening" line) — NOT that a schema_migrations table exists (E1's migrations dir is empty; that table is an E2 artifact).
7. docker compose stop api; assert the api container exited 0 (graceful drain, not a force-kill on timeout).
8. docker compose down -v
```

## 9. Architecture proposal (carried verbatim — downstream consumers read it from here)

The following validated `faff-contract:architecture-proposal` block is the architecture decided for this stack-establishing epic. Downstream epics (E2..E6) and the holdout env step read it from this spec; do not re-decide the stack.

```faff-contract:architecture-proposal
{
  "chosen_architecture": "Modular Go 1.22 service on the standard-library net/http router, backed by Postgres 16, packaged as a two-service docker-compose (api + db). Schema is owned by golang-migrate SQL migrations applied on api start; all config is read from the environment at startup; all error responses are structured JSON. Layered into cmd/ (entrypoint), internal/config, internal/httpapi (handlers incl. /healthz), internal/store (pgxpool), and migrations/ — not a single-file toy.",
  "rationale": "The brief fixes the stack envelope (Postgres datastore, docker-compose api+db, GET /healthz, env-config, a migration, structured errors, not a toy single-file) and offers TypeScript on Node 20 or Go 1.22 as the build-biased best fit. No infra profile is minable on this fresh repo (faff profile show reports no profile), so the choice is made from the brief alone. Go 1.22 is chosen: a single static binary yields the smallest api image and the fastest cold start, the direct lever on the born-verifiable DoD (api+db healthy and GET /healthz returns 200 within 60s of docker compose up). Go 1.22's enhanced net/http.ServeMux gives method-and-path routing from the standard library, so the HTTP surface needs no third-party web framework; pgx/pgxpool is the mature Postgres driver with pooling; golang-migrate applies the schema deterministically on start and gives later epics (the links table) a first-class place to land. Everything is built and run locally under docker-compose with no managed service, so the recommendation is build.",
  "adr_candidates": [
    { "title": "Go 1.22 on the standard-library net/http for the service runtime", "decision": "Build the api in Go 1.22 using net/http.ServeMux method-and-path routing; no third-party web framework", "rationale": "single static binary means the smallest image and fastest healthy startup, the 60s /healthz DoD lever; Go 1.22 routing patterns remove the need for a router dependency" },
    { "title": "Postgres 16 as the persistence datastore, accessed via pgx", "decision": "Use Postgres 16 as the only datastore, connection-pooled through pgxpool", "rationale": "the brief fixes a real persistent store; pgx is the mature, well-supported driver and provides the pooling a long-running service needs" },
    { "title": "Schema owned by golang-migrate, applied on api start", "decision": "Version the schema as golang-migrate SQL migrations and run them to latest at api startup before the server accepts traffic", "rationale": "deterministic, reviewable schema evolution; keeps the migration a first-class artifact rather than ad-hoc DDL and gives later epics a place to add the links table" },
    { "title": "Two-service docker-compose topology with healthchecks", "decision": "Package as docker-compose services api and db; db carries a pg_isready healthcheck, api depends_on db being healthy, api exposes GET /healthz as its own healthcheck", "rationale": "makes 'api+db healthy' born-verifiable and gives the env-compose provisioner a health-checked surface to point at" },
    { "title": "12-factor environment configuration with fail-fast validation", "decision": "Read all configuration (Postgres DSN or parts, HTTP listen port) from the environment at startup and fail fast on missing/invalid values; emit structured JSON errors", "rationale": "production-shaped per the brief; env-config is exactly what the env-compose step seeds and the holdout evaluator exercises" }
  ],
  "assumptions": [
    "No infra profile is minable on this fresh repo (faff profile show reports no profile); the architecture is fitted from the brief's stated stack preference alone.",
    "Postgres 16 is an acceptable concrete pin for the brief's generic 'Postgres' datastore; no acceptance criterion constrains the exact minor version.",
    "Traffic is single-instance and single-region for this slice; no clustering, read-replica, or horizontal-scale topology is implied."
  ],
  "recommendation": "build"
}
```

## ADR promotion intent

- **Go 1.22 on the standard-library net/http for the service runtime** (from section 6 / section 9) — decision: build the api in Go 1.22 with `net/http.ServeMux` routing, no web framework.
- **Postgres 16 as the persistence datastore via pgx** (section 9) — decision: Postgres 16 as the only datastore, pooled through pgxpool.
- **Schema owned by golang-migrate, applied on start** (section 6 / section 9) — decision: versioned SQL migrations run to latest at startup, `ErrNoChange` treated as success.
- **Two-service docker-compose topology with healthchecks** (section 4 / section 9) — decision: `api` + `db` services, `pg_isready` and `GET /healthz` healthchecks, `depends_on: service_healthy`.
- **12-factor env configuration with fail-fast validation and structured JSON errors** (section 3 / section 9) — decision: `DATABASE_URL` + `HTTP_ADDR` from the environment, validated at boot; `{"error":{"code","message"}}` envelope.

## Methodology critique

Methodology: faffter-dark-methodology-agile-delivery

- **Right-sized (principle 4):** No issues. E1 is a single, cohesive 1-3 day unit. Config, compose topology, and the migration runner are not independent concerns that could ship separately: the container topology is not healthy unless all three are present, so they form one deliverable, not a split candidate. It is not oversized (no product endpoints), and not a micro-fragment.
- **Workstream fit (principles 1 + 5):** No issues. E1 sits in Project P1 "Core shorten to redirect on a real datastore", an outcome-named container, as its foundation slice. The slice encodes one outcome (a runnable, health-checked stack), so it is cohesive.
- **Surfaced dependencies (principle 6):** No issues. E1 is blocked-by nothing (it is the first slice), and the downstream chain E2 blocked-by E1 is already explicit in the roadmap skeleton. No implicit dependency is hidden in the spec.
- **Risk profile (principle 7):** Positive. E1 front-loads the main integration risk of the whole service, the docker-compose api+db topology and migrate-on-start, into the first slice. That is exactly the early de-risking principle 7 calls for; no separate de-risking spike is warranted because E1 is itself the de-risking slice.

confidence: high

```faff-contract:spec-readiness
{ "confidence": "high",
  "decisions": [
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "chosen" },
    { "marker": "assumes" }
  ] }
```