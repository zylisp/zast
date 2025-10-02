package sexp

import (
	"errors"
	"fmt"
)

// Parser errors

// ErrNoSExpression is returned when no S-expression is found in input
var ErrNoSExpression = errors.New("no S-expression found")

// ErrExpectedList is returned when a list is expected but not found
func ErrExpectedList(got TokenType) error {
	return fmt.Errorf("expected list to start with '(', got %s", got)
}

// ErrParseErrors aggregates multiple parse errors
func ErrParseErrors(errs []string) error {
	return fmt.Errorf("parse errors: %s", JoinErrors(errs))
}

// ErrIllegalToken is returned when an illegal token is encountered
func ErrIllegalToken(literal string) error {
	return fmt.Errorf("illegal token: %s", literal)
}

// ErrUnexpectedToken is returned when an unexpected token is encountered
func ErrUnexpectedToken(tokenType TokenType) error {
	return fmt.Errorf("unexpected token: %v", tokenType)
}

// ErrUnterminatedList is returned when a list is not properly closed
var ErrUnterminatedList = errors.New("unterminated list")

// JoinErrors joins error messages with semicolons
func JoinErrors(errs []string) string {
	result := ""
	for i, err := range errs {
		if i > 0 {
			result += "; "
		}
		result += err
	}
	return result
}
