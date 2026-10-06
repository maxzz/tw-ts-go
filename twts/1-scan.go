package twts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunScan reads .ts and .tsx files under path.
// Check reports files that Transform would change and does not write.
// The default mode rewrites those files in place.
// Recursive walks subfolders. Directories named node_modules, dist, and .git are skipped.
func RunScan(path string, options ScanOptions) (ScanResult, error) {
	if strings.TrimSpace(path) == "" {
		return ScanResult{}, fmt.Errorf("a folder or file is required")
	}

	resolved, err := filepath.Abs(path)
	if err != nil {
		return ScanResult{}, fmt.Errorf("resolve path %q: %w", path, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return ScanResult{}, fmt.Errorf("path %q: %w", path, err)
	}

	root := resolved
	if !info.IsDir() {
		root = filepath.Dir(resolved)
	}

	var files []string
	if info.IsDir() {
		files, err = collectFiles(resolved, options.Recursive)
		if err != nil {
			return ScanResult{}, err
		}
	} else if hasExt(info.Name()) {
		files = []string{resolved}
	}

	result := ScanResult{
		FileCount: len(files),
		RootName:  filepath.Base(root),
	}

	for _, file := range files {
		originalBytes, err := os.ReadFile(file)
		if err != nil {
			return ScanResult{}, fmt.Errorf("read %s: %w", file, err)
		}
		original := string(originalBytes)
		next, err := Transform(original, strings.HasSuffix(file, ".tsx"))
		if err != nil {
			return ScanResult{}, fmt.Errorf("%s: %w", file, err)
		}
		if next == original {
			continue
		}

		rel, relErr := filepath.Rel(root, file)
		if relErr != nil {
			rel = filepath.Base(file)
		}
		result.Hits = append(result.Hits, Hit{
			Abs: file,
			Rel: filepath.ToSlash(rel),
		})
		if options.Check {
			continue
		}

		mode := os.FileMode(0o644)
		if st, statErr := os.Stat(file); statErr == nil {
			mode = st.Mode().Perm()
		}
		if err := os.WriteFile(file, []byte(next), mode); err != nil {
			return ScanResult{}, fmt.Errorf("write %s: %w", file, err)
		}
	}

	return result, nil
}

func collectFiles(root string, recursive bool) ([]string, error) {
	var files []string
	if err := walk(root, recursive, &files); err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func walk(dir string, recursive bool, files *[]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", dir, err)
	}
	for _, entry := range entries {
		name := entry.Name()
		full := filepath.Join(dir, name)
		if entry.IsDir() {
			if _, skip := ignoredDirectories[name]; skip {
				continue
			}
			if !recursive {
				continue
			}
			if err := walk(full, recursive, files); err != nil {
				return err
			}
			continue
		}
		if hasExt(name) {
			*files = append(*files, full)
		}
	}
	return nil
}

func hasExt(name string) bool {
	for _, ext := range defaultExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
