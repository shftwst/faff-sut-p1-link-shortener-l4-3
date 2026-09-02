# plot decompose log — link-shortener (nested autonomous ignition)

- Run: run-20260901-002455-lights-out (inherited; classify verdict `inherited-l4`)
- Skill: `/faff-plot --autonomous`, nested subagent (single-level nesting, producers run in-context)
- Mode: git-only (tracking.tracker = none). No tracker containers created.
- Methodology slot: faffter-dark-methodology-agile-delivery
- Appetite: full

## Ignition

- Target resolved (live): container link-shortener, source prd, repo null. Not hard-coded.
- SelfRef: repo null (tracking.repo unset) → is_self false.
- Outward signal: `faff run-outward` → outward=true, reason outward-adopter. Not self-directed.
- Classify: `faff run-record-prd --classify` → inherited-l4. Reused $FAFF_RUN_DIR; minted nothing; did not call faff lights-out.
- PRD readiness (code-blind read, validated via `faff contract prd-readiness`): verdict admissible, stop_conditions_verifiable true, creative_licence tight (explicit non-goals + "keep it to the spine"). conformant.
- Root container recorded: `faff run-record-prd --prd-creative-licence tight --prd-root-container link-shortener` → written:true (ledger sha 9573e0…→2d2b4f…).
- Run-start assertion: `faff run-start` signals {target_resolved,outward,prd_present,prd_ambiguous:false,prd_admissible,coverage_measurable,coverage_covered:false} → verdict plan, reason coverage-thin. Proceed to decompose.

## Decomposition (agile-delivery lens, in-context)

Outcome: a running, persistent URL-shortener verifiable end to end.

- shape-level initiative → 1 initiative (I1: Link-shortener service live end-to-end). Manufacturing more would be fantasy for a micro-service.
- shape-level project → 2 projects under I1, sequenced value x risk:
  - P1 Core shorten to redirect on a real datastore (thinnest viable slice, highest risk: persistence + packaging). Covers G1,G2,G4,G5.
  - P2 TTL expiry and operational hardening (additive; no shipped value without the redirect chain). Covers G3,G5. blocked-by P1.
- shape-level epic → first-slice epics, right-sized, deps surfaced:
  - P1: E1 skeleton+compose+healthz+env-config; E2 migration+persistence [bb E1]; E3 POST /shorten mint [bb E2]; E4 GET /{code} redirect+404+restart [bb E3].
  - P2: E5 TTL honoured on read [bb E4]; E6 structured JSON errors [bb E4].

## Stop-rule decisions

- Stopped at first-slice epics; leaves deferred to /faff-prep + bottom-up tributaries.
- No branch halted for want of discovery — every epic has a nameable deliverable from the PRD acceptance criteria.

## Writes (git-only)

- Intake skeleton: `.faff/intake/2026-09-01-link-shortener-roadmap.md` (nested checklist; gitkeys per epic; deps; creation-skipped note).
- Gitkeys: E1 gk-20260901-r99do5, E2 gk-20260901-vesiuq, E3 gk-20260901-ro9j0r, E4 gk-20260901-fab1lp, E5 gk-20260901-h4jfk1, E6 gk-20260901-icaiwz.

## Step 5b — Proposed PRDRs authored

- `faff prdr new --provenance loop --status Proposed`:
  - 0001 Core shorten to redirect on a real datastore (container link-shortener-core-redirect, PRD-goals G1,G2,G4,G5).
  - 0002 TTL expiry and operational hardening (container link-shortener-ttl-hardening, PRD-goals G3,G5).
- Definition of done authored on each (born-verifiable, cited to PRD goals). `faff prdr validate` → OK (proposed-uncommitted).

## Step 5c — admit or land each loop-PRDR

| PRDR | yagni upper | admit disposition | outcome |
|---|---|---|---|
| 0001 | admit (trace_to_goal true, over_scope false, Phase-2 survived) | admit | accepted → landing branch prdr/0001-core-shorten-to-redirect-on-a-real-datastore |
| 0002 | admit (trace_to_goal true, over_scope false, Phase-2 survived) | admit | accepted → landing branch prdr/0002-ttl-expiry-and-operational-hardening |

- yagni: `faff prdr yagni --proposal admit --serves-goal --within-scope --challenge survived` (Phase-1 methodology admit; Phase-2 adversarial challenge in-context did not overturn — functional MVP cuts, narrowly cited, no gold-plating).
- admit: `faff prdr admit --actor loop --upper <admit> --supersedes-provenance none` → disposition admit (lower.covered true by default; no goal dropped).
- accept: `faff prdr accept --actor loop --admit-verdict <verdict>` → Status Accepted committed on each landing branch; returned to main. No PR opened (forge step out of scope for this pass).

Both 5c PRDRs terminated in admit-and-land. No parks.

## Parked / ambiguous

- None. Zero cancel/delete ops. No branch halted.

## Notes

- `.faffrc.yaml` shows a harness-written sentry key (`run_elapsed_ceiling_skip_when_live: true`) modified during the run; not authored by plot, left untouched.
- Accepted PRDR files live on their prdr/NNNN landing branches (small-PR flow), not on main's working tree.
