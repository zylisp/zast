package writer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"zylisp/zast/builder"
	"zylisp/zast/sexp"
)

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
	writer := New(fset)
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
	bldr := builder.New()
	fset2, files2, err := bldr.BuildProgram(sexpNode)
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

	writer := New(fset)
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

// TestWriteFile tests writing a single file
func TestWriteFile(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	astFile := &ast.File{
		Package: file.Pos(1),
		Name: &ast.Ident{
			NamePos: file.Pos(9),
			Name:    "test",
		},
		Decls: []ast.Decl{},
	}

	writer := New(fset)
	output, err := writer.WriteFile(astFile)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	if !contains(output, "File") {
		t.Fatalf("expected 'File' in output")
	}
	if !contains(output, "test") {
		t.Fatalf("expected 'test' in output")
	}
}

// TestWriteFileWithComments tests writing a file with comments
func TestWriteFileWithComments(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	astFile := &ast.File{
		Package: file.Pos(1),
		Name: &ast.Ident{
			NamePos: file.Pos(9),
			Name:    "test",
		},
		Decls: []ast.Decl{},
		Comments: []*ast.CommentGroup{
			{
				List: []*ast.Comment{
					{
						Slash: file.Pos(1),
						Text:  "// Package comment",
					},
				},
			},
		},
	}

	writer := New(fset)
	output, err := writer.WriteFile(astFile)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	if !contains(output, "File") {
		t.Fatalf("expected 'File' in output")
	}
	if !contains(output, "Package comment") {
		t.Fatalf("expected comment in output")
	}
}
