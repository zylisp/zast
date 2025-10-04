package writer

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestWriteIdent(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	ident := &ast.Ident{
		NamePos: file.Pos(10),
		Name:    "main",
		Obj:     nil,
	}

	writer := New(fset)
	err := writer.writeIdent(ident)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	expected := `(Ident :namepos 11 :name "main" :obj nil)`
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestWriteIdentNil(t *testing.T) {
	writer := New(nil)
	err := writer.writeIdent(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	if output != "nil" {
		t.Fatalf("expected %q, got %q", "nil", output)
	}
}

func TestWriteBasicLit(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	lit := &ast.BasicLit{
		ValuePos: file.Pos(22),
		Kind:     token.STRING,
		Value:    `"fmt"`,
	}

	writer := New(fset)
	err := writer.writeBasicLit(lit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	expected := `(BasicLit :valuepos 23 :kind STRING :value "\"fmt\"")`
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestWriteSelectorExpr(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	sel := &ast.SelectorExpr{
		X: &ast.Ident{
			NamePos: file.Pos(46),
			Name:    "fmt",
		},
		Sel: &ast.Ident{
			NamePos: file.Pos(50),
			Name:    "Println",
		},
	}

	writer := New(fset)
	err := writer.writeSelectorExpr(sel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	// Should contain the nested structure
	if !contains(output, "SelectorExpr") || !contains(output, "fmt") || !contains(output, "Println") {
		t.Fatalf("output missing expected elements: %q", output)
	}
}

func TestWriteCallExpr(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	call := &ast.CallExpr{
		Fun: &ast.Ident{
			NamePos: file.Pos(10),
			Name:    "foo",
		},
		Lparen: file.Pos(13),
		Args: []ast.Expr{
			&ast.BasicLit{
				ValuePos: file.Pos(14),
				Kind:     token.STRING,
				Value:    `"hello"`,
			},
		},
		Ellipsis: 0,
		Rparen:   file.Pos(21),
	}

	writer := New(fset)
	err := writer.writeCallExpr(call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	if !contains(output, "CallExpr") || !contains(output, "foo") || !contains(output, "hello") {
		t.Fatalf("output missing expected elements: %q", output)
	}
}

func TestWriteExprNil(t *testing.T) {
	writer := New(nil)
	err := writer.writeExpr(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteExprUnknownType(t *testing.T) {
	writer := New(nil)

	// Create an unsupported expression type (BadExpr is never implemented)
	badExpr := &ast.BadExpr{}

	err := writer.writeExpr(badExpr)
	if err == nil {
		t.Fatalf("expected error for unknown expression type")
	}
}
