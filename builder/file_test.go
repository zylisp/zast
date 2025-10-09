package builder

import (
	"go/ast"
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestBuildFileInfo(t *testing.T) {
	input := `(FileInfo :name "main.go" :base 1 :size 78 :lines (1 14 27 42 78))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	fileInfo, err := builder.buildFileInfo(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if fileInfo.Name != "main.go" {
		t.Fatalf("expected name %q, got %q", "main.go", fileInfo.Name)
	}

	if fileInfo.Base != 1 {
		t.Fatalf("expected base %d, got %d", 1, fileInfo.Base)
	}

	if fileInfo.Size != 78 {
		t.Fatalf("expected size %d, got %d", 78, fileInfo.Size)
	}

	expectedLines := []int{1, 14, 27, 42, 78}
	if len(fileInfo.Lines) != len(expectedLines) {
		t.Fatalf("expected %d lines, got %d", len(expectedLines), len(fileInfo.Lines))
	}

	for i, expected := range expectedLines {
		if fileInfo.Lines[i] != expected {
			t.Fatalf("expected line[%d] = %d, got %d", i, expected, fileInfo.Lines[i])
		}
	}
}

func TestBuildFileSet(t *testing.T) {
	input := `(FileSet
		:base 1
		:files ((FileInfo :name "main.go" :base 1 :size 78 :lines (1 14 27))))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	fileSetInfo, err := builder.buildFileSet(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	// FileSet parsing is now simplified - we don't preserve position information
	// The FileSetInfo is minimal since positions can't be preserved through round-trip
	if fileSetInfo.Base != 1 {
		t.Fatalf("expected base %d, got %d", 1, fileSetInfo.Base)
	}

	// Files list should be nil since we don't preserve file metadata
	if fileSetInfo.Files != nil {
		t.Fatalf("expected nil files, got %d files", len(fileSetInfo.Files))
	}
}

func TestBuildFile(t *testing.T) {
	input := `(File
		:package 1
		:name (Ident :namepos 9 :name "main" :obj nil)
		:decls ((GenDecl :doc nil :tok IMPORT :tokpos 15 :lparen 0 :specs ((ImportSpec :doc nil :name nil :path (BasicLit :valuepos 22 :kind STRING :value "\"fmt\"") :comment nil :endpos 27)) :rparen 0))
		:scope nil
		:imports ()
		:unresolved ()
		:comments ())`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	file, err := builder.BuildFile(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if file.Package != token.Pos(1) {
		t.Fatalf("expected package pos %d, got %d", 1, file.Package)
	}

	if file.Name.Name != "main" {
		t.Fatalf("expected name %q, got %q", "main", file.Name.Name)
	}

	if len(file.Decls) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(file.Decls))
	}
}

func TestBuildCompleteProgram(t *testing.T) {
	input := `(Program
		:fileset (FileSet :base 1 :files ((FileInfo :name "main.go" :base 1 :size 78 :lines (1 14 27))))
		:files ((File
			:package 1
			:name (Ident :namepos 9 :name "main" :obj nil)
			:decls ((GenDecl :doc nil :tok IMPORT :tokpos 15 :lparen 0 :specs ((ImportSpec :doc nil :name nil :path (BasicLit :valuepos 22 :kind STRING :value "\"fmt\"") :comment nil :endpos 27)) :rparen 0))
			:scope nil
			:imports ()
			:unresolved ()
			:comments ())))`

	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	fset, files, err := builder.BuildProgram(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if fset == nil {
		t.Fatalf("expected non-nil fileset")
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].Name.Name != "main" {
		t.Fatalf("expected file name %q, got %q", "main", files[0].Name.Name)
	}
}

// Test full Hello World program
func TestBuildHelloWorld(t *testing.T) {
	input := `(Program
		:fileset (FileSet
			:base 1
			:files ((FileInfo
				:name "main.go"
				:base 1
				:size 78
				:lines (1 14 27 42 78))))
		:files ((File
			:package 1
			:name (Ident :namepos 9 :name "main" :obj nil)
			:decls (
				(GenDecl
					:doc nil
					:tok IMPORT
					:tokpos 15
					:lparen 0
					:specs ((ImportSpec
						:doc nil
						:name nil
						:path (BasicLit :valuepos 22 :kind STRING :value "\"fmt\"")
						:comment nil
						:endpos 27))
					:rparen 0)
				(FuncDecl
					:doc nil
					:recv nil
					:name (Ident :namepos 33 :name "main" :obj nil)
					:type (FuncType
						:func 28
						:params (FieldList :opening 37 :list () :closing 38)
						:results nil)
					:body (BlockStmt
						:lbrace 40
						:list ((ExprStmt
							:x (CallExpr
								:fun (SelectorExpr
									:x (Ident :namepos 46 :name "fmt" :obj nil)
									:sel (Ident :namepos 50 :name "Println" :obj nil))
								:lparen 57
								:args ((BasicLit :valuepos 58 :kind STRING :value "\"Hello, world!\""))
								:ellipsis 0
								:rparen 74)))
						:rbrace 76)))
			:scope nil
			:imports ()
			:unresolved ()
			:comments ())))`

	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	fset, files, err := builder.BuildProgram(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	// Verify FileSet
	if fset == nil {
		t.Fatalf("expected non-nil fileset")
	}

	// Verify we have 1 file
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	file := files[0]

	// Verify package name
	if file.Name.Name != "main" {
		t.Fatalf("expected package name %q, got %q", "main", file.Name.Name)
	}

	// Verify we have 2 declarations (import and func)
	if len(file.Decls) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(file.Decls))
	}

	// Verify first decl is GenDecl (import)
	genDecl, ok := file.Decls[0].(*ast.GenDecl)
	if !ok {
		t.Fatalf("expected first decl to be GenDecl, got %T", file.Decls[0])
	}
	if genDecl.Tok != token.IMPORT {
		t.Fatalf("expected IMPORT token, got %v", genDecl.Tok)
	}

	// Verify second decl is FuncDecl (main)
	funcDecl, ok := file.Decls[1].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected second decl to be FuncDecl, got %T", file.Decls[1])
	}
	if funcDecl.Name.Name != "main" {
		t.Fatalf("expected func name %q, got %q", "main", funcDecl.Name.Name)
	}

	// Verify function body has 1 statement
	if len(funcDecl.Body.List) != 1 {
		t.Fatalf("expected 1 statement in body, got %d", len(funcDecl.Body.List))
	}

	// Verify statement is ExprStmt with CallExpr
	exprStmt, ok := funcDecl.Body.List[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", funcDecl.Body.List[0])
	}

	callExpr, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.X)
	}

	// Verify it's fmt.Println
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", callExpr.Fun)
	}

	xIdent, ok := selExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "fmt" {
		t.Fatalf("expected selector X to be 'fmt', got %v", selExpr.X)
	}

	if selExpr.Sel.Name != "Println" {
		t.Fatalf("expected selector Sel to be 'Println', got %q", selExpr.Sel.Name)
	}

	// Verify argument
	if len(callExpr.Args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(callExpr.Args))
	}

	argLit, ok := callExpr.Args[0].(*ast.BasicLit)
	if !ok || argLit.Value != `"Hello, world!"` {
		t.Fatalf("expected argument to be 'Hello, world!', got %v", callExpr.Args[0])
	}
}

func TestBuildProgramWrongNodeType(t *testing.T) {
	builder := New()

	input := `(NotProgram :fileset (FileSet :base 1 :files ()) :files ())`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, _, err := builder.BuildProgram(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildFileWrongNodeType(t *testing.T) {
	builder := New()

	input := `(NotFile :package 1 :name (Ident :namepos 9 :name "main" :obj nil) :decls () :scope nil :imports () :unresolved () :comments ())`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.BuildFile(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}
