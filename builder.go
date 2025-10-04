package zast

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"zylisp/zast/sexp"
)

// BuilderConfig holds configuration for the AST builder
type BuilderConfig struct {
	// Strict mode: fail on unknown node types vs. skip them
	StrictMode bool // default: false

	// Maximum nesting depth to prevent stack overflow on malformed input
	MaxNestingDepth int // default: 1000

	// Collect all errors vs. fail on first error
	CollectAllErrors bool // default: true
}

// DefaultBuilderConfig returns the default builder configuration
func DefaultBuilderConfig() *BuilderConfig {
	return &BuilderConfig{
		StrictMode:       false,
		MaxNestingDepth:  1000,
		CollectAllErrors: true,
	}
}

// Builder builds Go AST nodes from S-expressions
type Builder struct {
	fset   *token.FileSet
	errors []string
	config *BuilderConfig
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
func NewBuilder() *Builder {
	return NewBuilderWithConfig(DefaultBuilderConfig())
}

// NewBuilderWithConfig creates a new AST builder with custom configuration
func NewBuilderWithConfig(config *BuilderConfig) *Builder {
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
		return fmt.Errorf("%w (%d)", ErrMaxNestingDepth, b.config.MaxNestingDepth)
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
		return 0, ErrWrongType("number", s)
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
		return "", ErrWrongType("string", s)
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
		return false, ErrWrongType("symbol for bool", s)
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
		return 0, ErrWrongType("symbol for ChanDir", s)
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
		return token.ILLEGAL, ErrWrongType("symbol for token", s)
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
		return token.ILLEGAL, ErrUnknownNodeType(sym.Value, "token")
	}
}

// buildIdent parses an Ident node
func (b *Builder) buildIdent(s sexp.SExp) (*ast.Ident, error) {
	list, ok := b.expectList(s, "Ident")
	if !ok {
		return nil, ErrNotList
	}

	if len(list.Elements) == 0 {
		return nil, ErrEmptyList
	}

	if !b.expectSymbol(list.Elements[0], "Ident") {
		return nil, ErrExpectedNodeType("Ident", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameposVal, ok := b.requireKeyword(args, "namepos", "Ident")
	if !ok {
		return nil, ErrMissingField("namepos")
	}

	nameVal, ok := b.requireKeyword(args, "name", "Ident")
	if !ok {
		return nil, ErrMissingField("name")
	}

	name, err := b.parseString(nameVal)
	if err != nil {
		return nil, ErrInvalidField("name", err)
	}

	ident := &ast.Ident{
		NamePos: b.parsePos(nameposVal),
		Name:    name,
		Obj:     nil, // Objects handled separately if needed
	}

	return ident, nil
}

// buildOptionalIdent builds Ident or returns nil
func (b *Builder) buildOptionalIdent(s sexp.SExp) (*ast.Ident, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildIdent(s)
}

// buildBasicLit parses a BasicLit node
func (b *Builder) buildBasicLit(s sexp.SExp) (*ast.BasicLit, error) {
	list, ok := b.expectList(s, "BasicLit")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BasicLit") {
		return nil, ErrExpectedNodeType("BasicLit", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	valueposVal, ok := b.requireKeyword(args, "valuepos", "BasicLit")
	if !ok {
		return nil, ErrMissingField("valuepos")
	}

	kindVal, ok := b.requireKeyword(args, "kind", "BasicLit")
	if !ok {
		return nil, ErrMissingField("kind")
	}

	valueVal, ok := b.requireKeyword(args, "value", "BasicLit")
	if !ok {
		return nil, ErrMissingField("value")
	}

	kind, err := b.parseToken(kindVal)
	if err != nil {
		return nil, ErrInvalidField("kind", err)
	}

	value, err := b.parseString(valueVal)
	if err != nil {
		return nil, ErrInvalidField("value", err)
	}

	return &ast.BasicLit{
		ValuePos: b.parsePos(valueposVal),
		Kind:     kind,
		Value:    value,
	}, nil
}

// buildSelectorExpr parses a SelectorExpr node
func (b *Builder) buildSelectorExpr(s sexp.SExp) (*ast.SelectorExpr, error) {
	list, ok := b.expectList(s, "SelectorExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SelectorExpr") {
		return nil, ErrExpectedNodeType("SelectorExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "SelectorExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	selVal, ok := b.requireKeyword(args, "sel", "SelectorExpr")
	if !ok {
		return nil, ErrMissingField("sel")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	sel, err := b.buildIdent(selVal)
	if err != nil {
		return nil, ErrInvalidField("sel", err)
	}

	return &ast.SelectorExpr{
		X:   x,
		Sel: sel,
	}, nil
}

// buildCallExpr parses a CallExpr node
func (b *Builder) buildCallExpr(s sexp.SExp) (*ast.CallExpr, error) {
	list, ok := b.expectList(s, "CallExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "CallExpr") {
		return nil, ErrExpectedNodeType("CallExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	funVal, ok := b.requireKeyword(args, "fun", "CallExpr")
	if !ok {
		return nil, ErrMissingField("fun")
	}

	lparenVal, ok := b.requireKeyword(args, "lparen", "CallExpr")
	if !ok {
		return nil, ErrMissingField("lparen")
	}

	argsVal, ok := b.requireKeyword(args, "args", "CallExpr")
	if !ok {
		return nil, ErrMissingField("args")
	}

	ellipsisVal, ok := b.requireKeyword(args, "ellipsis", "CallExpr")
	if !ok {
		return nil, ErrMissingField("ellipsis")
	}

	rparenVal, ok := b.requireKeyword(args, "rparen", "CallExpr")
	if !ok {
		return nil, ErrMissingField("rparen")
	}

	fun, err := b.buildExpr(funVal)
	if err != nil {
		return nil, ErrInvalidField("fun", err)
	}

	// Build args list
	var callArgs []ast.Expr
	argsList, ok := b.expectList(argsVal, "CallExpr args")
	if ok {
		for _, argSexp := range argsList.Elements {
			arg, err := b.buildExpr(argSexp)
			if err != nil {
				return nil, ErrInvalidField("arg", err)
			}
			callArgs = append(callArgs, arg)
		}
	}

	return &ast.CallExpr{
		Fun:      fun,
		Lparen:   b.parsePos(lparenVal),
		Args:     callArgs,
		Ellipsis: b.parsePos(ellipsisVal),
		Rparen:   b.parsePos(rparenVal),
	}, nil
}

// buildBinaryExpr parses a BinaryExpr node
func (b *Builder) buildBinaryExpr(s sexp.SExp) (*ast.BinaryExpr, error) {
	list, ok := b.expectList(s, "BinaryExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BinaryExpr") {
		return nil, ErrExpectedNodeType("BinaryExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "BinaryExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	opposVal, ok := b.requireKeyword(args, "oppos", "BinaryExpr")
	if !ok {
		return nil, ErrMissingField("oppos")
	}

	opVal, ok := b.requireKeyword(args, "op", "BinaryExpr")
	if !ok {
		return nil, ErrMissingField("op")
	}

	yVal, ok := b.requireKeyword(args, "y", "BinaryExpr")
	if !ok {
		return nil, ErrMissingField("y")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	op, err := b.parseToken(opVal)
	if err != nil {
		return nil, ErrInvalidField("op", err)
	}

	y, err := b.buildExpr(yVal)
	if err != nil {
		return nil, ErrInvalidField("y", err)
	}

	return &ast.BinaryExpr{
		X:     x,
		OpPos: b.parsePos(opposVal),
		Op:    op,
		Y:     y,
	}, nil
}

// buildParenExpr parses a ParenExpr node
func (b *Builder) buildParenExpr(s sexp.SExp) (*ast.ParenExpr, error) {
	list, ok := b.expectList(s, "ParenExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ParenExpr") {
		return nil, ErrExpectedNodeType("ParenExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lparenVal, ok := b.requireKeyword(args, "lparen", "ParenExpr")
	if !ok {
		return nil, ErrMissingField("lparen")
	}

	xVal, ok := b.requireKeyword(args, "x", "ParenExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	rparenVal, ok := b.requireKeyword(args, "rparen", "ParenExpr")
	if !ok {
		return nil, ErrMissingField("rparen")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	return &ast.ParenExpr{
		Lparen: b.parsePos(lparenVal),
		X:      x,
		Rparen: b.parsePos(rparenVal),
	}, nil
}

// buildStarExpr parses a StarExpr node
func (b *Builder) buildStarExpr(s sexp.SExp) (*ast.StarExpr, error) {
	list, ok := b.expectList(s, "StarExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "StarExpr") {
		return nil, ErrExpectedNodeType("StarExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	starVal, ok := b.requireKeyword(args, "star", "StarExpr")
	if !ok {
		return nil, ErrMissingField("star")
	}

	xVal, ok := b.requireKeyword(args, "x", "StarExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	return &ast.StarExpr{
		Star: b.parsePos(starVal),
		X:    x,
	}, nil
}

// buildIndexExpr parses an IndexExpr node
func (b *Builder) buildIndexExpr(s sexp.SExp) (*ast.IndexExpr, error) {
	list, ok := b.expectList(s, "IndexExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "IndexExpr") {
		return nil, ErrExpectedNodeType("IndexExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "IndexExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	lbrackVal, ok := b.requireKeyword(args, "lbrack", "IndexExpr")
	if !ok {
		return nil, ErrMissingField("lbrack")
	}

	indexVal, ok := b.requireKeyword(args, "index", "IndexExpr")
	if !ok {
		return nil, ErrMissingField("index")
	}

	rbrackVal, ok := b.requireKeyword(args, "rbrack", "IndexExpr")
	if !ok {
		return nil, ErrMissingField("rbrack")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	index, err := b.buildExpr(indexVal)
	if err != nil {
		return nil, ErrInvalidField("index", err)
	}

	return &ast.IndexExpr{
		X:      x,
		Lbrack: b.parsePos(lbrackVal),
		Index:  index,
		Rbrack: b.parsePos(rbrackVal),
	}, nil
}

// buildOptionalExpr builds an Expr or returns nil
func (b *Builder) buildOptionalExpr(s sexp.SExp) (ast.Expr, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildExpr(s)
}

// buildSliceExpr parses a SliceExpr node
func (b *Builder) buildSliceExpr(s sexp.SExp) (*ast.SliceExpr, error) {
	list, ok := b.expectList(s, "SliceExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SliceExpr") {
		return nil, ErrExpectedNodeType("SliceExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "SliceExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	lbrackVal, ok := b.requireKeyword(args, "lbrack", "SliceExpr")
	if !ok {
		return nil, ErrMissingField("lbrack")
	}

	lowVal, ok := b.requireKeyword(args, "low", "SliceExpr")
	if !ok {
		return nil, ErrMissingField("low")
	}

	highVal, ok := b.requireKeyword(args, "high", "SliceExpr")
	if !ok {
		return nil, ErrMissingField("high")
	}

	maxVal, ok := b.requireKeyword(args, "max", "SliceExpr")
	if !ok {
		return nil, ErrMissingField("max")
	}

	slice3Val, ok := b.requireKeyword(args, "slice3", "SliceExpr")
	if !ok {
		return nil, ErrMissingField("slice3")
	}

	rbrackVal, ok := b.requireKeyword(args, "rbrack", "SliceExpr")
	if !ok {
		return nil, ErrMissingField("rbrack")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	low, err := b.buildOptionalExpr(lowVal)
	if err != nil {
		return nil, ErrInvalidField("low", err)
	}

	high, err := b.buildOptionalExpr(highVal)
	if err != nil {
		return nil, ErrInvalidField("high", err)
	}

	max, err := b.buildOptionalExpr(maxVal)
	if err != nil {
		return nil, ErrInvalidField("max", err)
	}

	slice3, err := b.parseBool(slice3Val)
	if err != nil {
		return nil, ErrInvalidField("slice3", err)
	}

	return &ast.SliceExpr{
		X:      x,
		Lbrack: b.parsePos(lbrackVal),
		Low:    low,
		High:   high,
		Max:    max,
		Slice3: slice3,
		Rbrack: b.parsePos(rbrackVal),
	}, nil
}

// buildKeyValueExpr parses a KeyValueExpr node
func (b *Builder) buildKeyValueExpr(s sexp.SExp) (*ast.KeyValueExpr, error) {
	list, ok := b.expectList(s, "KeyValueExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "KeyValueExpr") {
		return nil, ErrExpectedNodeType("KeyValueExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	keyVal, ok := b.requireKeyword(args, "key", "KeyValueExpr")
	if !ok {
		return nil, ErrMissingField("key")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "KeyValueExpr")
	if !ok {
		return nil, ErrMissingField("colon")
	}

	valueVal, ok := b.requireKeyword(args, "value", "KeyValueExpr")
	if !ok {
		return nil, ErrMissingField("value")
	}

	key, err := b.buildExpr(keyVal)
	if err != nil {
		return nil, ErrInvalidField("key", err)
	}

	value, err := b.buildExpr(valueVal)
	if err != nil {
		return nil, ErrInvalidField("value", err)
	}

	return &ast.KeyValueExpr{
		Key:   key,
		Colon: b.parsePos(colonVal),
		Value: value,
	}, nil
}

// buildUnaryExpr parses a UnaryExpr node
func (b *Builder) buildUnaryExpr(s sexp.SExp) (*ast.UnaryExpr, error) {
	list, ok := b.expectList(s, "UnaryExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "UnaryExpr") {
		return nil, ErrExpectedNodeType("UnaryExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	opposVal, ok := b.requireKeyword(args, "oppos", "UnaryExpr")
	if !ok {
		return nil, ErrMissingField("oppos")
	}

	opVal, ok := b.requireKeyword(args, "op", "UnaryExpr")
	if !ok {
		return nil, ErrMissingField("op")
	}

	xVal, ok := b.requireKeyword(args, "x", "UnaryExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	op, err := b.parseToken(opVal)
	if err != nil {
		return nil, ErrInvalidField("op", err)
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	return &ast.UnaryExpr{
		OpPos: b.parsePos(opposVal),
		Op:    op,
		X:     x,
	}, nil
}

// buildTypeAssertExpr parses a TypeAssertExpr node
func (b *Builder) buildTypeAssertExpr(s sexp.SExp) (*ast.TypeAssertExpr, error) {
	list, ok := b.expectList(s, "TypeAssertExpr")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "TypeAssertExpr") {
		return nil, ErrExpectedNodeType("TypeAssertExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "TypeAssertExpr")
	if !ok {
		return nil, ErrMissingField("x")
	}

	lparenVal, ok := b.requireKeyword(args, "lparen", "TypeAssertExpr")
	if !ok {
		return nil, ErrMissingField("lparen")
	}

	typeVal, ok := b.requireKeyword(args, "type", "TypeAssertExpr")
	if !ok {
		return nil, ErrMissingField("type")
	}

	rparenVal, ok := b.requireKeyword(args, "rparen", "TypeAssertExpr")
	if !ok {
		return nil, ErrMissingField("rparen")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	// Type can be nil for x.(type) in type switches
	typ, err := b.buildOptionalExpr(typeVal)
	if err != nil {
		return nil, ErrInvalidField("type", err)
	}

	return &ast.TypeAssertExpr{
		X:      x,
		Lparen: b.parsePos(lparenVal),
		Type:   typ,
		Rparen: b.parsePos(rparenVal),
	}, nil
}

// buildExpr dispatches to appropriate expression builder
func (b *Builder) buildExpr(s sexp.SExp) (ast.Expr, error) {
	if err := b.enterDepth(); err != nil {
		return nil, err
	}
	defer b.exitDepth()

	list, ok := b.expectList(s, "expression")
	if !ok {
		return nil, ErrNotList
	}

	if len(list.Elements) == 0 {
		return nil, ErrEmptyList
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, ErrExpectedSymbol
	}

	switch sym.Value {
	case "Ident":
		return b.buildIdent(s)
	case "BasicLit":
		return b.buildBasicLit(s)
	case "CallExpr":
		return b.buildCallExpr(s)
	case "SelectorExpr":
		return b.buildSelectorExpr(s)
	case "UnaryExpr":
		return b.buildUnaryExpr(s)
	case "BinaryExpr":
		return b.buildBinaryExpr(s)
	case "ParenExpr":
		return b.buildParenExpr(s)
	case "StarExpr":
		return b.buildStarExpr(s)
	case "IndexExpr":
		return b.buildIndexExpr(s)
	case "SliceExpr":
		return b.buildSliceExpr(s)
	case "KeyValueExpr":
		return b.buildKeyValueExpr(s)
	case "ArrayType":
		return b.buildArrayType(s)
	case "MapType":
		return b.buildMapType(s)
	case "ChanType":
		return b.buildChanType(s)
	case "TypeAssertExpr":
		return b.buildTypeAssertExpr(s)
	default:
		return nil, ErrUnknownNodeType(sym.Value, "expression")
	}
}

// buildExprStmt parses an ExprStmt node
func (b *Builder) buildExprStmt(s sexp.SExp) (*ast.ExprStmt, error) {
	list, ok := b.expectList(s, "ExprStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ExprStmt") {
		return nil, ErrExpectedNodeType("ExprStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "ExprStmt")
	if !ok {
		return nil, ErrMissingField("x")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	return &ast.ExprStmt{X: x}, nil
}

// buildBlockStmt parses a BlockStmt node
func (b *Builder) buildBlockStmt(s sexp.SExp) (*ast.BlockStmt, error) {
	list, ok := b.expectList(s, "BlockStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BlockStmt") {
		return nil, ErrExpectedNodeType("BlockStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lbraceVal, ok := b.requireKeyword(args, "lbrace", "BlockStmt")
	if !ok {
		return nil, ErrMissingField("lbrace")
	}

	listVal, ok := b.requireKeyword(args, "list", "BlockStmt")
	if !ok {
		return nil, ErrMissingField("list")
	}

	rbraceVal, ok := b.requireKeyword(args, "rbrace", "BlockStmt")
	if !ok {
		return nil, ErrMissingField("rbrace")
	}

	// Build statement list
	var stmts []ast.Stmt
	stmtsList, ok := b.expectList(listVal, "BlockStmt list")
	if ok {
		for _, stmtSexp := range stmtsList.Elements {
			stmt, err := b.buildStmt(stmtSexp)
			if err != nil {
				return nil, ErrInvalidField("statement", err)
			}
			stmts = append(stmts, stmt)
		}
	}

	return &ast.BlockStmt{
		Lbrace: b.parsePos(lbraceVal),
		List:   stmts,
		Rbrace: b.parsePos(rbraceVal),
	}, nil
}

// buildOptionalBlockStmt builds BlockStmt or returns nil
func (b *Builder) buildOptionalBlockStmt(s sexp.SExp) (*ast.BlockStmt, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildBlockStmt(s)
}

// buildReturnStmt parses a ReturnStmt node
func (b *Builder) buildReturnStmt(s sexp.SExp) (*ast.ReturnStmt, error) {
	list, ok := b.expectList(s, "ReturnStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ReturnStmt") {
		return nil, ErrExpectedNodeType("ReturnStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	returnVal, ok := b.requireKeyword(args, "return", "ReturnStmt")
	if !ok {
		return nil, ErrMissingField("return")
	}

	resultsVal, ok := b.requireKeyword(args, "results", "ReturnStmt")
	if !ok {
		return nil, ErrMissingField("results")
	}

	// Build results list
	var results []ast.Expr
	resultsList, ok := b.expectList(resultsVal, "ReturnStmt results")
	if ok {
		for _, resultSexp := range resultsList.Elements {
			result, err := b.buildExpr(resultSexp)
			if err != nil {
				return nil, ErrInvalidField("result", err)
			}
			results = append(results, result)
		}
	}

	return &ast.ReturnStmt{
		Return:  b.parsePos(returnVal),
		Results: results,
	}, nil
}

// buildAssignStmt parses an AssignStmt node
func (b *Builder) buildAssignStmt(s sexp.SExp) (*ast.AssignStmt, error) {
	list, ok := b.expectList(s, "AssignStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "AssignStmt") {
		return nil, ErrExpectedNodeType("AssignStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lhsVal, ok := b.requireKeyword(args, "lhs", "AssignStmt")
	if !ok {
		return nil, ErrMissingField("lhs")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "AssignStmt")
	if !ok {
		return nil, ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "AssignStmt")
	if !ok {
		return nil, ErrMissingField("tok")
	}

	rhsVal, ok := b.requireKeyword(args, "rhs", "AssignStmt")
	if !ok {
		return nil, ErrMissingField("rhs")
	}

	// Build lhs list
	var lhs []ast.Expr
	lhsList, ok := b.expectList(lhsVal, "AssignStmt lhs")
	if ok {
		for _, lhsSexp := range lhsList.Elements {
			lhsExpr, err := b.buildExpr(lhsSexp)
			if err != nil {
				return nil, ErrInvalidField("lhs", err)
			}
			lhs = append(lhs, lhsExpr)
		}
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, ErrInvalidField("tok", err)
	}

	// Build rhs list
	var rhs []ast.Expr
	rhsList, ok := b.expectList(rhsVal, "AssignStmt rhs")
	if ok {
		for _, rhsSexp := range rhsList.Elements {
			rhsExpr, err := b.buildExpr(rhsSexp)
			if err != nil {
				return nil, ErrInvalidField("rhs", err)
			}
			rhs = append(rhs, rhsExpr)
		}
	}

	return &ast.AssignStmt{
		Lhs:    lhs,
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
		Rhs:    rhs,
	}, nil
}

// buildIncDecStmt parses an IncDecStmt node
func (b *Builder) buildIncDecStmt(s sexp.SExp) (*ast.IncDecStmt, error) {
	list, ok := b.expectList(s, "IncDecStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "IncDecStmt") {
		return nil, ErrExpectedNodeType("IncDecStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "IncDecStmt")
	if !ok {
		return nil, ErrMissingField("x")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "IncDecStmt")
	if !ok {
		return nil, ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "IncDecStmt")
	if !ok {
		return nil, ErrMissingField("tok")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, ErrInvalidField("tok", err)
	}

	return &ast.IncDecStmt{
		X:      x,
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
	}, nil
}

// buildBranchStmt parses a BranchStmt node
func (b *Builder) buildBranchStmt(s sexp.SExp) (*ast.BranchStmt, error) {
	list, ok := b.expectList(s, "BranchStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BranchStmt") {
		return nil, ErrExpectedNodeType("BranchStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	tokposVal, ok := b.requireKeyword(args, "tokpos", "BranchStmt")
	if !ok {
		return nil, ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "BranchStmt")
	if !ok {
		return nil, ErrMissingField("tok")
	}

	labelVal, ok := b.requireKeyword(args, "label", "BranchStmt")
	if !ok {
		return nil, ErrMissingField("label")
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, ErrInvalidField("tok", err)
	}

	label, err := b.buildOptionalIdent(labelVal)
	if err != nil {
		return nil, ErrInvalidField("label", err)
	}

	return &ast.BranchStmt{
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
		Label:  label,
	}, nil
}

// buildDeferStmt parses a DeferStmt node
func (b *Builder) buildDeferStmt(s sexp.SExp) (*ast.DeferStmt, error) {
	list, ok := b.expectList(s, "DeferStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "DeferStmt") {
		return nil, ErrExpectedNodeType("DeferStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	deferVal, ok := b.requireKeyword(args, "defer", "DeferStmt")
	if !ok {
		return nil, ErrMissingField("defer")
	}

	callVal, ok := b.requireKeyword(args, "call", "DeferStmt")
	if !ok {
		return nil, ErrMissingField("call")
	}

	call, err := b.buildCallExpr(callVal)
	if err != nil {
		return nil, ErrInvalidField("call", err)
	}

	return &ast.DeferStmt{
		Defer: b.parsePos(deferVal),
		Call:  call,
	}, nil
}

// buildGoStmt parses a GoStmt node
func (b *Builder) buildGoStmt(s sexp.SExp) (*ast.GoStmt, error) {
	list, ok := b.expectList(s, "GoStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "GoStmt") {
		return nil, ErrExpectedNodeType("GoStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	goVal, ok := b.requireKeyword(args, "go", "GoStmt")
	if !ok {
		return nil, ErrMissingField("go")
	}

	callVal, ok := b.requireKeyword(args, "call", "GoStmt")
	if !ok {
		return nil, ErrMissingField("call")
	}

	call, err := b.buildCallExpr(callVal)
	if err != nil {
		return nil, ErrInvalidField("call", err)
	}

	return &ast.GoStmt{
		Go:   b.parsePos(goVal),
		Call: call,
	}, nil
}

// buildSendStmt parses a SendStmt node
func (b *Builder) buildSendStmt(s sexp.SExp) (*ast.SendStmt, error) {
	list, ok := b.expectList(s, "SendStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SendStmt") {
		return nil, ErrExpectedNodeType("SendStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	chanVal, ok := b.requireKeyword(args, "chan", "SendStmt")
	if !ok {
		return nil, ErrMissingField("chan")
	}

	arrowVal, ok := b.requireKeyword(args, "arrow", "SendStmt")
	if !ok {
		return nil, ErrMissingField("arrow")
	}

	valueVal, ok := b.requireKeyword(args, "value", "SendStmt")
	if !ok {
		return nil, ErrMissingField("value")
	}

	chanExpr, err := b.buildExpr(chanVal)
	if err != nil {
		return nil, ErrInvalidField("chan", err)
	}

	value, err := b.buildExpr(valueVal)
	if err != nil {
		return nil, ErrInvalidField("value", err)
	}

	return &ast.SendStmt{
		Chan:  chanExpr,
		Arrow: b.parsePos(arrowVal),
		Value: value,
	}, nil
}

// buildEmptyStmt parses an EmptyStmt node
func (b *Builder) buildEmptyStmt(s sexp.SExp) (*ast.EmptyStmt, error) {
	list, ok := b.expectList(s, "EmptyStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "EmptyStmt") {
		return nil, ErrExpectedNodeType("EmptyStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	semicolonVal, ok := b.requireKeyword(args, "semicolon", "EmptyStmt")
	if !ok {
		return nil, ErrMissingField("semicolon")
	}

	implicitVal, ok := b.requireKeyword(args, "implicit", "EmptyStmt")
	if !ok {
		return nil, ErrMissingField("implicit")
	}

	implicit, err := b.parseBool(implicitVal)
	if err != nil {
		return nil, ErrInvalidField("implicit", err)
	}

	return &ast.EmptyStmt{
		Semicolon: b.parsePos(semicolonVal),
		Implicit:  implicit,
	}, nil
}

// buildLabeledStmt parses a LabeledStmt node
func (b *Builder) buildLabeledStmt(s sexp.SExp) (*ast.LabeledStmt, error) {
	list, ok := b.expectList(s, "LabeledStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "LabeledStmt") {
		return nil, ErrExpectedNodeType("LabeledStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	labelVal, ok := b.requireKeyword(args, "label", "LabeledStmt")
	if !ok {
		return nil, ErrMissingField("label")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "LabeledStmt")
	if !ok {
		return nil, ErrMissingField("colon")
	}

	stmtVal, ok := b.requireKeyword(args, "stmt", "LabeledStmt")
	if !ok {
		return nil, ErrMissingField("stmt")
	}

	label, err := b.buildIdent(labelVal)
	if err != nil {
		return nil, ErrInvalidField("label", err)
	}

	stmt, err := b.buildStmt(stmtVal)
	if err != nil {
		return nil, ErrInvalidField("stmt", err)
	}

	return &ast.LabeledStmt{
		Label: label,
		Colon: b.parsePos(colonVal),
		Stmt:  stmt,
	}, nil
}

// buildIfStmt parses an IfStmt node
func (b *Builder) buildIfStmt(s sexp.SExp) (*ast.IfStmt, error) {
	list, ok := b.expectList(s, "IfStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "IfStmt") {
		return nil, ErrExpectedNodeType("IfStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	ifVal, ok := b.requireKeyword(args, "if", "IfStmt")
	if !ok {
		return nil, ErrMissingField("if")
	}

	condVal, ok := b.requireKeyword(args, "cond", "IfStmt")
	if !ok {
		return nil, ErrMissingField("cond")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "IfStmt")
	if !ok {
		return nil, ErrMissingField("body")
	}

	cond, err := b.buildExpr(condVal)
	if err != nil {
		return nil, ErrInvalidField("cond", err)
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, ErrInvalidField("init", err)
		}
	}

	// Optional else
	var els ast.Stmt
	if elseVal, ok := args["else"]; ok && !b.parseNil(elseVal) {
		els, err = b.buildStmt(elseVal)
		if err != nil {
			return nil, ErrInvalidField("else", err)
		}
	}

	return &ast.IfStmt{
		If:   b.parsePos(ifVal),
		Init: init,
		Cond: cond,
		Body: body,
		Else: els,
	}, nil
}

// buildForStmt parses a ForStmt node
func (b *Builder) buildForStmt(s sexp.SExp) (*ast.ForStmt, error) {
	list, ok := b.expectList(s, "ForStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ForStmt") {
		return nil, ErrExpectedNodeType("ForStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	forVal, ok := b.requireKeyword(args, "for", "ForStmt")
	if !ok {
		return nil, ErrMissingField("for")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "ForStmt")
	if !ok {
		return nil, ErrMissingField("body")
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, ErrInvalidField("init", err)
		}
	}

	// Optional cond
	var cond ast.Expr
	if condVal, ok := args["cond"]; ok && !b.parseNil(condVal) {
		cond, err = b.buildExpr(condVal)
		if err != nil {
			return nil, ErrInvalidField("cond", err)
		}
	}

	// Optional post
	var post ast.Stmt
	if postVal, ok := args["post"]; ok && !b.parseNil(postVal) {
		post, err = b.buildStmt(postVal)
		if err != nil {
			return nil, ErrInvalidField("post", err)
		}
	}

	return &ast.ForStmt{
		For:  b.parsePos(forVal),
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}, nil
}

// buildRangeStmt parses a RangeStmt node
func (b *Builder) buildRangeStmt(s sexp.SExp) (*ast.RangeStmt, error) {
	list, ok := b.expectList(s, "RangeStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "RangeStmt") {
		return nil, ErrExpectedNodeType("RangeStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	forVal, ok := b.requireKeyword(args, "for", "RangeStmt")
	if !ok {
		return nil, ErrMissingField("for")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "RangeStmt")
	if !ok {
		return nil, ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "RangeStmt")
	if !ok {
		return nil, ErrMissingField("tok")
	}

	xVal, ok := b.requireKeyword(args, "x", "RangeStmt")
	if !ok {
		return nil, ErrMissingField("x")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "RangeStmt")
	if !ok {
		return nil, ErrMissingField("body")
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, ErrInvalidField("tok", err)
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, ErrInvalidField("x", err)
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, ErrInvalidField("body", err)
	}

	// Optional key
	var key ast.Expr
	if keyVal, ok := args["key"]; ok && !b.parseNil(keyVal) {
		key, err = b.buildExpr(keyVal)
		if err != nil {
			return nil, ErrInvalidField("key", err)
		}
	}

	// Optional value
	var value ast.Expr
	if valueVal, ok := args["value"]; ok && !b.parseNil(valueVal) {
		value, err = b.buildExpr(valueVal)
		if err != nil {
			return nil, ErrInvalidField("value", err)
		}
	}

	return &ast.RangeStmt{
		For:    b.parsePos(forVal),
		Key:    key,
		Value:  value,
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
		X:      x,
		Body:   body,
	}, nil
}

// buildSwitchStmt parses a SwitchStmt node
func (b *Builder) buildSwitchStmt(s sexp.SExp) (*ast.SwitchStmt, error) {
	list, ok := b.expectList(s, "SwitchStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SwitchStmt") {
		return nil, ErrExpectedNodeType("SwitchStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	switchVal, ok := b.requireKeyword(args, "switch", "SwitchStmt")
	if !ok {
		return nil, ErrMissingField("switch")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "SwitchStmt")
	if !ok {
		return nil, ErrMissingField("body")
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, ErrInvalidField("init", err)
		}
	}

	// Optional tag
	var tag ast.Expr
	if tagVal, ok := args["tag"]; ok && !b.parseNil(tagVal) {
		tag, err = b.buildExpr(tagVal)
		if err != nil {
			return nil, ErrInvalidField("tag", err)
		}
	}

	return &ast.SwitchStmt{
		Switch: b.parsePos(switchVal),
		Init:   init,
		Tag:    tag,
		Body:   body,
	}, nil
}

// buildCaseClause parses a CaseClause node
func (b *Builder) buildCaseClause(s sexp.SExp) (*ast.CaseClause, error) {
	list, ok := b.expectList(s, "CaseClause")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "CaseClause") {
		return nil, ErrExpectedNodeType("CaseClause", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	caseVal, ok := b.requireKeyword(args, "case", "CaseClause")
	if !ok {
		return nil, ErrMissingField("case")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "CaseClause")
	if !ok {
		return nil, ErrMissingField("colon")
	}

	listVal, ok := b.requireKeyword(args, "list", "CaseClause")
	if !ok {
		return nil, ErrMissingField("list")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "CaseClause")
	if !ok {
		return nil, ErrMissingField("body")
	}

	// Build list of expressions
	var exprs []ast.Expr
	exprsList, ok := b.expectList(listVal, "CaseClause list")
	if ok {
		for _, exprSexp := range exprsList.Elements {
			expr, err := b.buildExpr(exprSexp)
			if err != nil {
				return nil, ErrInvalidField("list expr", err)
			}
			exprs = append(exprs, expr)
		}
	}

	// Build body statements
	var stmts []ast.Stmt
	stmtsList, ok := b.expectList(bodyVal, "CaseClause body")
	if ok {
		for _, stmtSexp := range stmtsList.Elements {
			stmt, err := b.buildStmt(stmtSexp)
			if err != nil {
				return nil, ErrInvalidField("body stmt", err)
			}
			stmts = append(stmts, stmt)
		}
	}

	return &ast.CaseClause{
		Case:  b.parsePos(caseVal),
		List:  exprs,
		Colon: b.parsePos(colonVal),
		Body:  stmts,
	}, nil
}

// buildTypeSwitchStmt parses a TypeSwitchStmt node
func (b *Builder) buildTypeSwitchStmt(s sexp.SExp) (*ast.TypeSwitchStmt, error) {
	list, ok := b.expectList(s, "TypeSwitchStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "TypeSwitchStmt") {
		return nil, ErrExpectedNodeType("TypeSwitchStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	switchVal, ok := b.requireKeyword(args, "switch", "TypeSwitchStmt")
	if !ok {
		return nil, ErrMissingField("switch")
	}

	assignVal, ok := b.requireKeyword(args, "assign", "TypeSwitchStmt")
	if !ok {
		return nil, ErrMissingField("assign")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "TypeSwitchStmt")
	if !ok {
		return nil, ErrMissingField("body")
	}

	assign, err := b.buildStmt(assignVal)
	if err != nil {
		return nil, ErrInvalidField("assign", err)
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, ErrInvalidField("init", err)
		}
	}

	return &ast.TypeSwitchStmt{
		Switch: b.parsePos(switchVal),
		Init:   init,
		Assign: assign,
		Body:   body,
	}, nil
}

// buildSelectStmt parses a SelectStmt node
func (b *Builder) buildSelectStmt(s sexp.SExp) (*ast.SelectStmt, error) {
	list, ok := b.expectList(s, "SelectStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SelectStmt") {
		return nil, ErrExpectedNodeType("SelectStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	selectVal, ok := b.requireKeyword(args, "select", "SelectStmt")
	if !ok {
		return nil, ErrMissingField("select")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "SelectStmt")
	if !ok {
		return nil, ErrMissingField("body")
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, ErrInvalidField("body", err)
	}

	return &ast.SelectStmt{
		Select: b.parsePos(selectVal),
		Body:   body,
	}, nil
}

// buildCommClause parses a CommClause node
func (b *Builder) buildCommClause(s sexp.SExp) (*ast.CommClause, error) {
	list, ok := b.expectList(s, "CommClause")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "CommClause") {
		return nil, ErrExpectedNodeType("CommClause", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	caseVal, ok := b.requireKeyword(args, "case", "CommClause")
	if !ok {
		return nil, ErrMissingField("case")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "CommClause")
	if !ok {
		return nil, ErrMissingField("colon")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "CommClause")
	if !ok {
		return nil, ErrMissingField("body")
	}

	// Optional comm
	var comm ast.Stmt
	var err error
	if commVal, ok := args["comm"]; ok && !b.parseNil(commVal) {
		comm, err = b.buildStmt(commVal)
		if err != nil {
			return nil, ErrInvalidField("comm", err)
		}
	}

	// Build body statements
	var stmts []ast.Stmt
	stmtsList, ok := b.expectList(bodyVal, "CommClause body")
	if ok {
		for _, stmtSexp := range stmtsList.Elements {
			stmt, err := b.buildStmt(stmtSexp)
			if err != nil {
				return nil, ErrInvalidField("body stmt", err)
			}
			stmts = append(stmts, stmt)
		}
	}

	return &ast.CommClause{
		Case:  b.parsePos(caseVal),
		Comm:  comm,
		Colon: b.parsePos(colonVal),
		Body:  stmts,
	}, nil
}

// buildDeclStmt parses a DeclStmt node
func (b *Builder) buildDeclStmt(s sexp.SExp) (*ast.DeclStmt, error) {
	list, ok := b.expectList(s, "DeclStmt")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "DeclStmt") {
		return nil, ErrExpectedNodeType("DeclStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	declVal, ok := b.requireKeyword(args, "decl", "DeclStmt")
	if !ok {
		return nil, ErrMissingField("decl")
	}

	decl, err := b.buildDecl(declVal)
	if err != nil {
		return nil, ErrInvalidField("decl", err)
	}

	return &ast.DeclStmt{
		Decl: decl,
	}, nil
}

// buildStmt dispatches to appropriate statement builder
func (b *Builder) buildStmt(s sexp.SExp) (ast.Stmt, error) {
	list, ok := b.expectList(s, "statement")
	if !ok {
		return nil, ErrNotList
	}

	if len(list.Elements) == 0 {
		return nil, ErrEmptyList
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, ErrExpectedSymbol
	}

	switch sym.Value {
	case "ExprStmt":
		return b.buildExprStmt(s)
	case "BlockStmt":
		return b.buildBlockStmt(s)
	case "ReturnStmt":
		return b.buildReturnStmt(s)
	case "AssignStmt":
		return b.buildAssignStmt(s)
	case "IncDecStmt":
		return b.buildIncDecStmt(s)
	case "BranchStmt":
		return b.buildBranchStmt(s)
	case "DeferStmt":
		return b.buildDeferStmt(s)
	case "GoStmt":
		return b.buildGoStmt(s)
	case "SendStmt":
		return b.buildSendStmt(s)
	case "EmptyStmt":
		return b.buildEmptyStmt(s)
	case "LabeledStmt":
		return b.buildLabeledStmt(s)
	case "IfStmt":
		return b.buildIfStmt(s)
	case "ForStmt":
		return b.buildForStmt(s)
	case "RangeStmt":
		return b.buildRangeStmt(s)
	case "SwitchStmt":
		return b.buildSwitchStmt(s)
	case "CaseClause":
		return b.buildCaseClause(s)
	case "TypeSwitchStmt":
		return b.buildTypeSwitchStmt(s)
	case "SelectStmt":
		return b.buildSelectStmt(s)
	case "CommClause":
		return b.buildCommClause(s)
	case "DeclStmt":
		return b.buildDeclStmt(s)
	default:
		return nil, ErrUnknownNodeType(sym.Value, "statement")
	}
}

// buildField parses a Field node
func (b *Builder) buildField(s sexp.SExp) (*ast.Field, error) {
	list, ok := b.expectList(s, "Field")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "Field") {
		return nil, ErrExpectedNodeType("Field", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	typeVal, ok := b.requireKeyword(args, "type", "Field")
	if !ok {
		return nil, ErrMissingField("type")
	}

	fieldType, err := b.buildExpr(typeVal)
	if err != nil {
		return nil, ErrInvalidField("type", err)
	}

	// Build names list (optional)
	var names []*ast.Ident
	if namesVal, ok := args["names"]; ok {
		namesList, ok := b.expectList(namesVal, "Field names")
		if ok {
			for _, nameSexp := range namesList.Elements {
				ident, err := b.buildIdent(nameSexp)
				if err != nil {
					return nil, ErrInvalidField("name", err)
				}
				names = append(names, ident)
			}
		}
	}

	return &ast.Field{
		Names: names,
		Type:  fieldType,
	}, nil
}

// buildFieldList parses a FieldList node
func (b *Builder) buildFieldList(s sexp.SExp) (*ast.FieldList, error) {
	list, ok := b.expectList(s, "FieldList")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FieldList") {
		return nil, ErrExpectedNodeType("FieldList", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	openingVal, ok := b.requireKeyword(args, "opening", "FieldList")
	if !ok {
		return nil, ErrMissingField("opening")
	}

	listVal, ok := b.requireKeyword(args, "list", "FieldList")
	if !ok {
		return nil, ErrMissingField("list")
	}

	closingVal, ok := b.requireKeyword(args, "closing", "FieldList")
	if !ok {
		return nil, ErrMissingField("closing")
	}

	// Build field list
	var fields []*ast.Field
	fieldsList, ok := b.expectList(listVal, "FieldList list")
	if ok {
		for _, fieldSexp := range fieldsList.Elements {
			field, err := b.buildField(fieldSexp)
			if err != nil {
				return nil, ErrInvalidField("field", err)
			}
			fields = append(fields, field)
		}
	}

	return &ast.FieldList{
		Opening: b.parsePos(openingVal),
		List:    fields,
		Closing: b.parsePos(closingVal),
	}, nil
}

// buildOptionalFieldList builds FieldList or returns nil
func (b *Builder) buildOptionalFieldList(s sexp.SExp) (*ast.FieldList, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildFieldList(s)
}

// buildFuncType parses a FuncType node
func (b *Builder) buildFuncType(s sexp.SExp) (*ast.FuncType, error) {
	list, ok := b.expectList(s, "FuncType")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FuncType") {
		return nil, ErrExpectedNodeType("FuncType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	funcVal, ok := b.requireKeyword(args, "func", "FuncType")
	if !ok {
		return nil, ErrMissingField("func")
	}

	paramsVal, ok := b.requireKeyword(args, "params", "FuncType")
	if !ok {
		return nil, ErrMissingField("params")
	}

	resultsVal, _ := args["results"]

	params, err := b.buildFieldList(paramsVal)
	if err != nil {
		return nil, ErrInvalidField("params", err)
	}

	results, err := b.buildOptionalFieldList(resultsVal)
	if err != nil {
		return nil, ErrInvalidField("results", err)
	}

	return &ast.FuncType{
		Func:    b.parsePos(funcVal),
		Params:  params,
		Results: results,
	}, nil
}

// buildArrayType parses an ArrayType node
func (b *Builder) buildArrayType(s sexp.SExp) (*ast.ArrayType, error) {
	list, ok := b.expectList(s, "ArrayType")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ArrayType") {
		return nil, ErrExpectedNodeType("ArrayType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lbrackVal, ok := b.requireKeyword(args, "lbrack", "ArrayType")
	if !ok {
		return nil, ErrMissingField("lbrack")
	}

	lenVal, ok := b.requireKeyword(args, "len", "ArrayType")
	if !ok {
		return nil, ErrMissingField("len")
	}

	eltVal, ok := b.requireKeyword(args, "elt", "ArrayType")
	if !ok {
		return nil, ErrMissingField("elt")
	}

	len, err := b.buildOptionalExpr(lenVal)
	if err != nil {
		return nil, ErrInvalidField("len", err)
	}

	elt, err := b.buildExpr(eltVal)
	if err != nil {
		return nil, ErrInvalidField("elt", err)
	}

	return &ast.ArrayType{
		Lbrack: b.parsePos(lbrackVal),
		Len:    len,
		Elt:    elt,
	}, nil
}

// buildMapType parses a MapType node
func (b *Builder) buildMapType(s sexp.SExp) (*ast.MapType, error) {
	list, ok := b.expectList(s, "MapType")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "MapType") {
		return nil, ErrExpectedNodeType("MapType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	mapVal, ok := b.requireKeyword(args, "map", "MapType")
	if !ok {
		return nil, ErrMissingField("map")
	}

	keyVal, ok := b.requireKeyword(args, "key", "MapType")
	if !ok {
		return nil, ErrMissingField("key")
	}

	valueVal, ok := b.requireKeyword(args, "value", "MapType")
	if !ok {
		return nil, ErrMissingField("value")
	}

	key, err := b.buildExpr(keyVal)
	if err != nil {
		return nil, ErrInvalidField("key", err)
	}

	value, err := b.buildExpr(valueVal)
	if err != nil {
		return nil, ErrInvalidField("value", err)
	}

	return &ast.MapType{
		Map:   b.parsePos(mapVal),
		Key:   key,
		Value: value,
	}, nil
}

// buildChanType parses a ChanType node
func (b *Builder) buildChanType(s sexp.SExp) (*ast.ChanType, error) {
	list, ok := b.expectList(s, "ChanType")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ChanType") {
		return nil, ErrExpectedNodeType("ChanType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	beginVal, ok := b.requireKeyword(args, "begin", "ChanType")
	if !ok {
		return nil, ErrMissingField("begin")
	}

	arrowVal, ok := b.requireKeyword(args, "arrow", "ChanType")
	if !ok {
		return nil, ErrMissingField("arrow")
	}

	dirVal, ok := b.requireKeyword(args, "dir", "ChanType")
	if !ok {
		return nil, ErrMissingField("dir")
	}

	valueVal, ok := b.requireKeyword(args, "value", "ChanType")
	if !ok {
		return nil, ErrMissingField("value")
	}

	dir, err := b.parseChanDir(dirVal)
	if err != nil {
		return nil, ErrInvalidField("dir", err)
	}

	value, err := b.buildExpr(valueVal)
	if err != nil {
		return nil, ErrInvalidField("value", err)
	}

	return &ast.ChanType{
		Begin: b.parsePos(beginVal),
		Arrow: b.parsePos(arrowVal),
		Dir:   dir,
		Value: value,
	}, nil
}

// buildImportSpec parses an ImportSpec node
func (b *Builder) buildImportSpec(s sexp.SExp) (*ast.ImportSpec, error) {
	list, ok := b.expectList(s, "ImportSpec")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ImportSpec") {
		return nil, ErrExpectedNodeType("ImportSpec", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	pathVal, ok := b.requireKeyword(args, "path", "ImportSpec")
	if !ok {
		return nil, ErrMissingField("path")
	}

	path, err := b.buildBasicLit(pathVal)
	if err != nil {
		return nil, ErrInvalidField("path", err)
	}

	// Optional name
	var name *ast.Ident
	if nameVal, ok := args["name"]; ok {
		name, err = b.buildOptionalIdent(nameVal)
		if err != nil {
			return nil, ErrInvalidField("name", err)
		}
	}

	// Optional endpos
	var endPos token.Pos
	if endposVal, ok := args["endpos"]; ok {
		endPos = b.parsePos(endposVal)
	}

	return &ast.ImportSpec{
		Name:   name,
		Path:   path,
		EndPos: endPos,
	}, nil
}

// buildValueSpec parses a ValueSpec node
func (b *Builder) buildValueSpec(s sexp.SExp) (*ast.ValueSpec, error) {
	list, ok := b.expectList(s, "ValueSpec")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ValueSpec") {
		return nil, ErrExpectedNodeType("ValueSpec", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	namesVal, ok := b.requireKeyword(args, "names", "ValueSpec")
	if !ok {
		return nil, ErrMissingField("names")
	}

	typeVal, ok := b.requireKeyword(args, "type", "ValueSpec")
	if !ok {
		return nil, ErrMissingField("type")
	}

	valuesVal, ok := b.requireKeyword(args, "values", "ValueSpec")
	if !ok {
		return nil, ErrMissingField("values")
	}

	// Build names list
	var names []*ast.Ident
	namesList, ok := b.expectList(namesVal, "ValueSpec names")
	if ok {
		for _, nameSexp := range namesList.Elements {
			name, err := b.buildIdent(nameSexp)
			if err != nil {
				return nil, ErrInvalidField("name", err)
			}
			names = append(names, name)
		}
	}

	// Type is optional
	typ, err := b.buildOptionalExpr(typeVal)
	if err != nil {
		return nil, ErrInvalidField("type", err)
	}

	// Build values list
	var values []ast.Expr
	valuesList, ok := b.expectList(valuesVal, "ValueSpec values")
	if ok {
		for _, valueSexp := range valuesList.Elements {
			value, err := b.buildExpr(valueSexp)
			if err != nil {
				return nil, ErrInvalidField("value", err)
			}
			values = append(values, value)
		}
	}

	return &ast.ValueSpec{
		Names:  names,
		Type:   typ,
		Values: values,
	}, nil
}

// buildTypeSpec parses a TypeSpec node
func (b *Builder) buildTypeSpec(s sexp.SExp) (*ast.TypeSpec, error) {
	list, ok := b.expectList(s, "TypeSpec")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "TypeSpec") {
		return nil, ErrExpectedNodeType("TypeSpec", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "TypeSpec")
	if !ok {
		return nil, ErrMissingField("name")
	}

	typeparamsVal, ok := b.requireKeyword(args, "typeparams", "TypeSpec")
	if !ok {
		return nil, ErrMissingField("typeparams")
	}

	assignVal, ok := b.requireKeyword(args, "assign", "TypeSpec")
	if !ok {
		return nil, ErrMissingField("assign")
	}

	typeVal, ok := b.requireKeyword(args, "type", "TypeSpec")
	if !ok {
		return nil, ErrMissingField("type")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, ErrInvalidField("name", err)
	}

	// Type params is optional (for generics)
	typeparams, err := b.buildOptionalFieldList(typeparamsVal)
	if err != nil {
		return nil, ErrInvalidField("typeparams", err)
	}

	typ, err := b.buildExpr(typeVal)
	if err != nil {
		return nil, ErrInvalidField("type", err)
	}

	return &ast.TypeSpec{
		Name:       name,
		TypeParams: typeparams,
		Assign:     b.parsePos(assignVal),
		Type:       typ,
	}, nil
}

// buildSpec dispatches to appropriate spec builder
func (b *Builder) buildSpec(s sexp.SExp) (ast.Spec, error) {
	list, ok := b.expectList(s, "spec")
	if !ok {
		return nil, ErrNotList
	}

	if len(list.Elements) == 0 {
		return nil, ErrEmptyList
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, ErrExpectedSymbol
	}

	switch sym.Value {
	case "ImportSpec":
		return b.buildImportSpec(s)
	case "ValueSpec":
		return b.buildValueSpec(s)
	case "TypeSpec":
		return b.buildTypeSpec(s)
	default:
		return nil, ErrUnknownNodeType(sym.Value, "spec")
	}
}

// buildGenDecl parses a GenDecl node
func (b *Builder) buildGenDecl(s sexp.SExp) (*ast.GenDecl, error) {
	list, ok := b.expectList(s, "GenDecl")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "GenDecl") {
		return nil, ErrExpectedNodeType("GenDecl", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	tokVal, ok := b.requireKeyword(args, "tok", "GenDecl")
	if !ok {
		return nil, ErrMissingField("tok")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "GenDecl")
	if !ok {
		return nil, ErrMissingField("tokpos")
	}

	specsVal, ok := b.requireKeyword(args, "specs", "GenDecl")
	if !ok {
		return nil, ErrMissingField("specs")
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, ErrInvalidField("tok", err)
	}

	// Build specs list
	var specs []ast.Spec
	specsList, ok := b.expectList(specsVal, "GenDecl specs")
	if ok {
		for _, specSexp := range specsList.Elements {
			spec, err := b.buildSpec(specSexp)
			if err != nil {
				return nil, ErrInvalidField("spec", err)
			}
			specs = append(specs, spec)
		}
	}

	// Optional lparen/rparen
	var lparen, rparen token.Pos
	if lparenVal, ok := args["lparen"]; ok {
		lparen = b.parsePos(lparenVal)
	}
	if rparenVal, ok := args["rparen"]; ok {
		rparen = b.parsePos(rparenVal)
	}

	return &ast.GenDecl{
		Tok:    tok,
		TokPos: b.parsePos(tokposVal),
		Lparen: lparen,
		Specs:  specs,
		Rparen: rparen,
	}, nil
}

// buildFuncDecl parses a FuncDecl node
func (b *Builder) buildFuncDecl(s sexp.SExp) (*ast.FuncDecl, error) {
	list, ok := b.expectList(s, "FuncDecl")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FuncDecl") {
		return nil, ErrExpectedNodeType("FuncDecl", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "FuncDecl")
	if !ok {
		return nil, ErrMissingField("name")
	}

	typeVal, ok := b.requireKeyword(args, "type", "FuncDecl")
	if !ok {
		return nil, ErrMissingField("type")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, ErrInvalidField("name", err)
	}

	funcType, err := b.buildFuncType(typeVal)
	if err != nil {
		return nil, ErrInvalidField("type", err)
	}

	// Optional recv
	var recv *ast.FieldList
	if recvVal, ok := args["recv"]; ok {
		recv, err = b.buildOptionalFieldList(recvVal)
		if err != nil {
			return nil, ErrInvalidField("recv", err)
		}
	}

	// Optional body
	var body *ast.BlockStmt
	if bodyVal, ok := args["body"]; ok {
		body, err = b.buildOptionalBlockStmt(bodyVal)
		if err != nil {
			return nil, ErrInvalidField("body", err)
		}
	}

	return &ast.FuncDecl{
		Recv: recv,
		Name: name,
		Type: funcType,
		Body: body,
	}, nil
}

// buildDecl dispatches to appropriate declaration builder
func (b *Builder) buildDecl(s sexp.SExp) (ast.Decl, error) {
	list, ok := b.expectList(s, "declaration")
	if !ok {
		return nil, ErrNotList
	}

	if len(list.Elements) == 0 {
		return nil, ErrEmptyList
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, ErrExpectedSymbol
	}

	switch sym.Value {
	case "GenDecl":
		return b.buildGenDecl(s)
	case "FuncDecl":
		return b.buildFuncDecl(s)
	default:
		return nil, ErrUnknownNodeType(sym.Value, "declaration")
	}
}

// buildFileInfo parses a FileInfo node
func (b *Builder) buildFileInfo(s sexp.SExp) (*FileInfo, error) {
	list, ok := b.expectList(s, "FileInfo")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FileInfo") {
		return nil, ErrExpectedNodeType("FileInfo", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "FileInfo")
	if !ok {
		return nil, ErrMissingField("name")
	}

	baseVal, ok := b.requireKeyword(args, "base", "FileInfo")
	if !ok {
		return nil, ErrMissingField("base")
	}

	sizeVal, ok := b.requireKeyword(args, "size", "FileInfo")
	if !ok {
		return nil, ErrMissingField("size")
	}

	linesVal, ok := b.requireKeyword(args, "lines", "FileInfo")
	if !ok {
		return nil, ErrMissingField("lines")
	}

	name, err := b.parseString(nameVal)
	if err != nil {
		return nil, ErrInvalidField("name", err)
	}

	base, err := b.parseInt(baseVal)
	if err != nil {
		return nil, ErrInvalidField("base", err)
	}

	size, err := b.parseInt(sizeVal)
	if err != nil {
		return nil, ErrInvalidField("size", err)
	}

	// Parse lines list
	linesList, ok := b.expectList(linesVal, "FileInfo lines")
	if !ok {
		return nil, ErrInvalidField("lines", ErrNotList)
	}

	var lines []int
	for _, lineSexp := range linesList.Elements {
		line, err := b.parseInt(lineSexp)
		if err != nil {
			return nil, fmt.Errorf("invalid line offset: %v", err)
		}
		lines = append(lines, line)
	}

	return &FileInfo{
		Name:  name,
		Base:  base,
		Size:  size,
		Lines: lines,
	}, nil
}

// buildFileSet parses a FileSet node
func (b *Builder) buildFileSet(s sexp.SExp) (*FileSetInfo, error) {
	list, ok := b.expectList(s, "FileSet")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FileSet") {
		return nil, ErrExpectedNodeType("FileSet", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	baseVal, ok := b.requireKeyword(args, "base", "FileSet")
	if !ok {
		return nil, ErrMissingField("base")
	}

	filesVal, ok := b.requireKeyword(args, "files", "FileSet")
	if !ok {
		return nil, ErrMissingField("files")
	}

	base, err := b.parseInt(baseVal)
	if err != nil {
		return nil, ErrInvalidField("base", err)
	}

	// Parse files list
	filesList, ok := b.expectList(filesVal, "FileSet files")
	if !ok {
		return nil, ErrInvalidField("files", ErrNotList)
	}

	var files []FileInfo
	for _, fileSexp := range filesList.Elements {
		fileInfo, err := b.buildFileInfo(fileSexp)
		if err != nil {
			return nil, fmt.Errorf("invalid file info: %v", err)
		}
		files = append(files, *fileInfo)
	}

	return &FileSetInfo{
		Base:  base,
		Files: files,
	}, nil
}

// BuildFile converts a File s-expression to *ast.File
func (b *Builder) BuildFile(s sexp.SExp) (*ast.File, error) {
	list, ok := b.expectList(s, "File")
	if !ok {
		return nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "File") {
		return nil, ErrExpectedNodeType("File", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	packageVal, ok := b.requireKeyword(args, "package", "File")
	if !ok {
		return nil, ErrMissingField("package")
	}

	nameVal, ok := b.requireKeyword(args, "name", "File")
	if !ok {
		return nil, ErrMissingField("name")
	}

	declsVal, ok := b.requireKeyword(args, "decls", "File")
	if !ok {
		return nil, ErrMissingField("decls")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, ErrInvalidField("name", err)
	}

	// Build declarations list
	var decls []ast.Decl
	declsList, ok := b.expectList(declsVal, "File decls")
	if ok {
		for _, declSexp := range declsList.Elements {
			decl, err := b.buildDecl(declSexp)
			if err != nil {
				return nil, ErrInvalidField("declaration", err)
			}
			decls = append(decls, decl)
		}
	}

	// Optional imports, unresolved, comments - ignore for now

	file := &ast.File{
		Package: b.parsePos(packageVal),
		Name:    name,
		Decls:   decls,
	}

	return file, nil
}

// BuildProgram parses a Program s-expression and returns FileSet and Files
func (b *Builder) BuildProgram(s sexp.SExp) (*token.FileSet, []*ast.File, error) {
	list, ok := b.expectList(s, "Program")
	if !ok {
		return nil, nil, ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "Program") {
		return nil, nil, ErrExpectedNodeType("Program", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	filesetVal, ok := b.requireKeyword(args, "fileset", "Program")
	if !ok {
		return nil, nil, ErrMissingField("fileset")
	}

	filesVal, ok := b.requireKeyword(args, "files", "Program")
	if !ok {
		return nil, nil, ErrMissingField("files")
	}

	// Build FileSet
	fileSetInfo, err := b.buildFileSet(filesetVal)
	if err != nil {
		return nil, nil, ErrInvalidField("fileset", err)
	}

	// Create token.FileSet from FileSetInfo
	fset := token.NewFileSet()
	for _, fi := range fileSetInfo.Files {
		fset.AddFile(fi.Name, fi.Base, fi.Size)
	}
	b.fset = fset

	// Build files list
	var files []*ast.File
	filesList, ok := b.expectList(filesVal, "Program files")
	if ok {
		for _, fileSexp := range filesList.Elements {
			file, err := b.BuildFile(fileSexp)
			if err != nil {
				return nil, nil, ErrInvalidField("file", err)
			}
			files = append(files, file)
		}
	}

	if len(b.errors) > 0 {
		return nil, nil, fmt.Errorf("build errors: %s", strings.Join(b.errors, "; "))
	}

	return fset, files, nil
}
