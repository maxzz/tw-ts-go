package twts

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Token kinds match the TypeScript scanner values used by the import rewriter.
const (
	tokEOF        = 1
	tokString     = 10
	tokOpenBrace  = 18
	tokCloseBrace = 19
	tokOpenParen  = 20
	tokComma      = 27
	tokAsterisk   = 41
	tokIdent      = 79
	tokExport     = 94
	tokImport     = 101
	tokAs         = 129
	tokFrom       = 161
	tokOther      = 0
)

var headerRe = regexp.MustCompile(`(?m)^[ \t]*(import|export)\b`)
var pathSepRe = regexp.MustCompile(`[/\\]`)

// Transform rewrites import statements and re-exports.
// A statement is rewritten when its line starts with import or export.
// Module paths use double quotes, a trailing .ts or .tsx is removed, and a
// trailing /index, /index.ts, or /index.tsx imports the folder.
// .css, .js, .d.ts, .test.ts, and .test.tsx keep their suffix.
// import type { A, B } becomes import { type A, type B }.
// import type Name, import type * as Name, and export type * as Name stay type-only.
// An export that is not a re-export is left as written, including export function,
// export const, await import(), and import() calls that do not start the line.
// jsx selects the caller's language variant. Rewrites start at the statement
// keyword, so both variants use this scanner.
func Transform(text string, jsx bool) (string, error) {
	_ = jsx
	p := &parser{text: text, sc: &byteScanner{text: text}}
	for _, loc := range headerRe.FindAllStringSubmatchIndex(text, -1) {
		if p.err != nil {
			return "", p.err
		}
		p.jump(loc[2])
		tok := p.read()
		if tok.kind == tokImport {
			p.parseImport()
		} else if tok.kind == tokExport {
			p.parseExport()
		}
	}
	if p.err != nil {
		return "", p.err
	}
	return applyEdits(text, p.edits)
}

type edit struct {
	start int
	end   int
	text  string
}

type token struct {
	kind  int
	text  string
	start int
	end   int
}

type parser struct {
	text   string
	sc     *byteScanner
	primed bool
	look   token
	edits  []edit
	steps  int
	err    error
}

func (p *parser) jump(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(p.text) {
		pos = len(p.text)
	}
	p.sc.pos = pos
	p.primed = false
	p.steps = 0
}

func (p *parser) read() token {
	if p.err != nil {
		return token{kind: tokEOF, start: len(p.text), end: len(p.text)}
	}
	if !p.primed {
		p.look = p.sc.scan()
		p.primed = true
	}
	tok := p.look
	p.steps++
	if p.steps > len(p.text)+1000 {
		p.err = fmt.Errorf("scanner stalled at %d kind %d text %q", tok.start, tok.kind, tok.text)
		return token{kind: tokEOF, start: tok.start, end: tok.end}
	}
	next := p.sc.scan()
	if next.kind != tokEOF && next.start < tok.end {
		if tok.end > len(p.text) {
			p.sc.pos = len(p.text)
		} else {
			p.sc.pos = tok.end
		}
		next = p.sc.scan()
	}
	p.look = next
	return tok
}

func (p *parser) peekKind() int {
	return p.look.kind
}

func (p *parser) fixSpecifier(tok token) {
	if tok.kind != tokString {
		return
	}
	raw := p.text[tok.start:tok.end]
	next := fixModuleSpecifier(raw)
	if next != raw {
		p.edits = append(p.edits, edit{start: tok.start, end: tok.end, text: next})
	}
}

func (p *parser) removeTypeKeyword(typeTok token) {
	var before byte
	if typeTok.start > 0 {
		before = p.text[typeTok.start-1]
	}
	var after byte
	if typeTok.end < len(p.text) {
		after = p.text[typeTok.end]
	}
	if before == ' ' && (after == ' ' || after == '\n' || after == '\r') {
		p.edits = append(p.edits, edit{start: typeTok.start - 1, end: typeTok.end, text: ""})
		return
	}
	p.edits = append(p.edits, edit{start: typeTok.start, end: typeTok.end, text: ""})
}

func (p *parser) parseNamed(prefix bool) {
	tok := p.read()
	for tok.kind != tokCloseBrace && tok.kind != tokEOF {
		if tok.kind == tokComma {
			tok = p.read()
			continue
		}
		nextKind := p.peekKind()
		isTypeModifier := tok.text == "type" && nextKind != tokAs && nextKind != tokComma && nextKind != tokCloseBrace && nextKind != tokEOF
		if !isTypeModifier && prefix {
			p.edits = append(p.edits, edit{start: tok.start, end: tok.start, text: "type "})
		}
		if isTypeModifier {
			tok = p.read()
		}
		if p.peekKind() == tokAs {
			p.read()
			p.read()
		}
		tok = p.read()
	}
}

func (p *parser) parseImport() {
	tok := p.read()

	if tok.kind == tokOpenParen {
		maybe := p.read()
		if maybe.kind == tokString {
			p.fixSpecifier(maybe)
		}
		p.read()
		return
	}

	if tok.text == "." {
		p.read()
		return
	}

	if tok.kind == tokString {
		p.fixSpecifier(tok)
		p.read()
		return
	}

	prefixNamed := false

	if tok.text == "type" {
		next := p.peekKind()
		if next == tokOpenBrace {
			p.removeTypeKeyword(tok)
			prefixNamed = true
			tok = p.read()
		} else if next == tokAsterisk {
			tok = p.read()
		} else if next != tokFrom && next != tokComma {
			tok = p.read()
			if p.peekKind() == tokComma {
				p.read()
				tok = p.read()
				if tok.kind == tokOpenBrace {
					prefixNamed = true
				}
			}
		}
	}

	if tok.kind == tokOpenBrace {
		p.parseNamed(prefixNamed)
		tok = p.read()
	} else if tok.kind == tokAsterisk {
		if p.peekKind() == tokAs {
			p.read()
			p.read()
		}
		tok = p.read()
	} else if tok.kind != tokFrom {
		if p.peekKind() == tokComma {
			p.read()
			tok = p.read()
			if tok.kind == tokOpenBrace {
				p.parseNamed(prefixNamed)
				tok = p.read()
			} else if tok.kind == tokAsterisk {
				if p.peekKind() == tokAs {
					p.read()
					p.read()
				}
				tok = p.read()
			}
		} else {
			tok = p.read()
		}
	}

	if tok.kind == tokFrom {
		spec := p.read()
		if spec.kind == tokString {
			p.fixSpecifier(spec)
		}
		p.read()
	}
}

func (p *parser) parseExport() {
	tok := p.read()
	prefixNamed := false
	if tok.text == "type" {
		after := p.peekKind()
		if after == tokOpenBrace {
			p.removeTypeKeyword(tok)
			prefixNamed = true
			tok = p.read()
		} else if after == tokAsterisk {
			tok = p.read()
		} else {
			return
		}
	}
	if tok.kind != tokOpenBrace && tok.kind != tokAsterisk {
		return
	}

	if tok.kind == tokAsterisk {
		if p.peekKind() == tokAs {
			p.read()
			p.read()
		}
		tok = p.read()
	} else {
		p.parseNamed(prefixNamed)
		tok = p.read()
	}
	if tok.kind == tokFrom {
		spec := p.read()
		if spec.kind == tokString {
			p.fixSpecifier(spec)
		}
	}
}

type byteScanner struct {
	text string
	pos  int
}

func (s *byteScanner) scan() token {
	s.skipTrivia()
	if s.pos >= len(s.text) {
		return token{kind: tokEOF, start: len(s.text), end: len(s.text)}
	}
	start := s.pos
	ch := s.text[start]
	switch ch {
	case '{':
		return s.simple(tokOpenBrace)
	case '}':
		return s.simple(tokCloseBrace)
	case '(':
		return s.simple(tokOpenParen)
	case ',':
		return s.simple(tokComma)
	case '*':
		return s.simple(tokAsterisk)
	case '\'', '"':
		return s.scanString(ch)
	case '`':
		return s.scanTemplate()
	default:
		r, size := utf8.DecodeRuneInString(s.text[s.pos:])
		if isIdentStart(r) {
			return s.scanIdent()
		}
		if size <= 0 {
			size = 1
		}
		s.pos += size
		return token{kind: tokOther, text: s.text[start:s.pos], start: start, end: s.pos}
	}
}

func (s *byteScanner) simple(kind int) token {
	start := s.pos
	s.pos++
	return token{kind: kind, text: s.text[start:s.pos], start: start, end: s.pos}
}

func (s *byteScanner) scanIdent() token {
	start := s.pos
	_, size := utf8.DecodeRuneInString(s.text[s.pos:])
	s.pos += size
	for s.pos < len(s.text) {
		r, n := utf8.DecodeRuneInString(s.text[s.pos:])
		if !isIdentCont(r) {
			break
		}
		s.pos += n
	}
	text := s.text[start:s.pos]
	kind := tokIdent
	switch text {
	case "import":
		kind = tokImport
	case "export":
		kind = tokExport
	case "from":
		kind = tokFrom
	case "as":
		kind = tokAs
	}
	return token{kind: kind, text: text, start: start, end: s.pos}
}

func (s *byteScanner) scanString(quote byte) token {
	start := s.pos
	s.pos++
	for s.pos < len(s.text) {
		ch := s.text[s.pos]
		if ch == '\n' || ch == '\r' {
			break
		}
		s.pos++
		if ch == '\\' && s.pos < len(s.text) && s.text[s.pos] != '\n' && s.text[s.pos] != '\r' {
			s.pos++
			continue
		}
		if ch == quote {
			break
		}
	}
	return token{kind: tokString, text: s.text[start:s.pos], start: start, end: s.pos}
}

func (s *byteScanner) scanTemplate() token {
	start := s.pos
	s.pos++
	for s.pos < len(s.text) {
		ch := s.text[s.pos]
		if ch == '\\' && s.pos+1 < len(s.text) {
			s.pos += 2
			continue
		}
		s.pos++
		if ch == '`' {
			break
		}
	}
	return token{kind: tokOther, text: s.text[start:s.pos], start: start, end: s.pos}
}

func (s *byteScanner) skipTrivia() {
	for s.pos < len(s.text) {
		ch := s.text[s.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == '\v' || ch == '\f' {
			s.pos++
			continue
		}
		if ch == '/' && s.pos+1 < len(s.text) {
			next := s.text[s.pos+1]
			if next == '/' {
				s.pos += 2
				for s.pos < len(s.text) && s.text[s.pos] != '\n' {
					s.pos++
				}
				continue
			}
			if next == '*' {
				s.pos += 2
				for s.pos+1 < len(s.text) && !(s.text[s.pos] == '*' && s.text[s.pos+1] == '/') {
					s.pos++
				}
				if s.pos+1 < len(s.text) {
					s.pos += 2
				} else {
					s.pos = len(s.text)
				}
				continue
			}
		}
		return
	}
}

func isIdentStart(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r)
}

func isIdentCont(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func keepsTsExtension(value string) bool {
	return strings.HasSuffix(value, ".d.ts") || strings.HasSuffix(value, ".test.ts") || strings.HasSuffix(value, ".test.tsx")
}

func stripTrailingIndex(value string) string {
	parts := pathSepRe.Split(value, -1)
	if len(parts) < 2 {
		return value
	}
	last := parts[len(parts)-1]
	if last != "index" && last != "index.ts" && last != "index.tsx" {
		return value
	}
	parts = parts[:len(parts)-1]
	if len(parts) == 1 && parts[0] == "" {
		return "."
	}
	return strings.Join(parts, "/")
}

func fixModuleSpecifier(raw string) string {
	if len(raw) < 2 {
		return raw
	}
	quote := raw[0]
	if quote != '\'' && quote != '"' {
		return raw
	}
	value := unescapeSpecifier(raw[1 : len(raw)-1])
	if !keepsTsExtension(value) {
		if strings.HasSuffix(value, ".tsx") {
			value = value[:len(value)-4]
		} else if strings.HasSuffix(value, ".ts") {
			value = value[:len(value)-3]
		}
	}
	value = stripTrailingIndex(value)
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`) + `"`
}

func unescapeSpecifier(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			b.WriteByte(s[i+1])
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func applyEdits(text string, edits []edit) (string, error) {
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start > edits[j].start
		}
		return edits[i].end > edits[j].end
	})
	out := text
	for _, e := range edits {
		if e.start < 0 || e.end > len(out) || e.start > e.end {
			return "", fmt.Errorf("edit %d:%d outside text length %d", e.start, e.end, len(out))
		}
		out = out[:e.start] + e.text + out[e.end:]
	}
	return out, nil
}
