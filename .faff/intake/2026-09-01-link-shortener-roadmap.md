# Roadmap skeleton — link-shortener

- Planned by `/faff-plot --autonomous` (nested under run `run-20260901-002455-lights-out`)
- Source: PRD `docs/prd/link-shortener.md` (Active), brief `BRIEF.md`
- Methodology lens: `faffter-dark-methodology-agile-delivery` (value x risk, right-sized, surfaced deps)
- Mode: git-only (tracker `none`). No tracker containers were created. This file is the skeleton of record; first-slice epics carry stable gitkeys for `faff queue-state derive`.
- Root container recorded onto the run ledger: `link-shortener` (creative licence: tight)

## Outcome

A running, persistent URL-shortener micro-service that mints a 7-char base62 code for an absolute URL, redirects the code to the URL, honours an optional TTL, survives an api restart, and is packaged under docker-compose with a `/healthz` liveness check. Every acceptance scenario is covered by an automated test suite that passes.

## PRD goals (referenced by the project PRDRs)

- G1: mint a code for an absolute URL and redirect that code to the URL
- G2: codes survive an api container restart (proves the datastore is real)
- G3: an optional `ttl_seconds` is honoured — an expired code stops resolving
- G4: the packaged stack starts healthy and answers `/healthz` quickly
- G5: an automated test suite covers every acceptance scenario and passes

## Skeleton (initiatives -> projects -> first-slice epics)

- [ ] Initiative I1: Link-shortener service live end-to-end
  - [ ] Project P1: Core shorten to redirect on a real datastore
    - one-line: the MVP spine — mint, redirect, real Postgres persistence, docker-compose packaging, `/healthz`, survives restart. Covers G1, G2, G4, G5.
    - [ ] Epic E1: Packaged service skeleton, docker-compose (api + Postgres), `GET /healthz` -> 200, config from environment <!-- gitkey:gk-20260901-r99do5 -->
    - [ ] Epic E2: Schema migration and persistence layer (links table for code -> url with created/expiry columns; migration applied on start) <!-- gitkey:gk-20260901-vesiuq -->
    - [ ] Epic E3: `POST /shorten` mint endpoint (validate absolute http(s) URL, mint 7-char base62 code, persist, return 201 + `code`) <!-- gitkey:gk-20260901-ro9j0r -->
    - [ ] Epic E4: `GET /{code}` redirect, unknown -> 404, restart-persistence (302 with `Location` equal byte-for-byte; code resolves after api restart) <!-- gitkey:gk-20260901-fab1lp -->
  - [ ] Project P2: TTL expiry and operational hardening
    - one-line: additive to the working spine — honour `ttl_seconds` on read, structured JSON errors. Covers G3, G5.
    - [ ] Epic E5: TTL expiry honoured on read (`ttl_seconds` stored as expiry; expired code -> 404) <!-- gitkey:gk-20260901-h4jfk1 -->
    - [ ] Epic E6: Structured JSON error responses across endpoints <!-- gitkey:gk-20260901-icaiwz -->

## Dependencies (blocker / blocked-by)

- E2 blocked-by E1 (needs the running stack and datastore)
- E3 blocked-by E2 (needs persistence)
- E4 blocked-by E3 (needs the write path to store codes)
- E5 blocked-by E4 (needs the read path)
- E6 blocked-by E4 (hardens the live endpoints)
- Project P2 blocked-by Project P1 (P2 epics depend on P1's endpoints)

## Sequencing rationale (value x risk)

- E1 first: de-risks packaging and the container topology early (principle 7) and lights up the `/healthz` acceptance criterion.
- E2 -> E3 -> E4: builds the observable mint-and-redirect increment, crossing persistence, write, and read slices in order so the core value (shorten then follow) lights up as soon as E4 lands. E4 also carries the restart-persistence proof.
- P2 after P1: TTL and structured errors are additive to a working spine, sequenced second because they create no shipped value without the redirect chain.

## Stop-rule notes

- The skeleton stops at first-slice epics. Leaf tickets (individual handlers, test cases, config keys) grow later from `/faff-prep` specs and the bottom-up tributaries.
- No branch was halted for want of discovery: every epic above has a nameable deliverable derivable from the PRD acceptance criteria.

## Open questions (carried from the PRD)

- Language and runtime (TypeScript on Node 20 or Go 1.22) and the specific datastore driver are implementation choices resolved at the architecture step. They do not change the acceptance criteria.

## Hand-off

- Prep the first slice for build: highest-sequenced first-slice epic is E1 (`gk-20260901-r99do5`).
- Audit the roadmap join-up with `/faff-map` against this file.
