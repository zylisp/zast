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

// TestBuildBinaryExpr tests binary expression building
func TestBuildBinaryExpr(t *testing.T) {
	input := `(BinaryExpr
		:x (Ident :namepos 1 :name "a" :obj nil)
		:oppos 3
		:op ADD
		:y (Ident :namepos 5 :name "b" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	binExpr, err := builder.buildBinaryExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if binExpr.Op != token.ADD {
		t.Fatalf("expected op ADD, got %v", binExpr.Op)
	}

	xIdent, ok := binExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "a" {
		t.Fatalf("expected X to be Ident 'a', got %v", binExpr.X)
	}

	yIdent, ok := binExpr.Y.(*ast.Ident)
	if !ok || yIdent.Name != "b" {
		t.Fatalf("expected Y to be Ident 'b', got %v", binExpr.Y)
	}
}

// TestBuildUnaryExpr tests unary expression building
func TestBuildUnaryExpr(t *testing.T) {
	input := `(UnaryExpr
		:oppos 1
		:op SUB
		:x (Ident :namepos 2 :name "x" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	unaryExpr, err := builder.buildUnaryExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if unaryExpr.Op != token.SUB {
		t.Fatalf("expected op SUB, got %v", unaryExpr.Op)
	}

	xIdent, ok := unaryExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "x" {
		t.Fatalf("expected X to be Ident 'x', got %v", unaryExpr.X)
	}
}

// TestBuildParenExpr tests parenthesized expression building
func TestBuildParenExpr(t *testing.T) {
	input := `(ParenExpr
		:lparen 1
		:x (Ident :namepos 2 :name "x" :obj nil)
		:rparen 3)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	parenExpr, err := builder.buildParenExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := parenExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "x" {
		t.Fatalf("expected X to be Ident 'x', got %v", parenExpr.X)
	}
}

// TestBuildStarExpr tests star expression building
func TestBuildStarExpr(t *testing.T) {
	input := `(StarExpr
		:star 1
		:x (Ident :namepos 2 :name "int" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	starExpr, err := builder.buildStarExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := starExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "int" {
		t.Fatalf("expected X to be Ident 'int', got %v", starExpr.X)
	}
}

// TestBuildIndexExpr tests index expression building
func TestBuildIndexExpr(t *testing.T) {
	input := `(IndexExpr
		:x (Ident :namepos 1 :name "a" :obj nil)
		:lbrack 2
		:index (BasicLit :valuepos 3 :kind INT :value "0")
		:rbrack 4)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	indexExpr, err := builder.buildIndexExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := indexExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "a" {
		t.Fatalf("expected X to be Ident 'a', got %v", indexExpr.X)
	}

	indexLit, ok := indexExpr.Index.(*ast.BasicLit)
	if !ok || indexLit.Value != "0" {
		t.Fatalf("expected Index to be BasicLit '0', got %v", indexExpr.Index)
	}
}

// TestBuildSliceExpr tests slice expression building
func TestBuildSliceExpr(t *testing.T) {
	input := `(SliceExpr
		:x (Ident :namepos 1 :name "a" :obj nil)
		:lbrack 2
		:low (BasicLit :valuepos 3 :kind INT :value "1")
		:high (BasicLit :valuepos 5 :kind INT :value "3")
		:max nil
		:slice3 false
		:rbrack 6)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	sliceExpr, err := builder.buildSliceExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := sliceExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "a" {
		t.Fatalf("expected X to be Ident 'a', got %v", sliceExpr.X)
	}

	lowLit, ok := sliceExpr.Low.(*ast.BasicLit)
	if !ok || lowLit.Value != "1" {
		t.Fatalf("expected Low to be BasicLit '1', got %v", sliceExpr.Low)
	}

	highLit, ok := sliceExpr.High.(*ast.BasicLit)
	if !ok || highLit.Value != "3" {
		t.Fatalf("expected High to be BasicLit '3', got %v", sliceExpr.High)
	}
}

// TestBuildTypeAssertExpr tests type assertion expression building
func TestBuildTypeAssertExpr(t *testing.T) {
	input := `(TypeAssertExpr
		:x (Ident :namepos 1 :name "x" :obj nil)
		:lparen 2
		:type (Ident :namepos 3 :name "int" :obj nil)
		:rparen 6)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	typeAssert, err := builder.buildTypeAssertExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := typeAssert.X.(*ast.Ident)
	if !ok || xIdent.Name != "x" {
		t.Fatalf("expected X to be Ident 'x', got %v", typeAssert.X)
	}

	typeIdent, ok := typeAssert.Type.(*ast.Ident)
	if !ok || typeIdent.Name != "int" {
		t.Fatalf("expected Type to be Ident 'int', got %v", typeAssert.Type)
	}
}

// TestBuildKeyValueExpr tests key-value expression building
func TestBuildKeyValueExpr(t *testing.T) {
	input := `(KeyValueExpr
		:key (BasicLit :valuepos 1 :kind STRING :value "\"a\"")
		:colon 4
		:value (BasicLit :valuepos 6 :kind INT :value "1"))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	kvExpr, err := builder.buildKeyValueExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	keyLit, ok := kvExpr.Key.(*ast.BasicLit)
	if !ok || keyLit.Value != `"a"` {
		t.Fatalf("expected Key to be BasicLit '\"a\"', got %v", kvExpr.Key)
	}

	valueLit, ok := kvExpr.Value.(*ast.BasicLit)
	if !ok || valueLit.Value != "1" {
		t.Fatalf("expected Value to be BasicLit '1', got %v", kvExpr.Value)
	}
}

// TestBuildCompositeLit tests composite literal building
func TestBuildCompositeLit(t *testing.T) {
	input := `(CompositeLit
		:type (Ident :namepos 1 :name "Point" :obj nil)
		:lbrace 6
		:elts ()
		:rbrace 7
		:incomplete false)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	compLit, err := builder.buildCompositeLit(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	typeIdent, ok := compLit.Type.(*ast.Ident)
	if !ok || typeIdent.Name != "Point" {
		t.Fatalf("expected Type to be Ident 'Point', got %v", compLit.Type)
	}

	if len(compLit.Elts) != 0 {
		t.Fatalf("expected empty Elts, got %d elements", len(compLit.Elts))
	}
}

// TestBuildFuncLit tests function literal building
func TestBuildFuncLit(t *testing.T) {
	input := `(FuncLit
		:type (FuncType
			:func 1
			:typeparams nil
			:params (FieldList :opening 5 :list () :closing 6)
			:results nil)
		:body (BlockStmt :lbrace 8 :list () :rbrace 9))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	funcLit, err := builder.buildFuncLit(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if funcLit.Type == nil {
		t.Fatalf("expected Type to be non-nil")
	}

	if funcLit.Body == nil {
		t.Fatalf("expected Body to be non-nil")
	}
}

// TestBuildEllipsis tests ellipsis building
func TestBuildEllipsis(t *testing.T) {
	input := `(Ellipsis
		:ellipsis 1
		:elt (Ident :namepos 4 :name "int" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	ellipsis, err := builder.buildEllipsis(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	eltIdent, ok := ellipsis.Elt.(*ast.Ident)
	if !ok || eltIdent.Name != "int" {
		t.Fatalf("expected Elt to be Ident 'int', got %v", ellipsis.Elt)
	}
}

// TestBuildIndexListExpr tests index list expression building
func TestBuildIndexListExpr(t *testing.T) {
	input := `(IndexListExpr
		:x (Ident :namepos 1 :name "Generic" :obj nil)
		:lbrack 8
		:indices ((Ident :namepos 9 :name "int" :obj nil) (Ident :namepos 14 :name "string" :obj nil))
		:rbrack 20)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	indexList, err := builder.buildIndexListExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := indexList.X.(*ast.Ident)
	if !ok || xIdent.Name != "Generic" {
		t.Fatalf("expected X to be Ident 'Generic', got %v", indexList.X)
	}

	if len(indexList.Indices) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(indexList.Indices))
	}
}

// TestBuildBadExpr tests bad expression building
func TestBuildBadExpr(t *testing.T) {
	input := `(BadExpr :from 1 :to 10)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	badExpr, err := builder.buildBadExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if badExpr.From != token.Pos(1) {
		t.Fatalf("expected From to be 1, got %v", badExpr.From)
	}

	if badExpr.To != token.Pos(10) {
		t.Fatalf("expected To to be 10, got %v", badExpr.To)
	}
}
