package builder

import (
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestBuildFieldList(t *testing.T) {
	input := `(FieldList
		:opening 37
		:list ()
		:closing 38)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	fieldList, err := builder.buildFieldList(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if fieldList.Opening != token.Pos(37) {
		t.Fatalf("expected opening %d, got %d", 37, fieldList.Opening)
	}

	if fieldList.Closing != token.Pos(38) {
		t.Fatalf("expected closing %d, got %d", 38, fieldList.Closing)
	}

	if len(fieldList.List) != 0 {
		t.Fatalf("expected empty field list, got %d fields", len(fieldList.List))
	}
}

func TestBuildFuncType(t *testing.T) {
	input := `(FuncType
		:func 28
		:params (FieldList :opening 37 :list () :closing 38)
		:results nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	funcType, err := builder.buildFuncType(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if funcType.Func != token.Pos(28) {
		t.Fatalf("expected func pos %d, got %d", 28, funcType.Func)
	}

	if funcType.Params == nil {
		t.Fatalf("expected non-nil params")
	}

	if funcType.Results != nil {
		t.Fatalf("expected nil results, got %v", funcType.Results)
	}
}

// Test buildField with names
func TestBuildFieldWithNames(t *testing.T) {
	input := `(Field
		:doc nil
		:names ((Ident :namepos 10 :name "x" :obj nil) (Ident :namepos 13 :name "y" :obj nil))
		:type (Ident :namepos 15 :name "int" :obj nil)
		:tag nil
		:comment nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	field, err := builder.buildField(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if len(field.Names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(field.Names))
	}

	if field.Names[0].Name != "x" {
		t.Fatalf("expected first name %q, got %q", "x", field.Names[0].Name)
	}

	if field.Names[1].Name != "y" {
		t.Fatalf("expected second name %q, got %q", "y", field.Names[1].Name)
	}
}
