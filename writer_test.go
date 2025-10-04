package zast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestWriteString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `"hello"`},
		{`hello "world"`, `"hello \"world\""`},
		{"hello\nworld", `"hello\nworld"`},
		{`hello\world`, `"hello\\world"`},
		{"tab\there", `"tab\there"`},
		{"cr\rhere", `"cr\rhere"`},
	}

	for _, tt := range tests {
		writer := NewWriter(nil)
		writer.writeString(tt.input)
		result := writer.buf.String()
		if result != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, result)
		}
	}
}

func TestWriteTokenMapping(t *testing.T) {
	tests := []struct {
		tok      token.Token
		expected string
	}{
		{token.IMPORT, "IMPORT"},
		{token.CONST, "CONST"},
		{token.TYPE, "TYPE"},
		{token.VAR, "VAR"},
		{token.INT, "INT"},
		{token.FLOAT, "FLOAT"},
		{token.IMAG, "IMAG"},
		{token.CHAR, "CHAR"},
		{token.STRING, "STRING"},
	}

	for _, tt := range tests {
		writer := NewWriter(nil)
		writer.writeToken(tt.tok)
		result := writer.buf.String()
		if result != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, result)
		}
	}
}

func TestWriteIdent(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	ident := &ast.Ident{
		NamePos: file.Pos(10),
		Name:    "main",
		Obj:     nil,
	}

	writer := NewWriter(fset)
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
	writer := NewWriter(nil)
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

	writer := NewWriter(fset)
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

	writer := NewWriter(fset)
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

	writer := NewWriter(fset)
	err := writer.writeCallExpr(call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	if !contains(output, "CallExpr") || !contains(output, "foo") || !contains(output, "hello") {
		t.Fatalf("output missing expected elements: %q", output)
	}
}

func TestWriteEmptyLists(t *testing.T) {
	writer := NewWriter(nil)

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

func TestWriteFieldList(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	fieldList := &ast.FieldList{
		Opening: file.Pos(37),
		List:    []*ast.Field{},
		Closing: file.Pos(38),
	}

	writer := NewWriter(fset)
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
	writer := NewWriter(nil)
	err := writer.writeFieldList(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := writer.buf.String()
	if output != "nil" {
		t.Fatalf("expected %q, got %q", "nil", output)
	}
}

// Test round-trip: Go source -> AST -> S-expr -> AST -> comparison
func TestRoundTrip(t *testing.T) {
	helloWorldSource := `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`

	// 1. Parse Go source to AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", helloWorldSource, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// 2. Write AST to S-expression
	writer := NewWriter(fset)
	sexpText, err := writer.WriteProgram([]*ast.File{file})
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Verify it's not empty
	if len(sexpText) == 0 {
		t.Fatalf("empty S-expression output")
	}

	// 3. Parse S-expression
	sexpr := sexp.NewParser(sexpText)
	sexpNode, err := sexpr.Parse()
	if err != nil {
		t.Fatalf("sexp parse error: %v", err)
	}

	// Verify it's a Program
	list, ok := sexpNode.(*sexp.List)
	if !ok {
		t.Fatalf("expected list, got %T", sexpNode)
	}

	if len(list.Elements) == 0 {
		t.Fatalf("empty list")
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok || sym.Value != "Program" {
		t.Fatalf("expected Program symbol, got %v", list.Elements[0])
	}

	// 4. Build back to AST
	builder := NewBuilder()
	fset2, files2, err := builder.BuildProgram(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	// 5. Verify structure
	if fset2 == nil {
		t.Fatalf("nil fileset")
	}

	if len(files2) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files2))
	}

	if files2[0].Name.Name != "main" {
		t.Fatalf("expected package name %q, got %q", "main", files2[0].Name.Name)
	}

	// Verify declarations
	if len(files2[0].Decls) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(files2[0].Decls))
	}

	// Verify import
	genDecl, ok := files2[0].Decls[0].(*ast.GenDecl)
	if !ok || genDecl.Tok != token.IMPORT {
		t.Fatalf("expected import declaration")
	}

	// Verify function
	funcDecl, ok := files2[0].Decls[1].(*ast.FuncDecl)
	if !ok || funcDecl.Name.Name != "main" {
		t.Fatalf("expected main function")
	}
}

// Test writing and parsing back a minimal program
func TestWriteMinimalProgram(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	// Create minimal AST
	astFile := &ast.File{
		Package: file.Pos(1),
		Name: &ast.Ident{
			NamePos: file.Pos(9),
			Name:    "main",
		},
		Decls: []ast.Decl{},
	}

	writer := NewWriter(fset)
	output, err := writer.WriteProgram([]*ast.File{astFile})
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Parse it back
	sexpr := sexp.NewParser(output)
	sexpNode, err := sexpr.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Verify structure
	list, ok := sexpNode.(*sexp.List)
	if !ok {
		t.Fatalf("expected list")
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok || sym.Value != "Program" {
		t.Fatalf("expected Program")
	}
}

func TestWriteExprNil(t *testing.T) {
	writer := NewWriter(nil)
	err := writer.writeExpr(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteStmtNil(t *testing.T) {
	writer := NewWriter(nil)
	err := writer.writeStmt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteDeclNil(t *testing.T) {
	writer := NewWriter(nil)
	err := writer.writeDecl(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteBlockStmtNil(t *testing.T) {
	writer := NewWriter(nil)
	err := writer.writeBlockStmt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.buf.String() != "nil" {
		t.Fatalf("expected %q, got %q", "nil", writer.buf.String())
	}
}

func TestWriteExprUnknownType(t *testing.T) {
	writer := NewWriter(nil)

	// Create an unsupported expression type (e.g., FuncLit - not yet implemented)
	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{},
		Body: &ast.BlockStmt{},
	}

	err := writer.writeExpr(funcLit)
	if err == nil {
		t.Fatalf("expected error for unknown expression type")
	}
}

func TestWriteStmtUnknownType(t *testing.T) {
	writer := NewWriter(nil)

	// Create an unsupported statement type (BadStmt is not implemented)
	badStmt := &ast.BadStmt{}

	err := writer.writeStmt(badStmt)
	if err == nil {
		t.Fatalf("expected error for unknown statement type")
	}
}

func TestWriteDeclUnknownType(t *testing.T) {
	writer := NewWriter(nil)

	// Create an unsupported decl type (e.g., BadDecl)
	badDecl := &ast.BadDecl{}

	err := writer.writeDecl(badDecl)
	if err == nil {
		t.Fatalf("expected error for unknown declaration type")
	}
}

// Note: TestWriteSpecUnknownType removed because all spec types (ImportSpec, ValueSpec, TypeSpec)
// are now supported in Phase 2.

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
