// Package version provides build-time version information for phpvm.
package version

// These variables are overridden at build time via -ldflags.
var (
	// Version is the semantic version string (e.g. "0.1.0").
	Version = "dev"
	// Commit is the git commit SHA at build time.
	Commit = "none"
	// Date is the build timestamp in RFC3339 format.
	Date = "unknown"
	// BuiltBy identifies the build system (e.g. "goreleaser").
	BuiltBy = "manual"
)

// Info bundles all version fields into a single value.
type Info struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

// Get returns the current build Info.
func Get() Info {
	return Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
		BuiltBy: BuiltBy,
	}
}

// String returns a human-readable version string.
func (i Info) String() string {
	return "phpvm v" + i.Version + " (commit: " + i.Commit + ", built: " + i.Date + ")"
}
