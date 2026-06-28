package cmd

import (
	"fmt"

	"github.com/mahtdy/phpvm/pkg/version"
	"github.com/spf13/cobra"
)

// newVersionCmd returns the "phpvm version" command.
func newVersionCmd() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print phpvm version information",
		Long:  "Display the phpvm binary version, git commit, and build date.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			info := version.Get()
			if short {
				fmt.Println(info.Version)
				return
			}
			fmt.Println(info.String())
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "print only the version number")
	return cmd
}
