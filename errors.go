package zast

import (
	"errors"
	"fmt"
)

// Builder errors

// ErrMaxNestingDepth is returned when the maximum nesting depth is exceeded
var ErrMaxNestingDepth = errors.New("maximum nesting depth exceeded")

// ErrNotList is returned when a list is expected but not found
var ErrNotList = errors.New("expected list")

// ErrEmptyList is returned when a non-empty list is expected
var ErrEmptyList = errors.New("empty list")

// ErrExpectedSymbol is returned when a symbol is expected but not found
var ErrExpectedSymbol = errors.New("expected symbol as first element")

// ErrMissingField is returned when a required field is missing
func ErrMissingField(field string) error {
	return fmt.Errorf("missing %s", field)
}

// ErrInvalidField is returned when a field has an invalid value
func ErrInvalidField(field string, err error) error {
	return fmt.Errorf("invalid %s: %w", field, err)
}

// ErrUnknownNodeType is returned when an unknown node type is encountered
func ErrUnknownNodeType(nodeType, category string) error {
	return fmt.Errorf("unknown %s type: %s", category, nodeType)
}

// ErrExpectedNodeType is returned when a specific node type is expected
func ErrExpectedNodeType(expected, got string) error {
	return fmt.Errorf("expected %s node, got %s", expected, got)
}

// ErrWrongType is returned when a value has the wrong type
func ErrWrongType(expected string, got any) error {
	return fmt.Errorf("expected %s, got %T", expected, got)
}

// Writer errors

// ErrUnknownExprType is returned when an unknown expression type is encountered
func ErrUnknownExprType(expr any) error {
	return fmt.Errorf("unknown expression type: %T", expr)
}

// ErrUnknownStmtType is returned when an unknown statement type is encountered
func ErrUnknownStmtType(stmt any) error {
	return fmt.Errorf("unknown statement type: %T", stmt)
}

// ErrUnknownDeclType is returned when an unknown declaration type is encountered
func ErrUnknownDeclType(decl any) error {
	return fmt.Errorf("unknown declaration type: %T", decl)
}

// ErrUnknownSpecType is returned when an unknown spec type is encountered
func ErrUnknownSpecType(spec any) error {
	return fmt.Errorf("unknown spec type: %T", spec)
}

// Parser errors (in sexp package - kept here for reference, will move to sexp/errors.go)

// Note: Parser errors will be moved to a separate sexp/errors.go file
// to keep package boundaries clean
