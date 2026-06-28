// Package version resolves available PHP releases from official sources
// and constructs platform-specific download URLs.
package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/validation"
)

const (
	// windowsReleasesURL is the Windows PHP release listing page.
	windowsReleasesURL = "https://windows.php.net/downloads/releases/"
	// unixReleasesURL returns JSON release info for a given PHP major.minor.
	unixReleasesURL = "https://www.php.net/releases/index.php?json&version=%d.%d"
	// cacheMaxAge is how long resolved release lists are cached locally.
	cacheMaxAge = 1 * time.Hour
	// httpTimeout for release listing requests.
	httpTimeout = 30 * time.Second
)

// ReleaseInfo holds metadata about a downloadable PHP release.
type ReleaseInfo struct {
	Version  *validation.Version
	URL      string
	Filename string
	SHA256   string
	OS       string // "windows" | "linux" | "darwin"
	Arch     string // "x64" | "arm64"
	Variant  string // "ts" | "nts"
}

// Resolver fetches and caches PHP release information.
type Resolver struct {
	cacheDir        string
	httpClient      *http.Client
	platform        string
	downloadBaseURL string // overrides generated download URLs (for tests)
	fetchURL        func(url string) ([]byte, error)
}

// NewResolver creates a Resolver that caches data in cacheDir.
func NewResolver(cacheDir string) *Resolver {
	r := &Resolver{
		cacheDir: cacheDir,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
	r.fetchURL = r.defaultFetch
	return r
}

// Resolve returns all available releases for the given PHP major.minor version
// on the current platform.
func (r *Resolver) Resolve(v *validation.Version) ([]*ReleaseInfo, error) {
	platform := r.platform
	if platform == "" {
		platform = runtime.GOOS
	}
	switch platform {
	case "windows":
		return r.ResolveWindows(v)
	default:
		return r.ResolveUnix(v)
	}
}

// ResolveWindows returns Windows-specific releases (exported for tests).
func (r *Resolver) ResolveWindows(v *validation.Version) ([]*ReleaseInfo, error) {
	return r.resolveWindows(v)
}

// ResolveUnix returns Unix-specific releases (exported for tests).
func (r *Resolver) ResolveUnix(v *validation.Version) ([]*ReleaseInfo, error) {
	return r.resolveUnix(v)
}

// SetPlatform overrides the detected OS (for testing).
func (r *Resolver) SetPlatform(os string) {
	r.platform = os
}

// LatestPatch returns the highest available patch version matching v.
// v may be "8.3" (minor selector) or "8.3.10" (exact).
func (r *Resolver) LatestPatch(v *validation.Version) (*ReleaseInfo, error) {
	releases, err := r.Resolve(v)
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, cli.NewError(
			fmt.Sprintf("no releases found for PHP %s", v.String()), nil,
		)
	}
	// Releases are sorted desc; first entry is the latest.
	return releases[0], nil
}

// ----- Windows resolver -----

// winZipRe matches filenames like: php-8.3.10-nts-Win32-vs16-x64.zip
var winZipRe = regexp.MustCompile(
	`(?i)php-(\d+\.\d+\.\d+)-(nts|ts)-Win32-[a-z0-9]+-([a-z0-9]+)\.zip`,
)

func (r *Resolver) resolveWindows(v *validation.Version) ([]*ReleaseInfo, error) {
	cacheKey := fmt.Sprintf("windows_%d_%d.json", v.Major, v.Minor)
	body, err := r.cachedFetch(windowsReleasesURL, cacheKey)
	if err != nil {
		return nil, err
	}

	var releases []*ReleaseInfo
	page := string(body)

	for _, match := range winZipRe.FindAllStringSubmatch(page, -1) {
		full, verStr, variant, arch := match[0], match[1], strings.ToLower(match[2]), strings.ToLower(match[3])

		rv, err := validation.Parse(verStr)
		if err != nil {
			continue
		}
		// Filter: only the requested major.minor.
		if rv.Major != v.Major || rv.Minor != v.Minor {
			continue
		}

		releaseURL := windowsReleasesURL + full
		if r.downloadBaseURL != "" {
			releaseURL = r.downloadBaseURL
		}
		releases = append(releases, &ReleaseInfo{
			Version:  rv,
			URL:      releaseURL,
			Filename: full,
			OS:       "windows",
			Arch:     normaliseArch(arch),
			Variant:  variant,
		})
	}

	sortReleasesDesc(releases)
	return releases, nil
}

// ----- Unix resolver -----

// phpNetReleaseResponse is the JSON schema from php.net/releases API.
type phpNetReleaseResponse struct {
	Version  string `json:"version"`
	Filename struct {
		Tar struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"bz2"` // .tar.bz2 — php.net also has gz but bz2 is canonical
	} `json:"source"`
}

func (r *Resolver) resolveUnix(v *validation.Version) ([]*ReleaseInfo, error) {
	url := fmt.Sprintf(unixReleasesURL, v.Major, v.Minor)
	cacheKey := fmt.Sprintf("unix_%d_%d.json", v.Major, v.Minor)
	body, err := r.cachedFetch(url, cacheKey)
	if err != nil {
		return nil, err
	}

	// php.net returns a map of "version" -> {...}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, cli.NewError("failed to parse php.net release JSON", err)
	}

	var releases []*ReleaseInfo
	for _, vRaw := range raw {
		var rel phpNetReleaseResponse
		if err := json.Unmarshal(vRaw, &rel); err != nil {
			continue
		}
		rv, err := validation.Parse(rel.Version)
		if err != nil {
			continue
		}
		if rv.Major != v.Major || rv.Minor != v.Minor {
			continue
		}

		filename := fmt.Sprintf("php-%s.tar.gz", rv.String())
		downloadURL := fmt.Sprintf(
			"https://www.php.net/distributions/%s", filename,
		)

		releases = append(releases, &ReleaseInfo{
			Version:  rv,
			URL:      downloadURL,
			Filename: filename,
			SHA256:   rel.Filename.Tar.SHA256,
			OS:       runtime.GOOS,
			Arch:     defaultArch(),
			Variant:  "nts",
		})
	}

	sortReleasesDesc(releases)
	return releases, nil
}

// ----- Cache helpers -----

func (r *Resolver) cachedFetch(url, cacheKey string) ([]byte, error) {
	if err := os.MkdirAll(r.cacheDir, 0o750); err != nil {
		return nil, cli.NewError("cannot create cache dir", err)
	}

	cachePath := filepath.Join(r.cacheDir, cacheKey)
	if data, ok := r.readCache(cachePath); ok {
		return data, nil
	}

	data, err := r.fetchURL(url)
	if err != nil {
		return nil, err
	}

	_ = os.WriteFile(cachePath, data, 0o640)
	return data, nil
}

func (r *Resolver) readCache(path string) ([]byte, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > cacheMaxAge {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func (r *Resolver) defaultFetch(url string) ([]byte, error) {
	resp, err := r.httpClient.Get(url) //nolint:gosec // URL is constructed internally
	if err != nil {
		return nil, cli.NewNetworkError(
			fmt.Sprintf("failed to fetch %s", url), err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, cli.NewNetworkError(
			fmt.Sprintf("HTTP %d fetching %s", resp.StatusCode, url), nil,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, cli.NewNetworkError("failed to read response body", err)
	}
	return body, nil
}

// SetFetchURL replaces the HTTP fetcher (for testing).
func (r *Resolver) SetFetchURL(fn func(url string) ([]byte, error)) {
	r.fetchURL = fn
}

// SetDownloadBaseURL overrides the URL used for the actual file download.
// When set, any ReleaseInfo.URL built by the resolver is replaced with this value.
func (r *Resolver) SetDownloadBaseURL(url string) {
	r.downloadBaseURL = url
}

// ----- Helpers -----

func sortReleasesDesc(releases []*ReleaseInfo) {
	sort.Slice(releases, func(i, j int) bool {
		return releases[j].Version.Less(releases[i].Version)
	})
}

func normaliseArch(arch string) string {
	switch strings.ToLower(arch) {
	case "x64", "amd64":
		return "x64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return arch
	}
}

func defaultArch() string {
	if runtime.GOARCH == "arm64" {
		return "arm64"
	}
	return "x64"
}
