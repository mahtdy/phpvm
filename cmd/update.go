package cmd

import (
	"context"
	"fmt"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newUpdateCmd())
}

func newUpdateCmd() *cobra.Command {
	var (
		all        bool
		noProgress bool
	)

	cmd := &cobra.Command{
		Use:   "update [version]",
		Short: "Update PHP to the latest patch release",
		Long: `Update an installed PHP version to the latest patch release.

Examples:
  phpvm update           # update the current active version
  phpvm update 8.3       # update PHP 8.3.x to the latest patch
  phpvm update --all     # update all installed versions`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updater := php.NewUpdater(Cfg)

			if all {
				results, err := updater.UpdateAll(context.Background(), noProgress)
				if err != nil {
					return err
				}
				for _, r := range results {
					if r.Updated {
						fmt.Printf("  PHP %s → %s\n", r.OldVersion, r.NewVersion)
					} else {
						fmt.Printf("  PHP %s — already up to date\n", r.OldVersion)
					}
				}
				return nil
			}

			selector := ""
			if len(args) > 0 {
				selector = args[0]
			}

			result, err := updater.Update(context.Background(), selector, noProgress)
			if err != nil {
				return err
			}

			if result.Updated {
				fmt.Printf("PHP updated: %s → %s\n", result.OldVersion, result.NewVersion)
			} else {
				fmt.Printf("PHP %s is already up to date.\n", result.OldVersion)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "update all installed versions")
	cmd.Flags().BoolVar(&noProgress, "no-progress", false, "disable progress bar")
	return cmd
}
