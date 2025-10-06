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

// TestBuildArrayType tests array type building
func TestBuildArrayType(t *testing.T) {
	input := `(ArrayType
		:lbrack 1
		:len (BasicLit :valuepos 2 :kind INT :value "10")
		:elt (Ident :namepos 5 :name "int" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	arrayType, err := builder.buildArrayType(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if arrayType.Lbrack != token.Pos(1) {
		t.Fatalf("expected lbrack pos 1, got %v", arrayType.Lbrack)
	}
}

// TestBuildMapType tests map type building
func TestBuildMapType(t *testing.T) {
	input := `(MapType
		:map 1
		:key (Ident :namepos 5 :name "string" :obj nil)
		:value (Ident :namepos 12 :name "int" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	mapType, err := builder.buildMapType(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if mapType.Map != token.Pos(1) {
		t.Fatalf("expected map pos 1, got %v", mapType.Map)
	}
}

// TestBuildChanType tests channel type building
func TestBuildChanType(t *testing.T) {
	input := `(ChanType
		:begin 1
		:arrow 0
		:dir SEND
		:value (Ident :namepos 11 :name "int" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	chanType, err := builder.buildChanType(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if chanType.Begin != token.Pos(1) {
		t.Fatalf("expected begin pos 1, got %v", chanType.Begin)
	}
}

// TestBuildStructType tests struct type building
func TestBuildStructType(t *testing.T) {
	input := `(StructType
		:struct 1
		:fields (FieldList
			:opening 8
			:list ()
			:closing 9)
		:incomplete false)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	structType, err := builder.buildStructType(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if structType.Struct != token.Pos(1) {
		t.Fatalf("expected struct pos 1, got %v", structType.Struct)
	}

	if structType.Fields == nil {
		t.Fatalf("expected non-nil fields")
	}
}

// TestBuildInterfaceType tests interface type building
func TestBuildInterfaceType(t *testing.T) {
	input := `(InterfaceType
		:interface 1
		:methods (FieldList
			:opening 11
			:list ()
			:closing 12)
		:incomplete false)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	interfaceType, err := builder.buildInterfaceType(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if interfaceType.Interface != token.Pos(1) {
		t.Fatalf("expected interface pos 1, got %v", interfaceType.Interface)
	}

	if interfaceType.Methods == nil {
		t.Fatalf("expected non-nil methods")
	}
}
