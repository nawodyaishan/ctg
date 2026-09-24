// Package flags applies and clears filesystem immutability flags (schg)
// on Guardian and Cold Turkey files.
package flags

import "golang.org/x/sys/unix"

// SetImmutable applies the schg (SF_IMMUTABLE) flag to path.
// Only root can clear it again; this is friction, not a hard wall (tech spec §17, §25).
func SetImmutable(path string) error {
	return unix.Chflags(path, unix.SF_IMMUTABLE)
}

// ClearImmutable removes the schg (SF_IMMUTABLE) flag from path.
func ClearImmutable(path string) error {
	return unix.Chflags(path, 0)
}
