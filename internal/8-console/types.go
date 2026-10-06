package console

// UsageOption describes one CLI flag.
type UsageOption struct {
	Flag        string
	Description string
}

// UsageArg labels one token from the command line.
type UsageArg struct {
	Label string
	Value string
}

// UsageHelp is the content shown for help and for a missing path.
type UsageHelp struct {
	Message  string
	Syntax   string
	Options  []UsageOption
	Args     []UsageArg
	Examples []string
	ExitCode int
}

const (
	OperationCheck = "Check imports"
	OperationFix   = "Fix imports"
)

const (
	ProgramName        = "twts"
	ProgramDescription = "Rewrite TypeScript import and export statements in .ts and .tsx files."
)

const (
	ColorRed    = "\x1b[31m"
	ColorGreen  = "\x1b[32m"
	ColorYellow = "\x1b[33m"
	ColorGray   = "\x1b[90m"
	ColorCyan   = "\x1b[36m"
	ColorDim    = "\x1b[2m\x1b[90m"
	ColorReset  = "\x1b[0m"
)
