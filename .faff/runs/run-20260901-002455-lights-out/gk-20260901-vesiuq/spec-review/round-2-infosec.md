## Adversarial findings — openai/deepseek/deepseek-v4-flash-0731 (chain[1], host: config)

## Refutation — infosec

### [critical]: a missing `links` migration is still a successful no-op, so an API with no schema boots “healthy”

- claim: The startup path treats “the migrations directory contains no `.sql` files” as success. In E2 the `links` table is a hard prerequisite for every store method, but if the migration file is absent from the shipped image (stale build context, `.dockerignore`, deleted file, failed `COPY`, wrong relabeling), the service starts and binds the port with no table and no error. This is a failure-as-bypass: the one check that could prove the schema exists is silently skipped.
- evidence: `internal/migrate/migrate.go` contains `if noMigrationFiles(dir) { return "no change (empty migrations dir)", nil }`; `cmd/api/main.go` logs the status and proceeds to bind when `err == nil`. The E2 spec relies on the migration file being present but does not add any runtime assertion — the only guards are later manual `schema_migrations` queries in the smoke test.
- predicted_consequence: a production image can report healthy, answer `/healthz` with 200, and have no `links` table at all; every `Insert` and `ByCode` then fails, restarts do not repair it because no migration is recorded or attempted, and the E2 promise of “a place for code→url to live” is silently absent.
- spec_anchor: `4-how-behaviour`

### [minor] the integration smoke test’s “force a query error” step is destructive and contradicts its own restart step
- claim: the smoke procedure proposes forcing a query-path error with `DROP TABLE links`, for the same database instance used by the restart/idempotency assertions. If the test is run where `DATABASE_URL` points at any non-disposable database, this drops the E2 schema and all rows, and because version 1 is already recorded in `schema_migrations`, restarting the API will not recreate it.
- evidence: `PROCEDURE e2_smoke` step 8 says “Force a query error (e.g. ByCode after `DROP TABLE links`)” and step 9 still asserts the step-5 row is readable after restart — there is no step to recreate the table or restore the fixture.
- predicted_consequence: as written, the smoke test either fails at step 9 when run against an empty disposable DB, or destroys the `links` table and its data when run with `DATABASE_URL` pointed at a reusable environment.
- spec_anchor: `integration-smoke-test`
