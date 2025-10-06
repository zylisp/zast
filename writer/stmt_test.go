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

func TestWriteReturnStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.ReturnStmt{
		Results: []ast.Expr{&ast.Ident{Name: "x"}},
	}
	err := w.writeReturnStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "ReturnStmt") {
		t.Fatal("expected ReturnStmt in output")
	}
}

func TestWriteAssignStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "x"}},
		Tok: 0,
		Rhs: []ast.Expr{&ast.Ident{Name: "y"}},
	}
	err := w.writeAssignStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "AssignStmt") {
		t.Fatal("expected AssignStmt in output")
	}
}

func TestWriteIncDecStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.IncDecStmt{
		X:   &ast.Ident{Name: "x"},
		Tok: 0,
	}
	err := w.writeIncDecStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "IncDecStmt") {
		t.Fatal("expected IncDecStmt in output")
	}
}

func TestWriteBranchStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.BranchStmt{
		Tok:   0,
		Label: nil,
	}
	err := w.writeBranchStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "BranchStmt") {
		t.Fatal("expected BranchStmt in output")
	}
}

func TestWriteDeferStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "f"},
			Args: []ast.Expr{},
		},
	}
	err := w.writeDeferStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "DeferStmt") {
		t.Fatal("expected DeferStmt in output")
	}
}

func TestWriteGoStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.GoStmt{
		Call: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "f"},
			Args: []ast.Expr{},
		},
	}
	err := w.writeGoStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "GoStmt") {
		t.Fatal("expected GoStmt in output")
	}
}

func TestWriteSendStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.SendStmt{
		Chan:  &ast.Ident{Name: "ch"},
		Value: &ast.Ident{Name: "x"},
	}
	err := w.writeSendStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "SendStmt") {
		t.Fatal("expected SendStmt in output")
	}
}

func TestWriteEmptyStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.EmptyStmt{
		Implicit: false,
	}
	err := w.writeEmptyStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "EmptyStmt") {
		t.Fatal("expected EmptyStmt in output")
	}
}

func TestWriteLabeledStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.LabeledStmt{
		Label: &ast.Ident{Name: "loop"},
		Stmt:  &ast.EmptyStmt{},
	}
	err := w.writeLabeledStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "LabeledStmt") {
		t.Fatal("expected LabeledStmt in output")
	}
}

func TestWriteIfStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.IfStmt{
		Cond: &ast.Ident{Name: "true"},
		Body: &ast.BlockStmt{},
	}
	err := w.writeIfStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "IfStmt") {
		t.Fatal("expected IfStmt in output")
	}
}

func TestWriteForStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.ForStmt{
		Body: &ast.BlockStmt{},
	}
	err := w.writeForStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "ForStmt") {
		t.Fatal("expected ForStmt in output")
	}
}

func TestWriteRangeStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.RangeStmt{
		X:    &ast.Ident{Name: "a"},
		Body: &ast.BlockStmt{},
	}
	err := w.writeRangeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "RangeStmt") {
		t.Fatal("expected RangeStmt in output")
	}
}

func TestWriteSwitchStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.SwitchStmt{
		Body: &ast.BlockStmt{},
	}
	err := w.writeSwitchStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "SwitchStmt") {
		t.Fatal("expected SwitchStmt in output")
	}
}

func TestWriteCaseClause(t *testing.T) {
	w := New(nil)
	clause := &ast.CaseClause{
		List: []ast.Expr{},
		Body: []ast.Stmt{},
	}
	err := w.writeCaseClause(clause)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "CaseClause") {
		t.Fatal("expected CaseClause in output")
	}
}

func TestWriteTypeSwitchStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.TypeSwitchStmt{
		Assign: &ast.ExprStmt{X: &ast.Ident{Name: "x"}},
		Body:   &ast.BlockStmt{},
	}
	err := w.writeTypeSwitchStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "TypeSwitchStmt") {
		t.Fatal("expected TypeSwitchStmt in output")
	}
}

func TestWriteSelectStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.SelectStmt{
		Body: &ast.BlockStmt{},
	}
	err := w.writeSelectStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "SelectStmt") {
		t.Fatal("expected SelectStmt in output")
	}
}

func TestWriteCommClause(t *testing.T) {
	w := New(nil)
	clause := &ast.CommClause{
		Comm: nil,
		Body: []ast.Stmt{},
	}
	err := w.writeCommClause(clause)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "CommClause") {
		t.Fatal("expected CommClause in output")
	}
}

func TestWriteDeclStmt(t *testing.T) {
	w := New(nil)
	stmt := &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok:   0,
			Specs: []ast.Spec{},
		},
	}
	err := w.writeDeclStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(w.buf.String(), "DeclStmt") {
		t.Fatal("expected DeclStmt in output")
	}
}
