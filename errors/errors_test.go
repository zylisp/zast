package errors

import (
	"errors"
	"testing"
)

func TestErrMissingField(t *testing.T) {
	err := ErrMissingField("foo")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "missing foo"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrInvalidField(t *testing.T) {
	innerErr := errors.New("inner error")
	err := ErrInvalidField("bar", innerErr)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, innerErr) {
		t.Fatal("expected error to wrap inner error")
	}
}

func TestErrUnknownNodeType(t *testing.T) {
	err := ErrUnknownNodeType("FooExpr", "expression")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "unknown expression type: FooExpr"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrExpectedNodeType(t *testing.T) {
	err := ErrExpectedNodeType("Ident", "BasicLit")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "expected Ident node, got BasicLit"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrWrongType(t *testing.T) {
	err := ErrWrongType("string", 42)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "expected string, got int"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrUnknownExprType(t *testing.T) {
	type FakeExpr struct{}
	expr := &FakeExpr{}
	err := ErrUnknownExprType(expr)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "unknown expression type: *errors.FakeExpr"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrUnknownStmtType(t *testing.T) {
	type FakeStmt struct{}
	stmt := &FakeStmt{}
	err := ErrUnknownStmtType(stmt)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "unknown statement type: *errors.FakeStmt"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrUnknownDeclType(t *testing.T) {
	type FakeDecl struct{}
	decl := &FakeDecl{}
	err := ErrUnknownDeclType(decl)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "unknown declaration type: *errors.FakeDecl"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrUnknownSpecType(t *testing.T) {
	type FakeSpec struct{}
	spec := &FakeSpec{}
	err := ErrUnknownSpecType(spec)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	expected := "unknown spec type: *errors.FakeSpec"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are non-nil
	if ErrMaxNestingDepth == nil {
		t.Fatal("ErrMaxNestingDepth is nil")
	}
	if ErrNotList == nil {
		t.Fatal("ErrNotList is nil")
	}
	if ErrEmptyList == nil {
		t.Fatal("ErrEmptyList is nil")
	}
	if ErrExpectedSymbol == nil {
		t.Fatal("ErrExpectedSymbol is nil")
	}
}
