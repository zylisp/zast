package sexp

import (
	"fmt"
	"strings"
)

const (
	// NilLiteral is the string representation of nil
	NilLiteral = "nil"

	// ParenO is the opening parenthesis
	ParenO = "("
	// ParenC is the closing parenthesis
	ParenC = ")"
	// ParenOC is the empty list representation (open and close)
	ParenOC = "()"
)

// SExp is the interface all S-expression nodes implement
type SExp interface {
	sexp()
	Pos() Position
	String() string
}

// Position tracks source location
type Position struct {
	Offset int // byte offset
	Line   int // line number (1-based)
	Column int // column number (1-based)
}

// Symbol represents an identifier/symbol
type Symbol struct {
	Position Position
	Value    string
}

func (s *Symbol) sexp()           {}
func (s *Symbol) Pos() Position   { return s.Position }
func (s *Symbol) String() string  { return s.Value }

// Keyword represents a keyword argument (starts with :)
type Keyword struct {
	Position Position
	Name     string // without the : prefix
}

func (k *Keyword) sexp()          {}
func (k *Keyword) Pos() Position  { return k.Position }
func (k *Keyword) String() string { return ":" + k.Name }

// String represents a string literal
type String struct {
	Position Position
	Value    string // processed value (without quotes, escapes resolved)
}

func (s *String) sexp()          {}
func (s *String) Pos() Position  { return s.Position }
func (s *String) String() string { return s.Quoted() }

// Quoted returns the string value with surrounding quotes
func (s *String) Quoted() string {
	return fmt.Sprintf("%q", s.Value)
}

// Number represents a numeric literal
type Number struct {
	Position Position
	Value    string // string representation
}

func (n *Number) sexp()          {}
func (n *Number) Pos() Position  { return n.Position }
func (n *Number) String() string { return n.Value }

// Nil represents the nil value
type Nil struct {
	Position Position
}

func (n *Nil) sexp()          {}
func (n *Nil) Pos() Position  { return n.Position }
func (n *Nil) String() string { return NilLiteral }

// List represents a list of S-expressions
type List struct {
	Position Position // position of opening paren
	Elements []SExp
}

func (l *List) sexp() {}
func (l *List) Pos() Position { return l.Position }
func (l *List) String() string {
	if len(l.Elements) == 0 {
		return ParenOC
	}
	return ParenO + strings.Join(l.ElementStrings(), " ") + ParenC
}

// ElementStrings returns string representations of all elements
func (l *List) ElementStrings() []string {
	parts := make([]string, len(l.Elements))
	for i, elem := range l.Elements {
		parts[i] = elem.String()
	}
	return parts
}

// Parser parses S-expressions from token stream
type Parser struct {
	lexer   *Lexer
	current Token
	peek    Token
	errors  []string
}

// NewParser creates a new parser for the given input
func NewParser(input string) *Parser {
	p := &Parser{
		lexer:  NewLexer(input),
		errors: []string{},
	}
	// Read two tokens to initialize current and peek
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances the parser to the next token
func (p *Parser) nextToken() {
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

// currentTokenIs checks if current token matches the given type
func (p *Parser) currentTokenIs(t TokenType) bool {
	return p.current.Type == t
}

// peekTokenIs checks if peek token matches the given type
func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peek.Type == t
}

// expectPeek advances if peek matches expected type, otherwise adds error
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// addError appends a formatted error message
func (p *Parser) addError(msg string) {
	errMsg := fmt.Sprintf("line %d, column %d: %s",
		p.current.Line, p.current.Column, msg)
	p.errors = append(p.errors, errMsg)
}

// peekError adds an error for unexpected token type
func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peek.Type)
	p.addError(msg)
}

// Errors returns all parsing errors encountered
func (p *Parser) Errors() []string {
	return p.errors
}

// Parse parses the input and returns the S-expression tree
func (p *Parser) Parse() (SExp, error) {
	if p.currentTokenIs(EOF) {
		return nil, ErrNoSExpression
	}

	sexp := p.parseSExp()

	if len(p.errors) > 0 {
		return nil, ErrParseErrors(p.errors)
	}

	return sexp, nil
}

// ParseList parses input expecting a list and returns it
func (p *Parser) ParseList() (*List, error) {
	if !p.currentTokenIs(LPAREN) {
		return nil, ErrExpectedList(p.current.Type)
	}

	list := p.parseList()

	if len(p.errors) > 0 {
		return nil, ErrParseErrors(p.errors)
	}

	return list, nil
}

// parseSExp dispatches to appropriate parser based on token type
func (p *Parser) parseSExp() SExp {
	switch p.current.Type {
	case LPAREN:
		return p.parseList()
	case SYMBOL:
		return p.parseSymbol()
	case KEYWORD:
		return p.parseKeyword()
	case STRING:
		return p.parseString()
	case NUMBER:
		return p.parseNumber()
	case NIL:
		return p.parseNil()
	case ILLEGAL:
		p.addError(ErrIllegalToken(p.current.Literal).Error())
		return nil
	default:
		p.addError(ErrUnexpectedToken(p.current.Type).Error())
		return nil
	}
}

// parseList parses a list S-expression
func (p *Parser) parseList() *List {
	list := &List{
		Position: Position{
			Offset: p.current.Pos,
			Line:   p.current.Line,
			Column: p.current.Column,
		},
		Elements: []SExp{},
	}

	// Advance past LPAREN
	p.nextToken()

	// Parse elements until RPAREN or EOF
	for !p.currentTokenIs(RPAREN) && !p.currentTokenIs(EOF) {
		elem := p.parseSExp()
		if elem != nil {
			list.Elements = append(list.Elements, elem)
		}
		p.nextToken()
	}

	// Check we ended on RPAREN
	if !p.currentTokenIs(RPAREN) {
		p.addError(ErrUnterminatedList.Error())
		return list
	}

	return list
}

// parseSymbol parses a symbol
func (p *Parser) parseSymbol() *Symbol {
	return &Symbol{
		Position: Position{
			Offset: p.current.Pos,
			Line:   p.current.Line,
			Column: p.current.Column,
		},
		Value: p.current.Literal,
	}
}

// parseKeyword parses a keyword
func (p *Parser) parseKeyword() *Keyword {
	// Remove leading ':' from keyword name
	name := p.current.Literal
	if len(name) > 0 && name[0] == ':' {
		name = name[1:]
	}

	return &Keyword{
		Position: Position{
			Offset: p.current.Pos,
			Line:   p.current.Line,
			Column: p.current.Column,
		},
		Name: name,
	}
}

// parseString parses a string literal
func (p *Parser) parseString() *String {
	// Remove surrounding quotes and keep escaped content as-is
	// The lexer already processed escape sequences
	value := p.current.Literal
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	return &String{
		Position: Position{
			Offset: p.current.Pos,
			Line:   p.current.Line,
			Column: p.current.Column,
		},
		Value: value,
	}
}

// parseNumber parses a number literal
func (p *Parser) parseNumber() *Number {
	return &Number{
		Position: Position{
			Offset: p.current.Pos,
			Line:   p.current.Line,
			Column: p.current.Column,
		},
		Value: p.current.Literal,
	}
}

// parseNil parses nil
func (p *Parser) parseNil() *Nil {
	return &Nil{
		Position: Position{
			Offset: p.current.Pos,
			Line:   p.current.Line,
			Column: p.current.Column,
		},
	}
}
