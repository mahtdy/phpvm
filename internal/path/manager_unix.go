//go:build !windows

package path

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// unixManager manages PATH via shell profile files on Linux/macOS.
type unixManager struct{}

// New returns the platform-specific Manager (Unix).
func New() Manager {
	return &unixManager{}
}

// Add prepends dir to PATH by writing a marker block into the user's shell profile.
func (m *unixManager) Add(dir string) error {
	profile, err := m.detectProfile()
	if err != nil {
		return err
	}

	content, err := readOrEmpty(profile)
	if err != nil {
		return err
	}

	// Back up the profile before modification.
	if err := backupProfile(profile, content); err != nil {
		return err
	}

	newContent := replaceMarkerBlock(content, buildMarkerBlock(dir))
	return writeFileAtomic(profile, []byte(newContent), 0o644)
}

// Remove strips the phpvm marker block from the shell profile.
func (m *unixManager) Remove() error {
	profile, err := m.detectProfile()
	if err != nil {
		return err
	}

	content, err := readOrEmpty(profile)
	if err != nil {
		return err
	}

	newContent := removeMarkerBlock(content)
	return writeFileAtomic(profile, []byte(newContent), 0o644)
}

// Current returns PATH entries split from the current process environment.
func (m *unixManager) Current() ([]string, error) {
	pathStr := os.Getenv("PATH")
	parts := strings.Split(pathStr, ":")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			result = append(result, p)
		}
	}
	return result, nil
}

// Verify reports whether dir is in the current process PATH.
func (m *unixManager) Verify(dir string) (bool, error) {
	entries, err := m.Current()
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e == dir {
			return true, nil
		}
	}
	return false, nil
}

// detectProfile returns the most appropriate shell profile path.
// Priority: ZDOTDIR/.zshrc > ~/.zshrc > ~/.bashrc > ~/.profile
func (m *unixManager) detectProfile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("path: cannot determine home directory: %w", err)
	}

	shell := os.Getenv("SHELL")

	switch {
	case strings.Contains(shell, "zsh"):
		if zdotdir := os.Getenv("ZDOTDIR"); zdotdir != "" {
			return filepath.Join(zdotdir, ".zshrc"), nil
		}
		return filepath.Join(home, ".zshrc"), nil
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	case strings.Contains(shell, "bash"):
		// Prefer .bash_profile on macOS, .bashrc on Linux.
		rc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(rc); os.IsNotExist(err) {
			return filepath.Join(home, ".bash_profile"), nil
		}
		return rc, nil
	default:
		return filepath.Join(home, ".profile"), nil
	}
}

// readOrEmpty reads a file or returns empty string if it does not exist.
func readOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("path: read %s: %w", path, err)
	}
	return string(data), nil
}

// backupProfile writes a .bak copy of the profile before modification.
func backupProfile(profile, content string) error {
	if content == "" {
		return nil
	}
	bak := profile + ".phpvm.bak"
	if err := os.WriteFile(bak, []byte(content), 0o600); err != nil {
		return fmt.Errorf("path: backup profile: %w", err)
	}
	return nil
}
