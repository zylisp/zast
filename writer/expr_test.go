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

func TestWriteBinaryExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.BinaryExpr{
		X:  &ast.Ident{Name: "a"},
		Op: token.ADD,
		Y:  &ast.Ident{Name: "b"},
	}
	err := w.writeBinaryExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "BinaryExpr") {
		t.Fatal("expected BinaryExpr in output")
	}
}

func TestWriteUnaryExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.UnaryExpr{
		Op: token.SUB,
		X:  &ast.Ident{Name: "x"},
	}
	err := w.writeUnaryExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "UnaryExpr") {
		t.Fatal("expected UnaryExpr in output")
	}
}

func TestWriteParenExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.ParenExpr{X: &ast.Ident{Name: "x"}}
	err := w.writeParenExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "ParenExpr") {
		t.Fatal("expected ParenExpr in output")
	}
}

func TestWriteStarExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.StarExpr{X: &ast.Ident{Name: "int"}}
	err := w.writeStarExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "StarExpr") {
		t.Fatal("expected StarExpr in output")
	}
}

func TestWriteIndexExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "a"},
		Index: &ast.BasicLit{Kind: token.INT, Value: "0"},
	}
	err := w.writeIndexExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "IndexExpr") {
		t.Fatal("expected IndexExpr in output")
	}
}

func TestWriteSliceExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.SliceExpr{
		X:    &ast.Ident{Name: "a"},
		Low:  &ast.BasicLit{Kind: token.INT, Value: "1"},
		High: &ast.BasicLit{Kind: token.INT, Value: "3"},
	}
	err := w.writeSliceExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "SliceExpr") {
		t.Fatal("expected SliceExpr in output")
	}
}

func TestWriteTypeAssertExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.TypeAssertExpr{
		X:    &ast.Ident{Name: "x"},
		Type: &ast.Ident{Name: "int"},
	}
	err := w.writeTypeAssertExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "TypeAssertExpr") {
		t.Fatal("expected TypeAssertExpr in output")
	}
}

func TestWriteKeyValueExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.KeyValueExpr{
		Key:   &ast.BasicLit{Kind: token.STRING, Value: `"a"`},
		Value: &ast.BasicLit{Kind: token.INT, Value: "1"},
	}
	err := w.writeKeyValueExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "KeyValueExpr") {
		t.Fatal("expected KeyValueExpr in output")
	}
}

func TestWriteCompositeLit(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.CompositeLit{
		Type: &ast.Ident{Name: "Point"},
		Elts: []ast.Expr{},
	}
	err := w.writeCompositeLit(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "CompositeLit") {
		t.Fatal("expected CompositeLit in output")
	}
}

func TestWriteFuncLit(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.FuncLit{
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{},
	}
	err := w.writeFuncLit(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "FuncLit") {
		t.Fatal("expected FuncLit in output")
	}
}

func TestWriteEllipsis(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.Ellipsis{Elt: &ast.Ident{Name: "int"}}
	err := w.writeEllipsis(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "Ellipsis") {
		t.Fatal("expected Ellipsis in output")
	}
}

func TestWriteIndexListExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.IndexListExpr{
		X:       &ast.Ident{Name: "Generic"},
		Indices: []ast.Expr{&ast.Ident{Name: "int"}},
	}
	err := w.writeIndexListExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "IndexListExpr") {
		t.Fatal("expected IndexListExpr in output")
	}
}

func TestWriteBadExpr(t *testing.T) {
	fset := token.NewFileSet()
	w := New(fset)
	expr := &ast.BadExpr{From: 1, To: 10}
	err := w.writeBadExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "BadExpr") {
		t.Fatal("expected BadExpr in output")
	}
}
