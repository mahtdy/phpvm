package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newRemoveCmd())
}

func newRemoveCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "remove <version>",
		Aliases: []string{"rm", "uninstall"},
		Short:   "Remove an installed PHP version",
		Long: `Uninstall a PHP version from ~/.phpvm/versions/.

The currently active version cannot be removed.

Examples:
  phpvm remove 8.3
  phpvm remove 8.3.10 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			versionStr := args[0]

			if !yes {
				fmt.Printf("Remove PHP %s? [y/N] ", versionStr)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			remover := php.NewRemover(Cfg)
			iv, err := remover.Remove(versionStr)
			if err != nil {
				return err
			}
			fmt.Printf("PHP %s removed.\n", iv.Version.String())
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return cmd
}
