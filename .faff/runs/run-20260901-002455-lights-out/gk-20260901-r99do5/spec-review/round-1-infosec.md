## Adversarial findings — gemini/models/gemma-4-31b-it (chain[2], host: config)

<thought>Adversarial Infosec Spec Refuter.
Review a SPEC (diff) for security/safety risks.
Break the approach, find risks, do not rubber-stamp.

        *   Authn/Authz
        *   Secrets
        *   Input Surface
        *   Blast Radius
        *   Failure-as-bypass

        *   `## Ratified scope`: Defer to it.
        *   `### Ratified goals`: Defer to the goal itself, but critique the *implementation*.
        *   `critical`: Real exploit, data loss, fail-open.
        *   `major`: Genuine security defect.
        *   `minor`: Hardening gap.
        *   `observation`: Advisory.
One block, strongest first, specific format.

    *   *Goal:* A running service that mints codes for URLs and redirects.
    *   *Non-goals:* Authn/Authz, Custom aliases, Analytics, UI, Deduplication.
    *   *Context:* This is Epic E1 (Skeleton). It sets up the Docker Compose (API + Postgres), `/healthz`, env config, and migration runner. It *does not* implement the actual shortening/redirect logic yet.

    *   *Configuration:* `DATABASE_URL` (required), `HTTP_ADDR` (optional).
    *   *HTTP Surface:* `GET /healthz` (200 OK), others (404/405 structured JSON).
    *   *Startup Sequence:* Load config $\rightarrow$ Run migrations $\rightarrow$ Open DB pool $\rightarrow$ Ping DB $\rightarrow$ Start HTTP server.
    *   *Docker Compose:* `db` (Postgres 16) and `api` (Go 1.22). `api` depends on `db` being healthy.
    *   *Secrets Handling:* `.env.example` provided; real secrets not committed.

    *   *Authn/Authz:* Non-goal. No endpoints yet except `/healthz`. No risk here for E1.
    *   *Secrets:*
        *   `DATABASE_URL` contains credentials (`postgres://user:pass@db:5432/...`).
        *   The spec says: "A missing/blank database_url is a fatal config error — the process logs a structured message and exits non-zero".
        *   *Risk:* Does the "structured message" log the `DATABASE_URL` itself? If the config is "invalid" (e.g., wrong format but not empty), does it log the value?
        *   *Risk:* The `DATABASE_URL` is passed as an environment variable. In some container environments, env vars are visible to other processes or via `/proc`. (Standard practice, but worth noting).
        *   *Risk:* The `DATABASE_URL` is passed to `golang-migrate` and `pgxpool`.
    *   *Input Surface:*
        *   `GET /healthz`: No input.
        *   `HTTP_ADDR`: User-controlled env var. Could it be used for something weird? (e.g., binding to a privileged port if run as root, though it's in a container).
        *   `DATABASE_URL`: User-controlled env var. Could it be used for connection string injection? (Unlikely to be an exploit *against* the service, more of a config error).
    *   *Blast Radius:*
        *   The API runs as a process in a container.
        *   The DB runs in a container.
        *   If the API is compromised, it has the DB credentials.
    *   *Failure-as-bypass:*
        *   The spec emphasizes "fail-fast". If DB is down, API doesn't bind. This is "fail-safe" regarding the health check.
        *   Migration runner: If it fails, API doesn't start. Fail-safe.

    *   *Observation 1: Secret Logging.* The spec says "log a structured fatal message" when config is invalid. If the validation logic logs the invalid value of `DATABASE_URL` to stdout/stderr, the DB password ends up in the container logs. This is a classic secret leak.
    *   *Observation 2: Database URL in Env.* While standard, `DATABASE_URL` in env vars is less secure than secrets files/vaults. However, for a skeleton/dev setup, this is usually acceptable.
    *   *Observation 3: Root User in Container.* The Dockerfile is "multi-stage... minimal runtime image". It doesn't explicitly mention a non-root user. Running as root in a container increases blast radius if the API is compromised.

    *   *Secret Logging:* This is a `major` or `minor` risk. If the spec says "log a structured message" upon "invalid" config, and the config is the DSN, there's a high probability a developer will log the DSN to help debug why it's invalid.
    *   *Container User:* `minor`. Standard hardening.

    *   *Wait, look at the "Structured error envelope":* `{"error":{"code","message"}}`. The spec says "message: human-readable, safe to surface; no internal detail or secrets". This applies to *HTTP responses*.
    *   *But what about logs?* "log a structured fatal message". It doesn't explicitly forbid secrets in logs, but the HTTP section explicitly does. This inconsistency is a red flag.

    *   *Severity:* `major`.
    *   *Claim:* The spec mandates logging "structured fatal messages" when `DATABASE_URL` is invalid, but does not forbid logging the value of the invalid variable. Since `DATABASE_URL` contains the database password, this creates a high risk of secrets leaking into container logs.
    *   *Evidence:* Section 3 (Configuration) and Section 4 (Startup sequence).
    *   *Consequence:* Database credentials persisted in plaintext in log aggregation systems.
    *   *Anchor:* `configuration-read-from-the-environment-at-startup`.

    *   *Severity:* `minor`.
    *   *Claim:* The Dockerfile spec mentions a "minimal runtime image" but does not specify a non-root user.
    *   *Evidence:* Section 3 (Filesystem / package layout) -> Dockerfile.
    *   *Consequence:* Increased blast radius; a remote code execution (RCE) in the API would grant root privileges within the container.
    *   *Anchor:* `filesystem-package-layout`.

    *   *Authn/Authz:* Non-goal.
    *   *Secrets:* Covered (logging).
    *   *Input Surface:* `/healthz` is safe. `HTTP_ADDR` is internal config.
    *   *Blast Radius:* Covered (root user).
    *   *Failure-as-bypass:* The "fail-fast" approach is actually very secure here.

    *   Check anchors.
    *   Check severities.
    *   Check format.</thought>## Refutation — infosec

### major: Secret leakage via fatal configuration logs
- claim: The spec requires the API to "log a structured fatal message" and exit when the `DATABASE_URL` is missing or invalid. Because `DATABASE_URL` is a DSN containing a plaintext password, there is a high risk that the "structured message" will include the invalid value for debugging purposes, leaking database credentials into the container logs (stdout/stderr). While the spec explicitly forbids secrets in *HTTP* error responses, it provides no such constraint for *internal logs*.
- evidence: Section 3 (Configuration) states: "A missing/blank database_url is a fatal config error — the process logs a structured message and exits non-zero". Section 4 (Startup sequence) repeats: "IF cfg invalid... log a structured fatal message; EXIT non-zero."
- predicted_consequence: Database credentials are persisted in plaintext within log aggregation systems or container runtime logs.
- spec_anchor: configuration-read-from-the-environment-at-startup

### minor: Excessive privilege in runtime container
- claim: The spec defines a multi-stage Dockerfile for a "minimal runtime image" but does not specify the use of a non-root user. By default, Go binaries in minimal images (like `scratch` or `alpine`) run as root unless a user is explicitly created and switched to.
- evidence: Section 3 (Filesystem / package layout) describes the `Dockerfile` as "multi-stage: build static Go binary $\rightarrow$ minimal runtime image".
- predicted_consequence: Increased blast radius; any future vulnerability leading to remote code execution (RCE) in the API would grant the attacker root privileges within the container.
- spec_anchor: filesystem-package-layout
