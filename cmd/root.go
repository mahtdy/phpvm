// Package cmd implements all phpvm CLI commands using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	debug   bool
	noColor bool
	// Cfg is the loaded configuration, available to all sub-commands.
	Cfg *config.Config
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "phpvm",
	Short: "phpvm — PHP Version Manager",
	Long: `phpvm is a cross-platform PHP version manager.

Install, switch, and manage multiple PHP versions with a single tool.
Inspired by nvm, Volta, and Rustup.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config init for the version and completion commands.
		if cmd.Name() == "version" || cmd.Name() == "completion" {
			return nil
		}
		return initConfig()
	},
}

// Execute adds all child commands to the root command and runs the CLI.
// It returns the exit code.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", cli.UserMessage(err))
		return cli.ExitCodeFrom(err)
	}
	return cli.ExitSuccess
}

func init() {
	cobra.OnInitialize(setupLogger)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: ~/.phpvm/config.json)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false,
		"enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false,
		"disable colored output")

	// Bind flags to viper so env vars and config file can override them.
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Register sub-commands.
	rootCmd.AddCommand(newVersionCmd())
}

// setupLogger configures the global logger based on flags.
// Called by cobra.OnInitialize (before any command runs).
func setupLogger() {
	lvl := logger.LevelInfo
	if debug {
		lvl = logger.LevelDebug
	}
	logger.Init(logger.Options{
		Level:   lvl,
		NoColor: noColor,
	})
}

// initConfig loads the configuration file and environment variables.
func initConfig() error {
	var err error
	Cfg, err = config.Load(cfgFile)
	if err != nil {
		return cli.NewError("failed to load configuration", err)
	}

	// Re-apply log level from config if not set via flag.
	if !debug && Cfg.LogLevel == "debug" {
		logger.Reset()
		logger.Init(logger.Options{
			Level:   logger.LevelDebug,
			NoColor: noColor || Cfg.NoColor,
		})
	}
	return nil
}
