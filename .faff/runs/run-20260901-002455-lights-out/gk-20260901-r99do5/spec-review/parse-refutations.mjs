#!/usr/bin/env node
// Parse fan-out.mjs LensResult[] into aggregate.mjs Refutation[] entries.
// Maps review-call.mjs per-lens exit -> outcome, and parses exit-0 stdout
// (### [severity]: title  + claim/evidence/predicted_consequence/spec_anchor)
// into gating objections. Writes per-lens transcripts round-<n>-<lens>.md.
import { readFileSync, writeFileSync } from "node:fs";

const [resultPath, scratch, roundStr] = process.argv.slice(2);
const round = roundStr || "1";
const results = JSON.parse(readFileSync(resultPath, "utf8"));

// exit -> {outcome, kind}
function mapExit(code) {
  if (code === 0) return { outcome: "exit0" };
  if (code === 5 || code === 12) return { outcome: "unavailable", kind: "infra-configured" };
  if (code === 6 || code === 2 || code === 4 || code === 7) return { outcome: "unavailable", kind: "config-fault" };
  // Any other non-zero (1 OTHER, 8 DEADLINE, 9 MANDATORY_OUTAGE, 10 MALFORMED, 11) —
  // fail safe: treat as infra-configured unavailable (swing-capable outage, never silent approve).
  return { outcome: "unavailable", kind: "infra-configured" };
}

const SEV = /^#{2,4}\s*\[?(critical|major|minor|observation)\]?\s*:?\s*(.*)$/i;
const FIELD = /^\s*[-*]?\s*(claim|evidence|predicted_consequence|spec_anchor)\s*:\s*(.*)$/i;

function parseObjections(stdout) {
  // Drop the transport header line(s) starting with '## Adversarial findings'
  const lines = stdout.split("\n");
  const objs = [];
  let cur = null;
  let noObjection = false;
  for (const line of lines) {
    if (/no\s+\w+\s+objection\./i.test(line)) noObjection = true;
    const m = line.match(SEV);
    if (m) {
      if (cur) objs.push(cur);
      cur = { severity: m[1].toLowerCase(), summary: (m[2] || "").trim() };
      continue;
    }
    if (cur) {
      const f = line.match(FIELD);
      if (f) {
        const key = f[1].toLowerCase();
        cur[key] = f[2].trim();
      }
    }
  }
  if (cur) objs.push(cur);
  return { objs, noObjection };
}

const refutations = [];
for (const r of results) {
  const lens = r.lens;
  const m = mapExit(r.exit);
  // transcript
  const header = (r.stdout || "").split("\n").find((l) => l.startsWith("## Adversarial findings")) || "";
  try { writeFileSync(`${scratch}/round-${round}-${lens}.md`, r.stdout || `(exit ${r.exit})\n${r.stderr || ""}`); } catch {}
  let model = "unknown";
  const hm = header.match(/Adversarial findings\s+[—-]\s+(\S+)/);
  if (hm) model = hm[1];
  if (m.outcome === "unavailable") {
    refutations.push({ lens, outcome: "unavailable", kind: m.kind, objections: [], model });
    continue;
  }
  // exit 0 — parse
  const { objs, noObjection } = parseObjections(r.stdout || "");
  const gating = objs.filter((o) => ["critical", "major", "minor"].includes(o.severity));
  const outcome = gating.length > 0 ? "refuted" : "clear";
  const objections = objs.map((o) => {
    const out = { severity: o.severity };
    for (const k of ["claim", "evidence", "predicted_consequence", "spec_anchor", "summary"]) {
      if (typeof o[k] === "string" && o[k]) out[k] = o[k];
    }
    return out;
  });
  refutations.push({ lens, outcome, kind: undefined, objections, model });
}

process.stdout.write(JSON.stringify(refutations, null, 2));
