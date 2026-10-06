package twts

var defaultExtensions = []string{".ts", ".tsx"}

var ignoredDirectories = map[string]struct{}{
	"node_modules": {},
	"dist":         {},
	".git":         {},
}
