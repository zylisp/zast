package writer

import (
	"go/ast"
	"testing"
)

func TestWriteStmtNil(t *testing.T) {
	writer := New(nil)
	err := writer.writeStmt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteBlockStmtNil(t *testing.T) {
	writer := New(nil)
	err := writer.writeBlockStmt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteStmtUnknownType(t *testing.T) {
	writer := New(nil)

	// Create an unsupported statement type (BadStmt is not implemented)
	badStmt := &ast.BadStmt{}

	err := writer.writeStmt(badStmt)
	if err == nil {
		t.Fatalf("expected error for unknown statement type")
	}
}
