## Adversarial findings — gemini/models/gemma-4-31b-it (chain[2], host: config)

<thought>Adversarial QA spec refuter.
Review a SPEC (diff) for verifiability.
Break the spec's verifiability. Find why a build might be "correct" according to the spec but actually broken, or why "DONE" cannot be decided.

        *   Born-verifiable? (Pass/fail line)
        *   Scenario coverage? (Main behavior + edge cases)
        *   Acceptance gap? (DONE vs Scenarios)
        *   Oracle problem? (Concrete expected results)
        *   Regression surface? (Implicit assumptions not asserted)
        *   Defer to Ratified Scope/Goals.

    *   *Goal:* Packaged service skeleton (docker-compose: api + Postgres), `GET /healthz`, env config.
    *   *Key Requirements:*
        *   `docker compose up` -> healthy within 60s.
        *   `GET /healthz` -> 200 OK.
        *   Real Postgres (not in-memory).
        *   Env config (`DATABASE_URL` required, `HTTP_ADDR` optional).
        *   Migration runner (golang-migrate) runs at boot.
        *   Structured JSON errors (404/405).
        *   Fail-fast boot (no port bind if DB/config/migration fails).
        *   Graceful shutdown (SIGINT/SIGTERM).

    *   *Scenario 1:* `docker compose up -d --build` -> healthy within 60s, `/healthz` 200. (Covers main path).
    *   *Scenario 2:* `GET /healthz` -> 200, JSON body `{"status":"ok"}`. (Covers liveness).
    *   *Scenario 3 (Holdout):* `GET /does-not-exist` -> 404, structured JSON. (Covers error envelope).
    *   *Scenario 4 (Holdout):* `DATABASE_URL` unset -> fatal log, non-zero exit, no port bound. (Covers fail-fast config).

    *   *DONE Item:* "SIGINT/SIGTERM triggers a graceful HTTP shutdown."
        *   *Scenario:* None.
        *   *Oracle:* How do we verify "graceful"? Does it mean it closes connections? Does it mean it exits with 0? Does it mean it takes < X seconds?
        *   *Verdict:* Major gap.

    *   *DONE Item:* "A structured error envelope `{"error":{"code","message"}}` exists and is used by the 404 (and 405) fallback."
        *   *Scenario:* Only 404 is tested in the holdout. 405 (Method Not Allowed) is mentioned in the DONE list and the "WHAT" section but has no scenario.
        *   *Verdict:* Minor gap.

    *   *DONE Item:* "The api applies migrations at startup... golang-migrate's `schema_migrations` bookkeeping table is present in the db after `up`."
        *   *Scenario:* The "Integration smoke test" doesn't explicitly check the DB for the table. The main scenario just checks `/healthz`.
        *   *Oracle:* The spec says "observable via `docker compose ps`" and `/healthz`. But the *proof* of the migration runner is the table. If the developer just hardcoded `/healthz` to return 200 without actually running the migration runner, the main scenario passes.
        *   *Verdict:* Major. The "Integration smoke test" in section 8 is a list of steps, but the "Scenarios" section (section 5) is where the formal acceptance criteria usually live. The smoke test mentions `/healthz` and 404, but not the DB table.

    *   *DONE Item:* "The api does not bind its HTTP port unless config, migrations, and the DB connection all succeed."
        *   *Scenario:* Only `DATABASE_URL` unset is covered. What about a DB that is reachable but migrations fail? Or a DB that is unreachable (though `depends_on` handles some of this, the *api's* behavior on failure is what's being asserted).
        *   *Verdict:* Minor/Major. The "fail-fast" is a core design principle.

    *   *DONE Item:* "A missing/blank `DATABASE_URL` causes a structured fatal log..."
        *   *Oracle:* What is a "structured fatal log"? JSON? A specific prefix? If it's just `log.Fatal("missing db url")`, is that "structured"?
        *   *Verdict:* Minor.

    *   *Regression Surface:* The spec assumes that if `/healthz` is 200, the DB is reachable. It explicitly rejects a deep health check. However, if the DB goes down *after* boot, `/healthz` stays 200. This is acknowledged as an "Anti-pattern" and "Out of Scope", so I should not object to it as a missing feature, but I should check if the *boot* verification is tight.

    *   *Critical/Major 1: Graceful Shutdown.* No scenario, no definition of "graceful".
    *   *Major 2: Migration Runner Verification.* The "Integration smoke test" is a guide, but the formal "Scenarios" don't verify the `schema_migrations` table. A build could fake `/healthz` and pass.
    *   *Minor 3: 405 Method Not Allowed.* Defined in DONE/WHAT, but no scenario.
    *   *Minor 4: "Structured" Log.* Vague definition for the fail-fast log.

    *   *Graceful Shutdown:* `done-from-how-behaviour` -> no scenario.
    *   *Migration Table:* `done-from-how-behaviour` -> no scenario in section 5.
    *   *405 Error:* `done-from-what-http-surface-errors` -> no scenario.
    *   *Structured Log:* `done-from-what-config` -> "structured fatal log" is vague.</thought>## Refutation — QA

### major: Graceful shutdown is unverifiable
- claim: There is no observable criterion to determine if a shutdown is "graceful" or if it happened at all.
- evidence: The DONE item "SIGINT/SIGTERM triggers a graceful HTTP shutdown" has no corresponding scenario in section 5.
- predicted_consequence: A build that simply crashes or is killed instantly by Docker will satisfy the spec, as there is no test asserting the handling of signals or the closing of listeners.
- spec_anchor: done-from-how-behaviour

### major: Migration runner success is not asserted
- claim: The spec asserts the migration runner must execute, but the acceptance scenarios only check the HTTP surface.
- evidence: The DONE item "golang-migrate's `schema_migrations` bookkeeping table is present in the db after `up`" is not covered by any scenario in section 5.
- predicted_consequence: A developer could hardcode `GET /healthz` to return 200 without actually implementing the migration runner, and the build would pass all formal scenarios.
- spec_anchor: done-from-how-behaviour

### minor: Missing scenario for 405 Method Not Allowed
- claim: The spec requires a structured error for wrong methods, but provides no scenario to verify it.
- evidence: The DONE item "A structured error envelope... is used by the 404 (and 405) fallback" is only partially covered (404 is in the holdout scenario; 405 is missing).
- predicted_consequence: The 405 handler may be missing or return a non-structured response (e.g., plain text) without failing the build.
- spec_anchor: done-from-what-http-surface-errors

### minor: "Structured" log is undefined
- claim: The term "structured fatal log" is vague and lacks a pass/fail definition.
- evidence: The DONE item "A missing/blank `DATABASE_URL` causes a structured fatal log..."
- predicted_consequence: It is impossible to decide if a log is "structured" (e.g., does it require JSON? A specific key-value format? Or just a formatted string?).
- spec_anchor: done-from-what-config
