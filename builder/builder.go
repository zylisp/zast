package builder

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"zylisp/zast/errors"
	"zylisp/zast/sexp"
)

// Builder builds Go AST nodes from S-expressions
type Builder struct {
	fset   *token.FileSet
	errors []string
	config *Config
	depth  int // Current nesting depth
}

// FileSetInfo stores the parsed FileSet information
type FileSetInfo struct {
	Base  int
	Files []FileInfo
}

// FileInfo stores information about a source file
type FileInfo struct {
	Name  string
	Base  int
	Size  int
	Lines []int // byte offsets of line starts
}

// NewBuilder creates a new AST builder with default configuration
func New() *Builder {
	return NewWithConfig(DefaultConfig())
}

// NewBuilderWithConfig creates a new AST builder with custom configuration
func NewWithConfig(config *Config) *Builder {
	return &Builder{
		errors: []string{},
		config: config,
	}
}

// Errors returns accumulated errors
func (b *Builder) Errors() []string {
	return b.errors
}

// addError records an error
func (b *Builder) addError(format string, args ...interface{}) {
	b.errors = append(b.errors, fmt.Sprintf(format, args...))
}

// enterDepth increments nesting depth and checks limit
func (b *Builder) enterDepth() error {
	b.depth++
	if b.depth > b.config.MaxNestingDepth {
		return fmt.Errorf("%w (%d)", errors.ErrMaxNestingDepth, b.config.MaxNestingDepth)
	}
	return nil
}

// exitDepth decrements nesting depth
func (b *Builder) exitDepth() {
	b.depth--
}

// expectList verifies sexp is a List and returns it
func (b *Builder) expectList(s sexp.SExp, context string) (*sexp.List, bool) {
	list, ok := s.(*sexp.List)
	if !ok {
		b.addError("%s: expected list, got %T at line %d, column %d",
			context, s, s.Pos().Line, s.Pos().Column)
		return nil, false
	}
	return list, true
}

// expectSymbol verifies sexp is a Symbol with expected value
func (b *Builder) expectSymbol(s sexp.SExp, expected string) bool {
	sym, ok := s.(*sexp.Symbol)
	if !ok {
		b.addError("expected symbol %q, got %T", expected, s)
		return false
	}
	if sym.Value != expected {
		b.addError("expected symbol %q, got %q", expected, sym.Value)
		return false
	}
	return true
}

// parseKeywordArgs converts a list of alternating keywords and values into a map
func (b *Builder) parseKeywordArgs(elements []sexp.SExp) map[string]sexp.SExp {
	args := make(map[string]sexp.SExp)

	// Start at index 1 (skip the node type symbol)
	for i := 1; i < len(elements); i += 2 {
		if i+1 >= len(elements) {
			b.addError("keyword argument missing value at index %d", i)
			break
		}

		keyword, ok := elements[i].(*sexp.Keyword)
		if !ok {
			b.addError("expected keyword at index %d, got %T", i, elements[i])
			continue
		}

		args[keyword.Name] = elements[i+1]
	}

	return args
}

// getKeyword retrieves a keyword value from the args map
func (b *Builder) getKeyword(args map[string]sexp.SExp, name string) (sexp.SExp, bool) {
	val, ok := args[name]
	return val, ok
}

// requireKeyword gets a keyword value or adds an error if missing
func (b *Builder) requireKeyword(args map[string]sexp.SExp, name string, context string) (sexp.SExp, bool) {
	val, ok := args[name]
	if !ok {
		b.addError("%s: missing required field :%s", context, name)
		return nil, false
	}
	return val, true
}

// parseInt converts a Number or Symbol to int
func (b *Builder) parseInt(s sexp.SExp) (int, error) {
	switch v := s.(type) {
	case *sexp.Number:
		return strconv.Atoi(v.Value)
	case *sexp.Symbol:
		return strconv.Atoi(v.Value)
	default:
		return 0, errors.ErrWrongType("number", s)
	}
}

// parsePos converts a Number to token.Pos
func (b *Builder) parsePos(s sexp.SExp) token.Pos {
	n, err := b.parseInt(s)
	if err != nil {
		b.addError("invalid position: %v", err)
		return token.NoPos
	}
	return token.Pos(n)
}

// parseString extracts string value from String node
func (b *Builder) parseString(s sexp.SExp) (string, error) {
	str, ok := s.(*sexp.String)
	if !ok {
		return "", errors.ErrWrongType("string", s)
	}
	return str.Value, nil
}

// parseNil checks if value is nil
func (b *Builder) parseNil(s sexp.SExp) bool {
	_, ok := s.(*sexp.Nil)
	return ok
}

// parseBool converts a symbol to a boolean value
func (b *Builder) parseBool(s sexp.SExp) (bool, error) {
	sym, ok := s.(*sexp.Symbol)
	if !ok {
		return false, errors.ErrWrongType("symbol for bool", s)
	}

	switch sym.Value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool value: %s", sym.Value)
	}
}

// parseChanDir converts a symbol to a ChanDir value
func (b *Builder) parseChanDir(s sexp.SExp) (ast.ChanDir, error) {
	sym, ok := s.(*sexp.Symbol)
	if !ok {
		return 0, errors.ErrWrongType("symbol for ChanDir", s)
	}

	switch sym.Value {
	case "SEND":
		return ast.SEND, nil
	case "RECV":
		return ast.RECV, nil
	case "SEND_RECV":
		return ast.SEND | ast.RECV, nil
	default:
		return 0, fmt.Errorf("unknown ChanDir: %s", sym.Value)
	}
}

// parseToken converts a symbol to a token type
func (b *Builder) parseToken(s sexp.SExp) (token.Token, error) {
	sym, ok := s.(*sexp.Symbol)
	if !ok {
		return token.ILLEGAL, errors.ErrWrongType("symbol for token", s)
	}

	switch sym.Value {
	// Keywords
	case "IMPORT":
		return token.IMPORT, nil
	case "CONST":
		return token.CONST, nil
	case "TYPE":
		return token.TYPE, nil
	case "VAR":
		return token.VAR, nil
	case "BREAK":
		return token.BREAK, nil
	case "CONTINUE":
		return token.CONTINUE, nil
	case "GOTO":
		return token.GOTO, nil
	case "FALLTHROUGH":
		return token.FALLTHROUGH, nil

	// Literal types
	case "INT":
		return token.INT, nil
	case "FLOAT":
		return token.FLOAT, nil
	case "IMAG":
		return token.IMAG, nil
	case "CHAR":
		return token.CHAR, nil
	case "STRING":
		return token.STRING, nil

	// Operators
	case "ADD":
		return token.ADD, nil
	case "SUB":
		return token.SUB, nil
	case "MUL":
		return token.MUL, nil
	case "QUO":
		return token.QUO, nil
	case "REM":
		return token.REM, nil
	case "AND":
		return token.AND, nil
	case "OR":
		return token.OR, nil
	case "XOR":
		return token.XOR, nil
	case "SHL":
		return token.SHL, nil
	case "SHR":
		return token.SHR, nil
	case "AND_NOT":
		return token.AND_NOT, nil
	case "LAND":
		return token.LAND, nil
	case "LOR":
		return token.LOR, nil
	case "ARROW":
		return token.ARROW, nil
	case "INC":
		return token.INC, nil
	case "DEC":
		return token.DEC, nil

	// Comparison
	case "EQL":
		return token.EQL, nil
	case "LSS":
		return token.LSS, nil
	case "GTR":
		return token.GTR, nil
	case "ASSIGN":
		return token.ASSIGN, nil
	case "NOT":
		return token.NOT, nil
	case "NEQ":
		return token.NEQ, nil
	case "LEQ":
		return token.LEQ, nil
	case "GEQ":
		return token.GEQ, nil
	case "DEFINE":
		return token.DEFINE, nil

	// Assignment operators
	case "ADD_ASSIGN":
		return token.ADD_ASSIGN, nil
	case "SUB_ASSIGN":
		return token.SUB_ASSIGN, nil
	case "MUL_ASSIGN":
		return token.MUL_ASSIGN, nil
	case "QUO_ASSIGN":
		return token.QUO_ASSIGN, nil
	case "REM_ASSIGN":
		return token.REM_ASSIGN, nil
	case "AND_ASSIGN":
		return token.AND_ASSIGN, nil
	case "OR_ASSIGN":
		return token.OR_ASSIGN, nil
	case "XOR_ASSIGN":
		return token.XOR_ASSIGN, nil
	case "SHL_ASSIGN":
		return token.SHL_ASSIGN, nil
	case "SHR_ASSIGN":
		return token.SHR_ASSIGN, nil
	case "AND_NOT_ASSIGN":
		return token.AND_NOT_ASSIGN, nil
	case "ILLEGAL":
		return token.ILLEGAL, nil

	default:
		return token.ILLEGAL, errors.ErrUnknownNodeType(sym.Value, "token")
	}
}
