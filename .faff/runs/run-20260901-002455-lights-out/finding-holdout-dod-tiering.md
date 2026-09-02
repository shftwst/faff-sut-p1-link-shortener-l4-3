# Finding — the L4 code-blind holdout permanently refuses a store-layer epic (DoD-tiering mismatch)

- **Run:** `run-20260901-002455-lights-out` (L4 lights-out, git-only, link-shortener SUT)
- **Issue:** E2 `gk-20260901-vesiuq` — "schema migration + persistence (links table + store layer)"
- **Gate that refused:** L4 code-blind holdout merge-floor (`faff holdout` / `faff contract holdout-verdict`)
- **Class:** binding-constraint / first-failure-rung finding for the faff backlog
- **Disposition:** the gate behaved **correctly** (fail-closed, no fabricated verdict). The defect is upstream, in how the DoD was tiered.

## One line

A store-layer epic, split from the endpoints that exercise it, carries two DoD criteria that a code-blind evaluator (spec + running stack, never the source) can **never** satisfy at that stage — so the L4 holdout is structurally guaranteed to return `needs-human`, the fail-closed floor refuses the merge, and every downstream epic that depends on the store blocks. The code was fine; the acceptance contract was un-satisfiable by construction.

## What actually happened

E2 built cleanly and passed every gate up to the holdout:

- `gofmt -l` clean, `go vet ./...` exit 0, `go test ./...` green (run in `golang:1.22-alpine`, matching CI).
- **11/11 store integration tests** green against a real Postgres 16 (`STORE_TEST_DSN`).
- Adversarial review: **pass** (served by openrouter `google/gemma-4-31b-it`).

The code-blind holdout then stood up the compose stack (api + Postgres) and exercised the spec's criteria **against the running system only** — never the diff. Result: **10 of 12 born-verifiable criteria MET** with running-stack evidence (schema_migrations state, `links` columns, restart "no change", `DEFAULT now()`, `expires_at` µs round-trip, absent-code no-row, duplicate-PK rejection, UTF-8 byte-for-byte, empty-string persistence, plain-CREATE collision).

Two criteria forced `needs-human`, which under the canonical holdout rule (any `needs-human` ⇒ aggregate `needs-human`) refused the merge:

### S7 — down-migration correctness
The criterion asserts: applying the shipped `000001_create_links.down.sql` drops the table, and re-applying up restores it. **The running stack runs UP migrations only** at startup — there is no down trigger, and a code-blind evaluator has no file access to apply the down SQL itself. There is therefore **no running-stack surface** through which this can ever be observed. This is an integration/CI-tier property, not a running-stack one.

### S11 — query-path error carries no DSN
The criterion forces a store query-path error (Insert/ByCode against a dropped table) and inspects the returned error text for DSN/password leakage. **E2 exposes no HTTP endpoint that invokes the store** — the api serves only `/healthz` — so no query-path error can be triggered through the running stack. (The *connect-path* DSN scrubbing, E1's `SafeReason`, is startup-observable; the E2 *query* path is not.) This only becomes reachable once E3/E4 add the HTTP surface.

## Why this is structural, not a bug

The two criteria are not "failing" — they are **unobservable in the channel the holdout is restricted to**:

| Criterion | Verifiable by integration test? | Verifiable code-blind at E2's stage? | Why |
|-----------|:---:|:---:|-----|
| S7 down-migration | ✅ | ❌ **never** | running stack applies UP only; no file access |
| S11 query-path DSN scrub | ✅ | ❌ **not yet** | no HTTP endpoint invokes the store until E3/E4 |

S7 is *permanently* un-code-blind-verifiable. S11 is *prematurely* asked — it becomes verifiable one epic later.

## Root cause

**Two components disagree on what "born-verifiable" means.**

- The **spec-review QA lens** treats integration-test-level checks (down-migration correctness, error-text hygiene) as born-verifiable and writes them into the holdout DoD.
- The **code-blind holdout** can only verify **running-stack-observable** properties (spec + live endpoints/DSN, no source, no test suite).

The intersection is smaller than the QA lens assumes. When the lens emits a criterion outside that intersection, the L4 holdout is forced to `needs-human` on a criterion no code change could ever clear — producing a **permanently un-mergeable epic** at L4.

Aggravating factor: **granularity.** Splitting the store layer into its own epic, ahead of any endpoint that drives it, removes the very HTTP surface that would make S11 (and error-path criteria generally) running-stack-observable. A store slice is too fine-grained for independent code-blind holdout verification.

## Blast radius

- E2 did not merge to main.
- E3–E6 (`POST /shorten`, `GET /{code}` redirect, TTL expiry, structured errors) **all transitively depend on E2's store layer on main** → all four blocked, never admitted.
- Run terminated `product-incomplete` rather than claim the PRD done. Ground-truth reconcile: consistent (E1's shipped claim matches git; no phantom merge).

The fail-closed floor did **exactly** what L4 promises: no `met` verdict was fabricated to force a merge. This finding is about the DoD that fed the gate, not the gate.

## Shadow evidence (the corroboration channel the holdout can't see)

Both punted criteria **are** proven — just in the integration-test channel the code-blind holdout is forbidden from reading. On the unmerged branch `gk-20260901-vesiuq-links-table-and-store` (off main `da01a8a`), `internal/store/links_integration_test.go`:

- **`TestDownMigrationDropsTableAndUpRestoresIt`** → S7. Applies the shipped down file, asserts the table is gone, re-ups, asserts restored.
- **`TestQueryPathErrorCarriesNoDSN`** (+ `assertNoSecret`) → S11. Triggers a real query failure after `DROP TABLE`, asserts the error text contains neither the DSN nor its password.

Both passed in the build's Step 8. They are re-runnable:

```
git switch gk-20260901-vesiuq-links-table-and-store
STORE_TEST_DSN=postgres://…  go test ./internal/store/ \
  -run 'TestDownMigrationDropsTableAndUpRestoresIt|TestQueryPathErrorCarriesNoDSN' -v
```

**Caveat on equivalence** — this is corroborating, not identical, evidence, and the two differ:
- **S7:** the shadow test is a *full* substitute. A down-migration is inherently a CI/integration property; the code-blind holdout should never have owned it.
- **S11:** the shadow test proves the *store method* scrubs the DSN, but the holdout's criterion is about the error surfaced through a *live HTTP endpoint*, which doesn't exist until E3/E4. The clean resolution is to let E3's holdout **inherit** S11, not mark it done at E2.

## Recommended fixes (either resolves it; they compose)

1. **Tier the DoD by verification channel.** Have `faff dod classify` / the QA spec-review lens tag each criterion `running-stack-verifiable` vs `integration-tier`, and never route an integration-tier criterion into the code-blind holdout DoD. This is the durable, tool-level fix — it protects *every* future L4 run, not just this repo. (S7 is the canonical example; error-path criteria are the general case.)
2. **Re-slice at plot/decompose time.** Treat *store + first endpoint* as one holdout-verifiable epic, so the running stack always has a surface that drives the store. This removes the premature-S11 class of problem and keeps epics independently mergeable at L4.

## This run's unblock (tactical)

Confirm S7 is integration-test-covered and re-tier it out of E2's code-blind DoD; confirm S11 is inherited by E3's holdout. E2 then merges on its 10 met criteria and the E3–E6 chain proceeds. Alternatively, re-slice store + `POST /shorten` into one epic and rebuild.

## Evidence trail

- `.faff/runs/run-20260901-002455-lights-out/summary.md` — "Binding-constraint finding" section
- `…/gk-20260901-vesiuq/holdout.json` — `aggregate: needs-human`, `code_blind: true`, 10 met / 2 needs-human, `violations: []`
- `…/gk-20260901-vesiuq/holdout-evidence.txt` — per-criterion running-stack evidence, incl. the S7/S11 "NEEDS-HUMAN" rationale and the "independently verified by graft's integration test" notes
- `…/gk-20260901-vesiuq/park.md` — park record + root-cause statement
- `…/gk-20260901-vesiuq/graft.md` — Step 8 (11/11 integration tests) and Step 10 (holdout) records
- Branch `gk-20260901-vesiuq-links-table-and-store` — the unmerged E2 commits + the two shadow-evidence tests
