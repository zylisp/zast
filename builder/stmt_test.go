package builder

import (
	"go/ast"
	"go/token"
	"testing"

	"zylisp/zast/sexp"
)

func TestBuildExprStmt(t *testing.T) {
	input := `(ExprStmt :x (Ident :namepos 10 :name "foo" :obj nil))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
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

	builder := New()
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

func TestBuildStmtUnknownType(t *testing.T) {
	input := `(UnknownStmt :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	builder := New()
	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for unknown statement type")
	}
}

func TestBuildStmtEmptyList(t *testing.T) {
	builder := New()

	input := `()`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for empty list")
	}
}

func TestBuildStmtNonSymbolFirst(t *testing.T) {
	builder := New()

	input := `(42 :foo bar)`
	parser := sexp.NewParser(input)
	sexpNode, _ := parser.Parse()

	_, err := builder.buildStmt(sexpNode)
	if err == nil {
		t.Fatalf("expected error for non-symbol first element")
	}
}

// TestBuildReturnStmt tests return statement building
func TestBuildReturnStmt(t *testing.T) {
	input := `(ReturnStmt
		:return 1
		:results ((BasicLit :valuepos 8 :kind INT :value "42")))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	ret, err := builder.buildReturnStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if ret.Return != token.Pos(1) {
		t.Fatalf("expected return pos 1, got %v", ret.Return)
	}

	if len(ret.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(ret.Results))
	}
}

// TestBuildAssignStmt tests assignment statement building
func TestBuildAssignStmt(t *testing.T) {
	input := `(AssignStmt
		:lhs ((Ident :namepos 1 :name "x" :obj nil))
		:tokpos 3
		:tok ASSIGN
		:rhs ((BasicLit :valuepos 5 :kind INT :value "1")))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	assign, err := builder.buildAssignStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if assign.Tok != token.ASSIGN {
		t.Fatalf("expected tok ASSIGN, got %v", assign.Tok)
	}

	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		t.Fatalf("expected 1 lhs and 1 rhs, got %d and %d", len(assign.Lhs), len(assign.Rhs))
	}
}

// TestBuildIncDecStmt tests increment/decrement statement building
func TestBuildIncDecStmt(t *testing.T) {
	input := `(IncDecStmt
		:x (Ident :namepos 1 :name "x" :obj nil)
		:tokpos 2
		:tok INC)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	incDec, err := builder.buildIncDecStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if incDec.Tok != token.INC {
		t.Fatalf("expected tok INC, got %v", incDec.Tok)
	}
}

// TestBuildBranchStmt tests branch statement building
func TestBuildBranchStmt(t *testing.T) {
	input := `(BranchStmt
		:tokpos 1
		:tok BREAK
		:label nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	branch, err := builder.buildBranchStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if branch.Tok != token.BREAK {
		t.Fatalf("expected tok BREAK, got %v", branch.Tok)
	}

	if branch.Label != nil {
		t.Fatalf("expected nil label, got %v", branch.Label)
	}
}

// TestBuildDeferStmt tests defer statement building
func TestBuildDeferStmt(t *testing.T) {
	input := `(DeferStmt
		:defer 1
		:call (CallExpr
			:fun (Ident :namepos 7 :name "f" :obj nil)
			:lparen 8
			:args ()
			:ellipsis 0
			:rparen 9))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	deferStmt, err := builder.buildDeferStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if deferStmt.Defer != token.Pos(1) {
		t.Fatalf("expected defer pos 1, got %v", deferStmt.Defer)
	}

	if deferStmt.Call == nil {
		t.Fatalf("expected non-nil call")
	}
}

// TestBuildGoStmt tests go statement building
func TestBuildGoStmt(t *testing.T) {
	input := `(GoStmt
		:go 1
		:call (CallExpr
			:fun (Ident :namepos 4 :name "f" :obj nil)
			:lparen 5
			:args ()
			:ellipsis 0
			:rparen 6))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	goStmt, err := builder.buildGoStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if goStmt.Go != token.Pos(1) {
		t.Fatalf("expected go pos 1, got %v", goStmt.Go)
	}

	if goStmt.Call == nil {
		t.Fatalf("expected non-nil call")
	}
}

// TestBuildSendStmt tests channel send statement building
func TestBuildSendStmt(t *testing.T) {
	input := `(SendStmt
		:chan (Ident :namepos 1 :name "ch" :obj nil)
		:arrow 4
		:value (BasicLit :valuepos 7 :kind INT :value "1"))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	send, err := builder.buildSendStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	chanIdent, ok := send.Chan.(*ast.Ident)
	if !ok || chanIdent.Name != "ch" {
		t.Fatalf("expected chan to be Ident 'ch', got %v", send.Chan)
	}
}

// TestBuildEmptyStmt tests empty statement building
func TestBuildEmptyStmt(t *testing.T) {
	input := `(EmptyStmt :semicolon 1 :implicit false)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	empty, err := builder.buildEmptyStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if empty.Semicolon != token.Pos(1) {
		t.Fatalf("expected semicolon pos 1, got %v", empty.Semicolon)
	}
}

// TestBuildLabeledStmt tests labeled statement building
func TestBuildLabeledStmt(t *testing.T) {
	input := `(LabeledStmt
		:label (Ident :namepos 1 :name "loop" :obj nil)
		:colon 5
		:stmt (EmptyStmt :semicolon 6 :implicit true))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	labeled, err := builder.buildLabeledStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if labeled.Label.Name != "loop" {
		t.Fatalf("expected label 'loop', got %v", labeled.Label.Name)
	}
}

// TestBuildIfStmt tests if statement building
func TestBuildIfStmt(t *testing.T) {
	input := `(IfStmt
		:if 1
		:init nil
		:cond (Ident :namepos 4 :name "true" :obj nil)
		:body (BlockStmt :lbrace 9 :list () :rbrace 10)
		:else nil)`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	ifStmt, err := builder.buildIfStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if ifStmt.If != token.Pos(1) {
		t.Fatalf("expected if pos 1, got %v", ifStmt.If)
	}

	condIdent, ok := ifStmt.Cond.(*ast.Ident)
	if !ok || condIdent.Name != "true" {
		t.Fatalf("expected cond to be Ident 'true', got %v", ifStmt.Cond)
	}
}

// TestBuildForStmt tests for statement building
func TestBuildForStmt(t *testing.T) {
	input := `(ForStmt
		:for 1
		:init nil
		:cond nil
		:post nil
		:body (BlockStmt :lbrace 5 :list () :rbrace 6))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	forStmt, err := builder.buildForStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if forStmt.For != token.Pos(1) {
		t.Fatalf("expected for pos 1, got %v", forStmt.For)
	}
}

// TestBuildRangeStmt tests range statement building
func TestBuildRangeStmt(t *testing.T) {
	input := `(RangeStmt
		:for 1
		:key nil
		:value nil
		:tokpos 0
		:tok ILLEGAL
		:x (Ident :namepos 11 :name "a" :obj nil)
		:body (BlockStmt :lbrace 13 :list () :rbrace 14))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	rangeStmt, err := builder.buildRangeStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if rangeStmt.For != token.Pos(1) {
		t.Fatalf("expected for pos 1, got %v", rangeStmt.For)
	}

	xIdent, ok := rangeStmt.X.(*ast.Ident)
	if !ok || xIdent.Name != "a" {
		t.Fatalf("expected X to be Ident 'a', got %v", rangeStmt.X)
	}
}

// TestBuildSwitchStmt tests switch statement building
func TestBuildSwitchStmt(t *testing.T) {
	input := `(SwitchStmt
		:switch 1
		:init nil
		:tag nil
		:body (BlockStmt :lbrace 8 :list () :rbrace 9))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	switchStmt, err := builder.buildSwitchStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if switchStmt.Switch != token.Pos(1) {
		t.Fatalf("expected switch pos 1, got %v", switchStmt.Switch)
	}
}

// TestBuildCaseClause tests case clause building
func TestBuildCaseClause(t *testing.T) {
	input := `(CaseClause
		:case 1
		:list ()
		:colon 8
		:body ())`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	caseClause, err := builder.buildCaseClause(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if caseClause.Case != token.Pos(1) {
		t.Fatalf("expected case pos 1, got %v", caseClause.Case)
	}
}

// TestBuildTypeSwitchStmt tests type switch statement building
func TestBuildTypeSwitchStmt(t *testing.T) {
	input := `(TypeSwitchStmt
		:switch 1
		:init nil
		:assign (ExprStmt :x (Ident :namepos 10 :name "x" :obj nil))
		:body (BlockStmt :lbrace 20 :list () :rbrace 21))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	typeSwitch, err := builder.buildTypeSwitchStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if typeSwitch.Switch != token.Pos(1) {
		t.Fatalf("expected switch pos 1, got %v", typeSwitch.Switch)
	}
}

// TestBuildSelectStmt tests select statement building
func TestBuildSelectStmt(t *testing.T) {
	input := `(SelectStmt
		:select 1
		:body (BlockStmt :lbrace 8 :list () :rbrace 9))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	selectStmt, err := builder.buildSelectStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if selectStmt.Select != token.Pos(1) {
		t.Fatalf("expected select pos 1, got %v", selectStmt.Select)
	}
}

// TestBuildCommClause tests communication clause building
func TestBuildCommClause(t *testing.T) {
	input := `(CommClause
		:case 1
		:comm nil
		:colon 8
		:body ())`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	commClause, err := builder.buildCommClause(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	if commClause.Case != token.Pos(1) {
		t.Fatalf("expected case pos 1, got %v", commClause.Case)
	}
}

// TestBuildDeclStmt tests declaration statement building
func TestBuildDeclStmt(t *testing.T) {
	input := `(DeclStmt
		:decl (GenDecl
			:doc nil
			:tokpos 1
			:tok CONST
			:lparen 0
			:specs ((ValueSpec
				:doc nil
				:names ((Ident :namepos 7 :name "x" :obj nil))
				:type nil
				:values ((BasicLit :valuepos 11 :kind INT :value "1"))
				:comment nil))
			:rparen 0))`
	parser := sexp.NewParser(input)
	sexpNode, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	builder := New()
	declStmt, err := builder.buildDeclStmt(sexpNode)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	genDecl, ok := declStmt.Decl.(*ast.GenDecl)
	if !ok {
		t.Fatalf("expected decl to be GenDecl, got %T", declStmt.Decl)
	}

	if genDecl.Tok != token.CONST {
		t.Fatalf("expected tok CONST, got %v", genDecl.Tok)
	}
}
