package twts

// ScanOptions configures a file or directory run.
type ScanOptions struct {
	Recursive bool
	Check     bool
}

// Hit is one file whose import or export statements do not match the rules.
// In check mode the file is left unchanged. In the default mode it has been rewritten.
type Hit struct {
	Abs string
	Rel string
}

// ScanResult is returned by RunScan.
type ScanResult struct {
	FileCount int
	RootName  string
	Hits      []Hit
}

// TreeLine is one row of an indented folder tree.
type TreeLine struct {
	Text  string
	Dir   bool
	Depth int
}
