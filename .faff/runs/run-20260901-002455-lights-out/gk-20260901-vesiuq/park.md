# E2 (gk-20260901-vesiuq) — PARKED at L4 holdout merge-floor

Outcome: parked (holdout needs-human; L4 merge floor fail-closed, merge correctly refused).

## What passed
- Gates: gofmt/vet/test green; 11/11 store integration tests vs real Postgres.
- Adversarial review: pass (served by openrouter google/gemma-4-31b-it).
- Code-blind holdout: 10/12 born-verifiable criteria MET against the running DB (schema_migrations, links table columns, restart no-change, DEFAULT now(), expires_at round-trip, absent-code, duplicate-PK rejection, UTF-8 byte-for-byte, empty-string, plain-CREATE collision).

## Why parked (2 needs-human criteria)
- S7 down-migration correctness: the running stack runs UP migrations only; a code-blind evaluator (spec + running env, no files) can NEVER verify a down migration via the running stack. This is an integration/CI-tier criterion (E2's own integration test covers it), mis-classified into the code-blind holdout DoD by the spec-review QA lens.
- S11 query-path DSN-free error: E2 exposes no HTTP endpoint that invokes the store, so a query-path error cannot be triggered code-blind. Only verifiable once E3/E4 add the HTTP surface.

## Root cause (FINDING for faff backlog)
Spec-review's notion of "born-verifiable" (includes integration-test-level checks) is inconsistent with the code-blind holdout's (running-stack-only). A store-layer epic split from the endpoint epics is too fine-grained for independent code-blind holdout verification, and the QA spec-review lens can force a permanently-unverifiable criterion (S7) into the holdout DoD, producing a permanently-unmergeable L4 epic. The L4 fail-closed merge floor held correctly and refused the merge.

## Human action
Confirm S7 is integration-test-covered and re-tier it out of the code-blind holdout DoD; confirm S11 is inherited by E3's holdout. Then E2 merges on the 10 met criteria. Alternatively, re-slice so store+first-endpoint form one holdout-verifiable epic.
