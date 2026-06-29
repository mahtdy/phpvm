package cmd

import (
	"fmt"

	"github.com/mahtdy/phpvm/internal/extension"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newXdebugCmd())
}

func newXdebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xdebug",
		Short: "Manage Xdebug",
		Long:  "Enable, disable, and configure Xdebug for the active PHP version.",
	}
	cmd.AddCommand(
		newXdebugEnableCmd(),
		newXdebugDisableCmd(),
		newXdebugStatusCmd(),
		newXdebugSwitchCmd(),
	)
	return cmd
}

func newXdebugEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable Xdebug (mode=develop)",
		RunE: func(cmd *cobra.Command, args []string) error {
			xm, err := loadXdebugManager()
			if err != nil {
				return err
			}
			return xm.Enable()
		},
	}
}

func newXdebugDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable Xdebug (mode=off)",
		RunE: func(cmd *cobra.Command, args []string) error {
			xm, err := loadXdebugManager()
			if err != nil {
				return err
			}
			return xm.Disable()
		},
	}
}

func newXdebugStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Xdebug status",
		RunE: func(cmd *cobra.Command, args []string) error {
			xm, err := loadXdebugManager()
			if err != nil {
				return err
			}
			if !xm.IsInstalled() {
				fmt.Println("Xdebug is NOT installed.")
				fmt.Println("Run: phpvm xdebug install")
				return nil
			}
			mode := xm.CurrentMode()
			fmt.Printf("Xdebug is installed (mode=%s)\n", mode)
			return nil
		},
	}
}

func newXdebugSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <mode>",
		Short: "Switch Xdebug mode (develop|coverage|profile|trace|debug|off)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := args[0]
			if !extension.IsValidMode(mode) {
				return fmt.Errorf("invalid mode %q — valid modes: develop, coverage, profile, trace, debug, off", mode)
			}
			xm, err := loadXdebugManager()
			if err != nil {
				return err
			}
			return xm.SwitchMode(extension.XdebugMode(mode))
		},
	}
}

func loadXdebugManager() (*extension.XdebugManager, error) {
	iniPath, err := activeIniPath()
	if err != nil {
		return nil, err
	}
	return extension.NewXdebugManager(iniPath)
}
