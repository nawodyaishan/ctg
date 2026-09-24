package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCheckAndRestorePeer covers the file-restore half of T04 in isolation
// (no root, no launchd): the peer plist is copied back from ref when
// missing. The launchctl bootstrap call is expected to fail in this
// environment and is not the concern of this unit test — the real
// bootstrap/KeepAlive-respawn behavior needs the disposable VM the tasks
// doc specifies (docs/original/COLD_TURKEY_GUARDIAN_TASKS.md T04).
func TestCheckAndRestorePeer(t *testing.T) {
	dir := t.TempDir()
	origDaemons, origRef := launchDaemonsDir, refDir
	launchDaemonsDir = filepath.Join(dir, "LaunchDaemons")
	refDir = filepath.Join(dir, "ref")
	t.Cleanup(func() {
		launchDaemonsDir = origDaemons
		refDir = origRef
	})

	if err := os.MkdirAll(launchDaemonsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		t.Fatal(err)
	}

	want := []byte("<plist>reference copy</plist>")
	refPath := filepath.Join(refDir, plistName(roleTwin))
	if err := os.WriteFile(refPath, want, 0o644); err != nil {
		t.Fatal(err)
	}

	livePath := filepath.Join(launchDaemonsDir, plistName(roleTwin))
	if _, err := os.Stat(livePath); !os.IsNotExist(err) {
		t.Fatalf("precondition: expected %s absent, got err=%v", livePath, err)
	}

	// Peer plist absent -> should be restored from ref (bootstrap will fail
	// harmlessly since there's no launchd/root here; restore still happens
	// first).
	checkAndRestorePeer(roleTwin)

	got, err := os.ReadFile(livePath)
	if err != nil {
		t.Fatalf("expected %s restored, read failed: %v", livePath, err)
	}
	if string(got) != string(want) {
		t.Fatalf("restored content = %q, want %q", got, want)
	}

	// Peer plist present -> checkAndRestorePeer must not touch it.
	if err := os.WriteFile(livePath, []byte("live, do not overwrite"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkAndRestorePeer(roleTwin)
	got, err = os.ReadFile(livePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "live, do not overwrite" {
		t.Fatalf("present peer plist was overwritten: got %q", got)
	}
}

func TestPeerOf(t *testing.T) {
	if peerOf(rolePrimary) != roleTwin {
		t.Fatalf("peerOf(primary) = %q, want twin", peerOf(rolePrimary))
	}
	if peerOf(roleTwin) != rolePrimary {
		t.Fatalf("peerOf(twin) = %q, want primary", peerOf(roleTwin))
	}
}
