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

// TestParseBool tests boolean parsing
func TestParseBool(t *testing.T) {
	builder := New()

	tests := []struct {
		input    string
		expected bool
		hasError bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"TRUE", false, true}, // uppercase not supported
		{"42", false, true},   // non-bool
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseBool(sexpNode)
		if tt.hasError && err == nil {
			t.Fatalf("expected error for input %q", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Fatalf("unexpected error for input %q: %v", tt.input, err)
		}
		if !tt.hasError && result != tt.expected {
			t.Fatalf("expected %v, got %v", tt.expected, result)
		}
	}
}

// TestParseChanDir tests channel direction parsing
func TestParseChanDir(t *testing.T) {
	builder := New()

	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"SEND", 1, false},      // ast.SEND
		{"RECV", 2, false},      // ast.RECV
		{"INVALID", 0, true},    // Invalid direction
		{"42", 0, true},         // Non-symbol
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseChanDir(sexpNode)
		if tt.hasError && err == nil {
			t.Fatalf("expected error for input %q", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Fatalf("unexpected error for input %q: %v", tt.input, err)
		}
		if !tt.hasError && int(result) != tt.expected {
			t.Fatalf("expected %v, got %v", tt.expected, result)
		}
	}
}

// TestBuildComment tests comment building
func TestBuildComment(t *testing.T) {
	input := `(Comment :slash 1 :text "// comment")`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	comment, err := builder.buildComment(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if comment.Text != "// comment" {
		t.Fatalf("expected text %q, got %q", "// comment", comment.Text)
	}
}

// TestBuildCommentGroup tests comment group building
func TestBuildCommentGroup(t *testing.T) {
	input := `(CommentGroup :list ((Comment :slash 1 :text "// comment")))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	group, err := builder.buildCommentGroup(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if len(group.List) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(group.List))
	}

	if group.List[0].Text != "// comment" {
		t.Fatalf("expected text %q, got %q", "// comment", group.List[0].Text)
	}
}

// TestParseObjKind tests object kind parsing
func TestParseObjKind(t *testing.T) {
	builder := New()

	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"Bad", 0, false},   // ast.Bad
		{"Pkg", 1, false},   // ast.Pkg
		{"Con", 2, false},   // ast.Con
		{"Typ", 3, false},   // ast.Typ
		{"Var", 4, false},   // ast.Var
		{"Fun", 5, false},   // ast.Fun
		{"Lbl", 6, false},   // ast.Lbl
		{"INVALID", 0, true},// Invalid kind
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseObjKind(sexpNode)
		if tt.hasError && err == nil {
			t.Fatalf("expected error for input %q", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Fatalf("unexpected error for input %q: %v", tt.input, err)
		}
		if !tt.hasError && int(result) != tt.expected {
			t.Fatalf("expected %v, got %v", tt.expected, result)
		}
	}
}

// TestBuildObject tests object building
func TestBuildObject(t *testing.T) {
	input := `(Object
		:kind Var
		:name "x"
		:decl nil
		:data nil
		:type nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	obj, err := builder.buildObject(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if obj.Name != "x" {
		t.Fatalf("expected name %q, got %q", "x", obj.Name)
	}
}

// TestParseTokenAllTypes tests parsing all token types
func TestParseTokenAllTypes(t *testing.T) {
	builder := New()

	tests := []struct {
		input    string
		expected token.Token
	}{
		// Keywords (only those supported in parseToken)
		{"IMPORT", token.IMPORT},
		{"TYPE", token.TYPE},
		{"VAR", token.VAR},
		{"CONST", token.CONST},
		{"BREAK", token.BREAK},
		{"CONTINUE", token.CONTINUE},
		{"GOTO", token.GOTO},
		{"FALLTHROUGH", token.FALLTHROUGH},

		// Operators
		{"ADD", token.ADD},
		{"SUB", token.SUB},
		{"MUL", token.MUL},
		{"QUO", token.QUO},
		{"REM", token.REM},
		{"AND", token.AND},
		{"OR", token.OR},
		{"XOR", token.XOR},
		{"SHL", token.SHL},
		{"SHR", token.SHR},
		{"AND_NOT", token.AND_NOT},
		{"LAND", token.LAND},
		{"LOR", token.LOR},
		{"ARROW", token.ARROW},
		{"INC", token.INC},
		{"DEC", token.DEC},
		{"EQL", token.EQL},
		{"LSS", token.LSS},
		{"GTR", token.GTR},
		{"ASSIGN", token.ASSIGN},
		{"NOT", token.NOT},
		{"NEQ", token.NEQ},
		{"LEQ", token.LEQ},
		{"GEQ", token.GEQ},
		{"DEFINE", token.DEFINE},
		{"ADD_ASSIGN", token.ADD_ASSIGN},
		{"SUB_ASSIGN", token.SUB_ASSIGN},
		{"MUL_ASSIGN", token.MUL_ASSIGN},
		{"QUO_ASSIGN", token.QUO_ASSIGN},
		{"REM_ASSIGN", token.REM_ASSIGN},
		{"AND_ASSIGN", token.AND_ASSIGN},
		{"OR_ASSIGN", token.OR_ASSIGN},
		{"XOR_ASSIGN", token.XOR_ASSIGN},
		{"SHL_ASSIGN", token.SHL_ASSIGN},
		{"SHR_ASSIGN", token.SHR_ASSIGN},
		{"AND_NOT_ASSIGN", token.AND_NOT_ASSIGN},

		// Literal types
		{"INT", token.INT},
		{"FLOAT", token.FLOAT},
		{"IMAG", token.IMAG},
		{"STRING", token.STRING},
		{"CHAR", token.CHAR},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseToken(sexpNode)
		if err != nil {
			t.Fatalf("unexpected error for token %q: %v", tt.input, err)
		}
		if result != tt.expected {
			t.Fatalf("expected %v for %q, got %v", tt.expected, tt.input, result)
		}
	}
}
