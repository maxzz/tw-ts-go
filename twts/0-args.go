package twts

import (
	"fmt"
	"strings"
)

// ParsedArgs holds CLI arguments. Recursive and Tree start enabled.
type ParsedArgs struct {
	Path      string
	Recursive bool
	Check     bool
	Tree      bool
	Help      bool
}

// ParseArgs parses CLI arguments.
// A folder or file path is required unless --help is set.
// Recursive directory walks and tree output are enabled unless turned off.
func ParseArgs(argv []string) (ParsedArgs, error) {
	args := ParsedArgs{Recursive: true, Tree: true}
	var paths []string

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" {
			paths = append(paths, argv[i+1:]...)
			break
		}
		if arg == "--help" || arg == "-h" {
			args.Help = true
			continue
		}
		if arg == "--check" || arg == "-c" {
			args.Check = true
			continue
		}
		if matched, ok := applySwitch(arg, "recursive", &args.Recursive); matched {
			if !ok {
				return ParsedArgs{}, fmt.Errorf("invalid value for --recursive in %q", arg)
			}
			continue
		}
		if matched, ok := applySwitch(arg, "tree", &args.Tree); matched {
			if !ok {
				return ParsedArgs{}, fmt.Errorf("invalid value for --tree in %q", arg)
			}
			continue
		}
		if arg == "" {
			continue
		}
		if arg[0] == '-' {
			return ParsedArgs{}, fmt.Errorf("unknown option %q", arg)
		}
		paths = append(paths, arg)
	}

	if len(paths) > 1 {
		return ParsedArgs{}, fmt.Errorf("expected one folder or file, got %d paths", len(paths))
	}
	if len(paths) == 1 {
		args.Path = paths[0]
	}
	return args, nil
}

func applySwitch(arg, name string, value *bool) (matched bool, ok bool) {
	if arg == "--"+name {
		*value = true
		return true, true
	}
	if arg == "--no-"+name {
		*value = false
		return true, true
	}
	prefix := "--" + name + "="
	if !strings.HasPrefix(arg, prefix) {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(arg[len(prefix):])) {
	case "true", "1", "on", "yes":
		*value = true
	case "false", "0", "off", "no":
		*value = false
	default:
		return true, false
	}
	return true, true
}
