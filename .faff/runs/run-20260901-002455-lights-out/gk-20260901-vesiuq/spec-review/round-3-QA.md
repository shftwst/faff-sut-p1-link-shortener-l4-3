## Adversarial findings — openai/google/gemma-4-31b-it (chain[2], host: config)

## Refutation — QA

### major: `IF NOT EXISTS` anti-pattern is not verifiable
- claim: The spec explicitly prohibits the use of `IF NOT EXISTS` in the migration to ensure that "dirty-state or double-apply" bugs are exposed rather than masked, but no scenario verifies that the migration actually fails when the table already exists.
- evidence: "The migration is a plain `CREATE TABLE`, not `CREATE TABLE IF NOT EXISTS`." and "Anti-pattern: `CREATE TABLE IF NOT EXISTS links`. Why: it makes a double-apply silently succeed, hiding the exact `schema_migrations` bookkeeping bug..."
- predicted_consequence: A developer could use `IF NOT EXISTS` to make their local development smoother, violating the design principle. Because the provided scenarios only test "fresh boot" and "restart" (where the `schema_migrations` table correctly prevents the migration from running twice), the use of the forbidden `IF NOT EXISTS` clause would go undetected.
- spec_anchor: design-decision-rationale

### minor: "No validation" requirement for empty strings is not verified
- claim: The spec defines a specific design decision to *not* validate input (deferring this to E3), specifically mentioning that empty strings should be accepted. However, the scenarios only test "long" and "non-ASCII" inputs, leaving the empty-string boundary untested.
- evidence: "E2 does not reject an empty-string code or url... the DB `TEXT` columns accept any non-null string... Rejecting an empty or malformed URL... are E3's responsibilities."
- predicted_consequence: A developer might "helpfully" implement a check for empty strings (e.g., `if link.Code == "" { return err }`), violating the spec's boundary and pulling E3's responsibilities forward. This would not be caught by the "long/non-ASCII" scenario.
- spec_anchor: how-behaviour

### minor: `created_at` oracle is ambiguous
- claim: The oracle used to verify that `created_at` is DB-assigned is insufficient to distinguish between a database `DEFAULT` and a value assigned by the Go application immediately before the insert.
- evidence: "a fetched `CreatedAt` is not the zero time and falls within the `[just-before-insert, just-after-fetch]` window (proving `DEFAULT now()` fired, not a hard-coded value)."
- predicted_consequence: The test would pass even if the developer explicitly set `link.CreatedAt = time.Now()` in Go and included the column in the `INSERT` statement, failing to prove that the database's `DEFAULT` mechanism is actually being utilized as required.
- spec_anchor: how-behaviour
