## Adversarial findings — openai/google/gemma-4-31b-it (chain[2], host: config)

## Refutation — QA

### major: missing scenario for non-nil expiry round-trip
- claim: the round-trip integrity of the `expires_at` column for actual timestamps cannot be verified.
- evidence: the DONE item "a non-nil timestamp reads back equal" (under `how-behaviour`) and the scenario "Insert Link{... ExpiresAt:nil} ... fetched Link has ... a nil ExpiresAt".
- predicted_consequence: a bug where non-nil timestamps are truncated, shifted by timezone, or fail to scan into the `Link` struct would not be caught by the provided scenarios, as only the `NULL` case is exercised.
- spec_anchor: scenarios

### major: unverified DSN-free query errors
- claim: the requirement that query-path errors do not leak the DSN is a "promise" with no verifiable scenario.
- evidence: the DONE items "Insert returns ... a DSN-free wrapped error otherwise", "ByCode returns ... a DSN-free wrapped error otherwise", and "No query-path error returned by the store contains the DSN".
- predicted_consequence: the build could satisfy all scenarios while still leaking the DSN in the event of a Postgres error (e.g., a syntax error or connection drop during a query), violating the security discipline established in E1.
- spec_anchor: done-definition-of-done

### minor: weak oracle for `created_at`
- claim: the assertion that `created_at` is "non-zero" is too vague to prove the DB's `DEFAULT now()` is functioning correctly.
- evidence: the scenario "fetched Link has ... a non-zero CreatedAt".
- predicted_consequence: a build that accidentally hard-codes a static date or inserts a dummy value would pass the "non-zero" check, masking a failure of the database-level default.
- spec_anchor: scenarios
