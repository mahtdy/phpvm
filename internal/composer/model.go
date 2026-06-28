// Package composer manages Composer detection, installation, and updates.
package composer

import "fmt"

// Info holds metadata about a detected Composer installation.
type Info struct {
	// Path is the absolute path to the composer binary (composer or composer.phar).
	Path string
	// Version is the detected Composer version string (e.g. "2.7.0").
	Version string
	// Global is true when the binary is from the system PATH, not per-PHP-version.
	Global bool
}

// String returns a human-readable summary.
func (i *Info) String() string {
	scope := "per-version"
	if i.Global {
		scope = "global"
	}
	return fmt.Sprintf("Composer %s (%s) at %s", i.Version, scope, i.Path)
}
