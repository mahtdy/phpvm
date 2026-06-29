package php

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mahtdy/phpvm/internal/composer"
	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/extension"
	phpini "github.com/mahtdy/phpvm/internal/ini"
)

// CheckStatus is the result of one doctor check.
type CheckStatus int

const (
	CheckOK      CheckStatus = iota // ✓
	CheckWarn                       // ⚠
	CheckFail                       // ✗
)

// DoctorCheck represents a single environment health check.
type DoctorCheck struct {
	Name   string
	Status CheckStatus
	Detail string
}

// DoctorReport aggregates all health checks.
type DoctorReport struct {
	Checks []*DoctorCheck
}

// Issues returns only checks that are Warn or Fail.
func (r *DoctorReport) Issues() []*DoctorCheck {
	var out []*DoctorCheck
	for _, c := range r.Checks {
		if c.Status != CheckOK {
			out = append(out, c)
		}
	}
	return out
}

// Doctor runs all environment checks and returns a report.
type Doctor struct {
	cfg         *config.Config
	execCommand func(string, ...string) ([]byte, error)
}

// NewDoctor creates a Doctor for cfg.
func NewDoctor(cfg *config.Config) *Doctor {
	return &Doctor{cfg: cfg, execCommand: defaultDockerExec}
}

// SetExecCommand replaces the executor (for testing).
func (d *Doctor) SetExecCommand(fn func(string, ...string) ([]byte, error)) {
	d.execCommand = fn
}

// Run executes all checks and returns the report.
func (d *Doctor) Run() *DoctorReport {
	r := &DoctorReport{}

	// 1. PHP binary check.
	phpBin, phpDir := d.checkPHPBinary(r)

	// 2. PATH check.
	d.checkPATH(r, phpDir)

	// 3. php.ini check.
	iniPath := d.checkPHPIni(r, phpBin)

	// 4. memory_limit recommendation.
	if iniPath != "" {
		d.checkMemoryLimit(r, iniPath)
	}

	// 5. Composer check.
	d.checkComposer(r, phpDir)

	// 6. Key extensions.
	if phpBin != "" {
		d.checkExtensions(r, phpBin)
	}

	// 7. Xdebug check (warning only if missing).
	if iniPath != "" {
		d.checkXdebug(r, iniPath)
	}

	return r
}

// --- individual checks ---

func (d *Doctor) checkPHPBinary(r *DoctorReport) (phpBin, phpDir string) {
	detector := NewCurrentDetector(d.cfg.VersionsDir(), d.cfg.Current)
	v, binPath, err := detector.Detect()
	if err != nil {
		r.add("PHP binary", CheckFail, "PHP not found — run: phpvm use <version>")
		return "", ""
	}
	r.add("PHP binary", CheckOK, fmt.Sprintf("PHP %s at %s", v.String(), binPath))
	return binPath, filepath.Dir(binPath)
}

func (d *Doctor) checkPATH(r *DoctorReport, phpDir string) {
	if phpDir == "" {
		return
	}
	pathEnv := os.Getenv("PATH")
	if strings.Contains(pathEnv, phpDir) {
		r.add("PATH", CheckOK, fmt.Sprintf("%s is in PATH", phpDir))
	} else {
		r.add("PATH", CheckWarn, fmt.Sprintf("%s is not in PATH — run: phpvm use <version>", phpDir))
	}
}

func (d *Doctor) checkPHPIni(r *DoctorReport, phpBin string) string {
	if phpBin == "" {
		return ""
	}
	locator := phpini.NewLocator()
	locator.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return d.execCommand(name, args...)
	})
	iniPath, err := locator.Find(phpBin)
	if err != nil {
		r.add("php.ini", CheckWarn, "php.ini not found")
		return ""
	}
	r.add("php.ini", CheckOK, fmt.Sprintf("php.ini found at %s", iniPath))
	return iniPath
}

func (d *Doctor) checkMemoryLimit(r *DoctorReport, iniPath string) {
	f, err := phpini.Load(iniPath)
	if err != nil {
		return
	}
	val, ok := f.Get("memory_limit")
	if !ok {
		val = "128M"
	}
	// Simple check: if it looks like a small value warn.
	status := CheckOK
	detail := fmt.Sprintf("memory_limit = %s", val)
	if val == "128M" || val == "64M" || val == "32M" {
		status = CheckWarn
		detail = fmt.Sprintf("memory_limit is %s (recommended: 256M+)", val)
	}
	r.add("memory_limit", status, detail)
}

func (d *Doctor) checkComposer(r *DoctorReport, phpDir string) {
	det := composer.NewDetector()
	info, _ := det.Detect(phpDir)
	if info == nil {
		r.add("Composer", CheckWarn, "Composer not installed — run: phpvm composer install")
		return
	}
	r.add("Composer", CheckOK, fmt.Sprintf("Composer %s at %s", info.Version, info.Path))
}

func (d *Doctor) checkExtensions(r *DoctorReport, phpBin string) {
	required := []string{"openssl", "mbstring", "json", "curl"}
	out, err := d.execCommand(phpBin, "-m")
	if err != nil {
		r.add("Extensions", CheckWarn, "could not list PHP extensions")
		return
	}
	loaded := make(map[string]bool)
	for _, line := range strings.Split(string(out), "\n") {
		loaded[strings.ToLower(strings.TrimSpace(line))] = true
	}
	for _, ext := range required {
		if loaded[ext] {
			r.add(fmt.Sprintf("Extension: %s", ext), CheckOK,
				fmt.Sprintf("%s is enabled", ext))
		} else {
			r.add(fmt.Sprintf("Extension: %s", ext), CheckFail,
				fmt.Sprintf("%s extension not found", ext))
		}
	}
}

func (d *Doctor) checkXdebug(r *DoctorReport, iniPath string) {
	xm, err := extension.NewXdebugManager(iniPath)
	if err != nil {
		return
	}
	if xm.IsInstalled() {
		mode := xm.CurrentMode()
		r.add("Xdebug", CheckOK, fmt.Sprintf("Xdebug installed (mode=%s)", mode))
	} else {
		r.add("Xdebug", CheckWarn, "Xdebug not installed (optional)")
	}
}

func (r *DoctorReport) add(name string, status CheckStatus, detail string) {
	r.Checks = append(r.Checks, &DoctorCheck{
		Name:   name,
		Status: status,
		Detail: detail,
	})
}

func defaultDockerExec(name string, args ...string) ([]byte, error) {
	//nolint:gosec
	return exec.Command(name, args...).Output()
}
