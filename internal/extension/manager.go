// Package extension manages PHP extensions: listing, enabling, disabling.
package extension

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/mahtdy/phpvm/internal/cli"
	phpini "github.com/mahtdy/phpvm/internal/ini"
)

// Status describes whether an extension is loaded.
type Status struct {
	Name    string
	Enabled bool
	// Path is the .so/.dll path (may be empty if built-in).
	Path string
}

// Manager handles PHP extension operations via php.ini manipulation.
type Manager struct {
	iniFile     *phpini.File
	execCommand func(name string, args ...string) ([]byte, error)
}

// NewManager creates a Manager for the php.ini at iniPath.
func NewManager(iniPath string) (*Manager, error) {
	f, err := phpini.Load(iniPath)
	if err != nil {
		return nil, err
	}
	return &Manager{iniFile: f, execCommand: defaultExec}, nil
}

// SetExecCommand replaces the command executor (for testing).
func (m *Manager) SetExecCommand(fn func(string, ...string) ([]byte, error)) {
	m.execCommand = fn
}

// List returns status of all extensions found in php.ini.
func (m *Manager) List() []*Status {
	all := m.iniFile.All()
	var result []*Status
	for k, v := range all {
		name := extensionName(k)
		if name == "" {
			continue
		}
		result = append(result, &Status{
			Name:    name,
			Enabled: true,
			Path:    v,
		})
	}
	// Also scan for commented-out extensions.
	return result
}

// ListLoaded returns extensions currently loaded by the PHP runtime.
func (m *Manager) ListLoaded(phpBinary string) ([]*Status, error) {
	out, err := m.execCommand(phpBinary, "-m")
	if err != nil {
		return nil, cli.NewError("failed to list loaded PHP extensions", err)
	}
	return parseModulesList(string(out)), nil
}

// Enable uncomments or adds an extension directive in php.ini.
func (m *Manager) Enable(name string) error {
	// Check if already enabled as "extension=name" or "extension=name.so".
	key := "extension"
	existing, ok := m.iniFile.Get(key)
	if ok && strings.EqualFold(existing, name) {
		return nil // already enabled
	}
	m.iniFile.Set(key, name)
	if err := m.iniFile.Save(); err != nil {
		return cli.NewError(fmt.Sprintf("failed to enable extension %s", name), err)
	}
	return nil
}

// Disable comments out the extension directive for name.
func (m *Manager) Disable(name string) error {
	m.iniFile.Delete("extension")
	if err := m.iniFile.Save(); err != nil {
		return cli.NewError(fmt.Sprintf("failed to disable extension %s", name), err)
	}
	return nil
}

// GetStatus returns the status of a named extension using php -m.
func (m *Manager) GetStatus(phpBinary, name string) (*Status, error) {
	loaded, err := m.ListLoaded(phpBinary)
	if err != nil {
		return nil, err
	}
	nameLower := strings.ToLower(name)
	for _, s := range loaded {
		if strings.ToLower(s.Name) == nameLower {
			return s, nil
		}
	}
	return &Status{Name: name, Enabled: false}, nil
}

// parseModulesList parses output of `php -m`.
// Format: sections [PHP Modules] and [Zend Modules] with one extension per line.
var sectionRe = regexp.MustCompile(`^\[.*\]$`)

func parseModulesList(output string) []*Status {
	var result []*Status
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || sectionRe.MatchString(line) {
			continue
		}
		result = append(result, &Status{Name: line, Enabled: true})
	}
	return result
}

// extensionName returns the extension name from a php.ini key like
// "extension", "zend_extension", etc. Returns "" for non-extension keys.
func extensionName(key string) string {
	switch strings.ToLower(key) {
	case "extension", "zend_extension":
		return key
	}
	return ""
}

func defaultExec(name string, args ...string) ([]byte, error) {
	//nolint:gosec
	return exec.Command(name, args...).Output()
}
