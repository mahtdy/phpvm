package validation_test

import (
	"testing"

	"github.com/mahtdy/phpvm/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Parse tests ---

func TestParseValidFormats(t *testing.T) {
	cases := []struct {
		input         string
		major, minor, patch int
		variant       string
	}{
		{"8", 8, 0, 0, ""},
		{"8.3", 8, 3, 0, ""},
		{"8.3.10", 8, 3, 10, ""},
		{"8.3.10-nts", 8, 3, 10, "nts"},
		{"8.3.10-ts", 8, 3, 10, "ts"},
		{"8.4.1", 8, 4, 1, ""},
		{"  8.4.1  ", 8, 4, 1, ""},     // whitespace trimmed
		{"8.3.10-NTS", 8, 3, 10, "nts"}, // case-insensitive
	}
	for _, c := range cases {
		v, err := validation.Parse(c.input)
		require.NoError(t, err, "input=%q", c.input)
		assert.Equal(t, c.major, v.Major, "major mismatch for %q", c.input)
		assert.Equal(t, c.minor, v.Minor, "minor mismatch for %q", c.input)
		assert.Equal(t, c.patch, v.Patch, "patch mismatch for %q", c.input)
		assert.Equal(t, c.variant, v.Variant, "variant mismatch for %q", c.input)
	}
}

func TestParseInvalidFormats(t *testing.T) {
	cases := []string{
		"",
		"abc",
		"8.3.10.1",   // 4 components
		"7.4",        // below minimum supported
		"v8.3",       // leading 'v'
		"8.3.10-foo", // unknown variant
		"latest",
	}
	for _, input := range cases {
		_, err := validation.Parse(input)
		assert.Error(t, err, "expected error for input %q", input)
	}
}

// --- String / Full tests ---

func TestVersionString(t *testing.T) {
	assert.Equal(t, "8", validation.MustParse("8").String())
	assert.Equal(t, "8.3", validation.MustParse("8.3").String())
	assert.Equal(t, "8.3.10", validation.MustParse("8.3.10").String())
	assert.Equal(t, "8.3.10", validation.MustParse("8.3.10-nts").String())
}

func TestVersionFull(t *testing.T) {
	assert.Equal(t, "8.3.10-nts", validation.MustParse("8.3.10-nts").Full())
	assert.Equal(t, "8.3.10-ts", validation.MustParse("8.3.10-ts").Full())
	assert.Equal(t, "8.3.10", validation.MustParse("8.3.10").Full())
}

// --- Compare tests ---

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b     string
		expected int
	}{
		{"8.3", "8.4", -1},
		{"8.4", "8.3", 1},
		{"8.3", "8.3", 0},
		{"8.3.10", "8.3.9", 1},
		{"8.3.0", "8.3.10", -1},
		{"8.4.1", "8.4.1", 0},
	}
	for _, c := range cases {
		result := validation.MustParse(c.a).Compare(validation.MustParse(c.b))
		assert.Equal(t, c.expected, result, "%s vs %s", c.a, c.b)
	}
}

func TestLess(t *testing.T) {
	assert.True(t, validation.MustParse("8.3").Less(validation.MustParse("8.4")))
	assert.False(t, validation.MustParse("8.4").Less(validation.MustParse("8.3")))
	assert.False(t, validation.MustParse("8.3").Less(validation.MustParse("8.3")))
}

// --- Matches tests ---

func TestMatches(t *testing.T) {
	installed := validation.MustParse("8.3.10")
	assert.True(t, installed.Matches(validation.MustParse("8")))
	assert.True(t, installed.Matches(validation.MustParse("8.3")))
	assert.True(t, installed.Matches(validation.MustParse("8.3.10")))
	assert.False(t, installed.Matches(validation.MustParse("8.4")))
	assert.False(t, installed.Matches(validation.MustParse("8.3.9")))
}

// --- Variant tests ---

func TestVariants(t *testing.T) {
	nts := validation.MustParse("8.3.10-nts")
	assert.True(t, nts.IsNTS())
	assert.False(t, nts.IsTS())

	ts := validation.MustParse("8.3.10-ts")
	assert.True(t, ts.IsTS())
	assert.False(t, ts.IsNTS())

	plain := validation.MustParse("8.3.10")
	assert.True(t, plain.IsNTS()) // no variant defaults to NTS
}

// --- SortVersions tests ---

func TestSortVersions(t *testing.T) {
	vs := []*validation.Version{
		validation.MustParse("8.4.1"),
		validation.MustParse("8.2.20"),
		validation.MustParse("8.3.10"),
		validation.MustParse("8.5.0"),
	}
	validation.SortVersions(vs)
	expected := []string{"8.2.20", "8.3.10", "8.4.1", "8.5.0"}
	for i, v := range vs {
		assert.Equal(t, expected[i], v.String())
	}
}
