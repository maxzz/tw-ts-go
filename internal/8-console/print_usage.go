package console

import (
	"fmt"
	"strings"
)

const usageWrapWidth = 80

// PrintUsage shows colored help, then exits with help.ExitCode.
func PrintUsage(help UsageHelp) {
	fmt.Printf("%sUsage:%s\n", ColorGray, ColorReset)
	printWrappedLines(ColorGray, "  ", help.Syntax, usageWrapWidth)
	fmt.Println()

	if len(help.Options) > 0 {
		fmt.Printf("%sOptions:%s\n", ColorGray, ColorReset)
		maxFlagLen := 0
		for _, opt := range help.Options {
			if len(opt.Flag) > maxFlagLen {
				maxFlagLen = len(opt.Flag)
			}
		}
		optionIndent := 2 + maxFlagLen + 1
		for _, opt := range help.Options {
			padding := strings.Repeat(" ", maxFlagLen-len(opt.Flag))
			prefix := "  " + opt.Flag + padding + " "
			printWrappedLines(ColorGray, prefix, opt.Description, usageWrapWidth, strings.Repeat(" ", optionIndent))
		}
		fmt.Println()
	}

	if len(help.Args) > 0 {
		fmt.Printf("%sArguments:%s\n", ColorGray, ColorReset)
		for _, arg := range help.Args {
			fmt.Printf("  %s%s:%s %s\n", ColorGray, arg.Label, ColorReset, arg.Value)
		}
		fmt.Println()
	}

	if len(help.Examples) > 0 {
		fmt.Printf("%sExamples:%s\n", ColorGray, ColorReset)
		for _, example := range help.Examples {
			fmt.Printf("  %s%s%s\n", ColorGray, example, ColorReset)
		}
		fmt.Println()
	}

	printWrappedLines(ColorYellow, "", help.Message, usageWrapWidth)
	fmt.Println()
	finishUsage(help.ExitCode)
}

func printWrappedLines(color, firstIndent, text string, width int, continuationIndent ...string) {
	lines := wrapText(text, width-len(firstIndent))
	if len(lines) == 0 {
		return
	}

	contIndent := firstIndent
	if len(continuationIndent) > 0 {
		contIndent = continuationIndent[0]
	}

	fmt.Printf("%s%s%s%s\n", firstIndent, color, lines[0], ColorReset)
	for _, line := range lines[1:] {
		fmt.Printf("%s%s%s%s\n", contIndent, color, line, ColorReset)
	}
}

func wrapText(text string, width int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if width < 1 {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var line strings.Builder

	flush := func() {
		if line.Len() > 0 {
			lines = append(lines, line.String())
			line.Reset()
		}
	}

	for _, word := range words {
		if line.Len() == 0 {
			if len(word) <= width {
				line.WriteString(word)
			} else {
				lines = append(lines, word)
			}
			continue
		}

		if line.Len()+1+len(word) <= width {
			line.WriteByte(' ')
			line.WriteString(word)
			continue
		}

		flush()
		if len(word) <= width {
			line.WriteString(word)
		} else {
			lines = append(lines, word)
		}
	}

	flush()
	return lines
}
