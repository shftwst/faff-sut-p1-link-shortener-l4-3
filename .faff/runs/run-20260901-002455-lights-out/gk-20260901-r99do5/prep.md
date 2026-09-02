# Prep log — E1 (gk-20260901-r99do5)

- Run: run-20260901-002455-lights-out · mode: autonomous · level: L4 · git-only (tracker none)
- Path: autonomous Path 2 (fresh-spec, no existing spec found in `.faff/specs/`)
- Slots resolved: spec=faffter-dark-nlspec · spec_review=faffter-dark-spec-review · architecture=faffter-noon-architecture · methodology=faffter-dark-methodology-agile-delivery · appetite=full · adr.mode=offer

## Opening — issue outline

- Gloss: packaged Go+Postgres service skeleton under docker-compose with a GET /healthz liveness check and env config.
- Status: Backlog (git-only queue item, pending).
- What it is about: E1, the first slice of the link-shortener. Establishes the runnable stack (api + Postgres via docker-compose), a health-checked container topology, environment-driven config, the migration runner foundation, and structured JSON errors. Does not implement shorten/redirect/TTL (E2..E6).

## Step 0 — architecture proposal (conditional trigger FIRED)

- Trigger: new runnable surface, no established architecture to inherit, no sibling proposal exists. This is the stack-establishing epic, so the proposal is load-bearing for downstream E2..E6 and the holdout env step.
- Infra profile: `faff profile show` reports no profile (fresh repo) — proposed from the brief's stated stack preference alone, recorded as an assumption.
- Producer: faffter-noon-architecture (run in-context; single-level nesting as a subagent).
- Chosen stack: Go 1.22 on stdlib net/http (ServeMux method+path routing) + Postgres 16, two-service docker-compose (api + db), golang-migrate applied on start, 12-factor env config, structured JSON errors. recommendation: build.
- `faff contract architecture-proposal` → exit 0, violations: []. Block carried verbatim into the spec.

## Step 1 — spec production (faffter-dark-nlspec, in-context)

- Full nlspec spec produced for E1's slice ONLY: packaged skeleton, docker-compose (api + Postgres 16), GET /healthz liveness, env config (DATABASE_URL + HTTP_ADDR, fail-fast), golang-migrate runner foundation (empty migrations/ in E1; links table is E2), structured JSON error envelope.
- Out of scope recorded: POST /shorten (E3), GET /{code} redirect + 404 (E4), links table + persistence layer (E2), TTL (E5), deep readiness in /healthz (E6 extension point).
- Born-verifiable DoD: `docker compose up -d --build` brings api+db healthy and GET /healthz -> 200 within 60s (observable via `docker compose ps` + one HTTP GET).
- Architecture-proposal block carried VERBATIM in spec section 9 (+ ADR promotion intent) for downstream E2..E6 and the holdout env step.
- Self-review (in-context clean pass): no blocker, 0 major. Confidence: high.
- Provenance stamp + build-tier written under H1. build-tier: complex.
- Marker validation: `faff contract spec-readiness` -> exit 0, markers_valid:true (7 chosen + 1 assumes).

## Step 2 — already-shipped / premise gate

- Git-only, first epic, `queue-state derive` shows 0 terminal siblings. No Done work can supersede the premise. Proceed. No `## Already shipped against this surface` section.

## Methodology critique (faffter-dark-methodology-agile-delivery, issue-critique, in-context)

- Right-sized: no issues (single cohesive 1-3 day unit). Workstream fit: no issues (P1 foundation slice, outcome-named). Deps surfaced: no issues (first slice; E2 blocked-by E1 already explicit). Risk profile: positive (front-loads the compose+migrate integration risk).
- Block appended to spec. Does not block confidence-high promotion (autonomous).

## Attach (git-only)

- Spec written to `.faff/specs/gk-20260901-r99do5.md`; attach-state marker flipped attached:true; prepcheck state=attached.

## Spec-review gate (L4 adversarial, faffter-dark-spec-review)

- Lens selection: `faff spec-review-lenses` -> all four (architectural/infosec/methodology/QA), mode adversarial (L4 pinned, never narrowed).
- Ratified scope assembled (`faff ratified-scope --assemble --container link-shortener`) -> exit 0; PRD goals + non-goals supplied to the design lenses.
- Backend chain resolved (`faff spec-review-pin --resolve`): 3-link fallback (qwen local / deepseek-openrouter / gemini-gemma). API key envs present but placeholder-length (dummy fixtures).
- Transport: every dispatch (initial fan-out, single-lens probe with hard --deadline, full fan-out) exhausted the chain with NO first byte from any backend. Single-lens probe returned exit 9 (mandatory-chain-outage): all 3 backends exhausted their time slices. External egress is blocked in this sandbox, so the whole chain is unreachable.
- Per-lens outcome (all four): unavailable / infra-configured (DEADLINE/UNREACHABLE = transient retry-disposition).
- Aggregation (`aggregate.mjs --n 4`): transport floor -> verdict `unavailable` (swing-capable outage), 4 lenses major. round-1.json recorded.
- Consumer-fold: `faff contract spec-review-verdict` -> exit 0, conformant:true. Verdict `unavailable` routes to the outage disposition (NOT a spec-defect route; NEVER coerced to approve).
- disposition_unavailable (autonomous): in-turn transport retries exhausted (identical infra-configured chain-exhaustion each time). No prior hold -> outage_holds 0 -> HOLD 1/3 (< hold_limit 3).
  - Wrote `.faff/resume/gk-20260901-r99do5/spec-review-hold.json` (outage_holds:1, all four lenses outaged).
  - `faff label add faff-awaiting-spec-review` (git-only no-op, logged).
  - Status stays Backlog: spec attached in `.faff/specs/`, NOT promoted. In-flight marker closed.

## Outcome

- Return: `spec-review-held` (a hold, not a park; no faff-parked label; auto-resumes at review on the next drain when the reviewer chain is reachable).
- Spec confidence: high. Markers valid. Architecture-proposal landed verbatim.
- Only blocker to promotion: the spec-review backend outage in this environment. The spec artifact itself is high-confidence and build-ready pending the held review.

---

## Spec-review RESUME (this run — false-negative correction)

Resumed at the review gate per `faff-awaiting-spec-review` + carried hold (outage_holds=1). The prior hold was a FALSE NEGATIVE: it used a short `--deadline 80` that gave each of 3 backends a ~26s slice — too short for backend #1 (spark-qwen, first_byte_timeout 180) to produce a first byte, so the chain exhausted (exit 8 -> exit 9). Backends are in fact reachable. Cleared the stale outage round-1.json and re-ran the review properly.

Transport fix: `--deadline 800` (>=800 as directed), configured `--timeout 900`, `--num-predict 2000` (the actually-parsed output cap; review-call.mjs ignores `--max-tokens`, and the 12000 config value caused 6x-slower generation on a first attempt). Fan-out is genuinely slow here (~8-13 min/round: spark-qwen is slow under 4-way concurrent load), so each fan-out ran as a single backgrounded `fan-out.mjs` call and was awaited to completion within the turn.

Round 1 (full 4-lens fan-out): **verdict = revise** (aggregate.mjs, conformant). infosec REFUTED {major: DB creds could leak via the fatal config log; minor: root container user}; QA REFUTED {major: graceful shutdown unverifiable; major: migration runner not asserted by any scenario; minor: no 405 scenario; minor: "structured fatal log" undefined}; architectural served non-findings (exit-10 advance, treated clear per directive); methodology "no findings" (clear). round-1.json recorded, pin captured (spark chain[0]), contract exit 0.

Applied all 6 lensed fixes in place to `.faff/specs/gk-20260901-r99do5.md`: single-line-JSON fatal log that names the offending variable and never echoes the DSN; dedicated non-root runtime USER; born-verifiable graceful-shutdown / 405 / schema_migrations acceptance scenarios; concrete "structured fatal log" definition; +2 Chosen markers, DONE + smoke-test updates, revision provenance note. Readiness contract re-validated: conformant, confidence high, build-tier complex.

Round 2 (re-review of the revised spec; deadline 800, then 2 in-turn retries with larger deadlines for the outaged lenses):
- architectural (now SERVED by gemini): **major** — schema_migrations DoD may be unachievable: golang-migrate returns ErrNoChange and may not create the bookkeeping table with an EMPTY migrations/ dir (E1's case). A real soundness point on a DoD observable; a design call (baseline no-op migration vs change the observable).
- QA (re-review, gemini): **major** — no-secret-log constraint only tested for UNSET DATABASE_URL, not present-but-invalid; **minor** — no non-root-user scenario; **minor** — "without binding the port" lacks a concrete oracle.
- methodology: clear.
- infosec: could NOT be validly re-reviewed — gemini/gemma HTTP 429 (quota) then a degenerate exit-0 empty response, across 2 in-turn retries (limit reached). gemini's per-minute quota was depleted by the review runs and recovered intermittently (served architectural+QA, 429'd infosec).
- Deterministic **aggregate.mjs verdict = unavailable** (infosec infra-down and swing-capable; conformant via `faff contract spec-review-verdict`).

disposition_unavailable (autonomous): in-turn retries exhausted -> HOLD. outage_holds 1 -> 2 (< hold_limit 3). Wrote the hold store (outaged_lenses ["infosec"], pinned reviewer, full note). `faff-awaiting-spec-review` re-applied (git-only descriptor). Status stays Backlog; spec NOT promoted. In-flight marker closed.

## Outcome (resume)

- Return: **spec-review-held** (a hold, not a park; no faff-parked label; auto-resumes at review next drain).
- The round-1 `revise` was real and its fixes are applied. Round-2 surfaced real, un-fixed architectural + QA gating objections on the revised spec AND an infosec re-review outage, so the spec is NOT build-ready. The revised spec still needs another revise pass (architectural migration-table DoD + QA verification gaps); the infosec re-review needs a non-rate-limited gemini.
- Spec confidence: high; markers valid; architecture proposal intact. E1 remains at spec-review, held.

---

## Spec-review CONVERGENCE (coordinator resume — backends recovered)

Coordinator confirmed the round-2 429 was a transient per-minute gemini quota (recovered), directed continuation of the autonomous revise loop to a terminal verdict with paced fan-outs (2000 cap; per-lens 60s-backoff retry on 429 before concluding outage) and autonomous low-risk resolution of the outstanding objections.

Applied the outstanding round-2 objection fixes in place:
- architectural/major (empty `migrations/` can't create `schema_migrations`): RESCOPED — E1 now asserts the golang-migrate runner *ran* (empty dir → `ErrNoChange`), observable in startup logs ordered before the port-bound line; the `schema_migrations` table assertion is explicitly deferred to E2. Removed the table assertion from the scenario, DONE, failure-mode note, rationale, and smoke test.
- QA/major (no-secret-log only tested for unset DSN): added a present-but-invalid-DSN scenario asserting no secret token in logs.
- QA/minor (non-root, port oracle): added a non-root-user (`id -u` != 0) scenario and a connection-refused oracle for "port never bound".

Round 2 (full 4-lens, all served after paced per-lens retries): verdict = **revise**. architectural + methodology clear; infosec {major: DSN could leak via wrapped pgx/golang-migrate driver errors — deeper than round 1}; QA {major: failed-Ping path has no scenario; minor: migration-order oracle only checks log presence, not order}. Gating count 6→3, objecting lens-set steady {infosec,QA} → `faff spec-review-churn` = churn:false. round-2.json recorded, contract conformant.

Applied round-2 fixes:
- infosec/major: required ALL fatal boot logs (config/migrate/db-connect) to scrub the DSN from wrapped driver errors; stable failure-class events (config_invalid / migrate_failed / db_connect_failed).
- QA/major: added an unreachable-DB (`db_connect_failed`) scenario verifying the failed-Ping path + no secret leak.
- QA/minor: strengthened the migration oracle to assert the runner log precedes the port-bound log (ordering).

Round 3 (full 4-lens, all served after paced per-lens retries; gemini burst-429'd architectural/QA, cleared on solo backoff retries): **all four lenses clean** (non-findings). Deterministic aggregate.mjs verdict = **approve** (0 objections). `faff spec-review-churn` r2→r3 = churn:false (objecting set → empty). round-3.json recorded; `faff contract spec-review-verdict` conformant:true.

## Outcome (convergence)

- Verdict: **approve** (round 3). Total spec-review rounds: **3** (revise → revise → approve).
- Retained `spec-review: approve` on the spec provenance line (alongside `confidence: high`).
- Promotion (git-only): cleared the outage/awaiting hold (`spec-review-hold.json` removed), removed `faff-awaiting-spec-review`, closed the in-flight marker. `prepcheck` = state attached, disposition null (not parked). E1 is build-ready.
- Return: **promoted** — high-confidence spec attached, spec-review approve retained, build-eligible for the graft/build drain.
- No objection went unresolved; all round-1/round-2 objections were resolved autonomously with low-risk spec answers (no design call parked).
