# P1 Runbook — run + observe + score (L4 lights-out under claude-box)

Turnkey sequence for the L4 lights-out proof on this SUT, plus how to observe and score it.
Everything runs in the cage, genuinely git-only. Since FAFF-499 shipped, decompose and build happen
in a single arm cycle: `/faff:faff-beep-boop` plots the PRD and then falls through to the build lanes
in the same run. There is no separate plan stage any more.

## Operator framing (operator-only — NEVER pasted into the loop)
This restates what P1 measures and how it scores. It lives here, not in BRIEF.md, so it can never
reach the loop: feeding intent framing to the loop is teaching-to-the-test (it pre-warns the loop of
the exact failure modes the suite exists to catch — see FAFF-547). BRIEF.md is fed to the loop and
must stay a neutral build brief.

Intent: the DoD is 100% born-verifiable (every criterion is a real HTTP exchange), so the code-blind
evaluator should produce clean verdicts with zero needs-human punts. This is the first real exercise
of architecture → env → evaluate on a non-faff product.

Completion rubric (the scoring restatement — what "done" means for the operator, never for the loop):
1. `POST /shorten`, `GET /{code}`, and `GET /healthz` implemented and serving under docker-compose.
2. A schema migration creates the codes table; codes persist across an api container restart.
3. Expiry honoured: an expired code returns 404.
4. `docker compose up` brings api+db healthy; `GET /healthz` returns 200 within 60s.
5. An automated test suite covers every Scenario above and passes.

## 0. Pre-flight (Colima / Docker reachable from faff's shell)
    colima status            # running? if not: colima start
    docker context show      # should be 'colima'
    docker info  >/dev/null && echo "docker OK"
    docker compose version   # env-compose shells `docker compose`
If `docker info` fails here, env-compose will emit a `status: failed` handle (clean, visible) — fix Colima first.
Note: inside the cage, `faff env` uses claude-box's nested rootless docker engine, not host Colima;
this host pre-flight only matters when you drive a lane observably on the host (section 4).

Optional infra profile:

    faff profile show --json    # exit 3 = no profile → architecture proposes from BRIEF.md alone (fine for P1)
    # faff profile mine         # a fresh repo has little to mine; the BRIEF's stack-pref section is the real signal

## 1. Git-only via the tracker pin (FAFF-808)

`.faffrc.yaml` sets `tracking.tracker: none`, the reserved git-only sentinel. Every skill resolves
git-only from that pin before it attempts any MCP discovery, and a skill that reads `git-only` must
not upgrade to tracker-mode even when the Linear MCP is visible in the cage (`tracker.js`, FAFF-808).
Confirm it with a pure read, no cage needed:

    faff tracker probe            # prints: git-only

This replaces the old `--strict-mcp-config --mcp-config '{"mcpServers":{}}'` lever. The Linear MCP is
still bind-mounted into the cage via `~/.claude.json`, but the pin keeps plot, prep, and graft on the
git-only queue regardless, so nothing reaches your Faff Linear workspace. plot writes the skeleton to
`.faff/intake/` and PRDRs to `docs/prdr/` on disk; the drain works the git-only queue. (The
adversarial engines are direct-API, not MCP, so they are unaffected either way.)

## 2. Config — already dialled, nothing to add

`.faffrc.yaml` carries it all (`faff config check` is clean): the git-only pin above, budget ceiling
(`tokens: 30000000`), adversarial `review` + `spec_review`, and the `nvidia-glm` + `gemini-gemma`
engines (keys forwarded via `.env.claude-box`). `models:` is unset, so every stage runs on the
session model (Opus in the cage), a strong default; add a `models:` map only to route stages
elsewhere. `FAFF_RUN_DIR` is in `CLAUDE_BOX_EXTRA_VARS` so the drain adopts the armed L4 ledger (the
fix for the earlier L3 degrade; the mint always armed 8/8). `FAFF_RUN_HEARTBEAT_STALE_SECS=7200` is
forwarded too: a long plot/prep producer subagent does not tick the parent run heartbeat, so under
the larger adversarial deadlines (`adversarial.timeout: 600`, multi-round via the convergence yield)
a healthy producer can age the heartbeat past the 900s default and trip a false-positive sentry
abort. The 2h window clears the worst producer runtime; the token ceiling remains the runaway backstop.

## 3. The run — preflight, mint, drain

Run from the repo root. `faff lights-out` mints the armed L4 ledger and prints the run dir; the drain
adopts it via `FAFF_RUN_DIR`. Inside the drain, section 0a reads coverage: on this greenfield PRD it
is thin, so run-start returns `plan`, `/faff:faff-plot --autonomous` decomposes the PRD into
first-slice epics and authored PRDRs, and the run falls through to prep -> build in the same pass (FAFF-499, L4
unconditional). Architecture / env / evaluate fire, and the holdout gates the merge.

    # cage 1 — preflight + mint; write the run dir to a bind-mounted file.
    # `env -u FAFF_RUN_DIR` strips any stale value a prior run left exported in the
    # shell: cage 1 must mint with FAFF_RUN_DIR unset. A leftover value pointing at a
    # gone run dir makes `faff budget check` resolve a named-but-absent ledger, which
    # flips the budget guardrail to reachable:absent and REFUSES the whole preflight.
    env -u FAFF_RUN_DIR claude-box -p 'Run exactly these, nothing else, then stop:
      faff lights-out --check || exit 1
      faff lights-out --json | jq -r .run_dir > .faff/lights-out-rundir'

    # host — read the run dir back (exact-path mount => same path inside the cage)
    run_dir=$(cat .faff/lights-out-rundir)

    # cage 2 — the drain; FAFF_RUN_DIR forwarded via CLAUDE_BOX_EXTRA_VARS
    export FAFF_RUN_DIR="$run_dir"
    claude-box --model claude-opus-4-8 --effort high --verbose --output-format stream-json -p "/faff:faff-beep-boop" | tee run.log

    # host — the exit contract (pure reader; path is valid on the host too)
    faff disposition --run-dir "$run_dir"

The PRD at `docs/prd/link-shortener.md` is auto-discovered (`faff prd list` -> container
`link-shortener`); no jot, no target flag. The drain is the top-level caged agent, so claude-box
applies `--dangerously-skip-permissions` for you: no nested `claude`, no manual flag. Cage 1 must
mint with `FAFF_RUN_DIR` unset; claude-box forwards it only when non-empty (`libcage.sh`), so an
empty shell var is skipped, but a value a prior run left exported IS forwarded and refuses the mint
on the budget guardrail. The `env -u FAFF_RUN_DIR` on the cage-1 line above makes this
unconditional regardless of shell state.

Optional coverage read, once the run has been through section 0a's plot pass, to confirm the PRDRs
cite every PRD goal verbatim (`.covered` matches on exact goal strings; a paraphrased `prd_goal`
reports false). Pure CLI, no cage:

    faff prdr coverage --container link-shortener --prd-goals '<PRD goals as a JSON array>'

## 4. If a lane doesn't auto-fire, drive it directly (this is the value — establishing the wiring)
    faff env compose-gen --profile <p>     # → ProvisionPlan + compose file
    faff env up   --plan <plan>            # docker compose up -d + health-wait → env-handle (status: ready|failed)
    faff env seed --plan <plan>            # seed synthetic data
    faff holdout verdicts --association <json>   # pure bridge: reads the evaluator's persisted
                                                  # .faff/holdout/<key>.json verdicts (the
                                                  # evaluator slot already exercised the live
                                                  # endpoints to produce them) into prdr
                                                  # coverage's --dod-verdicts shape
    faff prdr coverage --prd-goals '<JSON array of DONE goals>' --dod-verdicts ...   # roll the verdicts into prd-satisfied
    faff env down --project <project>      # tear down (ephemeral)

## 5. Observe (the run's real signal)
    faff events read --run <id>     # the timeline
    faff audit <run-id>             # who/what/why forensics
    # The env-handle block (did docker stand up + health-check pass?)
    # The holdout-verdict block (did the evaluator HIT the endpoints, or read the code? prose → needs-human?)

## 6. Score P1 (B1/B2/B3/B6) — "did the behaviour occur + was the boundary respected", NOT "did it build it"
- [ ] B1 architecture: proposed a production-shaped stack (Node/Postgres-ish, compose-ready) — not a toy single-file
- [ ] B2 env: `faff env up` stood up api+db, health-check passed, seed ran
- [ ] B3 evaluate: holdout-verdict exercised the RUNNING endpoints (302/404/code-shape) — evidence shows HTTP calls, not source-reading
- [ ] B3-integrity: code-blind held (no diff/code in the evaluator's evidence)
- [ ] zero punts: a 100%-born-verifiable DoD produced NO needs-human (if it punted, why?)
- [ ] B6 terminate: the run reached a terminal run-done verdict, didn't stall/loop
- FIRST FAILURE RUNG = the binding constraint = the finding to take back to faff's backlog.

## Notes

- `faff disposition --run-dir "$run_dir"` reads the ledger the drain adopted. If section 0a's plan
  branch mints a distinct build run dir under `.faff/runs/`, read the newest one instead.
- `faff lights-out --check` (cage 1) mints nothing; confirm an 8/8 armed banner before the `--json`
  mint. Worktree root defaults to `$HOME/.faff/worktrees/<repo>`, outside the repo, so the isolation
  floor passes with no config.
- `FAFF_INTEGRITY_BOUNDARY` stays unset: with no boundary set, `faff lights-out --check` reports
  `corrective-integrity` as an **advisory degrade, not a refusal** (FAFF-525) — admission proceeds on
  the FAFF-518 digest custody floor. Setting a boundary is optional; it only buys the stronger
  mount-asserted basis once the cage read-only-mounts the integrity dirs (FAFF-517). Compose it with
  `faff integrity-boundary` if/when that lands (an automating cage supplies it; FAFF-514) — never a
  fabricated value (fabricating it to clear the gate is the lying attestation FAFF-525 exists to
  avoid). All three **dial-coherence** legs are already satisfied by this SUT's `.faffrc.yaml`: the
  two explicit dials (`slots.review`, `slots.spec_review`) plus `gates.fallback`, fail-closed by
  default (FAFF-522).
- `slots.transport` is intentionally unset: the default occupant `faffter-noon-transport-private-network`
  is zero-config and resolves the evaluator's base host from the runtime provision context. Under the
  cage's nested rootless docker the topology is `dind-in-cage`, so it rebases onto a shared
  user-defined docker network (not host-gateway, which does not route under rootless) and tears it
  down itself. `lights-out --check` does not gate on transport; this is not a gap.
