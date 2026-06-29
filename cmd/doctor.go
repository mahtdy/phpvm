package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newDoctorCmd())
}

func newDoctorCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check PHP environment health",
		Long: `Run a comprehensive environment health check.

Verifies PHP installation, PATH, php.ini, Composer,
required extensions, and Xdebug configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc := php.NewDoctor(Cfg)
			report := doc.Run()

			if jsonOutput {
				return printDoctorJSON(report)
			}
			return printDoctorText(report)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output report as JSON")
	return cmd
}

func printDoctorText(report *php.DoctorReport) error {
	fmt.Println("phpvm Doctor — Environment Check")
	fmt.Println("─────────────────────────────────")
	for _, c := range report.Checks {
		icon := "✓"
		switch c.Status {
		case php.CheckWarn:
			icon = "⚠"
		case php.CheckFail:
			icon = "✗"
		}
		fmt.Printf("%s  %s\n", icon, c.Detail)
	}
	fmt.Println()

	issues := report.Issues()
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}
	fmt.Printf("Issues found: %d\n", len(issues))
	os.Exit(1)
	return nil
}

// jsonCheck is the JSON-serialisable form of DoctorCheck.
type jsonCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

func printDoctorJSON(report *php.DoctorReport) error {
	checks := make([]jsonCheck, len(report.Checks))
	for i, c := range report.Checks {
		status := "ok"
		switch c.Status {
		case php.CheckWarn:
			status = "warn"
		case php.CheckFail:
			status = "fail"
		}
		checks[i] = jsonCheck{Name: c.Name, Status: status, Detail: c.Detail}
	}
	out, err := json.MarshalIndent(checks, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
