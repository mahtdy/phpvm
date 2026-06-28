package path_test

import (
	"strings"
	"testing"

	"github.com/mahtdy/phpvm/internal/path"
	"github.com/stretchr/testify/assert"
)

// These tests exercise the pure, platform-independent helpers in manager.go.
// Platform-specific integration tests live in manager_windows_test.go
// and manager_unix_test.go.

func TestDeduplicatePath(t *testing.T) {
	cases := []struct {
		input    string
		sep      string
		expected string
	}{
		{"/a:/b:/a:/c", ":", "/a:/b:/c"},
		{"C:\\Go;C:\\Go;C:\\foo", ";", "C:\\Go;C:\\foo"},
		{"/a", ":", "/a"},
		{"", ":", ""},
	}
	for _, c := range cases {
		got := path.DeduplicatePath(c.input, c.sep)
		assert.Equal(t, c.expected, got, "input=%q", c.input)
	}
}

func TestReplaceMarkerBlock_Append(t *testing.T) {
	original := "export FOO=bar\n"
	block := "# >>> phpvm initialize >>>\nexport PATH=/foo:$PATH\n# <<< phpvm initialize <<<"

	result := path.ReplaceMarkerBlock(original, block)

	assert.Contains(t, result, "export FOO=bar")
	assert.Contains(t, result, "# >>> phpvm initialize >>>")
	assert.Contains(t, result, "export PATH=/foo:$PATH")
	assert.Contains(t, result, "# <<< phpvm initialize <<<")
}

func TestReplaceMarkerBlock_Replace(t *testing.T) {
	original := "export FOO=bar\n# >>> phpvm initialize >>>\nexport PATH=/old:$PATH\n# <<< phpvm initialize <<<\nexport BAZ=qux\n"
	block := "# >>> phpvm initialize >>>\nexport PATH=/new:$PATH\n# <<< phpvm initialize <<<"

	result := path.ReplaceMarkerBlock(original, block)

	assert.Contains(t, result, "/new")
	assert.NotContains(t, result, "/old")
	assert.Contains(t, result, "export FOO=bar")
	assert.Contains(t, result, "export BAZ=qux")
}

func TestRemoveMarkerBlock(t *testing.T) {
	content := "before\n# >>> phpvm initialize >>>\nexport PATH=/foo:$PATH\n# <<< phpvm initialize <<<\nafter\n"

	result := path.RemoveMarkerBlock(content)

	assert.NotContains(t, result, "phpvm initialize")
	assert.NotContains(t, result, "/foo")
	assert.Contains(t, result, "before")
	assert.Contains(t, result, "after")
}

func TestRemoveMarkerBlock_NoBlock(t *testing.T) {
	content := "just some content\n"
	result := path.RemoveMarkerBlock(content)
	assert.Equal(t, content, result)
}

func TestBuildMarkerBlock_ContainsDir(t *testing.T) {
	block := path.BuildMarkerBlock("/home/user/.phpvm/versions/8.3.10")
	assert.True(t, strings.Contains(block, "/home/user/.phpvm/versions/8.3.10"))
	assert.Contains(t, block, "# >>> phpvm initialize >>>")
}
