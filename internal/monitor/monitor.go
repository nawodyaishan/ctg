// Package monitor watches Cold Turkey and Guardian twin processes, files,
// and launchd job state for tampering.
package monitor

import "github.com/fsnotify/fsnotify"

// NewFileWatcher returns a kqueue-backed watcher for parent-directory watches
// on Cold Turkey and Guardian critical files (tech spec §15).
func NewFileWatcher() (*fsnotify.Watcher, error) {
	return fsnotify.NewWatcher()
}
