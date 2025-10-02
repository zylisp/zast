package zast

import (
	"fmt"
	"go/ast"
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestBuildIdent(t *testing.T) {
	input := `(Ident :namepos 10 :name "main" :obj nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	ident, err := builder.buildIdent(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if ident.Name != "main" {
		t.Fatalf("expected name %q, got %q", "main", ident.Name)
	}

	if ident.NamePos != token.Pos(10) {
		t.Fatalf("expected namepos %d, got %d", 10, ident.NamePos)
	}

	if ident.Obj != nil {
		t.Fatalf("expected nil obj, got %v", ident.Obj)
	}
}

func TestBuildBasicLit(t *testing.T) {
	input := `(BasicLit :valuepos 22 :kind STRING :value "\"fmt\"")`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	lit, err := builder.buildBasicLit(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if lit.Kind != token.STRING {
		t.Fatalf("expected kind STRING, got %v", lit.Kind)
	}

	if lit.Value != `"fmt"` {
		t.Fatalf("expected value %q, got %q", `"fmt"`, lit.Value)
	}

	if lit.ValuePos != token.Pos(22) {
		t.Fatalf("expected valuepos %d, got %d", 22, lit.ValuePos)
	}
}

func TestBuildSelectorExpr(t *testing.T) {
	input := `(SelectorExpr
		:x (Ident :namepos 46 :name "fmt" :obj nil)
		:sel (Ident :namepos 50 :name "Println" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	sel, err := builder.buildSelectorExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	xIdent, ok := sel.X.(*ast.Ident)
	if !ok || xIdent.Name != "fmt" {
		t.Fatalf("expected X to be Ident 'fmt', got %v", sel.X)
	}

	if sel.Sel.Name != "Println" {
		t.Fatalf("expected Sel name %q, got %q", "Println", sel.Sel.Name)
	}
}

func TestBuildCallExpr(t *testing.T) {
	input := `(CallExpr
		:fun (Ident :namepos 10 :name "foo" :obj nil)
		:lparen 13
		:args ((BasicLit :valuepos 14 :kind STRING :value "\"hello\""))
		:ellipsis 0
		:rparen 21)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	call, err := builder.buildCallExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	funIdent, ok := call.Fun.(*ast.Ident)
	if !ok || funIdent.Name != "foo" {
		t.Fatalf("expected Fun to be Ident 'foo', got %v", call.Fun)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	argLit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || argLit.Value != `"hello"` {
		t.Fatalf("expected arg to be BasicLit '\"hello\"', got %v", call.Args[0])
	}
}

func TestBuildExprStmt(t *testing.T) {
	input := `(ExprStmt :x (Ident :namepos 10 :name "foo" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	stmt, err := builder.buildExprStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	ident, ok := stmt.X.(*ast.Ident)
	if !ok || ident.Name != "foo" {
		t.Fatalf("expected X to be Ident 'foo', got %v", stmt.X)
	}
}

func TestBuildBlockStmt(t *testing.T) {
	input := `(BlockStmt
		:lbrace 40
		:list ((ExprStmt :x (Ident :namepos 46 :name "foo" :obj nil)))
		:rbrace 76)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	block, err := builder.buildBlockStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if block.Lbrace != token.Pos(40) {
		t.Fatalf("expected lbrace %d, got %d", 40, block.Lbrace)
	}

	if block.Rbrace != token.Pos(76) {
		t.Fatalf("expected rbrace %d, got %d", 76, block.Rbrace)
	}

	if len(block.List) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(block.List))
	}
}

func TestBuildFieldList(t *testing.T) {
	input := `(FieldList
		:opening 37
		:list ()
		:closing 38)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	fieldList, err := builder.buildFieldList(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if fieldList.Opening != token.Pos(37) {
		t.Fatalf("expected opening %d, got %d", 37, fieldList.Opening)
	}

	if fieldList.Closing != token.Pos(38) {
		t.Fatalf("expected closing %d, got %d", 38, fieldList.Closing)
	}

	if len(fieldList.List) != 0 {
		t.Fatalf("expected empty field list, got %d fields", len(fieldList.List))
	}
}

func TestBuildFuncType(t *testing.T) {
	input := `(FuncType
		:func 28
		:params (FieldList :opening 37 :list () :closing 38)
		:results nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	funcType, err := builder.buildFuncType(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if funcType.Func != token.Pos(28) {
		t.Fatalf("expected func pos %d, got %d", 28, funcType.Func)
	}

	if funcType.Params == nil {
		t.Fatalf("expected non-nil params")
	}

	if funcType.Results != nil {
		t.Fatalf("expected nil results, got %v", funcType.Results)
	}
}

func TestBuildImportSpec(t *testing.T) {
	input := `(ImportSpec
		:doc nil
		:name nil
		:path (BasicLit :valuepos 22 :kind STRING :value "\"fmt\"")
		:comment nil
		:endpos 27)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	spec, err := builder.buildImportSpec(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if spec.Name != nil {
		t.Fatalf("expected nil name, got %v", spec.Name)
	}

	if spec.Path.Value != `"fmt"` {
		t.Fatalf("expected path %q, got %q", `"fmt"`, spec.Path.Value)
	}

	if spec.EndPos != token.Pos(27) {
		t.Fatalf("expected endpos %d, got %d", 27, spec.EndPos)
	}
}

func TestBuildGenDecl(t *testing.T) {
	input := `(GenDecl
		:doc nil
		:tok IMPORT
		:tokpos 15
		:lparen 0
		:specs ((ImportSpec :doc nil :name nil :path (BasicLit :valuepos 22 :kind STRING :value "\"fmt\"") :comment nil :endpos 27))
		:rparen 0)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	decl, err := builder.buildGenDecl(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if decl.Tok != token.IMPORT {
		t.Fatalf("expected tok IMPORT, got %v", decl.Tok)
	}

	if decl.TokPos != token.Pos(15) {
		t.Fatalf("expected tokpos %d, got %d", 15, decl.TokPos)
	}

	if len(decl.Specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(decl.Specs))
	}
}

func TestBuildFuncDecl(t *testing.T) {
	input := `(FuncDecl
		:doc nil
		:recv nil
		:name (Ident :namepos 33 :name "main" :obj nil)
		:type (FuncType :func 28 :params (FieldList :opening 37 :list () :closing 38) :results nil)
		:body (BlockStmt :lbrace 40 :list () :rbrace 76))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	decl, err := builder.buildFuncDecl(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if decl.Name.Name != "main" {
		t.Fatalf("expected name %q, got %q", "main", decl.Name.Name)
	}

	if decl.Recv != nil {
		t.Fatalf("expected nil recv, got %v", decl.Recv)
	}

	if decl.Type == nil {
		t.Fatalf("expected non-nil type")
	}

	if decl.Body == nil {
		t.Fatalf("expected non-nil body")
	}
}

func TestBuildFileInfo(t *testing.T) {
	input := `(FileInfo :name "main.go" :base 1 :size 78 :lines (1 14 27 42 78))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
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

	builder := NewBuilder()
	fileSetInfo, err := builder.buildFileSet(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if fileSetInfo.Base != 1 {
		t.Fatalf("expected base %d, got %d", 1, fileSetInfo.Base)
	}

	if len(fileSetInfo.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(fileSetInfo.Files))
	}

	if fileSetInfo.Files[0].Name != "main.go" {
		t.Fatalf("expected file name %q, got %q", "main.go", fileSetInfo.Files[0].Name)
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

	builder := NewBuilder()
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

	builder := NewBuilder()
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

func TestBuildIdentMissingName(t *testing.T) {
	input := `(Ident :namepos 10 :obj nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	_, err = builder.buildIdent(sexpNode)
	if err == nil {
		t.Fatalf("expected error for missing name field")
	}
}

func TestBuildOptionalNilHandling(t *testing.T) {
	input := `nil`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()

	// Test optional ident
	ident, err := builder.buildOptionalIdent(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ident != nil {
		t.Fatalf("expected nil ident, got %v", ident)
	}

	// Test optional field list
	fieldList, err := builder.buildOptionalFieldList(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fieldList != nil {
		t.Fatalf("expected nil field list, got %v", fieldList)
	}

	// Test optional block stmt
	block, err := builder.buildOptionalBlockStmt(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block != nil {
		t.Fatalf("expected nil block stmt, got %v", block)
	}
}

// Test helper methods
func TestParseInt(t *testing.T) {
	builder := NewBuilder()

	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{`42`, 42, false},
		{`-10`, -10, false},
		{`0`, 0, false},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseInt(sexpNode)
		if tt.hasError && err == nil {
			t.Fatalf("expected error for input %q", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Fatalf("unexpected error for input %q: %v", tt.input, err)
		}
		if !tt.hasError && result != tt.expected {
			t.Fatalf("expected %d, got %d", tt.expected, result)
		}
	}
}

func TestParseString(t *testing.T) {
	builder := NewBuilder()

	input := `"hello"`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	result, err := builder.parseString(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected %q, got %q", "hello", result)
	}

	// Test error case
	badInput := `42`
	parser = sexp.NewParser(badInput)
	sexpNode, _ = parser.Parse()

	_, err = builder.parseString(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-string input")
	}
}

func TestParseToken(t *testing.T) {
	builder := NewBuilder()

	tests := []struct {
		input    string
		expected token.Token
		hasError bool
	}{
		{"IMPORT", token.IMPORT, false},
		{"CONST", token.CONST, false},
		{"TYPE", token.TYPE, false},
		{"VAR", token.VAR, false},
		{"INT", token.INT, false},
		{"FLOAT", token.FLOAT, false},
		{"IMAG", token.IMAG, false},
		{"CHAR", token.CHAR, false},
		{"STRING", token.STRING, false},
		{"INVALID", token.ILLEGAL, true},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		result, err := builder.parseToken(sexpNode)
		if tt.hasError && err == nil {
			t.Fatalf("expected error for token %q", tt.input)
		}
		if !tt.hasError {
			if err != nil {
				t.Fatalf("unexpected error for token %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		}
	}
}

func TestParsePos(t *testing.T) {
	builder := NewBuilder()

	input := `123`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	result := builder.parsePos(sexpNode)
	if result != token.Pos(123) {
		t.Fatalf("expected pos %d, got %d", 123, result)
	}
}

func TestParseNil(t *testing.T) {
	builder := NewBuilder()

	input := `nil`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	if !builder.parseNil(sexpNode) {
		t.Fatalf("expected parseNil to return true for nil")
	}

	input2 := `foo`
	parser2 := sexp.NewParser(input2)
	sexpNode2, _ := parser2.Parse()

	if builder.parseNil(sexpNode2) {
		t.Fatalf("expected parseNil to return false for non-nil")
	}
}

// Test different literal types
func TestBuildBasicLitTypes(t *testing.T) {
	tests := []struct {
		input string
		kind  token.Token
		value string
	}{
		{`(BasicLit :valuepos 10 :kind INT :value "42")`, token.INT, "42"},
		{`(BasicLit :valuepos 10 :kind FLOAT :value "3.14")`, token.FLOAT, "3.14"},
		{`(BasicLit :valuepos 10 :kind IMAG :value "1.5i")`, token.IMAG, "1.5i"},
		{`(BasicLit :valuepos 10 :kind CHAR :value "'a'")`, token.CHAR, "'a'"},
		{`(BasicLit :valuepos 10 :kind STRING :value "\"hello\"")`, token.STRING, `"hello"`},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error for %q: %v", tt.input, err)
		}

		builder := NewBuilder()
		lit, err := builder.buildBasicLit(sexpNode)
		if err != nil {
			t.Fatalf("build error for %q: %v", tt.input, err)
		}

		if lit.Kind != tt.kind {
			t.Fatalf("expected kind %v, got %v", tt.kind, lit.Kind)
		}
		if lit.Value != tt.value {
			t.Fatalf("expected value %q, got %q", tt.value, lit.Value)
		}
	}
}

// Test different GenDecl token types
func TestBuildGenDeclTypes(t *testing.T) {
	tests := []struct {
		tokType  string
		expected token.Token
	}{
		{"IMPORT", token.IMPORT},
		{"CONST", token.CONST},
		{"TYPE", token.TYPE},
		{"VAR", token.VAR},
	}

	for _, tt := range tests {
		input := fmt.Sprintf(`(GenDecl :doc nil :tok %s :tokpos 10 :lparen 0 :specs () :rparen 0)`, tt.tokType)
		parser := sexp.NewParser(input)
		sexpNode, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		builder := NewBuilder()
		decl, err := builder.buildGenDecl(sexpNode)
		if err != nil {
			t.Fatalf("build error for tok %s: %v", tt.tokType, err)
		}

		if decl.Tok != tt.expected {
			t.Fatalf("expected tok %v, got %v", tt.expected, decl.Tok)
		}
	}
}

// Test buildField with names
func TestBuildFieldWithNames(t *testing.T) {
	input := `(Field
		:doc nil
		:names ((Ident :namepos 10 :name "x" :obj nil) (Ident :namepos 13 :name "y" :obj nil))
		:type (Ident :namepos 15 :name "int" :obj nil)
		:tag nil
		:comment nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	field, err := builder.buildField(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if len(field.Names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(field.Names))
	}

	if field.Names[0].Name != "x" {
		t.Fatalf("expected first name %q, got %q", "x", field.Names[0].Name)
	}

	if field.Names[1].Name != "y" {
		t.Fatalf("expected second name %q, got %q", "y", field.Names[1].Name)
	}
}

// Test nested CallExpr
func TestBuildNestedCallExpr(t *testing.T) {
	input := `(CallExpr
		:fun (SelectorExpr
			:x (CallExpr
				:fun (Ident :namepos 10 :name "foo" :obj nil)
				:lparen 13
				:args ()
				:ellipsis 0
				:rparen 14)
			:sel (Ident :namepos 16 :name "Bar" :obj nil))
		:lparen 19
		:args ()
		:ellipsis 0
		:rparen 20)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := NewBuilder()
	call, err := builder.buildCallExpr(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	// Verify it's a selector expr
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected Fun to be SelectorExpr, got %T", call.Fun)
	}

	// Verify the selector's X is a CallExpr
	innerCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected selector X to be CallExpr, got %T", sel.X)
	}

	// Verify the inner call's function is an Ident
	innerIdent, ok := innerCall.Fun.(*ast.Ident)
	if !ok || innerIdent.Name != "foo" {
		t.Fatalf("expected inner call fun to be Ident 'foo', got %v", innerCall.Fun)
	}
}

// Test error cases for builders
func TestBuildExprUnknownType(t *testing.T) {
	input := `(UnknownExpr :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	_, err := builder.buildExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown expression type")
	}
}

func TestBuildStmtUnknownType(t *testing.T) {
	input := `(UnknownStmt :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown statement type")
	}
}

func TestBuildDeclUnknownType(t *testing.T) {
	input := `(UnknownDecl :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	_, err := builder.buildDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown declaration type")
	}
}

func TestBuildSpecUnknownType(t *testing.T) {
	input := `(UnknownSpec :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	_, err := builder.buildSpec(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown spec type")
	}
}

func TestBuildBasicLitMissingFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing valuepos", `(BasicLit :kind STRING :value "foo")`},
		{"missing kind", `(BasicLit :valuepos 10 :value "foo")`},
		{"missing value", `(BasicLit :valuepos 10 :kind STRING)`},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		builder := NewBuilder()
		_, err := builder.buildBasicLit(sexpNode)
		if err == nil {
			t.Fatalf("%s: expected error", tt.name)
		}
	}
}

func TestBuildSelectorExprMissingFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing x", `(SelectorExpr :sel (Ident :namepos 10 :name "foo" :obj nil))`},
		{"missing sel", `(SelectorExpr :x (Ident :namepos 10 :name "foo" :obj nil))`},
	}

	for _, tt := range tests {
		parser := sexp.NewParser(tt.input)
		sexpNode, _ := parser.Parse()

		builder := NewBuilder()
		_, err := builder.buildSelectorExpr(sexpNode)
		if err == nil {
			t.Fatalf("%s: expected error", tt.name)
		}
	}
}

func TestBuildCallExprMissingFields(t *testing.T) {
	input := `(CallExpr :fun (Ident :namepos 10 :name "foo" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	_, err := builder.buildCallExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for missing fields")
	}
}

func TestBuildFuncDeclMissingFields(t *testing.T) {
	input := `(FuncDecl :name (Ident :namepos 10 :name "foo" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	_, err := builder.buildFuncDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for missing type field")
	}
}

func TestExpectListWithNonList(t *testing.T) {
	input := `foo`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	_, ok := builder.expectList(sexpNode, "test")
	if ok {
		t.Fatalf("expected expectList to return false for non-list")
	}

	if len(builder.errors) == 0 {
		t.Fatalf("expected error to be recorded")
	}
}

func TestExpectSymbolWithWrongSymbol(t *testing.T) {
	input := `foo`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	ok := builder.expectSymbol(sexpNode, "bar")
	if ok {
		t.Fatalf("expected expectSymbol to return false for wrong symbol")
	}
}

func TestExpectSymbolWithNonSymbol(t *testing.T) {
	input := `42`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := NewBuilder()
	ok := builder.expectSymbol(sexpNode, "foo")
	if ok {
		t.Fatalf("expected expectSymbol to return false for non-symbol")
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

	builder := NewBuilder()
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

func TestParseKeywordArgsEdgeCases(t *testing.T) {
	builder := NewBuilder()

	// Test odd number of elements (missing value)
	input := `(Foo :name "bar" :type)`
	parser := sexp.NewParser(input)
	list, _ := parser.Parse()
	listNode := list.(*sexp.List)

	args := builder.parseKeywordArgs(listNode.Elements)

	// Should still parse the first keyword correctly
	if _, ok := args["name"]; !ok {
		t.Fatalf("expected 'name' key to be present")
	}

	// Error should be recorded for missing value
	if len(builder.errors) == 0 {
		t.Fatalf("expected error for missing value")
	}
}

func TestRequireKeywordMissing(t *testing.T) {
	builder := NewBuilder()
	args := map[string]sexp.SExp{}

	_, ok := builder.requireKeyword(args, "missing", "TestNode")
	if ok {
		t.Fatalf("expected requireKeyword to return false")
	}

	if len(builder.errors) == 0 {
		t.Fatalf("expected error to be recorded")
	}
}

func TestGetKeyword(t *testing.T) {
	builder := NewBuilder()

	input := `(Foo :name "bar")`
	parser := sexp.NewParser(input)
	list, _ := parser.Parse()
	listNode := list.(*sexp.List)

	args := builder.parseKeywordArgs(listNode.Elements)

	// Test getting existing keyword
	val, ok := builder.getKeyword(args, "name")
	if !ok {
		t.Fatalf("expected to find 'name' keyword")
	}
	if val == nil {
		t.Fatalf("expected non-nil value")
	}

	// Test getting non-existing keyword
	_, ok = builder.getKeyword(args, "missing")
	if ok {
		t.Fatalf("expected not to find 'missing' keyword")
	}
}

func TestErrors(t *testing.T) {
	builder := NewBuilder()

	// Initially no errors
	if len(builder.Errors()) != 0 {
		t.Fatalf("expected no errors initially")
	}

	// Add an error
	builder.addError("test error")

	errors := builder.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	if errors[0] != "test error" {
		t.Fatalf("expected error %q, got %q", "test error", errors[0])
	}
}

func TestParseIntWithSymbol(t *testing.T) {
	builder := NewBuilder()

	// Test parsing int from symbol
	input := `42`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	result, err := builder.parseInt(sexpNode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
}

func TestParseIntError(t *testing.T) {
	builder := NewBuilder()

	// Test error case with invalid type
	input := `"not a number"`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.parseInt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-number input")
	}
}

func TestParsePosError(t *testing.T) {
	builder := NewBuilder()

	// Test with invalid position
	input := `"not a number"`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	pos := builder.parsePos(sexpNode)
	// Should return NoPos and record error
	if pos != 0 {
		t.Fatalf("expected NoPos (0), got %d", pos)
	}

	if len(builder.errors) == 0 {
		t.Fatalf("expected error to be recorded")
	}
}

func TestParseTokenWithNonSymbol(t *testing.T) {
	builder := NewBuilder()

	input := `42`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.parseToken(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol input")
	}
}

func TestBuildExprEmptyList(t *testing.T) {
	builder := NewBuilder()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildExprNonSymbolFirst(t *testing.T) {
	builder := NewBuilder()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

func TestBuildStmtEmptyList(t *testing.T) {
	builder := NewBuilder()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildStmtNonSymbolFirst(t *testing.T) {
	builder := NewBuilder()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

func TestBuildDeclEmptyList(t *testing.T) {
	builder := NewBuilder()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildDeclNonSymbolFirst(t *testing.T) {
	builder := NewBuilder()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildDecl(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

func TestBuildSpecEmptyList(t *testing.T) {
	builder := NewBuilder()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildSpec(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildSpecNonSymbolFirst(t *testing.T) {
	builder := NewBuilder()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildSpec(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

func TestBuildIdentWrongNodeType(t *testing.T) {
	builder := NewBuilder()

	input := `(NotIdent :namepos 10 :name "foo" :obj nil)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildIdent(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildBasicLitWrongNodeType(t *testing.T) {
	builder := NewBuilder()

	input := `(NotBasicLit :valuepos 10 :kind INT :value "42")`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildBasicLit(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildSelectorExprWrongNodeType(t *testing.T) {
	builder := NewBuilder()

	input := `(NotSelectorExpr :x (Ident :namepos 10 :name "foo" :obj nil) :sel (Ident :namepos 14 :name "bar" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildSelectorExpr(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildProgramWrongNodeType(t *testing.T) {
	builder := NewBuilder()

	input := `(NotProgram :fileset (FileSet :base 1 :files ()) :files ())`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, _, err := builder.BuildProgram(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestBuildFileWrongNodeType(t *testing.T) {
	builder := NewBuilder()

	input := `(NotFile :package 1 :name (Ident :namepos 9 :name "main" :obj nil) :decls () :scope nil :imports () :unresolved () :comments ())`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.BuildFile(sexpNode)
	if err == nil {
		t.Fatalf("expected error for wrong node type")
	}
}

func TestParseKeywordArgsWithNonKeyword(t *testing.T) {
	builder := NewBuilder()

	// Test with non-keyword where keyword expected
	input := `(Foo 42 "bar")`
	parser := sexp.NewParser(input)
	list, _ := parser.Parse()
	listNode := list.(*sexp.List)

	_ = builder.parseKeywordArgs(listNode.Elements)

	// Error should be recorded for non-keyword
	if len(builder.errors) == 0 {
		t.Fatalf("expected error for non-keyword")
	}
}
