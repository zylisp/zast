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
