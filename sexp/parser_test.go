package sexp

import (
	"testing"
)

// Test basic symbol parsing
func TestParseSymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo", "foo"},
		{"GenDecl", "GenDecl"},
		{"IMPORT", "IMPORT"},
	}

	for _, tt := range tests {
		parser := NewParser(tt.input)
		sexp, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		symbol, ok := sexp.(*Symbol)
		if !ok {
			t.Fatalf("expected Symbol, got %T", sexp)
		}

		if symbol.Value != tt.expected {
			t.Fatalf("expected value %q, got %q", tt.expected, symbol.Value)
		}
	}
}

// Test keyword parsing
func TestParseKeyword(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{":name", "name"},
		{":package", "package"},
		{":tok", "tok"},
	}

	for _, tt := range tests {
		parser := NewParser(tt.input)
		sexp, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		keyword, ok := sexp.(*Keyword)
		if !ok {
			t.Fatalf("expected Keyword, got %T", sexp)
		}

		if keyword.Name != tt.expected {
			t.Fatalf("expected name %q, got %q", tt.expected, keyword.Name)
		}
	}
}

// Test string parsing
func TestParseString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`"hello\nworld"`, "hello\nworld"},
		{`"hello\tworld"`, "hello\tworld"},
		{`"say \"hi\""`, `say "hi"`},
	}

	for _, tt := range tests {
		parser := NewParser(tt.input)
		sexp, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		str, ok := sexp.(*String)
		if !ok {
			t.Fatalf("expected String, got %T", sexp)
		}

		if str.Value != tt.expected {
			t.Fatalf("expected value %q, got %q", tt.expected, str.Value)
		}
	}
}

// Test number parsing
func TestParseNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"-10", "-10"},
		{"3.14", "3.14"},
		{"0", "0"},
		{"-0.5", "-0.5"},
	}

	for _, tt := range tests {
		parser := NewParser(tt.input)
		sexp, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		num, ok := sexp.(*Number)
		if !ok {
			t.Fatalf("expected Number, got %T", sexp)
		}

		if num.Value != tt.expected {
			t.Fatalf("expected value %q, got %q", tt.expected, num.Value)
		}
	}
}

// Test nil parsing
func TestParseNil(t *testing.T) {
	parser := NewParser("nil")
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	_, ok := sexp.(*Nil)
	if !ok {
		t.Fatalf("expected Nil, got %T", sexp)
	}
}

// Test empty list
func TestParseEmptyList(t *testing.T) {
	parser := NewParser("()")
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	list, ok := sexp.(*List)
	if !ok {
		t.Fatalf("expected List, got %T", sexp)
	}

	if len(list.Elements) != 0 {
		t.Fatalf("expected empty list, got %d elements", len(list.Elements))
	}
}

// Test simple list
func TestParseSimpleList(t *testing.T) {
	parser := NewParser("(foo bar)")
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	list, ok := sexp.(*List)
	if !ok {
		t.Fatalf("expected List, got %T", sexp)
	}

	if len(list.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(list.Elements))
	}

	sym1, ok := list.Elements[0].(*Symbol)
	if !ok || sym1.Value != "foo" {
		t.Fatalf("expected first element to be Symbol 'foo', got %v", list.Elements[0])
	}

	sym2, ok := list.Elements[1].(*Symbol)
	if !ok || sym2.Value != "bar" {
		t.Fatalf("expected second element to be Symbol 'bar', got %v", list.Elements[1])
	}
}

// Test nested list
func TestParseNestedList(t *testing.T) {
	parser := NewParser("(foo (bar baz))")
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	list, ok := sexp.(*List)
	if !ok {
		t.Fatalf("expected List, got %T", sexp)
	}

	if len(list.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(list.Elements))
	}

	sym, ok := list.Elements[0].(*Symbol)
	if !ok || sym.Value != "foo" {
		t.Fatalf("expected first element to be Symbol 'foo', got %v", list.Elements[0])
	}

	nested, ok := list.Elements[1].(*List)
	if !ok {
		t.Fatalf("expected second element to be List, got %T", list.Elements[1])
	}

	if len(nested.Elements) != 2 {
		t.Fatalf("expected nested list to have 2 elements, got %d", len(nested.Elements))
	}
}

// Test list with keywords
func TestParseListWithKeywords(t *testing.T) {
	parser := NewParser("(:name foo :type bar)")
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	list, ok := sexp.(*List)
	if !ok {
		t.Fatalf("expected List, got %T", sexp)
	}

	if len(list.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(list.Elements))
	}

	kw1, ok := list.Elements[0].(*Keyword)
	if !ok || kw1.Name != "name" {
		t.Fatalf("expected first element to be Keyword 'name', got %v", list.Elements[0])
	}

	sym1, ok := list.Elements[1].(*Symbol)
	if !ok || sym1.Value != "foo" {
		t.Fatalf("expected second element to be Symbol 'foo', got %v", list.Elements[1])
	}

	kw2, ok := list.Elements[2].(*Keyword)
	if !ok || kw2.Name != "type" {
		t.Fatalf("expected third element to be Keyword 'type', got %v", list.Elements[2])
	}

	sym2, ok := list.Elements[3].(*Symbol)
	if !ok || sym2.Value != "bar" {
		t.Fatalf("expected fourth element to be Symbol 'bar', got %v", list.Elements[3])
	}
}

// Test complex canonical format structure
func TestParseComplexStructure(t *testing.T) {
	input := `(File :package 1 :name (Ident :namepos 9 :name "main"))`
	parser := NewParser(input)
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	list, ok := sexp.(*List)
	if !ok {
		t.Fatalf("expected List, got %T", sexp)
	}

	if len(list.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(list.Elements))
	}

	// Check File symbol
	sym, ok := list.Elements[0].(*Symbol)
	if !ok || sym.Value != "File" {
		t.Fatalf("expected element 0 to be Symbol 'File', got %v", list.Elements[0])
	}

	// Check :package keyword
	kw1, ok := list.Elements[1].(*Keyword)
	if !ok || kw1.Name != "package" {
		t.Fatalf("expected element 1 to be Keyword 'package', got %v", list.Elements[1])
	}

	// Check 1 number
	num, ok := list.Elements[2].(*Number)
	if !ok || num.Value != "1" {
		t.Fatalf("expected element 2 to be Number '1', got %v", list.Elements[2])
	}

	// Check :name keyword
	kw2, ok := list.Elements[3].(*Keyword)
	if !ok || kw2.Name != "name" {
		t.Fatalf("expected element 3 to be Keyword 'name', got %v", list.Elements[3])
	}

	// Check nested list
	nested, ok := list.Elements[4].(*List)
	if !ok {
		t.Fatalf("expected element 4 to be List, got %T", list.Elements[4])
	}

	if len(nested.Elements) != 5 {
		t.Fatalf("expected nested list to have 5 elements, got %d", len(nested.Elements))
	}

	// Check nested Ident symbol
	identSym, ok := nested.Elements[0].(*Symbol)
	if !ok || identSym.Value != "Ident" {
		t.Fatalf("expected nested element 0 to be Symbol 'Ident', got %v", nested.Elements[0])
	}

	// Check nested :name string
	str, ok := nested.Elements[4].(*String)
	if !ok || str.Value != "main" {
		t.Fatalf("expected nested element 4 to be String 'main', got %v", nested.Elements[4])
	}
}

// Test unterminated list error
func TestParseUnterminatedList(t *testing.T) {
	parser := NewParser("(foo bar")
	_, err := parser.Parse()
	if err == nil {
		t.Fatalf("expected error for unterminated list")
	}
}

// Test unexpected token error
func TestParseUnexpectedToken(t *testing.T) {
	parser := NewParser(")")
	_, err := parser.Parse()
	if err == nil {
		t.Fatalf("expected error for unexpected token")
	}
}

// Test empty input error
func TestParseEmptyInput(t *testing.T) {
	parser := NewParser("")
	_, err := parser.Parse()
	if err == nil {
		t.Fatalf("expected error for empty input")
	}
}

// Test position tracking
func TestParsePositionTracking(t *testing.T) {
	input := "(foo\n  bar)"
	parser := NewParser(input)
	sexp, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	list, ok := sexp.(*List)
	if !ok {
		t.Fatalf("expected List, got %T", sexp)
	}

	// Check position of list (opening paren)
	if list.Position.Line != 1 || list.Position.Column != 1 {
		t.Fatalf("expected list position line 1, column 1, got line %d, column %d",
			list.Position.Line, list.Position.Column)
	}

	// Check position of "bar" symbol
	barSym, ok := list.Elements[1].(*Symbol)
	if !ok {
		t.Fatalf("expected second element to be Symbol, got %T", list.Elements[1])
	}

	if barSym.Position.Line != 2 {
		t.Fatalf("expected 'bar' at line 2, got line %d", barSym.Position.Line)
	}

	if barSym.Position.Column != 3 {
		t.Fatalf("expected 'bar' at column 3, got column %d", barSym.Position.Column)
	}
}

// Test ParseList method
func TestParseListMethod(t *testing.T) {
	input := "(foo bar)"
	parser := NewParser(input)
	list, err := parser.ParseList()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(list.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(list.Elements))
	}
}

// Test ParseList with non-list input
func TestParseListMethodError(t *testing.T) {
	input := "foo"
	parser := NewParser(input)
	_, err := parser.ParseList()
	if err == nil {
		t.Fatalf("expected error when ParseList called on non-list")
	}
}

// Test String methods for debugging
func TestStringMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo", "foo"},
		{":name", ":name"},
		{`"hello"`, `"hello"`},
		{"42", "42"},
		{"nil", "nil"},
		{"()", "()"},
		{"(foo bar)", "(foo bar)"},
	}

	for _, tt := range tests {
		parser := NewParser(tt.input)
		sexp, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error for %q: %v", tt.input, err)
		}

		str := sexp.String()
		if str != tt.expected {
			t.Fatalf("expected String() %q, got %q", tt.expected, str)
		}
	}
}
