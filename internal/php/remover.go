package php

import (
	"fmt"
	"os"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/logger"
	"github.com/mahtdy/phpvm/internal/validation"
)

// Remover handles PHP version uninstallation.
type Remover struct {
	cfg     *config.Config
	scanner *Scanner
}

// NewRemover creates a Remover backed by cfg.
func NewRemover(cfg *config.Config) *Remover {
	return &Remover{cfg: cfg, scanner: NewScanner(cfg.VersionsDir())}
}

// Remove uninstalls the best-matching installed PHP version.
// Returns an error if the version is currently active.
func (rem *Remover) Remove(selectorStr string) (*InstalledVersion, error) {
	sel, err := validation.Parse(selectorStr)
	if err != nil {
		return nil, err
	}

	iv, err := rem.scanner.FindBestMatch(sel)
	if err != nil {
		return nil, err
	}

	// Refuse to remove the currently active version.
	if iv.Version.String() == rem.cfg.Current {
		return nil, cli.NewValidationError(
			fmt.Sprintf(
				"cannot remove PHP %s because it is the active version — run 'phpvm use <other>' first",
				iv.Version.String(),
			),
			nil,
		)
	}

	logger.Info(fmt.Sprintf("Removing PHP %s from %s...", iv.Version.String(), iv.Dir))

	if err := os.RemoveAll(iv.Dir); err != nil {
		return nil, cli.NewPermissionError(
			fmt.Sprintf("failed to remove %s", iv.Dir), err,
		)
	}

	logger.Info(fmt.Sprintf("PHP %s removed.", iv.Version.String()))
	return iv, nil
}
