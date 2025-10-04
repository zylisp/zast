package builder

import (
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestParseInt(t *testing.T) {
	builder := New()

	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{`42`, 42, false},
		{`-10`, -10, false},
		{`0`, 0, false},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseInt(sexpNode)
		if tt.hasError && err == nil {
			t.Fatalf("expected error for input %q", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Fatalf("unexpected error for input %q: %v", tt.input, err)
		}
		if !tt.hasError && result != tt.expected {
			t.Fatalf("expected %d, got %d", tt.expected, result)
		}
	}
}

func TestParseString(t *testing.T) {
	builder := New()

	input := `"hello"`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	result, err := builder.parseString(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected %q, got %q", "hello", result)
	}

	// Test error case
	badInput := `42`
	parser = sexp.NewParser(badInput)
	sexpNode, _ = parser.Parse()

	_, err = builder.parseString(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-string input")
	}
}

func TestParseToken(t *testing.T) {
	builder := New()

	tests := []struct {
		input    string
		expected token.Token
		hasError bool
	}{
		{"IMPORT", token.IMPORT, false},
		{"CONST", token.CONST, false},
		{"TYPE", token.TYPE, false},
		{"VAR", token.VAR, false},
		{"INT", token.INT, false},
		{"FLOAT", token.FLOAT, false},
		{"IMAG", token.IMAG, false},
		{"CHAR", token.CHAR, false},
		{"STRING", token.STRING, false},
		{"INVALID", token.ILLEGAL, true},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseToken(sexpNode)
		if tt.hasError && err == nil {
			t.Fatalf("expected error for token %q", tt.input)
		}
		if !tt.hasError {
			if err != nil {
				t.Fatalf("unexpected error for token %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		}
	}
}

func TestParsePos(t *testing.T) {
	builder := New()

	input := `123`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	result := builder.parsePos(sexpNode)
	if result != token.Pos(123) {
		t.Fatalf("expected pos %d, got %d", 123, result)
	}
}

func TestParseNil(t *testing.T) {
	builder := New()

	input := `nil`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	if !builder.parseNil(sexpNode) {
		t.Fatalf("expected parseNil to return true for nil")
	}

	input2 := `foo`
	parser2 := sexp.NewParser(input2)
	sexpNode2, _ := parser2.Parse()

	if builder.parseNil(sexpNode2) {
		t.Fatalf("expected parseNil to return false for non-nil")
	}
}

func TestParseKeywordArgsEdgeCases(t *testing.T) {
	builder := New()

	// Test odd number of elements (missing value)
	input := `(Foo :name "bar" :type)`
	parser := sexp.NewParser(input)
	list, _ := parser.Parse()
	listNode := list.(*sexp.List)

	args := builder.parseKeywordArgs(listNode.Elements)

	// Should still parse the first keyword correctly
	if _, ok := args["name"]; !ok {
		t.Fatalf("expected 'name' key to be present")
	}

	// Error should be recorded for missing value
	if len(builder.errors) == 0 {
		t.Fatalf("expected error for missing value")
	}
}

func TestRequireKeywordMissing(t *testing.T) {
	builder := New()
	args := map[string]sexp.SExp{}

	_, ok := builder.requireKeyword(args, "missing", "TestNode")
	if ok {
		t.Fatalf("expected requireKeyword to return false")
	}

	if len(builder.errors) == 0 {
		t.Fatalf("expected error to be recorded")
	}
}

func TestGetKeyword(t *testing.T) {
	builder := New()

	input := `(Foo :name "bar")`
	parser := sexp.NewParser(input)
	list, _ := parser.Parse()
	listNode := list.(*sexp.List)

	args := builder.parseKeywordArgs(listNode.Elements)

	// Test getting existing keyword
	val, ok := builder.getKeyword(args, "name")
	if !ok {
		t.Fatalf("expected to find 'name' keyword")
	}
	if val == nil {
		t.Fatalf("expected non-nil value")
	}

	// Test getting non-existing keyword
	_, ok = builder.getKeyword(args, "missing")
	if ok {
		t.Fatalf("expected not to find 'missing' keyword")
	}
}

func TestErrors(t *testing.T) {
	builder := New()

	// Initially no errors
	if len(builder.Errors()) != 0 {
		t.Fatalf("expected no errors initially")
	}

	// Add an error
	builder.addError("test error")

	errors := builder.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	if errors[0] != "test error" {
		t.Fatalf("expected error %q, got %q", "test error", errors[0])
	}
}

func TestParseIntWithSymbol(t *testing.T) {
	builder := New()

	// Test parsing int from symbol
	input := `42`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	result, err := builder.parseInt(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
}

func TestParseIntError(t *testing.T) {
	builder := New()

	// Test error case with invalid type
	input := `"not a number"`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.parseInt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-number input")
	}
}

func TestParsePosError(t *testing.T) {
	builder := New()

	// Test with invalid position
	input := `"not a number"`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	pos := builder.parsePos(sexpNode)
	// Should return NoPos and record error
	if pos != 0 {
		t.Fatalf("expected NoPos (0), got %d", pos)
	}

	if len(builder.errors) == 0 {
		t.Fatalf("expected error to be recorded")
	}
}

func TestParseTokenWithNonSymbol(t *testing.T) {
	builder := New()

	input := `42`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.parseToken(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol input")
	}
}

func TestExpectListWithNonList(t *testing.T) {
	input := `foo`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, ok := builder.expectList(sexpNode, "test")
	if ok {
		t.Fatalf("expected expectList to return false for non-list")
	}

	if len(builder.errors) == 0 {
		t.Fatalf("expected error to be recorded")
	}
}

func TestExpectSymbolWithWrongSymbol(t *testing.T) {
	input := `foo`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	ok := builder.expectSymbol(sexpNode, "bar")
	if ok {
		t.Fatalf("expected expectSymbol to return false for wrong symbol")
	}
}

func TestExpectSymbolWithNonSymbol(t *testing.T) {
	input := `42`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	ok := builder.expectSymbol(sexpNode, "foo")
	if ok {
		t.Fatalf("expected expectSymbol to return false for non-symbol")
	}
}

func TestParseKeywordArgsWithNonKeyword(t *testing.T) {
	builder := New()

	// Test with non-keyword where keyword expected
	input := `(Foo 42 "bar")`
	parser := sexp.NewParser(input)
	list, _ := parser.Parse()
	listNode := list.(*sexp.List)

	_ = builder.parseKeywordArgs(listNode.Elements)

	// Error should be recorded for non-keyword
	if len(builder.errors) == 0 {
		t.Fatalf("expected error for non-keyword")
	}
}
