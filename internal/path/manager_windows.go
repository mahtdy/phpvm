//go:build windows

package path

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	registryPath = `Environment`
	registryKey  = `HKCU`
)

// windowsManager manages PATH via HKCU\Environment on Windows.
type windowsManager struct{}

// New returns the platform-specific Manager (Windows).
func New() Manager {
	return &windowsManager{}
}

// Add prepends dir to the user PATH in the Windows registry.
func (m *windowsManager) Add(dir string) error {
	current, err := m.readRegistry()
	if err != nil {
		return err
	}

	// Remove any previous phpvm-managed entry first.
	cleaned := m.removePHPVMEntries(current)

	// Prepend the new dir.
	var newPath string
	if cleaned != "" {
		newPath = dir + ";" + cleaned
	} else {
		newPath = dir
	}

	// Deduplicate.
	newPath = deduplicatePath(newPath, ";")

	return m.writeRegistry(newPath)
}

// Remove strips all phpvm-managed entries from the Windows registry PATH.
func (m *windowsManager) Remove() error {
	current, err := m.readRegistry()
	if err != nil {
		return err
	}
	cleaned := m.removePHPVMEntries(current)
	return m.writeRegistry(cleaned)
}

// Current returns the current PATH entries from the registry.
func (m *windowsManager) Current() ([]string, error) {
	pathStr, err := m.readRegistry()
	if err != nil {
		return nil, err
	}
	parts := strings.Split(pathStr, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			result = append(result, p)
		}
	}
	return result, nil
}

// Verify reports whether dir is in the registry PATH.
func (m *windowsManager) Verify(dir string) (bool, error) {
	entries, err := m.Current()
	if err != nil {
		return false, err
	}
	dirLower := strings.ToLower(dir)
	for _, e := range entries {
		if strings.ToLower(e) == dirLower {
			return true, nil
		}
	}
	return false, nil
}

// readRegistry reads the PATH value from HKCU\Environment.
func (m *windowsManager) readRegistry() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return os.Getenv("PATH"), nil // fallback to env if registry unavailable
	}
	defer k.Close()

	val, _, err := k.GetStringValue("PATH")
	if err != nil {
		// PATH key might not exist — not an error.
		return "", nil
	}
	return val, nil
}

// writeRegistry writes pathStr to HKCU\Environment\PATH.
func (m *windowsManager) writeRegistry(pathStr string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("path: open registry key: %w", err)
	}
	defer k.Close()

	if err := k.SetExpandStringValue("PATH", pathStr); err != nil {
		return fmt.Errorf("path: write registry PATH: %w", err)
	}
	return nil
}

// removePHPVMEntries removes entries under the phpvm versions directory from pathStr.
func (m *windowsManager) removePHPVMEntries(pathStr string) string {
	parts := strings.Split(pathStr, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		lower := strings.ToLower(p)
		// Remove any path that looks like ~/.phpvm/versions/...
		if strings.Contains(lower, ".phpvm") || strings.Contains(lower, "phpvm") {
			continue
		}
		result = append(result, p)
	}
	return strings.Join(result, ";")
}
