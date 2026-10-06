package console

import (
	"bytes"
	"strings"
	"testing"

	"twts/twts"
)

func sampleResult() twts.ScanResult {
	return twts.ScanResult{
		FileCount: 4,
		RootName:  "src",
		Hits: []twts.Hit{
			{Abs: `C:\proj\src\nested\child.tsx`, Rel: "nested/child.tsx"},
			{Abs: `C:\proj\src\sample.ts`, Rel: "sample.ts"},
		},
	}
}

func TestWriteReportTreeColors(t *testing.T) {
	var buf bytes.Buffer
	WriteReport(&buf, sampleResult(), OperationCheck, true)
	got := buf.String()

	for _, piece := range []string{ColorCyan, ColorYellow, ColorGray, ColorReset, "Check imports", "Scanned", "2 files do not match import rules."} {
		if !strings.Contains(got, piece) {
			t.Fatalf("missing %q in\n%s", piece, got)
		}
	}
	if !strings.Contains(got, "  "+ColorCyan+"nested"+ColorReset) {
		t.Fatalf("missing nested folder indent:\n%s", got)
	}
	if !strings.Contains(got, "    "+ColorYellow+"child.tsx"+ColorReset) {
		t.Fatalf("missing child file indent:\n%s", got)
	}
	if !strings.Contains(got, "  "+ColorYellow+"sample.ts"+ColorReset) {
		t.Fatalf("missing sample file indent:\n%s", got)
	}
	if strings.Contains(got, `C:\proj\src\sample.ts`) {
		t.Fatal("tree printed a full path")
	}
}

func TestWriteReportFlat(t *testing.T) {
	var buf bytes.Buffer
	WriteReport(&buf, sampleResult(), OperationCheck, false)
	got := buf.String()
	if !strings.Contains(got, ColorYellow+`C:\proj\src\nested\child.tsx`+ColorReset) {
		t.Fatalf("missing full path:\n%s", got)
	}
	if !strings.Contains(got, ColorYellow+`C:\proj\src\sample.ts`+ColorReset) {
		t.Fatalf("missing full path:\n%s", got)
	}
	if strings.Contains(got, "\n"+ColorCyan+"src"+ColorReset) || strings.Contains(got, "  "+ColorCyan+"nested") {
		t.Fatal("flat list printed a tree")
	}
	child := strings.Index(got, "child.tsx")
	sample := strings.Index(got, "sample.ts")
	if child < 0 || sample < 0 || child > sample {
		t.Fatalf("flat order child=%d sample=%d\n%s", child, sample, got)
	}
}

func TestWriteReportSuccessAndEmpty(t *testing.T) {
	var ok bytes.Buffer
	WriteReport(&ok, twts.ScanResult{FileCount: 2, RootName: "src"}, OperationCheck, true)
	if !strings.Contains(ok.String(), ColorGreen+"All import and export statements match."+ColorReset) {
		t.Fatalf("%s", ok.String())
	}

	var none bytes.Buffer
	WriteReport(&none, twts.ScanResult{}, OperationFix, true)
	if !strings.Contains(none.String(), ColorYellow+"No TypeScript files found."+ColorReset) {
		t.Fatalf("%s", none.String())
	}

	var fixed bytes.Buffer
	WriteReport(&fixed, sampleResult(), OperationFix, true)
	got := fixed.String()
	if !strings.Contains(got, "  "+ColorGreen+"sample.ts"+ColorReset) {
		t.Fatalf("updated file should be green:\n%s", got)
	}
	if !strings.Contains(got, "Updated "+ColorGreen+"2"+ColorReset) {
		t.Fatalf("missing update count:\n%s", got)
	}
}
