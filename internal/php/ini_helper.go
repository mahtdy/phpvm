package php

import (
	"os"
	"path/filepath"

	"github.com/mahtdy/phpvm/internal/cli"
)

// FindIniNextToBinary returns the php.ini path adjacent to the PHP binary.
func FindIniNextToBinary(phpBinary string) (string, error) {
	dir := filepath.Dir(phpBinary)
	candidate := filepath.Join(dir, "php.ini")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", cli.NewFileNotFoundError("php.ini not found next to "+phpBinary, nil)
}
