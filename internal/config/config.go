// Package config manages phpvm configuration loading, saving, and defaults.
// Config is stored at ~/.phpvm/config.json and can be overridden with
// PHPVM_* environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

const (
	// EnvPrefix is the prefix for all phpvm environment variables.
	EnvPrefix = "PHPVM"

	// DefaultConfigFile is the config file name (without extension).
	DefaultConfigFile = "config"

	// DefaultArch is the default CPU architecture.
	DefaultArch = "x64"
)

// Config holds all phpvm runtime settings.
type Config struct {
	// Root is the base directory for phpvm data (~/.phpvm).
	Root string `mapstructure:"root"`
	// Current is the active PHP version string (e.g. "8.3.10").
	Current string `mapstructure:"current"`
	// DefaultArch is the CPU architecture used for downloads (x64 / arm64).
	DefaultArch string `mapstructure:"default_arch"`
	// Telemetry enables anonymous usage reporting (opt-in only).
	Telemetry bool `mapstructure:"telemetry"`
	// LogLevel controls the verbosity (debug, info, warn, error).
	LogLevel string `mapstructure:"log_level"`
	// NoColor disables ANSI color output.
	NoColor bool `mapstructure:"no_color"`
}

// v is the package-level Viper instance.
var v *viper.Viper

// Load reads the config file and environment variables, returning a Config.
// If the config file does not exist, defaults are used without error.
func Load(cfgFile string) (*Config, error) {
	v = viper.New()

	setDefaults()

	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		root := defaultRoot()
		v.AddConfigPath(root)
		v.SetConfigName(DefaultConfigFile)
		v.SetConfigType("json")
	}

	if err := v.ReadInConfig(); err != nil {
		// Config file missing is not an error — use defaults.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Check for path-not-found errors too.
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("config: read config file: %w", err)
			}
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	return cfg, nil
}

// Save persists cfg to the config file.
func Save(cfg *Config) error {
	if v == nil {
		v = viper.New()
		setDefaults()
	}

	v.Set("root", cfg.Root)
	v.Set("current", cfg.Current)
	v.Set("default_arch", cfg.DefaultArch)
	v.Set("telemetry", cfg.Telemetry)
	v.Set("log_level", cfg.LogLevel)
	v.Set("no_color", cfg.NoColor)

	// Ensure config dir exists.
	if err := os.MkdirAll(cfg.Root, 0o750); err != nil {
		return fmt.Errorf("config: create root dir: %w", err)
	}

	cfgPath := filepath.Join(cfg.Root, DefaultConfigFile+".json")
	v.SetConfigFile(cfgPath)

	if err := v.WriteConfigAs(cfgPath); err != nil {
		return fmt.Errorf("config: write config: %w", err)
	}
	return nil
}

// VersionsDir returns the path where PHP versions are stored.
func (c *Config) VersionsDir() string {
	return filepath.Join(c.Root, "versions")
}

// DownloadsDir returns the path for cached download files.
func (c *Config) DownloadsDir() string {
	return filepath.Join(c.Root, "downloads")
}

// CacheDir returns the path for resolved-version caches.
func (c *Config) CacheDir() string {
	return filepath.Join(c.Root, "cache")
}

// PluginsDir returns the path where plugins are installed.
func (c *Config) PluginsDir() string {
	return filepath.Join(c.Root, "plugins")
}

// setDefaults registers default values in Viper.
func setDefaults() {
	v.SetDefault("root", defaultRoot())
	v.SetDefault("current", "")
	v.SetDefault("default_arch", defaultArch())
	v.SetDefault("telemetry", false)
	v.SetDefault("log_level", "info")
	v.SetDefault("no_color", false)
}

// defaultRoot returns the platform-appropriate phpvm root directory.
func defaultRoot() string {
	// PHPVM_ROOT env takes precedence.
	if r := os.Getenv("PHPVM_ROOT"); r != "" {
		return r
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".phpvm")
	}
	return filepath.Join(home, ".phpvm")
}

// defaultArch returns the host CPU architecture string.
func defaultArch() string {
	switch runtime.GOARCH {
	case "arm64":
		return "arm64"
	default:
		return "x64"
	}
}
