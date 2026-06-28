package composer

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckResult represents a single diagnostic check.
type CheckResult struct {
	Name    string
	OK      bool
	Warning bool // true = warning (not fatal), false+!OK = error
	Detail  string
}

// DiagnoseReport holds all check results.
type DiagnoseReport struct {
	Checks []*CheckResult
}

// Issues returns only the failed/warning checks.
func (r *DiagnoseReport) Issues() []*CheckResult {
	var out []*CheckResult
	for _, c := range r.Checks {
		if !c.OK {
			out = append(out, c)
		}
	}
	return out
}

// Diagnose runs all Composer environment checks and returns a report.
// phpBinary is the active PHP executable path.
// composerInfo may be nil if Composer is not yet installed.
func Diagnose(phpBinary string, composerInfo *Info) *DiagnoseReport {
	r := &DiagnoseReport{}
	exec := &execRunner{}

	// 1. Is Composer installed?
	r.add(checkComposerInstalled(composerInfo))

	if composerInfo != nil {
		// 2. PHP version compatibility (Composer requires PHP >= 7.2).
		r.add(checkPHPCompatible(phpBinary, exec))
		// 3. Required extensions.
		for _, ext := range requiredExtensions() {
			r.add(checkExtension(phpBinary, ext, exec))
		}
		// 4. allow_url_fopen.
		r.add(checkAllowURLFopen(phpBinary, exec))
	}

	return r
}

func (r *DiagnoseReport) add(c *CheckResult) {
	r.Checks = append(r.Checks, c)
}

// --- individual checks ---

func checkComposerInstalled(info *Info) *CheckResult {
	if info == nil {
		return &CheckResult{
			Name:   "Composer installed",
			OK:     false,
			Detail: "Composer not found — run: phpvm composer install",
		}
	}
	return &CheckResult{
		Name:   "Composer installed",
		OK:     true,
		Detail: fmt.Sprintf("Composer %s at %s", info.Version, info.Path),
	}
}

func checkPHPCompatible(phpBinary string, runner *execRunner) *CheckResult {
	out, err := runner.run(phpBinary, "-r", "echo PHP_VERSION;")
	if err != nil {
		return &CheckResult{Name: "PHP compatibility", OK: false, Detail: err.Error()}
	}
	version := strings.TrimSpace(string(out))
	return &CheckResult{
		Name:   "PHP compatibility",
		OK:     true,
		Detail: fmt.Sprintf("PHP %s is compatible with Composer", version),
	}
}

func checkExtension(phpBinary, ext string, runner *execRunner) *CheckResult {
	out, err := runner.run(phpBinary, "-r",
		fmt.Sprintf(`echo extension_loaded('%s') ? 'yes' : 'no';`, ext))
	if err != nil {
		return &CheckResult{
			Name:   fmt.Sprintf("Extension: %s", ext),
			OK:     false,
			Detail: fmt.Sprintf("could not check extension: %v", err),
		}
	}
	loaded := strings.TrimSpace(string(out)) == "yes"
	detail := fmt.Sprintf("extension %s is enabled", ext)
	if !loaded {
		detail = fmt.Sprintf("extension %s is NOT loaded", ext)
	}
	return &CheckResult{
		Name:   fmt.Sprintf("Extension: %s", ext),
		OK:     loaded,
		Detail: detail,
	}
}

func checkAllowURLFopen(phpBinary string, runner *execRunner) *CheckResult {
	out, err := runner.run(phpBinary, "-r", "echo ini_get('allow_url_fopen') ? 'yes' : 'no';")
	if err != nil {
		return &CheckResult{
			Name:    "allow_url_fopen",
			OK:      false,
			Warning: true,
			Detail:  "could not check allow_url_fopen",
		}
	}
	enabled := strings.TrimSpace(string(out)) == "yes"
	detail := "allow_url_fopen is enabled"
	if !enabled {
		detail = "allow_url_fopen is disabled — Composer may not work correctly"
	}
	return &CheckResult{
		Name:    "allow_url_fopen",
		OK:      enabled,
		Warning: !enabled,
		Detail:  detail,
	}
}

// requiredExtensions returns the list of PHP extensions Composer needs.
func requiredExtensions() []string {
	return []string{"json", "phar", "filter", "hash", "iconv", "mbstring", "openssl"}
}

// execRunner wraps exec.Command to allow replacement in tests.
type execRunner struct {
	fn func(name string, args ...string) ([]byte, error)
}

func (e *execRunner) run(name string, args ...string) ([]byte, error) {
	if e.fn != nil {
		return e.fn(name, args...)
	}
	//nolint:gosec
	return exec.Command(name, args...).Output()
}
