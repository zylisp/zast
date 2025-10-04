package builder

import (
	"fmt"
	"go/ast"
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestBuildIdent(t *testing.T) {
	input := `(Ident :namepos 10 :name "main" :obj nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	ident, err := builder.buildIdent(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if ident.Name != "main" {
		t.Fatalf("expected name %q, got %q", "main", ident.Name)
	}

	if ident.NamePos != token.Pos(10) {
		t.Fatalf("expected namepos %d, got %d", 10, ident.NamePos)
	}

	if ident.Obj != nil {
		t.Fatalf("expected nil obj, got %v", ident.Obj)
	}
}

func TestBuildBasicLit(t *testing.T) {
	input := `(BasicLit :valuepos 22 :kind STRING :value "\"fmt\"")`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	lit, err := builder.buildBasicLit(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if lit.Kind != token.STRING {
		t.Fatalf("expected kind STRING, got %v", lit.Kind)
	}

	if lit.Value != `"fmt"` {
		t.Fatalf("expected value %q, got %q", `"fmt"`, lit.Value)
	}

	if lit.ValuePos != token.Pos(22) {
		t.Fatalf("expected valuepos %d, got %d", 22, lit.ValuePos)
	}
}

func TestBuildSelectorExpr(t *testing.T) {
	input := `(SelectorExpr
		:x (Ident :namepos 46 :name "fmt" :obj nil)
		:sel (Ident :namepos 50 :name "Println" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	sel, err := builder.buildSelectorExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := sel.X.(*ast.Ident)
	if !ok || xIdent.Name != "fmt" {
		t.Fatalf("expected X to be Ident 'fmt', got %v", sel.X)
	}

	if sel.Sel.Name != "Println" {
		t.Fatalf("expected Sel name %q, got %q", "Println", sel.Sel.Name)
	}
}

func TestBuildCallExpr(t *testing.T) {
	input := `(CallExpr
		:fun (Ident :namepos 10 :name "foo" :obj nil)
		:lparen 13
		:args ((BasicLit :valuepos 14 :kind STRING :value "\"hello\""))
		:ellipsis 0
		:rparen 21)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	call, err := builder.buildCallExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	funIdent, ok := call.Fun.(*ast.Ident)
	if !ok || funIdent.Name != "foo" {
		t.Fatalf("expected Fun to be Ident 'foo', got %v", call.Fun)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	argLit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || argLit.Value != `"hello"` {
		t.Fatalf("expected arg to be BasicLit '\"hello\"', got %v", call.Args[0])
	}
}

func TestBuildIdentMissingName(t *testing.T) {
	input := `(Ident :namepos 10 :obj nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	_, err = builder.buildIdent(sexpNode)
	if err == nil {
		t.Fatalf("expected error for missing name field")
	}
}

// Test different literal types
func TestBuildBasicLitTypes(t *testing.T) {
	tests := []struct {
		input string
		kind  token.Token
		value string
	}{
		{`(BasicLit :valuepos 10 :kind INT :value "42")`, token.INT, "42"},
		{`(BasicLit :valuepos 10 :kind FLOAT :value "3.14")`, token.FLOAT, "3.14"},
		{`(BasicLit :valuepos 10 :kind IMAG :value "1.5i")`, token.IMAG, "1.5i"},
		{`(BasicLit :valuepos 10 :kind CHAR :value "'a'")`, token.CHAR, "'a'"},
		{`(BasicLit :valuepos 10 :kind STRING :value "\"hello\"")`, token.STRING, `"hello"`},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error for %q: %v", tt.input, err)
		}

		builder := New()
		lit, err := builder.buildBasicLit(sexpNode)
		if err != nil {
			t.Fatalf("build error for %q: %v", tt.input, err)
		}

		if lit.Kind != tt.kind {
			t.Fatalf("expected kind %v, got %v", tt.kind, lit.Kind)
		}
		if lit.Value != tt.value {
			t.Fatalf("expected value %q, got %q", tt.value, lit.Value)
		}
	}
}

// Test different GenDecl token types
func TestBuildGenDeclTypes(t *testing.T) {
	tests := []struct {
		tokType  string
		expected token.Token
	}{
		{"IMPORT", token.IMPORT},
		{"CONST", token.CONST},
		{"TYPE", token.TYPE},
		{"VAR", token.VAR},
	}

	for _, tt := range tests {
		input := fmt.Sprintf(`(GenDecl :doc nil :tok %s :tokpos 10 :lparen 0 :specs () :rparen 0)`, tt.tokType)
		parser := sexp.NewParser(input)
		sexpNode, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		builder := New()
		decl, err := builder.buildGenDecl(sexpNode)
		if err != nil {
			t.Fatalf("build error for tok %s: %v", tt.tokType, err)
		}

		if decl.Tok != tt.expected {
			t.Fatalf("expected tok %v, got %v", tt.expected, decl.Tok)
		}
	}
}

// Test nested CallExpr
func TestBuildNestedCallExpr(t *testing.T) {
	input := `(CallExpr
		:fun (SelectorExpr
			:x (CallExpr
				:fun (Ident :namepos 10 :name "foo" :obj nil)
				:lparen 13
				:args ()
				:ellipsis 0
				:rparen 14)
			:sel (Ident :namepos 16 :name "Bar" :obj nil))
		:lparen 19
		:args ()
		:ellipsis 0
		:rparen 20)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	call, err := builder.buildCallExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	// Verify it's a selector expr
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected Fun to be SelectorExpr, got %T", call.Fun)
	}

	// Verify the selector's X is a CallExpr
	innerCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected selector X to be CallExpr, got %T", sel.X)
	}

	// Verify the inner call's function is an Ident
	innerIdent, ok := innerCall.Fun.(*ast.Ident)
	if !ok || innerIdent.Name != "foo" {
		t.Fatalf("expected inner call fun to be Ident 'foo', got %v", innerCall.Fun)
	}
}

// Test error cases for builders
func TestBuildExprUnknownType(t *testing.T) {
	input := `(UnknownExpr :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, err := builder.buildExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown expression type")
	}
}

func TestBuildBasicLitMissingFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing valuepos", `(BasicLit :kind STRING :value "foo")`},
		{"missing kind", `(BasicLit :valuepos 10 :value "foo")`},
		{"missing value", `(BasicLit :valuepos 10 :kind STRING)`},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		builder := New()
		_, err := builder.buildBasicLit(sexpNode)
		if err == nil {
			t.Fatalf("%s: expected error", tt.name)
		}
	}
}

func TestBuildSelectorExprMissingFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing x", `(SelectorExpr :sel (Ident :namepos 10 :name "foo" :obj nil))`},
		{"missing sel", `(SelectorExpr :x (Ident :namepos 10 :name "foo" :obj nil))`},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		builder := New()
		_, err := builder.buildSelectorExpr(sexpNode)
		if err == nil {
			t.Fatalf("%s: expected error", tt.name)
		}
	}
}

func TestBuildCallExprMissingFields(t *testing.T) {
	input := `(CallExpr :fun (Ident :namepos 10 :name "foo" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, err := builder.buildCallExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for missing fields")
	}
}

func TestBuildExprEmptyList(t *testing.T) {
	builder := New()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildExprNonSymbolFirst(t *testing.T) {
	builder := New()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

func TestBuildIdentWrongNodeType(t *testing.T) {
	builder := New()

	input := `(NotIdent :namepos 10 :name "foo" :obj nil)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildIdent(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildBasicLitWrongNodeType(t *testing.T) {
	builder := New()

	input := `(NotBasicLit :valuepos 10 :kind INT :value "42")`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildBasicLit(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildSelectorExprWrongNodeType(t *testing.T) {
	builder := New()

	input := `(NotSelectorExpr :x (Ident :namepos 10 :name "foo" :obj nil) :sel (Ident :namepos 14 :name "bar" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildSelectorExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildOptionalNilHandling(t *testing.T) {
	input := `nil`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()

	// Test optional ident
	ident, err := builder.buildOptionalIdent(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ident != nil {
		t.Fatalf("expected nil ident, got %v", ident)
	}

	// Test optional field list
	fieldList, err := builder.buildOptionalFieldList(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fieldList != nil {
		t.Fatalf("expected nil field list, got %v", fieldList)
	}

	// Test optional block stmt
	block, err := builder.buildOptionalBlockStmt(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block != nil {
		t.Fatalf("expected nil block stmt, got %v", block)
	}
}
