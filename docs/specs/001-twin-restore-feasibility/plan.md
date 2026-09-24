# Plan 1 — Twin-Restore Feasibility (T04, T05)

**Companion:** `spec.md` (this directory). Technical basis: tech spec §14 (launchd
design), §16 (recovery strategy), §25 (protection of Guardian itself).

## Approach — T04 (twin-restore stub)

- One binary, `cmd/guardiand`, role selected by `--role=primary|twin` (spec §12).
- Two plists, `com.coldturkeyguardian.primary.plist` / `...twin.plist`, both
  `RunAtLoad + KeepAlive=true` (KeepAlive implies RunAtLoad — no need to set both
  explicitly, but tech spec §14's example sets both for clarity; keep that).
  `ThrottleInterval` left at a low override (5s) for this stub — production throttling
  is T14's job, not this slice's.
- Each twin, on a fixed poll (no fsnotify/kqueue yet — that's T08, out of scope):
  1. Checks the peer's plist exists at its known path; if missing, copies it back from
     an on-disk reference copy placed at install time, then
     `launchctl bootstrap system/... <plist>`.
  2. Does **not** need to check peer *process* liveness — `KeepAlive` + launchd already
     respawns a killed process; T04 only proves the *plist/bootstrap* leg of restore,
     which launchd cannot do on its own once a job is `bootout`.
- No baseline/manifest/schg yet (T10/T11) — the "on-disk copy" for this stub can be a
  plain file placed next to the plists at build/install time in `packaging/`.

## Approach — T05 (red-team harness)

- `test/redteam/burst.sh` (or Go test harness) issuing, individually then as a fast
  burst:
  1. `launchctl unload ~/Library/LaunchAgents/...` (N/A until Cold Turkey targets exist;
     for T05 substitute the Guardian stub's own agent-equivalent or run this against
     the twin daemons directly per tech spec §24's escalation path, since Cold Turkey
     discovery is a later milestone — confirm this substitution with the acceptance
     table's column headers before running).
  2. `launchctl bootout gui/$UID ...` / `system/...` against each twin.
  3. `pkill -9 guardiand` (both roles).
  4. `chflags noschg` (no-op at this stage — flags land in T11) → `bootout` both →
     `rm` both plist + binary, at increasing speed (e.g. 2s, 500ms, 100ms gaps).
- Capture: wall-clock timestamp per command, timestamp of observed restore (poll
  `launchctl print` for the job re-appearing), `sfltool dumpbtm` output immediately
  before and after each run.
- Separately: manually toggle the twin's Login Item / Background Item entry in System
  Settings (if it surfaces there — per spec.md's flagged research gap, confirm this
  first) and record whether that alone stops the daemon without any `launchctl` call.
- Output: a markdown table (attack → survived Y/N → restore latency ms) checked into
  `test/redteam/results-m1.md`, and the go/no-go decision line at the top.

## Decisions carried over from the tech spec (not re-litigated here)

- Twins, not a single daemon (tech spec §8: single daemon rejected as weak).
- `kickstart -k` + `bootstrap` are the correct launchd verbs for forced restart vs.
  re-registration respectively (confirmed against `launchctl(1)` — kickstart restarts
  an existing job on demand; bootstrap/bootout add/remove the job definition itself,
  which is what a deleted-plist scenario needs).

## Verification

- T04: `pkill -9` each role → observe respawn via launchd (`launchctl print`); `rm` each
  plist → observe peer restores + re-bootstraps within target window (no fixed SLA yet
  — T05 establishes the actual number).
- T05: scripted run per spec.md acceptance; `sfltool dumpbtm` diff before/after.

## Out of scope / explicitly deferred

- Any throttling, circuit breaker, `schg`, baseline manifest, or Cold Turkey-specific
  discovery — all later milestones. Building them into T04/T05 would make the go/no-go
  measurement about the wrong thing (a throttled, hardened system) instead of the raw
  question (can two bare KeepAlive daemons out-race a fast teardown at all).

## Unresolved boundary

None blocking — T04 has no approval requirement (Tasks.md: Risk M, Approval none).
T05 requires a **STOP after this task** human review of results before M2+ proceeds
(Tasks.md gate) — that gate is procedural (review the table), not a pre-authorization
for this planning/implementation work itself.
