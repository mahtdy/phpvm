// Package path manages system PATH modifications for phpvm across all platforms.
// On Windows it reads/writes the HKCU\Environment registry key.
// On Unix it modifies the user's shell profile files.
package path

import (
	"fmt"
	"os"
	"strings"
)

const (
	// markerBegin is the start of the phpvm-managed PATH block in shell profiles.
	markerBegin = "# >>> phpvm initialize >>>"
	// markerEnd is the end of the phpvm-managed PATH block.
	markerEnd = "# <<< phpvm initialize <<<"
)

// Manager handles PATH modifications for the current platform.
type Manager interface {
	// Add inserts dir into the system PATH (idempotent).
	Add(dir string) error
	// Remove removes any phpvm-managed PATH entry.
	Remove() error
	// Current returns the current PATH entries as a slice.
	Current() ([]string, error)
	// Verify reports whether dir is present in the current PATH.
	Verify(dir string) (bool, error)
}

// DeduplicatePath removes duplicate entries from a PATH-style colon/semicolon string.
func DeduplicatePath(pathStr, sep string) string {
	return deduplicatePath(pathStr, sep)
}

// ReplaceMarkerBlock is the exported wrapper for tests.
func ReplaceMarkerBlock(content, newBlock string) string {
	return replaceMarkerBlock(content, newBlock)
}

// RemoveMarkerBlock is the exported wrapper for tests.
func RemoveMarkerBlock(content string) string {
	return removeMarkerBlock(content)
}

// BuildMarkerBlock is the exported wrapper for tests.
func BuildMarkerBlock(dir string) string {
	return buildMarkerBlock(dir)
}
func RestartHint() string {
	return "Restart your shell or run: source ~/.bashrc (Linux/macOS) | . $PROFILE (PowerShell)"
}

// buildMarkerBlock builds the shell profile block that prepends dir to PATH.
func buildMarkerBlock(dir string) string {
	return fmt.Sprintf(
		"%s\nexport PATH=%s:$PATH\n%s",
		markerBegin,
		shellEscape(dir),
		markerEnd,
	)
}

// buildPowerShellBlock builds the PowerShell profile block.
func buildPowerShellBlock(dir string) string {
	return fmt.Sprintf(
		"%s\n$env:PATH = \"%s;\" + $env:PATH\n%s",
		markerBegin,
		dir,
		markerEnd,
	)
}

// replaceMarkerBlock replaces the phpvm block in content, or appends it.
func replaceMarkerBlock(content, newBlock string) string {
	start := strings.Index(content, markerBegin)
	end := strings.Index(content, markerEnd)

	if start >= 0 && end >= 0 && end > start {
		// Replace existing block.
		before := content[:start]
		after := content[end+len(markerEnd):]
		// Trim surrounding blank lines.
		after = strings.TrimLeft(after, "\n")
		return before + newBlock + "\n" + after
	}

	// Append new block.
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + newBlock + "\n"
}

// removeMarkerBlock strips the phpvm block from content.
func removeMarkerBlock(content string) string {
	start := strings.Index(content, markerBegin)
	end := strings.Index(content, markerEnd)

	if start < 0 || end < 0 || end <= start {
		return content
	}

	before := content[:start]
	after := content[end+len(markerEnd):]
	after = strings.TrimLeft(after, "\n")

	return strings.TrimRight(before, "\n") + "\n" + after
}

// deduplicatePath removes duplicate entries from a PATH-style colon/semicolon string.
func deduplicatePath(pathStr, sep string) string {
	parts := strings.Split(pathStr, sep)
	seen := make(map[string]bool, len(parts))
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		result = append(result, p)
	}
	return strings.Join(result, sep)
}

// shellEscape wraps dir in quotes if it contains spaces.
func shellEscape(dir string) string {
	if strings.ContainsAny(dir, " \t") {
		return fmt.Sprintf(`"%s"`, dir)
	}
	return dir
}

// writeFileAtomic writes data to path atomically via a temp file + rename.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".phpvm.tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("write temp file %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename %s -> %s: %w", tmp, path, err)
	}
	return nil
}
