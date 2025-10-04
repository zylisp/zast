package builder

import (
	"go/ast"
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestBuildExprStmt(t *testing.T) {
	input := `(ExprStmt :x (Ident :namepos 10 :name "foo" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	stmt, err := builder.buildExprStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	ident, ok := stmt.X.(*ast.Ident)
	if !ok || ident.Name != "foo" {
		t.Fatalf("expected X to be Ident 'foo', got %v", stmt.X)
	}
}

func TestBuildBlockStmt(t *testing.T) {
	input := `(BlockStmt
		:lbrace 40
		:list ((ExprStmt :x (Ident :namepos 46 :name "foo" :obj nil)))
		:rbrace 76)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	block, err := builder.buildBlockStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if block.Lbrace != token.Pos(40) {
		t.Fatalf("expected lbrace %d, got %d", 40, block.Lbrace)
	}

	if block.Rbrace != token.Pos(76) {
		t.Fatalf("expected rbrace %d, got %d", 76, block.Rbrace)
	}

	if len(block.List) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(block.List))
	}
}

func TestBuildStmtUnknownType(t *testing.T) {
	input := `(UnknownStmt :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown statement type")
	}
}

func TestBuildStmtEmptyList(t *testing.T) {
	builder := New()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildStmtNonSymbolFirst(t *testing.T) {
	builder := New()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}
