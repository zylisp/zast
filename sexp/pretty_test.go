package sexp

import (
	"strings"
	"testing"
)

func TestFormatAtomicValues(t *testing.T) {
	tests := []struct {
		name     string
		input    SExp
		expected string
	}{
		{"symbol", &Symbol{Value: "foo"}, "foo"},
		{"keyword", &Keyword{Name: "name"}, ":name"},
		{"string", &String{Value: "hello"}, `"hello"`},
		{"number", &Number{Value: "42"}, "42"},
		{"nil", &Nil{}, "nil"},
	}

	pp := NewPrettyPrinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pp.Format(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatEmptyList(t *testing.T) {
	pp := NewPrettyPrinter()
	input := &List{Elements: []SExp{}}
	result := pp.Format(input)

	if result != "()" {
		t.Errorf("expected %q, got %q", "()", result)
	}
}

func TestCompactFormatting(t *testing.T) {
	// Small list should stay on one line
	input := &List{Elements: []SExp{
		&Symbol{Value: "Ident"},
		&Keyword{Name: "namepos"},
		&Number{Value: "9"},
		&Keyword{Name: "name"},
		&String{Value: "main"},
		&Keyword{Name: "obj"},
		&Nil{},
	}}

	pp := NewPrettyPrinter()
	result := pp.Format(input)

	// Should be on one line (no newlines)
	if strings.Contains(result, "\n") {
		t.Errorf("expected compact format, but got:\n%s", result)
	}

	expected := `(Ident :namepos 9 :name "main" :obj nil)`
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestKeywordAlignment(t *testing.T) {
	input := &List{Elements: []SExp{
		&Symbol{Value: "File"},
		&Keyword{Name: "package"},
		&Number{Value: "1"},
		&Keyword{Name: "name"},
		&Symbol{Value: "main"},
		&Keyword{Name: "decls"},
		&List{Elements: []SExp{}},
	}}

	pp := NewPrettyPrinter()
	result := pp.Format(input)

	// Check that keywords are aligned
	lines := strings.Split(result, "\n")
	if len(lines) <= 3 {
		t.Fatalf("expected multiline output, got: %s", result)
	}

	// Extract keyword column positions
	var positions []int
	for _, line := range lines[1:] { // skip first line
		if idx := strings.Index(line, ":"); idx != -1 {
			positions = append(positions, idx)
		}
	}

	// All keywords should start at same column
	if len(positions) > 1 {
		first := positions[0]
		for i, pos := range positions[1:] {
			if pos != first {
				t.Errorf("Keywords not aligned: position %d is %d, expected %d\nOutput:\n%s",
					i+1, pos, first, result)
			}
		}
	}
}

func TestBodyFormatting(t *testing.T) {
	input := &List{Elements: []SExp{
		&Symbol{Value: "BlockStmt"},
		&Keyword{Name: "lbrace"},
		&Number{Value: "10"},
		&Keyword{Name: "list"},
		&List{Elements: []SExp{
			&List{Elements: []SExp{&Symbol{Value: "ExprStmt"}}},
			&List{Elements: []SExp{&Symbol{Value: "ReturnStmt"}}},
		}},
		&Keyword{Name: "rbrace"},
		&Number{Value: "20"},
	}}

	pp := NewPrettyPrinter()
	result := pp.Format(input)

	// Verify list elements are indented
	if !strings.Contains(result, "ExprStmt") {
		t.Errorf("expected ExprStmt in output")
	}
	if !strings.Contains(result, "ReturnStmt") {
		t.Errorf("expected ReturnStmt in output")
	}

	// Verify body formatting
	if !strings.Contains(result, ":list") {
		t.Errorf("expected :list keyword in output")
	}
}

func TestCustomConfiguration(t *testing.T) {
	config := &PrettyPrintConfig{
		IndentWidth:   4,
		MaxLineWidth:  40,
		AlignKeywords: false,
		CompactSmall:  false,
		CompactLimit:  20,
	}

	pp := NewPrettyPrinterWithConfig(config)

	input := &List{Elements: []SExp{
		&Symbol{Value: "Test"},
		&Keyword{Name: "a"},
		&Number{Value: "1"},
	}}

	result := pp.Format(input)

	// With 4-space indent and no compact formatting
	lines := strings.Split(result, "\n")
	if len(lines) > 1 {
		// Check indent is 4 spaces
		if !strings.HasPrefix(lines[1], "    ") {
			t.Errorf("expected 4-space indent, got: %q", lines[1])
		}
	}
}

func TestAlignmentDisabled(t *testing.T) {
	config := &PrettyPrintConfig{
		IndentWidth:   2,
		MaxLineWidth:  80,
		AlignKeywords: false,
		CompactSmall:  true,
		CompactLimit:  60,
	}

	pp := NewPrettyPrinterWithConfig(config)

	input := &List{Elements: []SExp{
		&Symbol{Value: "File"},
		&Keyword{Name: "package"},
		&Number{Value: "1"},
		&Keyword{Name: "name"},
		&Symbol{Value: "main"},
	}}

	result := pp.Format(input)

	// With alignment disabled, keywords should NOT be aligned
	lines := strings.Split(result, "\n")
	if len(lines) > 2 {
		// Each line should have keyword followed by single space
		for _, line := range lines[1:] {
			if strings.Contains(line, ":") {
				idx := strings.Index(line, ":")
				if idx+1 < len(line) {
					// Find next non-space character position
					nextIdx := idx + 1
					for nextIdx < len(line) && line[nextIdx] != ' ' {
						nextIdx++
					}
					// Should be exactly one space after keyword
					if nextIdx < len(line) && line[nextIdx] == ' ' {
						// Count spaces
						spaces := 0
						for i := nextIdx; i < len(line) && line[i] == ' '; i++ {
							spaces++
						}
						if spaces > 1 {
							t.Errorf("expected single space after keyword with alignment disabled, got %d spaces in: %q",
								spaces, line)
						}
					}
				}
			}
		}
	}
}

func TestCompactLimit(t *testing.T) {
	config := &PrettyPrintConfig{
		IndentWidth:   2,
		MaxLineWidth:  80,
		AlignKeywords: true,
		CompactSmall:  true,
		CompactLimit:  10, // Very small limit
	}

	pp := NewPrettyPrinterWithConfig(config)

	// Use a known type (File) which has StyleKeywordPairs
	// This should exceed the compact limit
	input := &List{Elements: []SExp{
		&Symbol{Value: "File"},
		&Keyword{Name: "package"},
		&Number{Value: "1"},
		&Keyword{Name: "name"},
		&List{Elements: []SExp{
			&Symbol{Value: "Ident"},
			&Keyword{Name: "name"},
			&String{Value: "main"},
		}},
	}}

	result := pp.Format(input)

	// Should be multiline due to compact limit
	if !strings.Contains(result, "\n") {
		t.Errorf("expected multiline output due to compact limit, got: %s", result)
	}
}

func TestNestedLists(t *testing.T) {
	input := &List{Elements: []SExp{
		&Symbol{Value: "Outer"},
		&Keyword{Name: "inner"},
		&List{Elements: []SExp{
			&Symbol{Value: "Inner"},
			&Keyword{Name: "value"},
			&Number{Value: "42"},
		}},
	}}

	pp := NewPrettyPrinter()
	result := pp.Format(input)

	// Should contain nested structure
	if !strings.Contains(result, "Outer") {
		t.Errorf("expected Outer in output")
	}
	if !strings.Contains(result, "Inner") {
		t.Errorf("expected Inner in output")
	}
}

func TestRoundTrip(t *testing.T) {
	// Parse compact S-expression
	input := `(File :package 1 :name (Ident :namepos 9 :name "main" :obj nil))`
	parser := NewParser(input)
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Pretty print it
	pp := NewPrettyPrinter()
	pretty := pp.Format(sexp)

	// Parse the pretty-printed version
	parser2 := NewParser(pretty)
	sexp2, err := parser2.Parse()
	if err != nil {
		t.Fatalf("parse pretty error: %v", err)
	}

	// Should produce same structure (compare string representation)
	if sexp.String() != sexp2.String() {
		t.Errorf("round-trip failed:\nOriginal: %s\nPretty:   %s\nReparsed: %s",
			input, pretty, sexp2.String())
	}
}

func TestSimpleListStyle(t *testing.T) {
	input := &List{Elements: []SExp{
		&Symbol{Value: "CommentGroup"},
		&List{Elements: []SExp{
			&Symbol{Value: "Comment1"},
		}},
		&List{Elements: []SExp{
			&Symbol{Value: "Comment2"},
		}},
	}}

	pp := NewPrettyPrinter()
	result := pp.Format(input)

	// CommentGroup uses StyleList
	if !strings.Contains(result, "Comment1") {
		t.Errorf("expected Comment1 in output")
	}
	if !strings.Contains(result, "Comment2") {
		t.Errorf("expected Comment2 in output")
	}
}

func TestDefaultStyle(t *testing.T) {
	// Unknown node type should use default formatting
	input := &List{Elements: []SExp{
		&Symbol{Value: "UnknownNodeType"},
		&Symbol{Value: "arg1"},
		&List{Elements: []SExp{
			&Symbol{Value: "nested"},
		}},
		&Symbol{Value: "arg2"},
	}}

	pp := NewPrettyPrinter()
	result := pp.Format(input)

	// Default style will try compact first, and if that fails,
	// should break before nested lists
	// Just verify it contains the expected elements
	if !strings.Contains(result, "UnknownNodeType") {
		t.Errorf("expected UnknownNodeType in output")
	}
	if !strings.Contains(result, "nested") {
		t.Errorf("expected nested in output")
	}
}

func TestMaxLineWidth(t *testing.T) {
	config := &PrettyPrintConfig{
		IndentWidth:   2,
		MaxLineWidth:  30, // Small line width
		AlignKeywords: true,
		CompactSmall:  true,
		CompactLimit:  60,
	}

	pp := NewPrettyPrinterWithConfig(config)

	// Use File (StyleKeywordPairs) with long content that would exceed line width
	input := &List{Elements: []SExp{
		&Symbol{Value: "File"},
		&Keyword{Name: "package"},
		&Number{Value: "1"},
		&Keyword{Name: "name"},
		&List{Elements: []SExp{
			&Symbol{Value: "Ident"},
			&Keyword{Name: "name"},
			&String{Value: "veryLongPackageNameHere"},
		}},
	}}

	result := pp.Format(input)

	// File uses StyleKeywordPairs which always formats multiline
	if !strings.Contains(result, "\n") {
		t.Errorf("expected multiline output, got: %s", result)
	}
}

func TestFormatComplexStructure(t *testing.T) {
	// Build a more complex structure
	input := &List{Elements: []SExp{
		&Symbol{Value: "FuncDecl"},
		&Keyword{Name: "name"},
		&List{Elements: []SExp{
			&Symbol{Value: "Ident"},
			&Keyword{Name: "name"},
			&String{Value: "main"},
		}},
		&Keyword{Name: "type"},
		&List{Elements: []SExp{
			&Symbol{Value: "FuncType"},
			&Keyword{Name: "params"},
			&List{Elements: []SExp{}},
		}},
		&Keyword{Name: "body"},
		&List{Elements: []SExp{
			&Symbol{Value: "BlockStmt"},
			&Keyword{Name: "list"},
			&List{Elements: []SExp{}},
		}},
	}}

	pp := NewPrettyPrinter()
	result := pp.Format(input)

	// Verify all parts are present
	expectedParts := []string{"FuncDecl", "Ident", "FuncType", "BlockStmt"}
	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("expected %q in output", part)
		}
	}

	// Should be parseable
	parser := NewParser(result)
	_, err := parser.Parse()
	if err != nil {
		t.Errorf("pretty-printed output should be parseable: %v", err)
	}
}
