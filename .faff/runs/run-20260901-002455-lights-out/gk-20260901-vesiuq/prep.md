# Prep — gk-20260901-vesiuq (Epic E2: schema migration + persistence layer)

Outline: E2 — links table (code→url, created/expires) + store insert/fetch, grounded in merged E1.
Status: Backlog → build-ready (git-only; no tracker column move).
Mode: autonomous L4 lights-out, git-only. Producer: faffter-dark-nlspec.

## Outcome
- Spec attached: .faff/specs/gk-20260901-vesiuq.md
- Confidence: high (mechanical schema+store slice; no open punts; one Assumes, all validated against merged E1).
- Build-tier: complex.
- Architecture-proposal step: did NOT fire (E2 inherits E1's fixed stack; no new runnable surface).
- Methodology critique (agile-delivery): no principle violations (right-sized, cohesive, deps surfaced, low risk).
- Spec-review (L4 adversarial, 4 independent lens refuters): APPROVE at round 4.
- Final prep outcome: promoted. E2 is build-ready.

## Spec-review rounds
- R1 revise: QA — expires_at non-nil round-trip, DSN-free query-error scenario, created_at oracle. (arch/infosec/methodology clear)
- R2 revise: QA — down.sql correctness, non-empty over-claim, alphabet/length. (churn false; convergence flat)
- R3 revise: QA — IF NOT EXISTS verification, empty-string boundary, created_at DEFAULT-vs-app-set. (churn false)
- R4 approve: all four lenses clear. QA: observation: no findings.
- All QA objections were genuine born-verifiable test-coverage gaps closed in place; no scope expanded into E3/E4/E5.
- Transport notes: QA lens empty-out (exit 11) at max_tokens 2000 on R2/R3 (reasoning==budget knee); cleared with solo retries and max_tokens 4000 (config invariant max_tokens>reasoning_budget; honours "never 12000"). gemini persistently 429 (backed off 60s / dropped from retry chain). No genuine transport outage (never exit 5/12/6/2/4/7 terminal for a served lens).

## Unresolved objections
None.
