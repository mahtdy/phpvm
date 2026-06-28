// Package validation provides input validation for phpvm CLI commands.
package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mahtdy/phpvm/internal/cli"
)

// minSupportedMajor is the lowest PHP major version phpvm manages.
const minSupportedMajor = 8

// versionRe matches: 8, 8.3, 8.3.10, 8.3.10-nts, 8.3.10-ts
// Capture groups: (major)(minor)(patch)(variant)
var versionRe = regexp.MustCompile(
	`^(\d+)(?:\.(\d+)(?:\.(\d+))?)?(?:-(nts|ts))?$`,
)

// Version represents a parsed, validated PHP version.
type Version struct {
	Major   int
	Minor   int
	Patch   int
	Variant string // "ts", "nts", or ""
	Raw     string // original input string
}

// Parse parses and validates a PHP version string.
// Accepted formats: "8", "8.3", "8.3.10", "8.3.10-nts", "8.3.10-ts".
// Returns a validation error for any other format.
func Parse(input string) (*Version, error) {
	s := strings.TrimSpace(strings.ToLower(input))
	if s == "" {
		return nil, cli.NewValidationError("version string must not be empty", nil)
	}

	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return nil, cli.NewValidationError(
			fmt.Sprintf("invalid version format %q — expected formats: 8, 8.3, 8.3.10, 8.3.10-nts", input),
			nil,
		)
	}

	major, _ := strconv.Atoi(m[1])
	if major < minSupportedMajor {
		return nil, cli.NewValidationError(
			fmt.Sprintf("PHP %d is not supported — phpvm requires PHP %d.0 or later", major, minSupportedMajor),
			nil,
		)
	}

	v := &Version{
		Major:   major,
		Raw:     input,
		Variant: m[4],
	}
	if m[2] != "" {
		v.Minor, _ = strconv.Atoi(m[2])
	}
	if m[3] != "" {
		v.Patch, _ = strconv.Atoi(m[3])
	}

	return v, nil
}

// String returns the canonical dot-separated version string (without variant).
func (v *Version) String() string {
	switch {
	case v.Patch > 0 || (v.Minor >= 0 && v.Patch >= 0 && strings.Count(v.Raw, ".") >= 2):
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	case v.Minor > 0 || strings.Count(v.Raw, ".") >= 1:
		return fmt.Sprintf("%d.%d", v.Major, v.Minor)
	default:
		return fmt.Sprintf("%d", v.Major)
	}
}

// Full returns the full version string including variant suffix if present.
func (v *Version) Full() string {
	s := v.String()
	if v.Variant != "" {
		return s + "-" + v.Variant
	}
	return s
}

// Compare compares two Versions semantically.
// Returns -1 if v < other, 0 if equal, +1 if v > other.
// Variant is ignored for ordering.
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		return cmp(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return cmp(v.Minor, other.Minor)
	}
	return cmp(v.Patch, other.Patch)
}

// Less reports whether v is strictly less than other.
func (v *Version) Less(other *Version) bool {
	return v.Compare(other) < 0
}

// Matches reports whether v is a prefix-match for selector.
// e.g. "8.3" matches "8.3.10", "8.3.0"; "8" matches "8.1.0", "8.4.2".
func (v *Version) Matches(selector *Version) bool {
	if v.Major != selector.Major {
		return false
	}
	// selector has minor component — must also match
	rawDots := strings.Count(selector.Raw, ".")
	if rawDots >= 1 && v.Minor != selector.Minor {
		return false
	}
	// selector has patch component — must also match
	if rawDots >= 2 && v.Patch != selector.Patch {
		return false
	}
	return true
}

// IsNTS reports whether the variant is Non-Thread Safe.
func (v *Version) IsNTS() bool { return v.Variant == "nts" || v.Variant == "" }

// IsTS reports whether the variant is Thread Safe.
func (v *Version) IsTS() bool { return v.Variant == "ts" }

// MustParse panics if the version string is invalid. Use only in tests.
func MustParse(s string) *Version {
	v, err := Parse(s)
	if err != nil {
		panic(fmt.Sprintf("validation.MustParse(%q): %v", s, err))
	}
	return v
}

func cmp(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// SortVersions sorts a slice of *Version in ascending semver order (in-place).
func SortVersions(versions []*Version) {
	n := len(versions)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && versions[j].Less(versions[j-1]); j-- {
			versions[j], versions[j-1] = versions[j-1], versions[j]
		}
	}
}
