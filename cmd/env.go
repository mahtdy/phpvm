package cmd

import (
	"fmt"
	"sort"

	phpini "github.com/mahtdy/phpvm/internal/ini"
	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newEnvCmd())
}

func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage php.ini settings",
		Long:  "Read and write php.ini configuration for the active PHP version.",
	}
	cmd.AddCommand(newEnvGetCmd(), newEnvSetCmd(), newEnvListCmd())
	return cmd
}

func newEnvGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a php.ini value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := loadActiveIni()
			if err != nil {
				return err
			}
			val, ok := f.Get(args[0])
			if !ok {
				return fmt.Errorf("key %q not found in php.ini", args[0])
			}
			fmt.Println(val)
			return nil
		},
	}
}

func newEnvSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a php.ini value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := loadActiveIni()
			if err != nil {
				return err
			}
			_, err = f.Backup()
			if err != nil {
				return err
			}
			f.Set(args[0], args[1])
			if err := f.Save(); err != nil {
				return err
			}
			fmt.Printf("Set %s = %s in %s\n", args[0], args[1], f.Path())
			return nil
		},
	}
}

func newEnvListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all php.ini settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := loadActiveIni()
			if err != nil {
				return err
			}
			all := f.All()
			keys := make([]string, 0, len(all))
			for k := range all {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%-40s = %s\n", k, all[k])
			}
			return nil
		},
	}
}

// loadActiveIni returns a parsed php.ini for the currently active PHP version.
func loadActiveIni() (*phpini.File, error) {
	phpBin, err := activePHPBinary()
	if err != nil {
		return nil, err
	}
	locator := phpini.NewLocator()
	iniPath, err := locator.Find(phpBin)
	if err != nil {
		// Fallback: look next to the binary.
		iniPath, err = php.FindIniNextToBinary(phpBin)
		if err != nil {
			return nil, err
		}
	}
	return phpini.Load(iniPath)
}
