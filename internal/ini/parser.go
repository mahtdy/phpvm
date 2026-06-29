// Package ini provides php.ini file parsing, writing, and backup operations.
package ini

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/mahtdy/phpvm/internal/cli"
)

// keyValueRe matches "key = value" or "key=value", ignoring leading whitespace.
var keyValueRe = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z0-9_.]*)\s*=\s*(.*)$`)

// Entry represents a single php.ini directive.
type Entry struct {
	Key     string
	Value   string
	Comment string // inline comment (everything after ';' on the value line)
	Raw     string // original line, preserved for unknown lines
}

// File represents a parsed php.ini file.
type File struct {
	path    string
	entries []*Entry // ordered list preserving file structure
	index   map[string]int // key → index in entries
}

// Load parses the php.ini file at path.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, cli.NewFileNotFoundError(fmt.Sprintf("cannot read php.ini: %s", path), err)
	}
	f := &File{path: path, index: make(map[string]int)}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		f.parseLine(line)
	}
	return f, nil
}

// Get returns the value for key, or ("", false) if not present.
func (f *File) Get(key string) (string, bool) {
	if i, ok := f.index[strings.ToLower(key)]; ok {
		return f.entries[i].Value, true
	}
	return "", false
}

// Set updates an existing key or appends a new directive.
func (f *File) Set(key, value string) {
	lk := strings.ToLower(key)
	if i, ok := f.index[lk]; ok {
		f.entries[i].Value = value
		f.entries[i].Raw = ""
		return
	}
	idx := len(f.entries)
	f.entries = append(f.entries, &Entry{Key: key, Value: value})
	f.index[lk] = idx
}

// Delete removes a directive by key (no-op if absent).
func (f *File) Delete(key string) {
	lk := strings.ToLower(key)
	if _, ok := f.index[lk]; !ok {
		return
	}
	// Mark as deleted by zeroing key/value and clearing raw.
	i := f.index[lk]
	f.entries[i] = &Entry{Raw: ""}
	delete(f.index, lk)
}

// All returns all key-value pairs as a map.
func (f *File) All() map[string]string {
	out := make(map[string]string, len(f.index))
	for k, i := range f.index {
		out[k] = f.entries[i].Value
	}
	return out
}

// Path returns the file path.
func (f *File) Path() string { return f.path }

// Save writes the modified php.ini back to disk (atomic).
func (f *File) Save() error {
	return f.SaveTo(f.path)
}

// SaveTo writes to an explicit path (used for backup-then-save patterns).
func (f *File) SaveTo(path string) error {
	var sb strings.Builder
	for _, e := range f.entries {
		if e.Raw != "" || (e.Key == "" && e.Value == "") {
			sb.WriteString(e.Raw)
		} else {
			sb.WriteString(e.Key + " = " + e.Value)
			if e.Comment != "" {
				sb.WriteString(" ;" + e.Comment)
			}
		}
		sb.WriteByte('\n')
	}

	tmp := path + ".phpvm.tmp"
	if err := os.WriteFile(tmp, []byte(sb.String()), 0o640); err != nil {
		return cli.NewPermissionError("cannot write php.ini temp file", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return cli.NewPermissionError("cannot replace php.ini", err)
	}
	return nil
}

// Backup creates a timestamped backup of the file.
func (f *File) Backup() (string, error) {
	ts := time.Now().Format("20060102-150405")
	backupPath := f.path + "." + ts + ".bak"
	data, err := os.ReadFile(f.path)
	if err != nil {
		return "", cli.NewFileNotFoundError("cannot read php.ini for backup", err)
	}
	if err := os.WriteFile(backupPath, data, 0o640); err != nil {
		return "", cli.NewPermissionError("cannot write php.ini backup", err)
	}
	return backupPath, nil
}

// parseLine classifies and appends a line from the ini file.
func (f *File) parseLine(line string) {
	// Comments and section headers — preserve as-is.
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "[") {
		f.entries = append(f.entries, &Entry{Raw: line})
		return
	}
	m := keyValueRe.FindStringSubmatch(line)
	if m == nil {
		f.entries = append(f.entries, &Entry{Raw: line})
		return
	}
	key := m[1]
	val := strings.TrimSpace(m[2])
	comment := ""
	// Split inline comment.
	if idx := strings.Index(val, " ;"); idx >= 0 {
		comment = strings.TrimSpace(val[idx+2:])
		val = strings.TrimSpace(val[:idx])
	}

	idx := len(f.entries)
	e := &Entry{Key: key, Value: val, Comment: comment}
	f.entries = append(f.entries, e)
	f.index[strings.ToLower(key)] = idx
}

// Locator finds the php.ini for a given PHP version directory.
type Locator struct {
	execCommand func(name string, args ...string) ([]byte, error)
}

// NewLocator creates a Locator.
func NewLocator() *Locator {
	return &Locator{execCommand: defaultExec}
}

// SetExecCommand replaces the executor (for testing).
func (l *Locator) SetExecCommand(fn func(string, ...string) ([]byte, error)) {
	l.execCommand = fn
}

// Find returns the php.ini path for the given PHP binary.
// It queries `php --ini` and parses the "Loaded Configuration File" line.
func (l *Locator) Find(phpBinary string) (string, error) {
	out, err := l.execCommand(phpBinary, "--ini")
	if err != nil {
		return "", cli.NewError("failed to query php --ini", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Loaded Configuration File") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				p := strings.TrimSpace(parts[1])
				if p != "(none)" && p != "" {
					return p, nil
				}
			}
		}
	}
	// Fallback: look for php.ini next to the binary.
	dir := filepath.Dir(phpBinary)
	candidate := filepath.Join(dir, "php.ini")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", cli.NewFileNotFoundError("php.ini not found for "+phpBinary, nil)
}

// FindForVersion locates php.ini using the PHP binary in versionDir.
func (l *Locator) FindForVersion(versionDir string) (string, error) {
	bin := filepath.Join(versionDir, phpBin())
	return l.Find(bin)
}

func phpBin() string {
	if runtime.GOOS == "windows" {
		return "php.exe"
	}
	return "php"
}

func defaultExec(name string, args ...string) ([]byte, error) {
	//nolint:gosec
	return exec.Command(name, args...).Output()
}
