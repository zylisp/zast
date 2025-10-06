package writer

import (
	"go/ast"
	"testing"
)

func TestWriteDeclNil(t *testing.T) {
	writer := New(nil)
	err := writer.writeDecl(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteDeclUnknownType(t *testing.T) {
	writer := New(nil)

	// Create an unsupported decl type (e.g., BadDecl)
	badDecl := &ast.BadDecl{}

	err := writer.writeDecl(badDecl)
	if err == nil {
		t.Fatalf("expected error for unknown declaration type")
	}
}

func TestWriteValueSpec(t *testing.T) {
	w := New(nil)
	spec := &ast.ValueSpec{
		Names:  []*ast.Ident{{Name: "x"}},
		Type:   &ast.Ident{Name: "int"},
		Values: []ast.Expr{&ast.BasicLit{Value: "1"}},
	}
	err := w.writeValueSpec(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "ValueSpec") {
		t.Fatal("expected ValueSpec in output")
	}
}

func TestWriteTypeSpec(t *testing.T) {
	w := New(nil)
	spec := &ast.TypeSpec{
		Name: &ast.Ident{Name: "MyInt"},
		Type: &ast.Ident{Name: "int"},
	}
	err := w.writeTypeSpec(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "TypeSpec") {
		t.Fatal("expected TypeSpec in output")
	}
}
