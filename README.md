# Cold Turkey Guardian

Two mutually-restoring root LaunchDaemons that wrap [Cold Turkey
Blocker](https://getcoldturkey.com/), auto-reversing the four commands people habitually
use to disable it during an urge, and gate the *only* clean unlock behind an
accountability partner's cryptographic key. See `docs/original/` for the full rationale,
threat model, and design.

**Status:** pre-MVP, milestone M0/M1 (module skeleton + go/no-go feasibility spike).
This is a personal-use friction tool, not a security product against a determined
attacker — see `docs/original/COLD_TURKEY_GUARDIAN_HIGH_LEVEL_TECH_SPEC.md` §2–3, §33
for explicit non-goals and accepted limits.

## Docs

- `docs/original/COLD_TURKEY_GUARDIAN_HIGH_LEVEL_TECH_SPEC.md` — architecture, threat
  model, decisions.
- `docs/original/COLD_TURKEY_GUARDIAN_TASKS.md` — the authoritative task breakdown
  (T01–T34) with acceptance/verification/status per task.
- `docs/specs/001-twin-restore-feasibility/` — scope and approach for the M1 go/no-go
  slice (T04–T05).
- `AGENTS.md` — canonical doc pointers, verification commands, and approval boundaries
  for agents working in this repo.

## Layout

```
cmd/guardiand/   enforcement daemon (--role=primary|twin)
cmd/guardian/    CLI client (status, lock, unlock, break-glass, ...)
internal/*       discovery, policy, monitor, restore, baseline, flags,
                 crypto, lockstate, ipc, audit — see tech spec §12
packaging/       LaunchDaemon plists, installer scripts, pkg metadata
test/            red-team and chaos test harnesses
```

## Build

Requires Go 1.26+ on macOS.

```sh
make build          # go build ./cmd/guardiand and ./cmd/guardian into dist/
make vet             # go vet ./...
make test            # go test ./...
```

## Unsigned dev package

`make pkg-unsigned` builds an **unsigned** `.pkg` (`dist/guardian-dev.pkg`) for
fast dev-loop installs on a disposable macOS VM — never run this against a
production machine; it installs root LaunchDaemons.

```sh
make pkg-unsigned
sudo installer -pkg dist/guardian-dev.pkg -target /
launchctl print system/com.coldturkeyguardian.primary
```

Signed, notarized packaging is a later milestone (T23) requiring Apple Developer
Program enrollment.

## Requirements/approvals

See `AGENTS.md` for the dependency allowlist, no-cgo constraint, and the list of
changes that require explicit approval before implementation (schg on Cold Turkey's own
files, the break-glass default delay, enabling the `os_log` cgo build, etc).
