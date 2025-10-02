package sexp

import (
	"fmt"
	"strings"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	EOF TokenType = iota
	ILLEGAL

	// Delimiters
	LPAREN // (
	RPAREN // )

	// Literals
	SYMBOL  // foo, bar, GenDecl, etc.
	STRING  // "hello"
	NUMBER  // 42, 3.14, -10
	KEYWORD // :name, :package, :tok
	NIL     // nil
)

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case EOF:
		return "EOF"
	case ILLEGAL:
		return "ILLEGAL"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case SYMBOL:
		return "SYMBOL"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case KEYWORD:
		return "KEYWORD"
	case NIL:
		return "NIL"
	default:
		return "UNKNOWN"
	}
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Pos     int // byte offset in input
	Line    int // line number (1-based)
	Column  int // column number (1-based)
}

// String returns a string representation of the token
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Pos: %d, Line: %d, Column: %d}",
		t.Type, t.Literal, t.Pos, t.Line, t.Column)
}

// Lexer tokenizes S-expression input
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position (after current char)
	ch           byte // current char under examination
	line         int  // current line number (1-based)
	column       int  // current column number (1-based)
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar() // Initialize first character
	return l
}

// readChar advances the lexer to the next character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // null byte indicates EOF
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++

	// Update line and column tracking
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else if l.ch != 0 {
		l.column++
	}
}

// peekChar returns the next character without advancing
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment skips characters until end of line
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// isLetter checks if a byte is a letter
func (l *Lexer) isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

// isDigit checks if a byte is a digit
func (l *Lexer) isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// isSymbolChar checks if a byte is allowed in a symbol
func (l *Lexer) isSymbolChar(ch byte) bool {
	return l.isLetter(ch) || l.isDigit(ch) ||
		ch == '_' || ch == '-' || ch == '+' || ch == '*' ||
		ch == '/' || ch == '<' || ch == '>' || ch == '=' ||
		ch == '!' || ch == '?'
}

// readSymbol reads a symbol from the input
func (l *Lexer) readSymbol() string {
	startPos := l.position
	for l.isSymbolChar(l.ch) {
		l.readChar()
	}
	return l.input[startPos:l.position]
}

// readKeyword reads a keyword from the input (starts with :)
func (l *Lexer) readKeyword() string {
	startPos := l.position
	l.readChar() // skip the ':'
	for l.isSymbolChar(l.ch) {
		l.readChar()
	}
	return l.input[startPos:l.position]
}

// readString reads a string literal from the input
func (l *Lexer) readString() string {
	var result strings.Builder
	result.WriteByte('"') // Include opening quote

	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			default:
				// Unknown escape, keep the backslash
				result.WriteByte('\\')
				result.WriteByte(l.ch)
			}
			l.readChar()
		} else {
			result.WriteByte(l.ch)
			l.readChar()
		}
	}

	if l.ch == '"' {
		result.WriteByte('"') // Include closing quote
		l.readChar()          // skip closing quote
		return result.String()
	}

	// Unterminated string
	return ""
}

// readNumber reads a number from the input
func (l *Lexer) readNumber() string {
	startPos := l.position

	// Handle optional sign
	if l.ch == '-' || l.ch == '+' {
		l.readChar()
	}

	// Read integer part
	for l.isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && l.isDigit(l.peekChar()) {
		l.readChar() // skip '.'
		for l.isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[startPos:l.position]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	// Handle comments
	if l.ch == ';' {
		l.skipComment()
		l.skipWhitespace()
	}

	tok.Pos = l.position
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case 0:
		tok.Type = EOF
		tok.Literal = ""
	case '(':
		tok.Type = LPAREN
		tok.Literal = "("
		l.readChar()
	case ')':
		tok.Type = RPAREN
		tok.Literal = ")"
		l.readChar()
	case ':':
		tok.Type = KEYWORD
		tok.Literal = l.readKeyword()
	case '"':
		str := l.readString()
		if str == "" {
			tok.Type = ILLEGAL
			tok.Literal = "unterminated string"
		} else {
			tok.Type = STRING
			tok.Literal = str
		}
	default:
		if l.isLetter(l.ch) {
			tok.Literal = l.readSymbol()
			// Check for special symbol 'nil'
			if tok.Literal == "nil" {
				tok.Type = NIL
			} else {
				tok.Type = SYMBOL
			}
			return tok
		} else if l.isDigit(l.ch) || (l.ch == '-' || l.ch == '+') && l.isDigit(l.peekChar()) {
			tok.Type = NUMBER
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok.Type = ILLEGAL
			tok.Literal = string(l.ch)
			l.readChar()
		}
	}

	return tok
}

// Peek returns the next token without consuming it
func (l *Lexer) Peek() Token {
	// Save current state
	savedPos := l.position
	savedReadPos := l.readPosition
	savedCh := l.ch
	savedLine := l.line
	savedColumn := l.column

	// Get next token
	tok := l.NextToken()

	// Restore state
	l.position = savedPos
	l.readPosition = savedReadPos
	l.ch = savedCh
	l.line = savedLine
	l.column = savedColumn

	return tok
}
