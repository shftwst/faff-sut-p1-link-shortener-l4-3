# graft — gk-20260901-vesiuq (E2: links table + store methods)

Mode: autonomous L4 lights-out, git-only (no remote; merge is local to main; PR = 0 sentinel).
Worktree: /Users/shftwst/.faff/worktrees/faff-sut-p1-link-shortener-l4-3/gk-20260901-vesiuq-links-table-and-store
Branch: gk-20260901-vesiuq-links-table-and-store (off main da01a8a)

## Gates (Step 2)
- admissibility (lights-out): admissible=true (12 born-verifiable scenarios, 19 DONE items, 0 vague). R3 prose-DONE warnings advisory only.
- eligibility / intake: git-only (tracker: none) — no tracker labels; gates N/A by construction. Issue already admitted in armed L4 run-ledger.

## Build (Step 7)
Files added:
- migrations/000001_create_links.up.sql — plain CREATE TABLE links (code TEXT PK, url TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), expires_at TIMESTAMPTZ NULL)
- migrations/000001_create_links.down.sql — DROP TABLE IF EXISTS links
- internal/store/links.go — Link record, Links repo, NewLinks, Insert, ByCode, ErrLinkNotFound, ErrDuplicateCode
- internal/store/links_integration_test.go — store round-trips + migration file correctness + DSN-free query error (env-gated: STORE_TEST_DSN/DATABASE_URL, skips without a DB)
E1 files unchanged: store.go (Open/SafeReason), migrate.go (Run), cmd/api/main.go, router.go.

## Engineering gate ladder (Step 7.5) — via golang:1.22-alpine (matches CI)
- gofmt -l .: clean
- go vet ./...: exit 0
- go test ./... (no DB): all packages ok (store integration tests skip)
Signal: pass.

## AC verification (Step 8)
Integration tests against a real Postgres (STORE_TEST_DSN) — 11/11 PASS:
- Insert/ByCode round trip, nil expiry: URL preserved, ExpiresAt nil, CreatedAt DB-assigned within window (proves DEFAULT now()).
- Non-nil expires_at round trip: equal at µs precision, no tz shift.
- ByCode unknown -> ErrLinkNotFound + zero Link.
- Duplicate code -> ErrDuplicateCode, existing row unchanged.
- Long/UTF-8 code+url round trip byte-for-byte.
- Empty-string code+url persisted (E2 does not validate).
- Query-path error carries no DSN/password (real query failure after DROP TABLE).
- Down migration drops table; re-up restores it.
- Up migration plain CREATE TABLE: raw double-apply errors "already exists".
- E1 SafeReason tests still pass (unchanged).

Live compose stack (docker compose, from worktree) verified the startup migration path:
- api startup log: migrate_done "applied"; on restart "no change".
- schema_migrations: version 1, dirty=f (unchanged after restart).
- links columns: code text NOT NULL PK, url text NOT NULL, created_at timestamptz NOT NULL default now(), expires_at timestamptz nullable.
AC checklist: all_verified=true (ac-checklist.json).

Scope boundaries respected: no HTTP handler, main.go/router unchanged, no TTL/expiry logic, no input validation.

## Review (Step 9) — adversarial, MANDATORY at L4
Phase 1 (structural): pass — AC coverage complete, no obvious bugs, in-scope (no HTTP/TTL/validation), spec-faithful.
Phase 2 (adversarial fan-out via review-call.mjs, --timeout 900 --deadline 800 --num-predict 2000):
  chain: qwen-3-8 (empty->skip), deepseek-v4-flash (empty->skip), openrouter google/gemma-4-31b-it SERVED -> "### observation: no findings".
  backends that served: openrouter-gemma (google/gemma-4-31b-it).
Verdict: pass (conformant via faff contract review-verdict). review-verdict.json written.

## Holdout gate (Step 10) — L4 code-blind, MANDATORY
env: docker compose api+db (project e2holdout) stood up from the worktree; api healthy; db endpoint via env DSN. Torn down after.
Exercised code-blind against the RUNNING stack (spec + endpoints/DSN only, never diff/code). See holdout-evidence.txt.
Result: 10 born-verifiable criteria MET with running-stack evidence (schema_migrations 1/false; links 4 cols; restart "no change"; insert-omitting-created_at DEFAULT now(); non-nil expires_at us round-trip; absent-code no-row; duplicate PK rejected + row unchanged; long UTF-8 byte-for-byte; empty-string persisted; plain-CREATE collision).
NEEDS-HUMAN punts (2) — no code-blind running-stack surface at E2's pre-HTTP-endpoint stage:
  - S7 down-migration file correctness (api runs UP only; no down trigger; no file access code-blind).
  - S11 query-path DSN-free error (api exposes only /healthz; no endpoint invokes Insert/ByCode, so no query-path error is triggerable). Connect-path scrubbing is E1's, startup-observable; the E2 query path is not.
  Both ARE independently verified by graft's internal/store integration tests (which passed in Step 8), but the code-blind holdout cannot reach them without an HTTP surface (arrives in E3/E4).
holdout-verdict: aggregate=needs-human, code_blind=true, conformant (faff contract holdout-verdict, no violations). faff holdout verdict --issue -> gate:block (exit 1).

## Merge decision
L4 merge floor: AC verified (yes) + gates green (yes) + review pass (yes) + holdout meets-spec (NO -> needs-human).
Holdout is a merge precondition and returned needs-human -> MERGE REFUSED (fail-closed). E2 NOT merged.
git-only (no remote; PR=0 sentinel): branch gk-20260901-vesiuq-links-table-and-store holds the E2 commits, unmerged, awaiting human confirmation of S7/S11 (or E3/E4's HTTP surface).
Disposition: pr-open-for-human (holdout-block) -> in git-only maps to a needs-human park (no PR).
