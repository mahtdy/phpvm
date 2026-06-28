package cli_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestPhpvmErrorMessage(t *testing.T) {
	cause := errors.New("disk full")
	err := cli.NewError("write failed", cause)
	assert.Contains(t, err.Error(), "write failed")
	assert.Contains(t, err.Error(), "disk full")
}

func TestPhpvmErrorNoCause(t *testing.T) {
	err := cli.NewValidationError("invalid version format", nil)
	assert.Equal(t, "invalid version format", err.Error())
}

func TestExitCodes(t *testing.T) {
	cases := []struct {
		err      *cli.PhpvmError
		expected int
	}{
		{cli.NewPermissionError("perm", nil), cli.ExitPermission},
		{cli.NewNetworkError("net", nil), cli.ExitNetwork},
		{cli.NewFileNotFoundError("notfound", nil), cli.ExitNotFound},
		{cli.NewArchiveError("arch", nil), cli.ExitArchive},
		{cli.NewPathConflictError("path", nil), cli.ExitPathConflict},
		{cli.NewRegistryError("reg", nil), cli.ExitRegistry},
		{cli.NewValidationError("val", nil), cli.ExitValidation},
		{cli.NewNotInstalledError("8.4"), cli.ExitNotInstalled},
		{cli.NewError("gen", nil), cli.ExitGeneralErr},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, c.err.ExitCode(), c.err.Message)
	}
}

func TestErrorsIs(t *testing.T) {
	err := cli.NewNetworkError("connection refused", errors.New("EOF"))
	assert.True(t, errors.Is(err, cli.ErrNetwork))
	assert.False(t, errors.Is(err, cli.ErrPermission))
}

func TestErrorsAs(t *testing.T) {
	wrapped := fmt.Errorf("outer: %w", cli.NewValidationError("bad input", nil))
	var pe *cli.PhpvmError
	assert.True(t, errors.As(wrapped, &pe))
	assert.Equal(t, cli.KindValidation, pe.Kind)
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := cli.NewError("msg", cause)
	assert.Equal(t, cause, errors.Unwrap(err))
}

func TestExitCodeFrom(t *testing.T) {
	assert.Equal(t, cli.ExitSuccess, cli.ExitCodeFrom(nil))
	assert.Equal(t, cli.ExitNetwork, cli.ExitCodeFrom(cli.NewNetworkError("x", nil)))
	assert.Equal(t, cli.ExitGeneralErr, cli.ExitCodeFrom(errors.New("plain")))
}

func TestUserMessage(t *testing.T) {
	assert.Equal(t, "", cli.UserMessage(nil))
	assert.Equal(t, "bad version", cli.UserMessage(cli.NewValidationError("bad version", nil)))
	assert.Equal(t, "plain error", cli.UserMessage(errors.New("plain error")))
}

func TestNotInstalledMessage(t *testing.T) {
	err := cli.NewNotInstalledError("8.3")
	assert.Contains(t, err.Error(), "8.3")
	assert.Contains(t, err.Error(), "phpvm install")
}
