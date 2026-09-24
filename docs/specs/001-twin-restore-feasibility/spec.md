# Spec 1 — Twin-Restore Feasibility (M1 go/no-go)

**Status:** Draft v1.0
**Source of truth:** `docs/original/COLD_TURKEY_GUARDIAN_HIGH_LEVEL_TECH_SPEC.md` §14, §16, §25, §33;
task IDs T04–T05 in `docs/original/COLD_TURKEY_GUARDIAN_TASKS.md`. This document adds
nothing those don't already authorize — it scopes the slice and records research findings.
Tasks and their acceptance/verification stay in the Tasks doc; not duplicated here.

## Outcome

Answer the single question the whole MVP architecture rests on (tech spec §33.1):

> Can a scripted burst tear down both twin `guardiand` daemons before mutual restore
> completes — or does the twin pattern survive fast enough to be worth building on?

This is a GO/NO-GO gate, not a feature. Output is a decision, backed by a measured
latency/survive table, on whether M2+ proceeds as designed or the roadmap pulls in
Endpoint Security sooner (plan §11.1 reference in Tasks.md — that companion plan file
is not present in this repo; treat the citation as unresolved, see `AGENTS.md`).

## Scope

In scope (= T04 + T05):

- A minimal two-process stub of `guardiand` (`--role primary|twin`), same binary,
  `KeepAlive` LaunchDaemons, each twin able to rewrite the other's plist from an
  on-disk copy and `bootstrap` it back if missing.
- A red-team harness scripting the four documented bypass commands (tech spec §1, §24)
  plus the escalation sequence (`chflags noschg` → `bootout` both → `rm` both) at
  increasing speed.
- A separate check of whether disabling a daemon via **System Settings ▸ Login Items ▸
  Allow in the Background**, or clearing its BTM "Allowed" bit, defeats it outside the
  scripted-command path.
- Capturing `sfltool dumpbtm` before/after each run.

Out of scope (deferred to later slices, already covered by later tasks in the Tasks doc):

- Discovery/monitoring of Cold Turkey itself (T06–T09).
- Baseline store, `schg`, throttle/circuit breaker (T10, T11, T14) — the M1 stub restores
  unconditionally and unthrottled; that is intentional so the burst test measures the
  twin pattern's raw ceiling, not a throttled production version.
- Crypto/lock, IPC, CLI, packaging, audit (M4–M7).

## Acceptance (already stated in Tasks.md T04/T05; restated for this slice's readiness)

- T04: killing either stub daemon → launchd respawns it; deleting either plist → the
  peer restores + bootstraps it.
- T05: a written result table — for each of the 4 commands, the burst, and the
  escalation sequence at varying speed — whether the pair survived and the restore
  latency. `sfltool dumpbtm` captured before/after.

## Research findings feeding this slice (Exa, 2026-09-24)

- **launchd facts confirmed:** `KeepAlive: true` implies `RunAtLoad`; default
  `ThrottleInterval` is 10s but is an overridable ceiling (the plan's example value of
  5s is a valid override, not an error). `launchctl kickstart -k` kills-then-restarts
  on demand; `bootstrap`/`bootout` add/remove a service from a domain. These are the
  primitives T04/T05 exercise. Source: Apple launchd.plist(5), launchctl(1).
- **BTM / Login Items nuance (flag for T05, don't assume):** Apple's own guidance and
  third-party analysis (Eclectic Light Co., `sfltool` man page) describe **System
  Settings ▸ General ▸ Login Items ▸ Background Items** and BTM (`sfltool dumpbtm`) as
  primarily surfacing *login items and launch agents* (often user-domain or app-bundle
  registered via `SMAppService`). Whether a root **LaunchDaemon** in
  `/Library/LaunchDaemons` (Guardian's actual persistence mechanism) is independently
  toggleable from that same UI, versus only appearing in the BTM dump without a UI
  disable affordance, is **not confirmed** by these sources and must be observed
  directly on the test VM as part of T05 — do not assume the UI-disable bypass vector
  applies to system daemons the way it does to login items/agents.

## Edge cases relevant to the go/no-go call

- Restore latency variance if the burst script itself is slow to invoke `launchctl`
  serially vs. backgrounded/parallel commands — record actual command timing, not just
  wall-clock task duration, so the result isn't an artifact of the harness.
- A twin restoring the other mid-restore (races) — acceptable for this slice; not a
  failure unless it causes a crash loop (that's what T14's breaker exists for, later).

## Readiness

Sufficient to proceed directly to `agentic-sdd-plan` / implementation for T04, then T05.
No further clarification needed — acceptance and verification are already testable as
written in the Tasks doc.
