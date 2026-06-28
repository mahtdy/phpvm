package cmd

import (
	"fmt"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCurrentCmd())
}

func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the active PHP version",
		Long:  "Display the currently active PHP version by checking PATH and config.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			detector := php.NewCurrentDetector(Cfg.VersionsDir(), Cfg.Current)
			v, _, err := detector.Detect()
			if err != nil {
				return err
			}
			fmt.Println(v.String())
			return nil
		},
	}
}
