// Package cli provides shared CLI infrastructure including typed errors and
// exit-code conventions used throughout phpvm.
package cli

import (
	"errors"
	"fmt"
)

// Exit codes used by phpvm commands.
const (
	ExitSuccess     = 0
	ExitGeneralErr  = 1
	ExitUsageErr    = 2
	ExitPermission  = 3
	ExitNetwork     = 4
	ExitNotFound    = 5
	ExitArchive     = 6
	ExitPathConflict = 7
	ExitRegistry    = 8
	ExitValidation  = 9
	ExitNotInstalled = 10
)

// ErrorKind classifies the type of error for consistent UX and exit codes.
type ErrorKind int

const (
	KindGeneral      ErrorKind = iota
	KindPermission             // insufficient permissions
	KindNetwork                // connectivity / download failures
	KindFileNotFound           // file or directory missing
	KindArchive                // zip/tar extraction problems
	KindPathConflict           // PATH manipulation conflict
	KindRegistry               // Windows registry error
	KindValidation             // invalid user input
	KindNotInstalled           // requested PHP version not installed
)

// PhpvmError is the structured error type for all phpvm operations.
// It carries a user-facing message, the underlying cause, and metadata
// needed to select the correct exit code and debug output.
type PhpvmError struct {
	Kind    ErrorKind
	Message string // user-friendly message (English)
	Cause   error  // original error (may be nil)
}

// Error implements the error interface.
func (e *PhpvmError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap allows errors.Is / errors.As to traverse the chain.
func (e *PhpvmError) Unwrap() error {
	return e.Cause
}

// ExitCode returns the process exit code appropriate for this error kind.
func (e *PhpvmError) ExitCode() int {
	switch e.Kind {
	case KindPermission:
		return ExitPermission
	case KindNetwork:
		return ExitNetwork
	case KindFileNotFound:
		return ExitNotFound
	case KindArchive:
		return ExitArchive
	case KindPathConflict:
		return ExitPathConflict
	case KindRegistry:
		return ExitRegistry
	case KindValidation:
		return ExitValidation
	case KindNotInstalled:
		return ExitNotInstalled
	default:
		return ExitGeneralErr
	}
}

// --- Constructor helpers ---

// NewError creates a general PhpvmError.
func NewError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindGeneral, Message: msg, Cause: cause}
}

// NewPermissionError creates a permission-related error.
func NewPermissionError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindPermission, Message: msg, Cause: cause}
}

// NewNetworkError creates a network-related error.
func NewNetworkError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindNetwork, Message: msg, Cause: cause}
}

// NewFileNotFoundError creates a file-not-found error.
func NewFileNotFoundError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindFileNotFound, Message: msg, Cause: cause}
}

// NewArchiveError creates an archive extraction error.
func NewArchiveError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindArchive, Message: msg, Cause: cause}
}

// NewPathConflictError creates a PATH conflict error.
func NewPathConflictError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindPathConflict, Message: msg, Cause: cause}
}

// NewRegistryError creates a Windows Registry error.
func NewRegistryError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindRegistry, Message: msg, Cause: cause}
}

// NewValidationError creates a user-input validation error.
func NewValidationError(msg string, cause error) *PhpvmError {
	return &PhpvmError{Kind: KindValidation, Message: msg, Cause: cause}
}

// NewNotInstalledError creates a "version not installed" error.
func NewNotInstalledError(version string) *PhpvmError {
	return &PhpvmError{
		Kind:    KindNotInstalled,
		Message: fmt.Sprintf("PHP %s is not installed — run: phpvm install %s", version, version),
	}
}

// --- Sentinel errors for errors.Is matching ---

var (
	ErrPermission   = errors.New("permission denied")
	ErrNetwork      = errors.New("network error")
	ErrFileNotFound = errors.New("file not found")
	ErrArchive      = errors.New("archive error")
	ErrPathConflict = errors.New("path conflict")
	ErrRegistry     = errors.New("registry error")
	ErrValidation   = errors.New("validation error")
	ErrNotInstalled = errors.New("version not installed")
)

// kindToSentinel maps ErrorKind to its sentinel for errors.Is support.
var kindToSentinel = map[ErrorKind]error{
	KindPermission:   ErrPermission,
	KindNetwork:      ErrNetwork,
	KindFileNotFound: ErrFileNotFound,
	KindArchive:      ErrArchive,
	KindPathConflict: ErrPathConflict,
	KindRegistry:     ErrRegistry,
	KindValidation:   ErrValidation,
	KindNotInstalled: ErrNotInstalled,
}

// Is enables errors.Is to match PhpvmError against sentinel errors.
func (e *PhpvmError) Is(target error) bool {
	if sentinel, ok := kindToSentinel[e.Kind]; ok {
		return errors.Is(sentinel, target) || sentinel == target
	}
	return false
}

// ExitCodeFrom returns the exit code from an error.
// If err is nil it returns ExitSuccess.
// If err is not a *PhpvmError it returns ExitGeneralErr.
func ExitCodeFrom(err error) int {
	if err == nil {
		return ExitSuccess
	}
	var pe *PhpvmError
	if errors.As(err, &pe) {
		return pe.ExitCode()
	}
	return ExitGeneralErr
}

// UserMessage returns a user-friendly string from any error.
// For *PhpvmError it returns the Message field; otherwise err.Error().
func UserMessage(err error) string {
	if err == nil {
		return ""
	}
	var pe *PhpvmError
	if errors.As(err, &pe) {
		return pe.Message
	}
	return err.Error()
}
