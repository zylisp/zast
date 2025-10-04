package writer

import (
	"go/ast"
	"testing"
)

func TestWriteEmptyLists(t *testing.T) {
	writer := New(nil)

	// Test empty expression list
	err := writer.writeExprList([]ast.Expr{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	if output != "()" {
		t.Fatalf("expected %q, got %q", "()", output)
	}
}
