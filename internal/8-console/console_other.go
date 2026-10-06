//go:build !windows

package console

import (
	"fmt"
	"os"
)

// PrintError writes err to stderr in red and exits.
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ColorRed, err, ColorReset)
	os.Exit(1)
}

// WaitAndExit exits immediately without waiting for a key.
func WaitAndExit(code int) {
	os.Exit(code)
}

func finishUsage(code int) {
	os.Exit(code)
}
