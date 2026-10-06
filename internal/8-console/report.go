package console

import (
	"fmt"
	"io"
	"os"
	"strings"

	"twts/twts"
)

// PrintVersion writes the program name, description, and version at startup.
func PrintVersion(version string) {
	fmt.Printf("%s — %s", ProgramName, ProgramDescription)
	fmt.Printf(" %s(version %s)%s\n\n", ColorGray, version, ColorReset)
}

// PrintReport writes a colored check or fix summary.
// tree prints folders and files with indentation. Otherwise each hit is a full path.
func PrintReport(result twts.ScanResult, operation string, tree bool) {
	WriteReport(os.Stdout, result, operation, tree)
}

// WriteReport writes the same report to w.
func WriteReport(w io.Writer, result twts.ScanResult, operation string, tree bool) {
	fmt.Fprintf(w, "%s%s%s\n\n", ColorCyan, operation, ColorReset)

	if result.FileCount == 0 {
		fmt.Fprintf(w, "%sNo TypeScript files found.%s\n", ColorYellow, ColorReset)
		return
	}

	fmt.Fprintf(w, "Scanned %s%d%s file%s\n", ColorGray, result.FileCount, ColorReset, pluralS(result.FileCount))

	if len(result.Hits) == 0 {
		fmt.Fprintf(w, "%sAll import and export statements match.%s\n", ColorGreen, ColorReset)
		return
	}

	fmt.Fprintln(w)
	if tree {
		writeTree(w, result, operation)
	} else {
		writeFlat(w, result, operation)
	}
	fmt.Fprintln(w)

	n := len(result.Hits)
	if operation == OperationCheck {
		verb := "files do"
		if n == 1 {
			verb = "file does"
		}
		fmt.Fprintf(w, "%s%d %s not match import rules.%s\n", ColorYellow, n, verb, ColorReset)
		return
	}

	fmt.Fprintf(
		w,
		"Updated %s%d%s of %s%d%s file%s\n",
		ColorGreen,
		n,
		ColorReset,
		ColorGray,
		result.FileCount,
		ColorReset,
		pluralS(result.FileCount),
	)
}

func writeTree(w io.Writer, result twts.ScanResult, operation string) {
	rels := make([]string, len(result.Hits))
	for i, hit := range result.Hits {
		rels[i] = hit.Rel
	}
	fileColor := ColorYellow
	if operation != OperationCheck {
		fileColor = ColorGreen
	}
	for _, line := range twts.TreeLines(result.RootName, rels) {
		color := fileColor
		if line.Dir {
			color = ColorCyan
		}
		fmt.Fprintf(w, "%s%s%s%s\n", strings.Repeat("  ", line.Depth), color, line.Text, ColorReset)
	}
}

func writeFlat(w io.Writer, result twts.ScanResult, operation string) {
	color := ColorYellow
	if operation != OperationCheck {
		color = ColorGreen
	}
	for _, path := range twts.FlatPaths(result.Hits) {
		fmt.Fprintf(w, "%s%s%s\n", color, path, ColorReset)
	}
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
