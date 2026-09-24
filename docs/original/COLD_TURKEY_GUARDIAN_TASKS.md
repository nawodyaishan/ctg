# Cold Turkey Guardian — Implementation Tasks

**Companion to:** `COLD_TURKEY_GUARDIAN_HIGH_LEVEL_TECH_SPEC.md` + `COLD_TURKEY_GUARDIAN_IMPLEMENTATION_PLAN.md`
**Status:** Tasks v1.0

Each task lists: **objective · allowed paths · acceptance · verification · deps · risk · approval · status**. Risk = L/M/H. Status starts `TODO`. All tasks are for one focused pass. Verify on a disposable macOS VM (14/15/26) with Cold Turkey installed, unless noted.

> **Golden rule:** work top-to-bottom by milestone. **Do not skip M1** — it decides whether the whole architecture holds.

---

## Track summary

34 tasks across M0–M7. Critical path: T01→T02→T05→(T10,T11,T12,T13)→(T14,T15,T16)→T18→(T20,T21,T22)→T25→T27→T30→T33. Parallel-safe groups listed in §"Parallel-safe groups".

## Prerequisites

- Go 1.26 toolchain (macOS 12–26 supported). **[Fact]**
- A disposable macOS VM (Apple silicon + at least one Intel pass) with Cold Turkey Blocker installed and a **locked** test block.
- For M5+: Apple Developer Program membership; Developer ID Application + Developer ID Installer certs in the keychain; a `notarytool` keychain profile. **[Fact]** (Stop condition — plan §11.4.)

---

## M0 — Skeleton & toolchain

### T01 — Repo & module skeleton
- **Objective:** create the Go module and directory layout from plan §5.
- **Allowed:** whole repo (greenfield).
- **Acceptance:** `go build ./...` succeeds with empty stubs for `cmd/guardiand`, `cmd/guardian`, and each `internal/*` package.
- **Verification:** `go vet ./... && go build ./...`.
- **Deps:** none. **Risk:** L. **Approval:** none. **Status:** TODO.

### T02 — Dependency pinning
- **Objective:** add only the approved deps: `golang.org/x/sys`, `github.com/fsnotify/fsnotify`, `howett.net/plist`.
- **Allowed:** `go.mod`, `go.sum`.
- **Acceptance:** module graph contains exactly those three (plus stdlib); no cgo.
- **Verification:** `go mod tidy && go list -deps ./... | sort -u` review; `CGO_ENABLED=0 go build ./...` succeeds.
- **Deps:** T01. **Risk:** L. **Approval:** required if any 4th dep appears (plan §11.2). **Status:** TODO.

### T03 — Local build/sign Makefile (unsigned path)
- **Objective:** `make build`, `make pkg-unsigned` (pkgbuild/productbuild without signing) for fast dev-loop installs.
- **Allowed:** `packaging/Makefile`, `packaging/*.plist`, `packaging/distribution.xml`, `packaging/scripts/*`.
- **Acceptance:** produces a `.pkg` that installs the two daemons + CLI to dev locations.
- **Verification:** `sudo installer -pkg dist/guardian-dev.pkg -target /`; `launchctl print system/com.coldturkeyguardian.primary` shows the job. **[Fact: launchctl print]**
- **Deps:** T01. **Risk:** M (touches install/boot state — rollback: `guardian uninstall` or manual `bootout`+`rm`). **Approval:** none. **Status:** TODO.

---

## M1 — Feasibility spike (GO/NO-GO)

### T04 — Twin-restore stub
- **Objective:** minimal two-daemon stub (same binary, `--role`) with `KeepAlive`, each rewriting the other's plist from an on-disk copy and `bootstrap`-ing it if missing.
- **Allowed:** `cmd/guardiand` (stub), `packaging/*.plist`.
- **Acceptance:** killing either daemon → launchd respawns it; deleting either plist → the peer restores + bootstraps it.
- **Verification:** scripted `pkill`/`rm`; observe both back within target window.
- **Deps:** T03. **Risk:** M. **Approval:** none. **Status:** TODO.

### T05 — Red-team burst harness + BTM test  ★ decision gate
- **Objective:** script the four bypass commands and escalation bursts (`chflags noschg` → `bootout` both → `rm` both) at increasing speed; separately test disabling the daemon via **System Settings ▸ Login Items ▸ Allow in the Background** and via clearing the BTM "Allowed" bit.
- **Allowed:** `test/redteam/*`.
- **Acceptance:** a written result table: for each attack, did the pair survive, and restore latency.
- **Verification:** run on VM; capture `sfltool dumpbtm` before/after. **[Fact: BTM/sfltool]**
- **Deps:** T04. **Risk:** H (this validates the load-bearing assumption). **Approval:** **STOP after this task** — review results with owner; decide whether to proceed with M2+ or pull ES forward (plan §11.1). **Status:** TODO.

---

## M2 — Discovery & monitoring

### T06 — Cold Turkey discovery
- **Objective:** at install/runtime, discover Cold Turkey's launchd labels, plist paths, app bundle, support dir; never hardcode. Persist into policy.
- **Allowed:** `internal/discovery`, `internal/policy`.
- **Acceptance:** on a real install, returns the actual labels/paths; tolerant of version differences and missing pieces.
- **Verification:** unit test with fixture plists (`howett.net/plist` decode) + live run comparing to `launchctl print gui/$UID`. **[Fact: plist, launchctl]**
- **Deps:** T02. **Risk:** M. **Approval:** none. **Status:** TODO.

### T07 — Process monitor (kern.proc + NOTE_EXIT)
- **Objective:** detect Cold Turkey / twin process death.
- **Allowed:** `internal/monitor`.
- **Acceptance:** emits an event within grace window when a target PID exits; name-match falls back to exec path when `P_comm` ≥15 chars.
- **Verification:** unit test around `SysctlKinfoProcSlice("kern.proc.all")`; integration killing Cold Turkey. Use `EVFILT_PROC`/`NOTE_EXIT` for the twin. **[Fact: x/sys/unix]**
- **Deps:** T06. **Risk:** M. **Approval:** none. **Status:** TODO.

### T08 — File monitor (kqueue via fsnotify, parent-dir watches)
- **Objective:** detect deletion/rename of Cold Turkey + Guardian critical files.
- **Allowed:** `internal/monitor`.
- **Acceptance:** deleting the support dir or a plist fires a Remove/Rename event; FD usage bounded (watch parent dirs, not every file).
- **Verification:** integration `rm -rf` the support dir → event; check FD count vs `kern.maxfilesperproc`. **[Fact: fsnotify kqueue]**
- **Deps:** T06. **Risk:** M. **Approval:** none. **Status:** TODO. **Parallel-safe with T07.**

### T09 — Job-state monitor (launchctl reconcile)
- **Objective:** detect `unload`/`bootout` of Cold Turkey jobs via periodic `launchctl print` reconciliation.
- **Allowed:** `internal/monitor`, `internal/restore` (shared launchctl wrapper).
- **Acceptance:** unloading a Cold Turkey agent is detected within the poll interval.
- **Verification:** integration `launchctl bootout gui/$UID/<label>` → detection event. **[Fact: launchctl]**
- **Deps:** T06. **Risk:** M. **Approval:** none. **Status:** TODO. **Parallel-safe with T07, T08.**

---

## M3 — Restore & baseline

### T10 — Baseline store + manifest
- **Objective:** create read-only, root-owned baseline copies of Cold Turkey structural files + Guardian bin/plists, with a SHA-256 manifest; refresh only when unlocked.
- **Allowed:** `internal/baseline`, `packaging/scripts/postinstall`.
- **Acceptance:** manifest verifies; refresh refused while locked.
- **Verification:** unit test manifest verify; integration confirm refuse-when-locked.
- **Deps:** T06. **Risk:** M. **Approval:** none. **Status:** TODO.

### T11 — chflags schg helpers
- **Objective:** apply/clear `schg` on Guardian (and optionally Cold Turkey) files.
- **Allowed:** `internal/flags`.
- **Acceptance:** files show `schg`; documented that root can clear it (friction only).
- **Verification:** `ls -lO`; unit test `Chflags`/`SF_IMMUTABLE`. **[Fact: x/sys/unix]**
- **Deps:** T02. **Risk:** M. **Approval:** **required** before applying `schg` to Cold Turkey's own files (plan §11.3 — updater conflict risk). **Status:** TODO. **Parallel-safe with T10.**

### T12 — Restore engine (launchctl actions)
- **Objective:** re-`bootstrap`/`kickstart -k` unloaded jobs; relaunch dead app after grace.
- **Allowed:** `internal/restore`.
- **Acceptance:** unloaded/booted-out job returns; killed app returns.
- **Verification:** integration for commands #1–#3. **[Fact: launchctl verbs]**
- **Deps:** T09, T10. **Risk:** M. **Approval:** none. **Status:** TODO.

### T13 — File restore (data-preserving rule)
- **Objective:** restore deleted structural files from baseline + re-apply `schg`; for Cold Turkey's mutable DB, restore only on absence (never overwrite a present, non-empty DB).
- **Allowed:** `internal/restore`, `internal/baseline`.
- **Acceptance:** `rm -rf` support dir → restored; a present DB is never clobbered.
- **Verification:** integration command #4; assert DB contents preserved when present. **[Assumption: restore-while-running safe — record result]**
- **Deps:** T10, T11. **Risk:** H (data safety). **Approval:** none, but flag test result. **Status:** TODO.

### T14 — Throttle + circuit breaker
- **Objective:** token-bucket + breaker on all restore actions (guard respawn storm / shutdown hang).
- **Allowed:** `internal/restore`.
- **Acceptance:** exceeding threshold → `STORM_DETECTED` + slow mode, no CPU spin.
- **Verification:** unit test rate math; chaos kill-loop shows breaker trips.
- **Deps:** T12, T13. **Risk:** M. **Approval:** none. **Status:** TODO.

### T15 — Twin mutual-restore (production)
- **Objective:** promote the M1 stub to real: each twin restores the other's binary + plist from `schg` baseline and re-bootstraps, throttled.
- **Allowed:** `internal/restore`, `internal/monitor`, `cmd/guardiand`.
- **Acceptance:** kill/`rm` one twin (after `noschg`) → peer restores it.
- **Verification:** re-run T05 burst; compare latency to the M1 baseline result.
- **Deps:** T07, T11, T14. **Risk:** H. **Approval:** none. **Status:** TODO.

---

## M4 — Lock, crypto & IPC

### T16 — Ed25519 crypto core
- **Objective:** keypair handling (pinned partner pubkey), challenge = `base32(nonce‖deviceID‖lock_epoch)`, verify token, single-use time-boxed nonce store.
- **Allowed:** `internal/crypto`.
- **Acceptance:** valid token verifies; stale/replayed/wrong-epoch tokens rejected.
- **Verification:** unit tests incl. replay + epoch-mismatch. **[Fact: stdlib crypto/ed25519]**
- **Deps:** T02. **Risk:** M. **Approval:** none. **Status:** TODO. **Parallel-safe with M3 tasks.**

### T17 — Signed lock-state record
- **Objective:** append-only signed lock record as source of truth; `state` in `policy.json` re-derived on read; monotonic `lock_epoch`.
- **Allowed:** `internal/lockstate`, `internal/policy`.
- **Acceptance:** editing `policy.json` to `unlocked` is inert; authoritative state comes from the record.
- **Verification:** integration edit JSON → daemon reverts + logs `TAMPER_POLICY`.
- **Deps:** T16. **Risk:** M. **Approval:** none. **Status:** TODO.

### T18 — Break-glass timer
- **Objective:** `break-glass` starts an enforced delay (default 48 h, min 12 h) using monotonic time + persisted `not_before`; auto-unlock at expiry; ignore wall-clock rollback.
- **Allowed:** `internal/lockstate`.
- **Acceptance:** clock rollback does not shorten the wait; completes on/after boot if elapsed.
- **Verification:** unit test with fake clock; integration reboot mid-window.
- **Deps:** T17. **Risk:** M. **Approval:** **required** on default value (plan §11.5). **Status:** TODO.

### T19 — UDS server + peer-cred auth
- **Objective:** root-owned Unix socket; authenticate CLI via `GetsockoptXucred` (UID/GID) + `LOCAL_PEERPID`; verify caller code-signing identity.
- **Allowed:** `internal/ipc`.
- **Acceptance:** only permitted callers connect; mutating ops still require a partner token.
- **Verification:** integration from an unauthorized process → rejected. **[Fact: x/sys/unix peercred]**
- **Deps:** T02. **Risk:** M. **Approval:** none. **Status:** TODO. **Parallel-safe with T16/T17.**

### T20 — CLI commands
- **Objective:** `status`, `lock`, `request-unlock`, `unlock`, `break-glass`, `logs`, `doctor`, `uninstall` over UDS.
- **Allowed:** `cmd/guardian`, `internal/ipc`.
- **Acceptance:** each command works; `request-unlock` prints a challenge; `unlock --token` transitions state.
- **Verification:** end-to-end offline: challenge → partner-tool sign → `unlock`.
- **Deps:** T17, T18, T19. **Risk:** M. **Approval:** none. **Status:** TODO.

### T21 — Partner signer tool
- **Objective:** tiny standalone binary/script: input challenge, output Ed25519 token; trivial for a non-technical partner.
- **Allowed:** `partner-tool/`.
- **Acceptance:** produces tokens the daemon accepts; runs on macOS/Linux/Windows.
- **Verification:** cross-check token verifies in T16 tests.
- **Deps:** T16. **Risk:** L. **Approval:** none. **Status:** TODO. **Parallel-safe with T20.**

---

## M5 — Packaging

### T22 — LaunchDaemon plists + install scripts
- **Objective:** finalize both plists (`RunAtLoad`, `KeepAlive`, `ThrottleInterval`, role args) and `preinstall`/`postinstall` (discovery, `schg`, bootstrap, `doctor` verify).
- **Allowed:** `packaging/*.plist`, `packaging/scripts/*`.
- **Acceptance:** clean install loads both twins at boot; `postinstall` sets flags + bootstraps.
- **Verification:** fresh-VM install → reboot → both jobs `print` as running. **[Fact: launchd keys]**
- **Deps:** T15, T20. **Risk:** M (boot state; rollback via uninstall). **Approval:** none. **Status:** TODO.

### T23 — Codesign + notarize + staple pipeline
- **Objective:** sign binaries (`--options runtime --timestamp`, Developer ID Application); build pkg (`pkgbuild`→`productbuild --sign "Developer ID Installer"`); `notarytool submit --wait`; `stapler staple` the pkg.
- **Allowed:** `packaging/Makefile`, `packaging/notarize.sh`.
- **Acceptance:** `spctl`/`pkgutil --check-signature` pass; stapled pkg installs without Gatekeeper prompt.
- **Verification:** `xcrun notarytool submit … --wait` accepted; `xcrun stapler staple` OK; install on a fresh VM. **[Fact: notarization workflow]**
- **Deps:** T22. **Risk:** M. **Approval:** **required** — needs Apple Developer enrollment (plan §11.4). **Status:** TODO.

### T24 — Partner-gated uninstall + locked-update refusal
- **Objective:** `guardian uninstall` requires unlocked/break-glass then clears `schg`, boots out twins, removes files/baseline, logs; installer refuses to replace core components while locked without a token.
- **Allowed:** `cmd/guardian`, `packaging/scripts/*`.
- **Acceptance:** uninstall blocked while locked; succeeds after unlock; locked pkg upgrade refused.
- **Verification:** integration both paths.
- **Deps:** T20, T22. **Risk:** M. **Approval:** none. **Status:** TODO.

---

## M6 — Hardening, fail-safe & audit

### T25 — Policy load/validate + SAFE-PERMISSIVE
- **Objective:** robust `policy.json` parsing; on unreadable/invalid policy enter SAFE-PERMISSIVE (keep running, stop destructive restores, log loudly).
- **Allowed:** `internal/policy`.
- **Acceptance:** corrupt policy → machine stays usable, boot/login unaffected.
- **Verification:** integration inject bad policy → SAFE-PERMISSIVE, no brick.
- **Deps:** T17. **Risk:** M. **Approval:** none. **Status:** TODO.

### T26 — Audit logger
- **Objective:** JSON-lines audit log (event taxonomy spec §20), root-only, `schg` on dir; `os_log` mirror behind an optional cgo build tag (default off).
- **Allowed:** `internal/audit`.
- **Acceptance:** all tamper/lock/restore/error events recorded with actor where knowable.
- **Verification:** run red-team → inspect log entries.
- **Deps:** T14. **Risk:** L. **Approval:** **required** to enable the `os_log` cgo build (plan §11.2). **Status:** TODO.

### T27 — `guardian doctor`
- **Objective:** self-diagnosis: resolved paths, jobs loaded, `schg` status, baseline validity, lock/break-glass state, FD headroom.
- **Allowed:** `cmd/guardian`.
- **Acceptance:** accurate one-shot health report.
- **Verification:** compare against manual `launchctl print` / `ls -lO` / `sfltool dumpbtm`.
- **Deps:** T20, T25. **Risk:** L. **Approval:** none. **Status:** TODO.

### T28 — Chaos & stability suite
- **Objective:** respawn-storm, reboot mid-lock, sleep/wake, clock rollback, shutdown-hang check.
- **Allowed:** `test/`.
- **Acceptance:** breaker prevents CPU spin; no shutdown hang; lock survives reboot.
- **Verification:** scripted; observe logs + `shutdown` timing.
- **Deps:** T14, T18, T22. **Risk:** M. **Approval:** none. **Status:** TODO.

### T29 — Compatibility matrix
- **Objective:** run acceptance on macOS 14, 15, 26 (Apple silicon + one Intel); test Cold Turkey running vs stopped during restore; App Management/FDA behavior for bundle restore.
- **Allowed:** `test/`.
- **Acceptance:** documented pass/fail per OS; App Management workaround noted if needed.
- **Verification:** matrix run. **[Assumption: App Management — record result]**
- **Deps:** T13, T22. **Risk:** M. **Approval:** none. **Status:** TODO.

---

## M7 — Docs & release

### T30 — Red-team acceptance report
- **Objective:** final run of the four commands + escalations across the matrix; publish latency + survive/defeat table = the product's honest efficacy claim.
- **Allowed:** `docs/`, `test/redteam/`.
- **Acceptance:** report matches spec §24 expectations (each command self-reverts; bypass requires deliberate multi-step teardown or Recovery).
- **Verification:** attach logs.
- **Deps:** T15, T24, T28. **Risk:** M. **Approval:** none. **Status:** TODO.

### T31 — Owner runbook
- **Objective:** install, lock, request-unlock, break-glass, uninstall, and Recovery-mode manual removal steps.
- **Allowed:** `docs/`.
- **Acceptance:** a second person can operate Guardian from the doc alone.
- **Deps:** T27. **Risk:** L. **Approval:** none. **Status:** TODO. **Parallel-safe with T32, T33.**

### T32 — Partner guide
- **Objective:** how the accountability partner receives a challenge and returns a token; key-custody guidance.
- **Allowed:** `docs/`, `partner-tool/README`.
- **Acceptance:** non-technical partner can complete an unlock.
- **Deps:** T21. **Risk:** L. **Approval:** none. **Status:** TODO.

### T33 — Recovery/lockout guide
- **Objective:** exact break-glass timing + Recovery-mode `chflags -h noschg` + delete steps; reassure no permanent lockout.
- **Allowed:** `docs/`.
- **Acceptance:** verified by performing a Recovery-mode removal on the VM.
- **Deps:** T24. **Risk:** L. **Approval:** none. **Status:** TODO.

### T34 — Release checklist
- **Objective:** version, notarized pkg artifact, checksums, changelog, tag.
- **Allowed:** `docs/`, `packaging/`.
- **Acceptance:** reproducible signed build from a clean checkout.
- **Deps:** T23, T30, T31, T32, T33. **Risk:** L. **Approval:** none. **Status:** TODO.

---

## Dependency order (critical path)

`T01 → T02 → T03 → T04 → T05★ → T06 → {T07,T08,T09} → T10 → T11⚑ → {T12,T13} → T14 → T15 → T22 → T23⚑ → T24 → T30 → T34`
(with the M4 lock chain `T16 → T17 → T18⚑ → T19 → T20 → T21` merging into T22)

★ = go/no-go gate · ⚑ = approval-required

## Parallel-safe groups (disjoint write sets)

- **G1 (monitoring):** T07, T08, T09 — all in `internal/monitor` but separable by file; treat as sequential if edited in one file, parallel if split per-source.
- **G2 (independent cores):** the M4 crypto/IPC chain (T16, T19, T21) can proceed in parallel with the M3 restore chain (T10–T15).
- **G3 (docs):** T31, T32, T33 in parallel once their deps land.

## Verification matrix (acceptance highlights)

| Capability | Task(s) | How verified |
|---|---|---|
| Detect all 4 tamper types | T07–T09 | correct event per command |
| Auto-revert all 4 commands | T12, T13, T15 | red-team single-command self-heal |
| Twin survives burst teardown | T05, T15 | latency/survive table |
| Partner unlock works offline | T16, T20, T21 | challenge→token→unlock |
| Edited JSON is inert | T17 | `TAMPER_POLICY` + revert |
| No permanent lockout | T18, T33 | break-glass + Recovery removal |
| Notarized install/boot/uninstall | T22–T24 | fresh-VM lifecycle |
| No shutdown hang / storm | T14, T28 | chaos suite |
| macOS 14/15/26 compatible | T29 | matrix run |

## Blocked / approval-required work

- **T05 gate** — review results before M2+ (plan §11.1).
- **T11** — `schg` on Cold Turkey files (updater conflict) (plan §11.3).
- **T18** — break-glass default value (plan §11.5).
- **T23** — requires Apple Developer enrollment (plan §11.4); otherwise stops at unsigned dev pkg from T03.
- **T26** — enabling `os_log` cgo (plan §11.2).

## First safe task to start now

**T01 — Repo & module skeleton** (no deps, no approvals, risk L).
