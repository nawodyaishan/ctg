// Package discovery locates Cold Turkey's launchd labels, plist paths,
// app bundle, and support directory at install/runtime.
package discovery

import (
	"os"

	"howett.net/plist"
)

// DecodePlist reads and decodes the property list at path into v.
func DecodePlist(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return plist.NewDecoder(f).Decode(v)
}
