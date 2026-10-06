package main

import (
	"fmt"
	"os"
	"path/filepath"

	console "twts/internal/8-console"
	"twts/twts"
)

func main() {
	args, err := twts.ParseArgs(os.Args[1:])
	if err != nil {
		console.PrintError(err)
	}
	if args.Help {
		printUsage(0, "Rewrite TypeScript import and export statements in .ts and .tsx files.")
	}
	if args.Path == "" {
		printUsage(1, "A folder or file path is required.")
	}

	console.PrintVersion(twts.Version)

	path, err := resolvePath(args.Path)
	if err != nil {
		console.PrintError(err)
	}

	operation := console.OperationFix
	if args.Check {
		operation = console.OperationCheck
	}

	result, err := twts.RunScan(path, twts.ScanOptions{
		Recursive: args.Recursive,
		Check:     args.Check,
	})
	if err != nil {
		console.PrintError(err)
	}
	console.PrintReport(result, operation, args.Tree)

	exitCode := 0
	if args.Check && len(result.Hits) > 0 {
		exitCode = 1
	}
	console.WaitAndExit(exitCode)
}

func printUsage(code int, message string) {
	console.PrintUsage(console.UsageHelp{
		Message: message,
		Syntax:  "twts [options] <path>",
		Options: []console.UsageOption{
			{
				Flag:        "--check, -c",
				Description: "Report files that break the import rules and do not rewrite them (default: off)",
			},
			{
				Flag:        "--recursive",
				Description: "Walk subfolders (default: on)",
			},
			{
				Flag:        "--no-recursive",
				Description: "Scan only the selected folder",
			},
			{
				Flag:        "--tree",
				Description: "Print matching files as an indented folder tree (default: on)",
			},
			{
				Flag:        "--no-tree",
				Description: "Print each matching file as a full path",
			},
			{
				Flag:        "--help, -h",
				Description: "Show this help message (default: off)",
			},
		},
		Args: []console.UsageArg{
			{
				Label: "path",
				Value: "Folder or file to process (required)",
			},
		},
		Examples: []string{
			"twts src",
			"twts -c src",
			"twts --no-recursive src",
			"twts -c --no-tree src/App.tsx",
		},
		ExitCode: code,
	})
}

func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("path %q: %w", path, err)
	}
	return abs, nil
}
