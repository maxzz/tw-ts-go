package twts

import (
	"sort"
	"strings"
)

type treeNode struct {
	name string
	dir  bool
	kids map[string]*treeNode
}

// TreeLines builds an indented tree. rootName is the top folder.
// rels are slash-separated paths of the files to show, relative to that folder.
func TreeLines(rootName string, rels []string) []TreeLine {
	root := &treeNode{name: rootName, dir: rootName != "", kids: map[string]*treeNode{}}
	for _, rel := range rels {
		rel = strings.TrimPrefix(filepathToSlash(rel), "./")
		if rel == "" || rel == "." {
			continue
		}
		insertTree(root, rel)
	}

	var lines []TreeLine
	walkTree(root, 0, &lines)
	return lines
}

func insertTree(root *treeNode, rel string) {
	parts := strings.Split(rel, "/")
	cur := root
	for i, part := range parts {
		if part == "" {
			continue
		}
		child := cur.kids[part]
		if child == nil {
			child = &treeNode{name: part, kids: map[string]*treeNode{}}
			cur.kids[part] = child
		}
		if i < len(parts)-1 {
			child.dir = true
		}
		cur = child
	}
}

func walkTree(n *treeNode, depth int, out *[]TreeLine) {
	if n.name != "" {
		*out = append(*out, TreeLine{Text: n.name, Dir: n.dir || len(n.kids) > 0, Depth: depth})
		depth++
	}
	names := make([]string, 0, len(n.kids))
	for name := range n.kids {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		walkTree(n.kids[name], depth, out)
	}
}

// FlatPaths returns the absolute paths of hits, sorted.
func FlatPaths(hits []Hit) []string {
	paths := make([]string, len(hits))
	for i, hit := range hits {
		paths[i] = hit.Abs
	}
	sort.Strings(paths)
	return paths
}

func filepathToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
