package version_test

import (
	"strings"
	"testing"

	"github.com/mahtdy/phpvm/pkg/version"
)

func TestGet(t *testing.T) {
	info := version.Get()
	if info.Version == "" {
		t.Error("Version must not be empty")
	}
}

func TestInfoString(t *testing.T) {
	version.Version = "1.2.3"
	version.Commit = "abc123"
	version.Date = "2025-01-01"

	info := version.Get()
	s := info.String()

	if !strings.Contains(s, "phpvm v1.2.3") {
		t.Errorf("expected version in string, got: %s", s)
	}
	if !strings.Contains(s, "abc123") {
		t.Errorf("expected commit in string, got: %s", s)
	}
}
