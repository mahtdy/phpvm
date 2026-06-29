package cmd

import (
	"fmt"

	"github.com/mahtdy/phpvm/internal/extension"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newExtCmd())
}

func newExtCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ext",
		Short: "Manage PHP extensions",
		Long:  "List, enable, or disable PHP extensions via php.ini.",
	}
	cmd.AddCommand(newExtListCmd(), newExtEnableCmd(), newExtDisableCmd(), newExtStatusCmd())
	return cmd
}

func newExtListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List loaded PHP extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			phpBin, err := activePHPBinary()
			if err != nil {
				return err
			}
			iniPath, err := activeIniPath()
			if err != nil {
				return err
			}
			m, err := extension.NewManager(iniPath)
			if err != nil {
				return err
			}
			loaded, err := m.ListLoaded(phpBin)
			if err != nil {
				return err
			}
			fmt.Printf("Loaded extensions (%d):\n", len(loaded))
			for _, s := range loaded {
				fmt.Printf("  • %s\n", s.Name)
			}
			return nil
		},
	}
}

func newExtEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <name>",
		Short: "Enable a PHP extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			iniPath, err := activeIniPath()
			if err != nil {
				return err
			}
			m, err := extension.NewManager(iniPath)
			if err != nil {
				return err
			}
			if err := m.Enable(args[0]); err != nil {
				return err
			}
			fmt.Printf("Extension %q enabled in %s\n", args[0], iniPath)
			fmt.Println("Restart your web server / PHP-FPM for the change to take effect.")
			return nil
		},
	}
}

func newExtDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <name>",
		Short: "Disable a PHP extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			iniPath, err := activeIniPath()
			if err != nil {
				return err
			}
			m, err := extension.NewManager(iniPath)
			if err != nil {
				return err
			}
			if err := m.Disable(args[0]); err != nil {
				return err
			}
			fmt.Printf("Extension %q disabled in %s\n", args[0], iniPath)
			return nil
		},
	}
}

func newExtStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <name>",
		Short: "Check status of a PHP extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phpBin, err := activePHPBinary()
			if err != nil {
				return err
			}
			iniPath, err := activeIniPath()
			if err != nil {
				return err
			}
			m, err := extension.NewManager(iniPath)
			if err != nil {
				return err
			}
			s, err := m.GetStatus(phpBin, args[0])
			if err != nil {
				return err
			}
			if s.Enabled {
				fmt.Printf("✓ %s is enabled\n", args[0])
			} else {
				fmt.Printf("✗ %s is NOT enabled\n", args[0])
			}
			return nil
		},
	}
}

// activeIniPath returns the php.ini path for the active PHP version.
func activeIniPath() (string, error) {
	f, err := loadActiveIni()
	if err != nil {
		return "", err
	}
	return f.Path(), nil
}
