// Command guardiand is the Cold Turkey Guardian enforcement daemon.
// It runs as one of two mutually-restoring launchd twins, selected by --role.
//
// T04 scope (docs/specs/001-twin-restore-feasibility/plan.md): each twin polls
// for its peer's LaunchDaemon plist and, if missing, restores it from a
// read-only reference copy and re-bootstraps the job. Process liveness after
// a plain kill/crash is launchd's own KeepAlive job — this daemon only covers
// the case launchd cannot: a bootout'd/deleted job definition. Baseline
// manifests, schg, throttling, and Cold Turkey-specific monitoring are later
// milestones (T08, T10, T11, T14) and are intentionally absent here so the
// M1 red-team measurement reflects the twin pattern's raw ceiling.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// role identifies which twin this process is acting as.
type role string

const (
	rolePrimary role = "primary"
	roleTwin    role = "twin"

	pollInterval = 5 * time.Second
)

// launchDaemonsDir and refDir are vars, not consts, so tests can redirect
// them into a temp directory instead of touching the real /Library paths.
var (
	launchDaemonsDir = "/Library/LaunchDaemons"
	refDir           = "/Library/Application Support/ColdTurkeyGuardian/ref"
)

func peerOf(r role) role {
	if r == rolePrimary {
		return roleTwin
	}
	return rolePrimary
}

func plistName(r role) string {
	return "com.coldturkeyguardian." + string(r) + ".plist"
}

func main() {
	roleFlag := flag.String("role", "", `guardiand role: "primary" or "twin"`)
	flag.Parse()

	r := role(*roleFlag)
	if r != rolePrimary && r != roleTwin {
		log.Fatalf("guardiand: --role must be %q or %q, got %q", rolePrimary, roleTwin, *roleFlag)
	}
	peer := peerOf(r)

	log.Printf("guardiand: starting role=%s peer=%s", r, peer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	watchPeer(ctx, peer)
}

// watchPeer polls for the peer twin's plist and restores it on absence.
func watchPeer(ctx context.Context, peer role) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	checkAndRestorePeer(peer)

	for {
		select {
		case <-ctx.Done():
			log.Printf("guardiand: shutting down")
			return
		case <-ticker.C:
			checkAndRestorePeer(peer)
		}
	}
}

func checkAndRestorePeer(peer role) {
	name := plistName(peer)
	live := filepath.Join(launchDaemonsDir, name)
	ref := filepath.Join(refDir, name)

	if _, err := os.Stat(live); err == nil {
		return
	} else if !os.IsNotExist(err) {
		log.Printf("guardiand: stat %s: %v", live, err)
		return
	}

	log.Printf("guardiand: peer plist missing: %s — restoring from %s", live, ref)

	if err := copyFile(ref, live); err != nil {
		log.Printf("guardiand: restore copy %s -> %s failed: %v", ref, live, err)
		return
	}

	if err := exec.Command("launchctl", "bootstrap", "system", live).Run(); err != nil {
		log.Printf("guardiand: bootstrap %s failed: %v", live, err)
		return
	}

	log.Printf("guardiand: restored + bootstrapped %s", live)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
