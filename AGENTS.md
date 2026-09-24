# AGENTS.md

## Canonical documents
- Technical spec: `docs/original/COLD_TURKEY_GUARDIAN_HIGH_LEVEL_TECH_SPEC.md`
- Task breakdown / status: `docs/original/COLD_TURKEY_GUARDIAN_TASKS.md`
- `COLD_TURKEY_GUARDIAN_IMPLEMENTATION_PLAN.md` is referenced by the tasks doc as a
  companion but does not exist in this repo — treat plan-section citations (`plan §…`)
  in the tasks doc as unresolved until that file is supplied or reconstructed.

## Verification commands
- `go vet ./...`
- `go build ./...`
- `CGO_ENABLED=0 go build ./...` (no-cgo constraint, spec §NFR4/plan §11.2)

## Boundaries / authorization source
- Work proceeds top-to-bottom by milestone (M0–M7) per the tasks doc; do not skip **M1**
  (T05, the twin-restore red-team gate) — it decides whether the architecture holds.
- Approval-required items (see tasks doc "Blocked / approval-required work"): a 4th
  dependency beyond `golang.org/x/sys`, `fsnotify`, `howett.net/plist` (T02); applying
  `schg` to Cold Turkey's own files (T11); the break-glass default delay (T18); enabling
  the `os_log` cgo build (T26); Apple Developer enrollment for notarization (T23).
- Module: `github.com/nawodyaishan/ctg`, Go 1.26 toolchain, no cgo.

## Slice specs
- Slice 1 (M1 twin-restore go/no-go, T04–T05): `docs/specs/001-twin-restore-feasibility/{spec.md,plan.md}`.
  Acceptance/verification for T04/T05 stay authoritative in the Tasks doc; these files
  scope the slice and record supporting research, not a duplicate task list.

## Status
T01 done (module + package skeleton builds clean). See the tasks doc for current
per-task status; slice 1 above is the next ready work (T02 dependency pinning, then T03,
T04, T05).
