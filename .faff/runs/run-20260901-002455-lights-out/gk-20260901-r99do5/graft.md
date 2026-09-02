# graft — gk-20260901-r99do5 (E1: packaged service skeleton)

Run: run-20260901-002455-lights-out · L4 lights-out · git-only · autonomous
Branch: gk-20260901-r99do5-e1-skeleton (worktree under ~/.faff/worktrees)

## Build
Go 1.22 stdlib net/http + pgxpool + golang-migrate, packaged as docker-compose (api + Postgres 16).
Layout: cmd/api, internal/{config,httpapi,store,migrate}, migrations/ (empty), Dockerfile (multi-stage,
non-root uid 10001), docker-compose.yml, .env.example, CI workflow + scripts driving the gate ladder.

Deps pinned for the Go 1.22 target: pgx v5.5.4, golang-migrate v4.18.1 (latest require Go 1.24+).

Build-time bug caught by fail-fast and fixed: golang-migrate reports "file does not exist" (not
ErrNoChange) for a truly empty migrations dir; the runner now treats a present-but-empty dir as a no-op.

## Gates (Step 7.5) — pass
FORMAT (gofmt) / LINT (go vet) / UNIT (go test) all pass, run via Docker (host has no Go/make).
Declared through a .github/workflows/ci.yml the `faff gates discover` ladder reads.

## AC verification (Step 8) — all verified against the running compose stack
- compose up -d --build → both db and api healthy; GET /healthz 200 within 60s.
- GET /healthz → 200, application/json, {"status":"ok"}.
- GET /does-not-exist → 404, application/json, {"error":{"code":"not_found","message":...}}.
- POST /healthz → 405, application/json, {"error":{"code":"method_not_allowed","message":...}}.
- Startup logs: migrate_done ("no change (empty migrations dir)") strictly before listening.
- SIGTERM (compose stop -t 10) → clean exit 0 within timeout; shutdown log present.
- Unset DATABASE_URL → config_invalid fatal, exit 1, no value echoed.
- Malformed DSN (token sup3rSecret) → db_connect_failed fatal, exit 1, 0 secret occurrences.
- Unreachable DB (token sup3rSecret) → db_connect_failed fatal, exit 1, 0 secret occurrences (DSN scrubbed).
- Non-root uid: id -u = 10001.

## Review (Step 9) — verdict: pass (adversarial, L4 mandatory)
Phase 1 (structural): pass — AC covered, scope = E1 only, spec-faithful.
Phase 2 (adversarial second opinion): configured chain is spark-qwen (502 unreachable), openrouter-deepseek,
openrouter-gemma, gemini-gemma (HTTP 429 hard quota). deepseek returned findings on the first pass;
gemini quota-dead and qwen down, so re-review used openrouter-gemma (a configured chain backend).

Terminal Phase-2 result: openrouter-gemma (google/gemma-4-31b-it), exit 0, "no findings" on the fixed code.
No critical in the terminal review → no autonomous escalation.

Disposition of deepseek's first-pass findings (collapse-and-log):
- critical "main.go incomplete / won't compile / logFatal,logInfo undefined" — FALSE POSITIVE (hallucinated
  truncation). Disproof: go build ./... and go vet both exit 0; the container built and served
  /healthz/404/405 and shut down cleanly; main.go is complete and defines logFatal/logInfo/writeLine.
- major "migrate.Run masks real errors on an empty dir" — VALID, FIXED. Since E1's dir is always empty,
  the noMigrationFiles arm would have swallowed any migrate.New/Up error. Reworked to short-circuit only
  the deterministic empty-source case and surface every genuine error.
- minor "scrub may miss encoded password forms" — VALID (low risk), FIXED. scrub now also redacts
  QueryEscape/PathEscape forms of the password.
- observation "no empty-dir+valid-DSN success test" — covered by the integration smoke test (compose up
  shows "no change (empty migrations dir)" then healthy).
- observation "healthz is shallow" — the spec's deliberate E1 design (deep readiness is a later-epic
  extension point).

## Merge gate (Step 10) — L4 floor: AC verified + gates green + review pass + holdout meets-spec.

## Outcome — SHIPPED
- Merge: git-only local fast-forward to main. merge-gate --local --execute → merge-ok.
  Floor: CI ci-green (fresh gate ladder on branch tip), integrity custody-trusted, AC verified,
  review pass, holdout meets-spec.
- merge_sha: da01a8a18742a854d3643a3491adb521a4980086 (merge-record.json pr:0 sentinel, merged:true).
- Post-merge verification: verified-ok (go test ./... exit 0 at merged sha).
- Run ledger: outcomes{gk-20260901-r99do5: shipped}, owner done.
- Housekeeping: worktree removed, merged branch deleted, main operator config (openrouter-gemma,
  !run.log) preserved via stash/pop.

## Human follow-ups (non-blocking)
- ADR materialisation (graft Step 4b) deferred: adr.mode=offer and the spec carries an ADR promotion
  intent (5 decisions: Go 1.22 stdlib net/http, Postgres 16 via pgx, golang-migrate on start, compose
  topology, 12-factor env config). Not gating the merge floor; deferred rather than spend flaky LLM
  slot calls unattended. Materialise via the adr slot when convenient.
- Adversarial backend health: spark-qwen (tailnet) returned HTTP 502 both attempts; gemini free tier
  is hard quota-exhausted (HTTP 429). Terminal review served by openrouter-gemma; openrouter-deepseek
  is finicky (empties below ~5k max_tokens, overruns a <10min window at 5k). Worth revisiting backend
  tuning.
