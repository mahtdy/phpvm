package version_test

import (
	"fmt"
	"testing"

	"github.com/mahtdy/phpvm/internal/validation"
	"github.com/mahtdy/phpvm/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeWindowsPage is a minimal excerpt of windows.php.net listing HTML.
const fakeWindowsPage = `
<a href="php-8.3.10-nts-Win32-vs16-x64.zip">php-8.3.10-nts-Win32-vs16-x64.zip</a>
<a href="php-8.3.10-ts-Win32-vs16-x64.zip">php-8.3.10-ts-Win32-vs16-x64.zip</a>
<a href="php-8.3.9-nts-Win32-vs16-x64.zip">php-8.3.9-nts-Win32-vs16-x64.zip</a>
<a href="php-8.4.1-nts-Win32-vs16-x64.zip">php-8.4.1-nts-Win32-vs16-x64.zip</a>
`

// fakeUnixJSON is a minimal php.net releases API response.
const fakeUnixJSON = `{
  "8.3.10": {"version":"8.3.10","source":{"bz2":{"path":"php-8.3.10.tar.bz2","sha256":"abc123"}}},
  "8.3.9":  {"version":"8.3.9", "source":{"bz2":{"path":"php-8.3.9.tar.bz2","sha256":"def456"}}}
}`

func newTestResolver(t *testing.T, responseBody string) *version.Resolver {
	t.Helper()
	r := version.NewResolver(t.TempDir())
	r.SetFetchURL(func(url string) ([]byte, error) {
		return []byte(responseBody), nil
	})
	return r
}

func TestResolveWindows_ParsesReleases(t *testing.T) {
	r := newTestResolver(t, fakeWindowsPage)

	sel := validation.MustParse("8.3")
	releases, err := r.ResolveWindows(sel)
	require.NoError(t, err)

	// Should find 8.3.10 (nts+ts) and 8.3.9 (nts) = 3 total
	assert.GreaterOrEqual(t, len(releases), 2)

	// First release should be newest (8.3.10)
	assert.Equal(t, "8.3.10", releases[0].Version.String())
}

func TestResolveWindows_SkipsOtherMinor(t *testing.T) {
	r := newTestResolver(t, fakeWindowsPage)

	sel := validation.MustParse("8.3")
	releases, err := r.ResolveWindows(sel)
	require.NoError(t, err)

	for _, rel := range releases {
		assert.Equal(t, 8, rel.Version.Major)
		assert.Equal(t, 3, rel.Version.Minor)
	}
}

func TestResolveUnix_ParsesReleases(t *testing.T) {
	r := newTestResolver(t, fakeUnixJSON)

	sel := validation.MustParse("8.3")
	releases, err := r.ResolveUnix(sel)
	require.NoError(t, err)

	require.Len(t, releases, 2)
	assert.Equal(t, "8.3.10", releases[0].Version.String())
	assert.Equal(t, "8.3.9", releases[1].Version.String())
}

func TestLatestPatch_ReturnsHighest(t *testing.T) {
	r := newTestResolver(t, fakeUnixJSON)
	r.SetPlatform("linux")

	sel := validation.MustParse("8.3")
	rel, err := r.LatestPatch(sel)
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", rel.Version.String())
}

func TestCacheIsUsed(t *testing.T) {
	callCount := 0
	r := version.NewResolver(t.TempDir())
	r.SetFetchURL(func(url string) ([]byte, error) {
		callCount++
		return []byte(fakeUnixJSON), nil
	})
	r.SetPlatform("linux")

	sel := validation.MustParse("8.3")
	_, err := r.LatestPatch(sel)
	require.NoError(t, err)

	// Second call — cache should be hit, not HTTP.
	_, err = r.LatestPatch(sel)
	require.NoError(t, err)

	assert.Equal(t, 1, callCount, "expected cache to prevent second HTTP call")
}

func TestReleaseInfo_URL(t *testing.T) {
	r := newTestResolver(t, fakeWindowsPage)

	sel := validation.MustParse("8.3")
	releases, err := r.ResolveWindows(sel)
	require.NoError(t, err)
	require.NotEmpty(t, releases)

	assert.Contains(t, releases[0].URL, fmt.Sprintf("%s", "php-8.3.10"))
	assert.True(t, releases[0].Filename != "")
}
