// phpvm — PHP Version Manager
// A cross-platform CLI tool to install, switch, and manage PHP versions.
//
// Usage:
//
//	phpvm [command] [flags]
//
// Run 'phpvm --help' for a list of available commands.
package main

import (
	"os"

	"github.com/mahtdy/phpvm/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
