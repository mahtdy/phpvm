package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mahtdy/phpvm/internal/composer"
	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newComposerCmd())
}

func newComposerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "composer <command>",
		Short: "Manage Composer",
		Long:  "Install, update, or diagnose Composer for the active PHP version.",
	}

	cmd.AddCommand(
		newComposerInstallCmd(),
		newComposerUpdateCmd(),
		newComposerDiagnoseCmd(),
	)
	return cmd
}

func newComposerInstallCmd() *cobra.Command {
	var skipVerify bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Download and install Composer",
		RunE: func(cmd *cobra.Command, args []string) error {
			phpBin, err := activePHPBinary()
			if err != nil {
				return err
			}

			inst := composer.NewInstaller()
			info, err := inst.Install(context.Background(), composer.InstallOptions{
				DestDir:             Cfg.VersionsDir(),
				PHPBinary:           phpBin,
				SkipSignatureVerify: skipVerify,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Composer %s installed at %s\n", info.Version, info.Path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "skip SHA-384 signature verification")
	return cmd
}

func newComposerUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update Composer to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			phpBin, err := activePHPBinary()
			if err != nil {
				return err
			}

			u := composer.NewUpdater()
			info, err := u.Update(context.Background(), Cfg.VersionsDir(), phpBin)
			if err != nil {
				return err
			}
			fmt.Printf("Composer %s\n", info.Version)
			return nil
		},
	}
}

func newComposerDiagnoseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diagnose",
		Short: "Check Composer and PHP environment health",
		RunE: func(cmd *cobra.Command, args []string) error {
			phpBin, _ := activePHPBinary()

			detector := composer.NewDetector()
			info, _ := detector.Detect(Cfg.VersionsDir())

			report := composer.Diagnose(phpBin, info)

			fmt.Println("Composer Diagnostics")
			fmt.Println("────────────────────")
			for _, c := range report.Checks {
				icon := "✓"
				if !c.OK && c.Warning {
					icon = "⚠"
				} else if !c.OK {
					icon = "✗"
				}
				fmt.Printf("  %s  %s\n", icon, c.Detail)
			}

			issues := report.Issues()
			fmt.Println()
			if len(issues) == 0 {
				fmt.Println("No issues found.")
			} else {
				fmt.Printf("%d issue(s) found.\n", len(issues))
				os.Exit(1)
			}
			return nil
		},
	}
}

// activePHPBinary returns the path to the current active PHP binary.
func activePHPBinary() (string, error) {
	detector := php.NewCurrentDetector(Cfg.VersionsDir(), Cfg.Current)
	v, binPath, err := detector.Detect()
	if err != nil {
		return "", err
	}
	_ = v
	return binPath, nil
}
