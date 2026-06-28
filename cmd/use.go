package cmd

import (
	"fmt"

	"github.com/mahtdy/phpvm/internal/path"
	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newUseCmd())
}

func newUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <version>",
		Short: "Switch to an installed PHP version",
		Long: `Switch the active PHP version.

Examples:
  phpvm use 8.4        # switch to highest installed 8.4.x
  phpvm use 8.4.1      # switch to exactly 8.4.1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switcher := php.NewSwitcher(Cfg)
			iv, err := switcher.Switch(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Switched to PHP %s\n", iv.Version.String())
			fmt.Println(path.RestartHint())
			return nil
		},
	}
}
