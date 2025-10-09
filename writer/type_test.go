package writer

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestWriteFieldList(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	fieldList := &ast.FieldList{
		Opening: file.Pos(37),
		List:    []*ast.Field{},
		Closing: file.Pos(38),
	}

	writer := New(fset)
	err := writer.writeFieldList(fieldList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	if !contains(output, "FieldList") || !contains(output, ":list ()") {
		t.Fatalf("output missing expected elements: %q", output)
	}
}

func TestWriteFieldListNil(t *testing.T) {
	writer := New(nil)
	err := writer.writeFieldList(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	if output != "nil" {
		t.Fatalf("expected %q, got %q", "nil", output)
	}
}

func TestWriteField(t *testing.T) {
	w := New(nil)
	field := &ast.Field{
		Names: []*ast.Ident{{Name: "x"}},
		Type:  &ast.Ident{Name: "int"},
	}
	err := w.writeField(field)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "Field") {
		t.Fatal("expected Field in output")
	}
}

func TestWriteArrayType(t *testing.T) {
	w := New(nil)
	arrayType := &ast.ArrayType{
		Len: &ast.BasicLit{Kind: token.INT, Value: "10"},
		Elt: &ast.Ident{Name: "int"},
	}
	err := w.writeArrayType(arrayType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "ArrayType") {
		t.Fatal("expected ArrayType in output")
	}
}

func TestWriteMapType(t *testing.T) {
	w := New(nil)
	mapType := &ast.MapType{
		Key:   &ast.Ident{Name: "string"},
		Value: &ast.Ident{Name: "int"},
	}
	err := w.writeMapType(mapType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "MapType") {
		t.Fatal("expected MapType in output")
	}
}

func TestWriteChanType(t *testing.T) {
	w := New(nil)
	chanType := &ast.ChanType{
		Dir:   ast.SEND,
		Value: &ast.Ident{Name: "int"},
	}
	err := w.writeChanType(chanType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "ChanType") {
		t.Fatal("expected ChanType in output")
	}
}

func TestWriteStructType(t *testing.T) {
	w := New(nil)
	structType := &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{},
		},
		Incomplete: false,
	}
	err := w.writeStructType(structType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "StructType") {
		t.Fatal("expected StructType in output")
	}
}

func TestWriteInterfaceType(t *testing.T) {
	w := New(nil)
	interfaceType := &ast.InterfaceType{
		Methods: &ast.FieldList{
			List: []*ast.Field{},
		},
		Incomplete: false,
	}
	err := w.writeInterfaceType(interfaceType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "InterfaceType") {
		t.Fatal("expected InterfaceType in output")
	}
}
