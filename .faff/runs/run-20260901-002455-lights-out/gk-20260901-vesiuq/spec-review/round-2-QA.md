## Adversarial findings — openai/google/gemma-4-31b-it (chain[2], host: config)

## Refutation — QA

### major: `.down.sql` correctness is unverifiable
- claim: The requirement that the down migration actually drops the table cannot be verified.
- evidence: The DONE criteria explicitly requires that "the down file drops the table", but there is no corresponding scenario to execute the down migration and verify the table is gone.
- predicted_consequence: A build could ship with a syntactically invalid or logically incorrect `.down.sql` (e.g., a typo in the table name), satisfying the "file exists" check but failing the functional requirement.
- spec_anchor: done-definition-of-done

### minor: "non-empty" constraints are not verified
- claim: The assertion that `Code` and `URL` are "required, non-empty" is not testable.
- evidence: The `Link` record definition in `types-the-store-surface-added-to-internal-store` marks both fields as "required, non-empty", but the schema only specifies `NOT NULL` (which allows empty strings `''` in Postgres), and no scenario tests the behavior of `Insert` when provided with empty strings.
- predicted_consequence: The store may allow the persistence of empty strings, violating the record's defined constraints without triggering an error.
- spec_anchor: types-the-store-surface-added-to-internal-store

### minor: alphabet and length boundaries are not exercised
- claim: The claim that the system "presumes no particular length or alphabet" for codes is not verified.
- evidence: The spec claims no particular length or alphabet is presumed, but all provided scenarios use short, alphanumeric ASCII strings (e.g., `"abc1234"`, `"exp0001"`).
- predicted_consequence: Potential encoding issues with UTF-8 codes or truncation/performance issues with extremely long URLs would not be caught by the integration tests.
- spec_anchor: types-the-store-surface-added-to-internal-store
