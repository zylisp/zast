package writer

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestWriteString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `"hello"`},
		{`hello "world"`, `"hello \"world\""`},
		{"hello\nworld", `"hello\nworld"`},
		{`hello\world`, `"hello\\world"`},
		{"tab\there", `"tab\there"`},
		{"cr\rhere", `"cr\rhere"`},
	}

	for _, tt := range tests {
		writer := New(nil)
		writer.writeString(tt.input)
		result := writer.buf.String()
		if result != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, result)
		}
	}
}

func TestWriteTokenMapping(t *testing.T) {
	tests := []struct {
		tok      token.Token
		expected string
	}{
		{token.IMPORT, "IMPORT"},
		{token.CONST, "CONST"},
		{token.TYPE, "TYPE"},
		{token.VAR, "VAR"},
		{token.INT, "INT"},
		{token.FLOAT, "FLOAT"},
		{token.IMAG, "IMAG"},
		{token.CHAR, "CHAR"},
		{token.STRING, "STRING"},
	}

	for _, tt := range tests {
		writer := New(nil)
		writer.writeToken(tt.tok)
		result := writer.buf.String()
		if result != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, result)
		}
	}
}

// TestWriteBool tests boolean writing
func TestWriteBool(t *testing.T) {
	tests := []struct {
		input    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		writer := New(nil)
		writer.writeBool(tt.input)
		result := writer.buf.String()
		if result != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, result)
		}
	}
}

// TestWriteChanDir tests channel direction writing
func TestWriteChanDir(t *testing.T) {
	tests := []struct {
		input    ast.ChanDir
		expected string
	}{
		{ast.SEND, "SEND"},
		{ast.RECV, "RECV"},
		{ast.SEND | ast.RECV, "SEND_RECV"},
	}

	for _, tt := range tests {
		writer := New(nil)
		writer.writeChanDir(tt.input)
		result := writer.buf.String()
		if result != tt.expected {
			t.Fatalf("expected %q for dir %d, got %q", tt.expected, tt.input, result)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
