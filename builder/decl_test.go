package builder

import (
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestBuildImportSpec(t *testing.T) {
	input := `(ImportSpec
		:doc nil
		:name nil
		:path (BasicLit :valuepos 22 :kind STRING :value "\"fmt\"")
		:comment nil
		:endpos 27)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	spec, err := builder.buildImportSpec(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if spec.Name != nil {
		t.Fatalf("expected nil name, got %v", spec.Name)
	}

	if spec.Path.Value != `"fmt"` {
		t.Fatalf("expected path %q, got %q", `"fmt"`, spec.Path.Value)
	}

	if spec.EndPos != token.Pos(27) {
		t.Fatalf("expected endpos %d, got %d", 27, spec.EndPos)
	}
}

func TestBuildGenDecl(t *testing.T) {
	input := `(GenDecl
		:doc nil
		:tok IMPORT
		:tokpos 15
		:lparen 0
		:specs ((ImportSpec :doc nil :name nil :path (BasicLit :valuepos 22 :kind STRING :value "\"fmt\"") :comment nil :endpos 27))
		:rparen 0)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	decl, err := builder.buildGenDecl(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if decl.Tok != token.IMPORT {
		t.Fatalf("expected tok IMPORT, got %v", decl.Tok)
	}

	if decl.TokPos != token.Pos(15) {
		t.Fatalf("expected tokpos %d, got %d", 15, decl.TokPos)
	}

	if len(decl.Specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(decl.Specs))
	}
}

func TestBuildFuncDecl(t *testing.T) {
	input := `(FuncDecl
		:doc nil
		:recv nil
		:name (Ident :namepos 33 :name "main" :obj nil)
		:type (FuncType :func 28 :params (FieldList :opening 37 :list () :closing 38) :results nil)
		:body (BlockStmt :lbrace 40 :list () :rbrace 76))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	decl, err := builder.buildFuncDecl(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if decl.Name.Name != "main" {
		t.Fatalf("expected name %q, got %q", "main", decl.Name.Name)
	}

	if decl.Recv != nil {
		t.Fatalf("expected nil recv, got %v", decl.Recv)
	}

	if decl.Type == nil {
		t.Fatalf("expected non-nil type")
	}

	if decl.Body == nil {
		t.Fatalf("expected non-nil body")
	}
}

func TestBuildDeclUnknownType(t *testing.T) {
	input := `(UnknownDecl :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, err := builder.buildDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown declaration type")
	}
}

func TestBuildSpecUnknownType(t *testing.T) {
	input := `(UnknownSpec :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, err := builder.buildSpec(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown spec type")
	}
}

func TestBuildFuncDeclMissingFields(t *testing.T) {
	input := `(FuncDecl :name (Ident :namepos 10 :name "foo" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, err := builder.buildFuncDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for missing type field")
	}
}

func TestBuildDeclEmptyList(t *testing.T) {
	builder := New()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildSpecNonSymbolFirst(t *testing.T) {
	builder := New()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildSpec(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

func TestBuildDeclNonSymbolFirst(t *testing.T) {
	builder := New()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

func TestBuildSpecEmptyList(t *testing.T) {
	builder := New()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildSpec(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

// TestBuildValueSpec tests value spec building
func TestBuildValueSpec(t *testing.T) {
	input := `(ValueSpec
		:doc nil
		:names ((Ident :namepos 5 :name "x" :obj nil))
		:type (Ident :namepos 7 :name "int" :obj nil)
		:values ((BasicLit :valuepos 11 :kind INT :value "1"))
		:comment nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	spec, err := builder.buildValueSpec(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if len(spec.Names) != 1 || spec.Names[0].Name != "x" {
		t.Fatalf("expected name 'x', got %v", spec.Names)
	}

	if len(spec.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(spec.Values))
	}
}

// TestBuildTypeSpec tests type spec building
func TestBuildTypeSpec(t *testing.T) {
	input := `(TypeSpec
		:doc nil
		:name (Ident :namepos 6 :name "MyInt" :obj nil)
		:typeparams nil
		:assign 0
		:type (Ident :namepos 12 :name "int" :obj nil)
		:comment nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	spec, err := builder.buildTypeSpec(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if spec.Name.Name != "MyInt" {
		t.Fatalf("expected name 'MyInt', got %v", spec.Name.Name)
	}
}
