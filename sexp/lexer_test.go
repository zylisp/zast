package sexp

import (
	"testing"
)

func TestBasicTokens(t *testing.T) {
	input := "() nil"
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LPAREN, "("},
		{RPAREN, ")"},
		{NIL, "nil"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestSymbols(t *testing.T) {
	input := "File GenDecl IMPORT main foo-bar test_123"
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SYMBOL, "File"},
		{SYMBOL, "GenDecl"},
		{SYMBOL, "IMPORT"},
		{SYMBOL, "main"},
		{SYMBOL, "foo-bar"},
		{SYMBOL, "test_123"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := ":name :package :tok :doc"
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{KEYWORD, ":name"},
		{KEYWORD, ":package"},
		{KEYWORD, ":tok"},
		{KEYWORD, ":doc"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		input           string
		expectedType    TokenType
		expectedLiteral string
	}{
		{`"hello"`, STRING, `"hello"`},
		{`"hello world"`, STRING, `"hello world"`},
		{`"hello\nworld"`, STRING, "\"hello\nworld\""},
		{`"hello\tworld"`, STRING, "\"hello\tworld\""},
		{`"say \"hi\""`, STRING, `"say "hi""`},
		{`"backslash\\"`, STRING, `"backslash\"`},
		{`"unterminated`, ILLEGAL, "unterminated string"},
	}

	for i, tt := range tests {
		l := NewLexer(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNumbers(t *testing.T) {
	input := "42 -10 3.14 0 -0.5 +123"
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{NUMBER, "42"},
		{NUMBER, "-10"},
		{NUMBER, "3.14"},
		{NUMBER, "0"},
		{NUMBER, "-0.5"},
		{NUMBER, "+123"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := "; this is a comment\n(File ; inline comment\n)"
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LPAREN, "("},
		{SYMBOL, "File"},
		{RPAREN, ")"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestWhitespace(t *testing.T) {
	input := "(  foo  \n  bar\t)"
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LPAREN, "("},
		{SYMBOL, "foo"},
		{SYMBOL, "bar"},
		{RPAREN, ")"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestPositionTracking(t *testing.T) {
	input := "(File\n  :name)"
	tests := []struct {
		expectedType   TokenType
		expectedLine   int
		expectedColumn int
	}{
		{LPAREN, 1, 1},
		{SYMBOL, 1, 2},
		{KEYWORD, 2, 3},
		{RPAREN, 2, 8},
		{EOF, 2, 8},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Line != tt.expectedLine {
			t.Fatalf("tests[%d] - line wrong. expected=%d, got=%d",
				i, tt.expectedLine, tok.Line)
		}
		if tok.Column != tt.expectedColumn {
			t.Fatalf("tests[%d] - column wrong. expected=%d, got=%d",
				i, tt.expectedColumn, tok.Column)
		}
	}
}

func TestComplexSExpression(t *testing.T) {
	input := `(File :package 1 :name (Ident :namepos 9 :name "main"))`
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LPAREN, "("},
		{SYMBOL, "File"},
		{KEYWORD, ":package"},
		{NUMBER, "1"},
		{KEYWORD, ":name"},
		{LPAREN, "("},
		{SYMBOL, "Ident"},
		{KEYWORD, ":namepos"},
		{NUMBER, "9"},
		{KEYWORD, ":name"},
		{STRING, `"main"`},
		{RPAREN, ")"},
		{RPAREN, ")"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestPeek(t *testing.T) {
	input := "(File :name)"
	l := NewLexer(input)

	// Peek should return first token without consuming
	peek1 := l.Peek()
	if peek1.Type != LPAREN {
		t.Fatalf("peek1 - tokentype wrong. expected=%q, got=%q", LPAREN, peek1.Type)
	}

	// Peeking again should return same token
	peek2 := l.Peek()
	if peek2.Type != LPAREN {
		t.Fatalf("peek2 - tokentype wrong. expected=%q, got=%q", LPAREN, peek2.Type)
	}

	// NextToken should return the peeked token
	tok := l.NextToken()
	if tok.Type != LPAREN {
		t.Fatalf("tok - tokentype wrong. expected=%q, got=%q", LPAREN, tok.Type)
	}

	// Now peek should return next token
	peek3 := l.Peek()
	if peek3.Type != SYMBOL || peek3.Literal != "File" {
		t.Fatalf("peek3 - wrong. expected SYMBOL 'File', got=%q %q", peek3.Type, peek3.Literal)
	}
}

func TestEmptyInput(t *testing.T) {
	input := ""
	l := NewLexer(input)
	tok := l.NextToken()
	if tok.Type != EOF {
		t.Fatalf("empty input should return EOF, got=%q", tok.Type)
	}
}

func TestEmptyList(t *testing.T) {
	input := "()"
	tests := []struct {
		expectedType TokenType
	}{
		{LPAREN},
		{RPAREN},
		{EOF},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
	}
}

func TestIllegalCharacter(t *testing.T) {
	input := "(@)"
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LPAREN, "("},
		{ILLEGAL, "@"},
		{RPAREN, ")"},
		{EOF, ""},
	}

	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// Test String methods for coverage
func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{EOF, "EOF"},
		{ILLEGAL, "ILLEGAL"},
		{LPAREN, "LPAREN"},
		{RPAREN, "RPAREN"},
		{SYMBOL, "SYMBOL"},
		{STRING, "STRING"},
		{NUMBER, "NUMBER"},
		{KEYWORD, "KEYWORD"},
		{NIL, "NIL"},
	}

	for _, tt := range tests {
		result := tt.tokenType.String()
		if result != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, result)
		}
	}

	// Test unknown token type
	unknownToken := TokenType(999)
	if unknownToken.String() != "UNKNOWN" {
		t.Fatalf("expected UNKNOWN for invalid token type")
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{
		Type:    SYMBOL,
		Literal: "foo",
		Pos:     10,
		Line:    1,
		Column:  5,
	}

	str := tok.String()
	if str == "" {
		t.Fatalf("expected non-empty string representation")
	}
}

func TestStringEscapeSequences(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"tab\there"`, "\"tab\there\""},
		{`"cr\rhere"`, "\"cr\rhere\""},
		{`"unknown\xescape"`, "\"unknown\\xescape\""},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tok := lexer.NextToken()

		if tok.Type != STRING {
			t.Fatalf("expected STRING token, got %v", tok.Type)
		}

		if tok.Literal != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, tok.Literal)
		}
	}
}

func TestPeekCharAtEOF(t *testing.T) {
	input := "x"
	lexer := NewLexer(input)

	// Read the character
	lexer.readChar()

	// Now we're at EOF, peek should return 0
	peeked := lexer.peekChar()
	if peeked != 0 {
		t.Fatalf("expected peekChar to return 0 at EOF, got %v", peeked)
	}
}
