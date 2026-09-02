## Adversarial findings — gemini/models/gemma-4-31b-it (chain[2], host: config)

<thought>Adversarial QA spec refuter.
Review a SPEC (diff) for verifiability.
Break the spec's verifiability. Find gaps in testability, scenario coverage, acceptance gaps, oracle problems, and regression surfaces.
`## Ratified scope` (goals and non-goals).

    *   *Goal:* Running service, mints code, redirects. (E1 is just the skeleton, so this is the long-term goal).
    *   *Goal:* Codes survive restart (datastore is real). (E1 sets up the datastore).
    *   *Goal:* TTL honored. (E1 excludes this).
    *   *Goal:* Packaged stack starts healthy, answers liveness check quickly. (E1's primary focus).
    *   *Goal:* Automated test suite covers every acceptance scenario. (E1's primary focus).
    *   *Non-goals:* Auth, custom aliases, analytics, UI, deduplication.

    *   *Purpose:* E1 is a "packaged service skeleton".
    *   *Key Components:* `docker-compose` (api + Postgres), `GET /healthz`, env config, migration runner, structured JSON errors.
    *   *Out of Scope:* `POST /shorten`, `GET /{code}`, `links` table, TTL.
    *   *Config:* `DATABASE_URL` (required), `HTTP_ADDR` (optional).
    *   *HTTP Surface:* `GET /healthz` (200), others (404), wrong method (405).
    *   *Error Envelope:* `{"error":{"code","message"}}`.
    *   *Startup Sequence:* Load config $\rightarrow$ Migrate $\rightarrow$ Open DB $\rightarrow$ Serve.
    *   *Graceful Shutdown:* SIGINT/SIGTERM $\rightarrow$ drain $\rightarrow$ exit 0 within 10s.
    *   *Security:* Non-root user, no secrets in logs.

    *   *Scenario 1 (Happy Path):* `docker compose up` $\rightarrow$ healthy within 60s $\rightarrow$ `/healthz` 200. (Verifiable).
    *   *Scenario 2 (Healthz):* `GET /healthz` $\rightarrow$ 200, JSON, `{"status":"ok"}`. (Verifiable).
    *   *Scenario 3 (404):* Undefined route $\rightarrow$ 404, JSON, structured error. (Verifiable).
    *   *Scenario 4 (405):* Defined path, wrong method $\rightarrow$ 405, JSON, structured error. (Verifiable).
    *   *Scenario 5 (Migration Runner):* Startup logs show runner ran (or "no change") before bind. (Verifiable via logs).
    *   *Scenario 6 (Graceful Shutdown):* SIGTERM $\rightarrow$ exit 0 within 10s. (Verifiable via exit code).
    *   *Scenario 7 (Secret Leak/Invalid DSN):* Invalid DSN $\rightarrow$ fatal log, non-zero exit, no port bound, NO secret in logs. (Verifiable).
    *   *Scenario 8 (Non-root):* `id -u` != 0. (Verifiable).
    *   *Holdout (Missing DSN):* Unset `DATABASE_URL` $\rightarrow$ fatal log, non-zero exit, port refused. (Verifiable).

    *   *Born-verifiable?* Most items are concrete. "Structured fatal message" is defined as a JSON object. "Graceful" is defined as exit 0 within 10s.
    *   *Scenario coverage?*
        *   Empty `migrations/` is covered.
        *   Invalid DSN is covered.
        *   Missing DSN is covered.
        *   DB not ready is covered (via `depends_on` and `restart: on-failure`).
        *   *Wait:* What about a DB that is *reachable* but the *credentials* are wrong? The spec says "invalid DSN" (syntactically malformed) and "missing DSN". It doesn't explicitly have a scenario for "Correct syntax, wrong password". The "Open DB" step says "IF Ping fails: log structured fatal; EXIT non-zero". This is a failure mode, but is there a scenario to verify it?
    *   *Acceptance gap?*
        *   DONE item: "The api establishes a pgxpool and Pings Postgres successfully at boot; a failed Ping is a non-zero exit."
        *   Scenario: There is no scenario that *triggers* a failed Ping (e.g., wrong password) to verify the non-zero exit and structured log.
    *   *Oracle problem?*
        *   "Startup logs show the golang-migrate runner ran to completion... before the HTTP server bound."
        *   Oracle: `docker compose logs api | grep -i "migrat"`. This is a bit fuzzy. Does it check the *order*? The spec says "before the HTTP server bound". To verify the *order*, the log must have a timestamp or a specific sequence. The oracle provided in the smoke test just greps for the string. It doesn't verify the sequence.
    *   *Regression surface?*
        *   The spec asserts that a healthy API implies a reachable DB. It does this by making the DB connection a prerequisite for binding the port.
        *   If the DB goes down *after* boot, `/healthz` still returns 200. The spec explicitly calls this out as an "Anti-pattern" and "Out of Scope" (deep readiness). This is a conscious decision, not a gap.

    *   *Objection 1 (Major):* Missing scenario for DB connection failure (wrong credentials). The DONE criteria require that a failed Ping results in a non-zero exit and structured log, but no scenario tests this.
    *   *Objection 2 (Minor):* The "migration runner ran *before* bind" oracle is weak. Grepping for "migrat" doesn't prove it happened *before* the server started listening.

    *   *Objection 1:*
        *   Claim: DB connection failure (e.g., wrong password) cannot be verified.
        *   Evidence: DONE item "The api establishes a pgxpool and Pings Postgres successfully at boot; a failed Ping is a non-zero exit."
        *   Consequence: A build could fail to Ping the DB but still bind the port (violating the "healthy implies reachable" invariant) and the test suite wouldn't catch it because no scenario triggers a Ping failure.
        *   Anchor: `done-definition-of-done` (or `how-behaviour`). Let's use `done-definition-of-done`.

    *   *Objection 2:*
        *   Claim: The sequence of migration $\rightarrow$ bind is not concretely verifiable.
        *   Evidence: Scenario "api's startup logs show the golang-migrate runner ran... before the HTTP server bound" and Smoke Test step 6.
        *   Consequence: The build could bind the port first and then run migrations in a goroutine; the `grep` would still pass, but the "fail-fast" invariant is broken.
        *   Anchor: `scenarios`.

    *   *Check against Ratified Scope:*
        *   Goal: "A running service... redirects". (E1 is just the skeleton).
        *   Goal: "Codes survive restart... datastore is real". (E1 sets up the datastore).
        *   Goal: "Automated test suite covers every acceptance scenario". (This is where my objections land).

    *   *Refining Objection 1:* Is "invalid DSN" the same as "wrong password"? No. "Syntactically malformed" (Scenario 7) is a config validation error. "Wrong password" is a connection error (Ping failure). These are different stages of the startup sequence.

    *   *Refining Objection 2:* The spec says "the observable is... the api startup logs". If the logs are timestamped, it's verifiable. But the smoke test just says `grep`. I'll flag it as a minor oracle problem.

    *   *Wait, look at the "Failure modes" section:*
        *   "The failure: the migration runner is wired but never actually executes... How you'd know: the api startup logs would not show the golang-migrate step running... before the server binds."
        *   This confirms the *intent* is to verify the sequence. The *scenario* and *smoke test* just use `grep`. This is the gap.</thought>## Refutation — QA

### major: missing scenario for database connection failure
- claim: the requirement that a failed database Ping results in a non-zero exit and structured log cannot be verified because no scenario triggers a connection failure.
- evidence: the DONE item "The api establishes a pgxpool and Pings Postgres successfully at boot; a failed Ping is a non-zero exit" has no corresponding scenario.
- predicted_consequence: a build could implement the DB connection as a background task or ignore Ping failures (binding the port anyway), violating the "healthy api implies reachable db" invariant, and the test suite would still pass.
- spec_anchor: done-definition-of-done

### minor: weak oracle for migration sequence
- claim: the assertion that migrations run *before* the server binds is not concretely verifiable via the provided oracle.
- evidence: the scenario "api's startup logs show the golang-migrate runner ran... before the HTTP server bound" and the smoke test step 6 (`docker compose logs api | grep -i "migrat"`) only verify the *presence* of the log, not the *sequence* relative to the port binding.
- predicted_consequence: a build could bind the port and then run migrations asynchronously; the `grep` would pass, but the "fail-fast" boot sequence is broken.
- spec_anchor: scenarios
