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
