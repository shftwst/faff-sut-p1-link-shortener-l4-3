# faff SUT — P1 link-shortener (L4 lights-out) + run evidence

A system-under-test for [faff](https://github.com/shftwst/faff)'s **L4 lights-out** autonomous
lane: a small Go link-shortener built greenfield from a PRD, plus the **complete `.faff/` run
evidence tree** from the run that built it. The repo is kept intact — code *and* evidence — so a
single autonomous run can be inspected end to end.

## Start here

- **Finding write-up** → [`.faff/runs/run-20260901-002455-lights-out/finding-holdout-dod-tiering.md`](.faff/runs/run-20260901-002455-lights-out/finding-holdout-dod-tiering.md)
  — why the L4 code-blind holdout *permanently refused* a store-layer epic, traced to a
  spec-review DoD-tiering mismatch. The interesting artifact in this repo.
- **Run summary** → [`.faff/runs/run-20260901-002455-lights-out/summary.md`](.faff/runs/run-20260901-002455-lights-out/summary.md)
  — the beep-boop run report: 1 shipped, 1 parked, 4 blocked, stop reason `product-incomplete`.

## The run in one line

`run-20260901-002455-lights-out` (L4, git-only, ~6h55m, 41.45M tokens / ~$40). E1 shipped
end-to-end through the full L4 gate; **E2 was built correctly and refused at the fail-closed
holdout merge floor** (10/12 criteria met code-blind, 2 not verifiable at the store-layer stage);
E3–E6 blocked on E2's non-merge. The run escalated rather than claim the PRD done — which is the
behaviour under test.

## What's in the tree

| Path | What it is |
|------|-----------|
| `cmd/`, `internal/`, `migrations/`, `docker-compose.yml`, `Dockerfile` | E1's shipped service (health-check skeleton + Postgres wiring), merged to `main` |
| `docs/prd/`, `docs/specs/`, `BRIEF.md`, `RUNBOOK.md` | the PRD the run decomposed + the shipped design spec |
| `.faffrc.yaml` | the L4 lights-out dial config (adversarial review + spec_review slots, budget ceilings, backends) |
| `.faff/runs/<run>/` | run ledger, event chain, and per-issue `graft.md` / `holdout.json` / `holdout-evidence.txt` / `park.md` / spec-review fan-out |
| `.faff/{anchors,bundles,holdout,specs,prep,intake}/` | the rest of the evidence chain (redaction bundles, anchors, specs, gate inputs) |
| `run.log` | the full headless run transcript (stream-json) |

E2's store layer is **not** on `main` — it lives, unmerged, on branch
`gk-20260901-vesiuq-links-table-and-store`, which is where the two "shadow evidence" integration
tests (`TestDownMigrationDropsTableAndUpRestoresIt`, `TestQueryPathErrorCarriesNoDSN`) that
corroborate the parked holdout criteria live.

## Notes

- **git-only run** — no forge was involved during the build; merges were local to `main` and PRs
  used the `0` sentinel. This remote was added afterwards purely to publish the evidence.
- **Fixtures, not secrets** — the Postgres DSNs in the review diffs (`linkshortener:linkshortener`,
  `u:p`, `user:sup3rSecret`) are throwaway compose/dev values and DSN-scrubber *test* strings for
  this disposable SUT. No real credentials.
