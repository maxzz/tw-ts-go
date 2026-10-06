package twts

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgsDefaults(t *testing.T) {
	args, err := ParseArgs([]string{"src"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Path != "src" || args.Check || !args.Recursive || !args.Tree || args.Help {
		t.Fatalf("%+v", args)
	}
}

func TestParseArgsFlags(t *testing.T) {
	args, err := ParseArgs([]string{"-c", "--no-recursive", "--no-tree", "src"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Path != "src" || !args.Check || args.Recursive || args.Tree {
		t.Fatalf("%+v", args)
	}

	on, err := ParseArgs([]string{"--recursive=false", "--tree=off", "--check", "lib"})
	if err != nil {
		t.Fatal(err)
	}
	if on.Recursive || on.Tree || !on.Check || on.Path != "lib" {
		t.Fatalf("%+v", on)
	}

	last, err := ParseArgs([]string{"--no-tree", "--tree", "src"})
	if err != nil {
		t.Fatal(err)
	}
	if !last.Tree {
		t.Fatal("expected the last tree switch to enable tree output")
	}
}

func TestParseArgsErrors(t *testing.T) {
	if _, err := ParseArgs([]string{"--nope", "src"}); err == nil {
		t.Fatal("expected unknown option error")
	}
	if _, err := ParseArgs([]string{"a", "b"}); err == nil {
		t.Fatal("expected two-path error")
	}
	if _, err := ParseArgs([]string{"--tree=maybe", "src"}); err == nil {
		t.Fatal("expected invalid tree value error")
	}
	args, err := ParseArgs([]string{"--", "-weird"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Path != "-weird" {
		t.Fatalf("path = %q", args.Path)
	}
	help, err := ParseArgs([]string{"--help"})
	if err != nil {
		t.Fatal(err)
	}
	if !help.Help || help.Path != "" {
		t.Fatalf("%+v", help)
	}
}

func TestTreeLines(t *testing.T) {
	lines := TreeLines("src", []string{"b.ts", "a/c.ts", "a.ts"})
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(strings.Repeat("  ", line.Depth))
		b.WriteString(line.Text)
		b.WriteByte('\n')
	}
	want := "src\n  a\n    c.ts\n  a.ts\n  b.ts\n"
	if b.String() != want {
		t.Fatalf("tree =\n%s\nwant\n%s", b.String(), want)
	}
	if !lines[1].Dir || lines[2].Dir {
		t.Fatalf("dir flags = %+v", lines)
	}
}

func TestFlatPathsSorts(t *testing.T) {
	got := FlatPaths([]Hit{{Abs: "b"}, {Abs: "a"}})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("%v", got)
	}
}

func TestRunScanCheckDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.ts")
	original := "import { a } from './a.ts';\n"
	if err := os.WriteFile(file, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(dir, "sub", "b.ts")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte("import type { B } from './b.ts';\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := RunScan(dir, ScanOptions{Recursive: true, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.FileCount != 2 || len(res.Hits) != 2 {
		t.Fatalf("count=%d hits=%d", res.FileCount, len(res.Hits))
	}
	if res.RootName != filepath.Base(dir) {
		t.Fatalf("root = %q", res.RootName)
	}
	for _, hit := range res.Hits {
		if strings.Contains(hit.Rel, `\`) {
			t.Fatalf("rel %q", hit.Rel)
		}
	}
	after, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != original {
		t.Fatalf("check wrote the file: %s", after)
	}
}

func TestRunScanFixAndNonRecursive(t *testing.T) {
	dir := t.TempDir()
	top := filepath.Join(dir, "app.ts")
	nested := filepath.Join(dir, "sub", "b.tsx")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatal(err)
	}
	topSrc := "import { a } from './a.ts';\nexport const kept = \"import type { No } from './no.ts'\";\n"
	nestedSrc := "import { StreamsSelector } from '@/editor/6-streams/index.tsx';\nexport function F(){ return <div className=\"from './nope.ts'\">x</div>; }\n"
	if err := os.WriteFile(top, []byte(topSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte(nestedSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	shallow, err := RunScan(dir, ScanOptions{Recursive: false, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(shallow.Hits) != 1 || shallow.Hits[0].Rel != "app.ts" {
		t.Fatalf("shallow hits = %+v", shallow.Hits)
	}

	fixed, err := RunScan(dir, ScanOptions{Recursive: true, Check: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(fixed.Hits) != 2 {
		t.Fatalf("fixed hits = %+v", fixed.Hits)
	}
	gotTop, err := os.ReadFile(top)
	if err != nil {
		t.Fatal(err)
	}
	wantTop := "import { a } from \"./a\";\nexport const kept = \"import type { No } from './no.ts'\";\n"
	if string(gotTop) != wantTop {
		t.Fatalf("top =\n%s", gotTop)
	}
	gotNested, err := os.ReadFile(nested)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(gotNested, []byte(`from "@/editor/6-streams"`)) || !bytes.Contains(gotNested, []byte(`from './nope.ts'`)) {
		t.Fatalf("nested =\n%s", gotNested)
	}

	again, err := RunScan(dir, ScanOptions{Recursive: true, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Hits) != 0 {
		t.Fatalf("still failing: %+v", again.Hits)
	}
}

func TestRunScanSkipsVendorDirs(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bad := "import { a } from './a.ts';\n"
	write("app.ts", bad)
	write("node_modules/pkg/index.ts", bad)
	write("dist/out.ts", bad)
	write(".git/x.ts", bad)

	res, err := RunScan(dir, ScanOptions{Recursive: true, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.FileCount != 1 || len(res.Hits) != 1 || res.Hits[0].Rel != "app.ts" {
		t.Fatalf("%+v", res)
	}
}

func TestRunScanSingleFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "only.ts")
	if err := os.WriteFile(file, []byte("export type { Y } from './e.ts';\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := RunScan(file, ScanOptions{Recursive: false, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Hits) != 1 || res.Hits[0].Rel != "only.ts" {
		t.Fatalf("%+v", res.Hits)
	}
}

func TestTestdataSamples(t *testing.T) {
	root := filepath.Join("..", "testdata")
	before := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		before[path] = body
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := RunScan(root, ScanOptions{Recursive: true, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.RootName != "testdata" || res.FileCount != 4 || len(res.Hits) != 3 {
		t.Fatalf("root=%s files=%d hits=%+v", res.RootName, res.FileCount, res.Hits)
	}
	got := map[string]bool{}
	for _, hit := range res.Hits {
		got[hit.Rel] = true
	}
	for _, rel := range []string{"sample.ts", "nested/child.tsx", "nested/deeper/leaf.ts"} {
		if !got[rel] {
			t.Fatalf("missing %s in %+v", rel, res.Hits)
		}
	}
	if got["ok.ts"] {
		t.Fatal("ok.ts was reported")
	}

	shallow, err := RunScan(root, ScanOptions{Recursive: false, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(shallow.Hits) != 1 || shallow.Hits[0].Rel != "sample.ts" {
		t.Fatalf("shallow = %+v", shallow.Hits)
	}

	for path, body := range before {
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(body, after) {
			t.Fatalf("check modified %s", path)
		}
	}
}
