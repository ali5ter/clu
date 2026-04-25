// clu — explore the commandlineuser.com CLI catalog from your terminal.
//
// Launches a TUI when stdout is a terminal; emits JSON when piped or --json is set.
package main

import "github.com/ali5ter/clu/cmd"

func main() {
	cmd.Execute()
}
