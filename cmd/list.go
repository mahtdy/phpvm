package cmd

import (
	"fmt"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newListCmd())
}

func newListCmd() *cobra.Command {
	var plain bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed PHP versions",
		Long:    "Show all PHP versions installed under ~/.phpvm/versions/.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			scanner := php.NewScanner(Cfg.VersionsDir())
			versions, err := scanner.ScanInstalled()
			if err != nil {
				return err
			}

			if len(versions) == 0 {
				fmt.Println("No PHP versions installed.")
				fmt.Println("Run: phpvm install <version>")
				return nil
			}

			fmt.Println("Installed PHP versions:")
			for _, iv := range versions {
				if plain {
					fmt.Println(iv.Version.String())
					continue
				}
				marker := "  "
				if iv.Version.String() == Cfg.Current {
					marker = "→ "
				}
				fmt.Printf("  %s%s\n", marker, iv.Version.String())
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&plain, "plain", false, "print only version strings, one per line")
	return cmd
}
