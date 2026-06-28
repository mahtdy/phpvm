package cmd

import (
	"context"
	"fmt"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newInstallCmd())
}

func newInstallCmd() *cobra.Command {
	var (
		ts         bool
		nts        bool
		arch       string
		force      bool
		noProgress bool
	)

	cmd := &cobra.Command{
		Use:   "install <version>",
		Short: "Download and install a PHP version",
		Long: `Download and install a PHP version from php.net.

Examples:
  phpvm install 8.4          # install latest 8.4.x
  phpvm install 8.4.1        # install exact version
  phpvm install 8.4.1 --ts   # Thread Safe build
  phpvm install latest       # install latest stable`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			variant := ""
			if ts {
				variant = "ts"
			} else if nts {
				variant = "nts"
			}

			inst := php.NewInstaller(Cfg)
			iv, err := inst.Install(context.Background(), php.InstallOptions{
				VersionStr: args[0],
				Variant:    variant,
				Arch:       arch,
				Force:      force,
				NoProgress: noProgress,
			})
			if err != nil {
				return err
			}
			fmt.Printf("\nPHP %s installed.\nRun: phpvm use %s\n",
				iv.Version.String(), iv.Version.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&ts, "ts", false, "Thread Safe build")
	cmd.Flags().BoolVar(&nts, "nts", false, "Non-Thread Safe build (default)")
	cmd.Flags().StringVar(&arch, "arch", "", "CPU architecture (x64, arm64)")
	cmd.Flags().BoolVar(&force, "force", false, "reinstall even if already installed")
	cmd.Flags().BoolVar(&noProgress, "no-progress", false, "disable progress bar")

	return cmd
}
