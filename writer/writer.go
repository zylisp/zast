package writer

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strings"
)

// Writer writes Go AST nodes as S-expressions
type Writer struct {
	fset   *token.FileSet
	buf    strings.Builder
	config *Config
}

// NewWriter creates a new S-expression writer with default configuration
func New(fset *token.FileSet) *Writer {
	return NewWithConfig(fset, DefaultConfig())
}

// NewWithConfig creates a new S-expression writer with custom configuration
func NewWithConfig(fset *token.FileSet, config *Config) *Writer {
	return &Writer{
		fset:   fset,
		config: config,
	}
}

// Helper methods for writing primitives

func (w *Writer) writeString(s string) {
	w.buf.WriteString(`"`)
	for _, ch := range s {
		switch ch {
		case '"':
			w.buf.WriteString(`\"`)
		case '\\':
			w.buf.WriteString(`\\`)
		case '\n':
			w.buf.WriteString(`\n`)
		case '\t':
			w.buf.WriteString(`\t`)
		case '\r':
			w.buf.WriteString(`\r`)
		default:
			w.buf.WriteRune(ch)
		}
	}
	w.buf.WriteString(`"`)
}

func (w *Writer) writeSymbol(s string) {
	w.buf.WriteString(s)
}

func (w *Writer) writeKeyword(name string) {
	w.buf.WriteString(":")
	w.buf.WriteString(name)
}

func (w *Writer) writePos(pos token.Pos) {
	w.buf.WriteString(fmt.Sprintf("%d", pos))
}

func (w *Writer) writeBool(b bool) {
	if b {
		w.writeSymbol("true")
	} else {
		w.writeSymbol("false")
	}
}

func (w *Writer) writeChanDir(dir ast.ChanDir) {
	switch dir {
	case ast.SEND:
		w.writeSymbol("SEND")
	case ast.RECV:
		w.writeSymbol("RECV")
	case ast.SEND | ast.RECV:
		w.writeSymbol("SEND_RECV")
	default:
		w.writeSymbol("SEND_RECV")
	}
}

func (w *Writer) writeToken(tok token.Token) {
	switch tok {
	// Keywords
	case token.IMPORT:
		w.writeSymbol("IMPORT")
	case token.CONST:
		w.writeSymbol("CONST")
	case token.TYPE:
		w.writeSymbol("TYPE")
	case token.VAR:
		w.writeSymbol("VAR")
	case token.BREAK:
		w.writeSymbol("BREAK")
	case token.CONTINUE:
		w.writeSymbol("CONTINUE")
	case token.GOTO:
		w.writeSymbol("GOTO")
	case token.FALLTHROUGH:
		w.writeSymbol("FALLTHROUGH")

	// Literal types
	case token.INT:
		w.writeSymbol("INT")
	case token.FLOAT:
		w.writeSymbol("FLOAT")
	case token.IMAG:
		w.writeSymbol("IMAG")
	case token.CHAR:
		w.writeSymbol("CHAR")
	case token.STRING:
		w.writeSymbol("STRING")

	// Operators
	case token.ADD:
		w.writeSymbol("ADD")
	case token.SUB:
		w.writeSymbol("SUB")
	case token.MUL:
		w.writeSymbol("MUL")
	case token.QUO:
		w.writeSymbol("QUO")
	case token.REM:
		w.writeSymbol("REM")
	case token.AND:
		w.writeSymbol("AND")
	case token.OR:
		w.writeSymbol("OR")
	case token.XOR:
		w.writeSymbol("XOR")
	case token.SHL:
		w.writeSymbol("SHL")
	case token.SHR:
		w.writeSymbol("SHR")
	case token.AND_NOT:
		w.writeSymbol("AND_NOT")
	case token.LAND:
		w.writeSymbol("LAND")
	case token.LOR:
		w.writeSymbol("LOR")
	case token.ARROW:
		w.writeSymbol("ARROW")
	case token.INC:
		w.writeSymbol("INC")
	case token.DEC:
		w.writeSymbol("DEC")

	// Comparison
	case token.EQL:
		w.writeSymbol("EQL")
	case token.LSS:
		w.writeSymbol("LSS")
	case token.GTR:
		w.writeSymbol("GTR")
	case token.ASSIGN:
		w.writeSymbol("ASSIGN")
	case token.NOT:
		w.writeSymbol("NOT")
	case token.NEQ:
		w.writeSymbol("NEQ")
	case token.LEQ:
		w.writeSymbol("LEQ")
	case token.GEQ:
		w.writeSymbol("GEQ")
	case token.DEFINE:
		w.writeSymbol("DEFINE")

	// Assignment operators
	case token.ADD_ASSIGN:
		w.writeSymbol("ADD_ASSIGN")
	case token.SUB_ASSIGN:
		w.writeSymbol("SUB_ASSIGN")
	case token.MUL_ASSIGN:
		w.writeSymbol("MUL_ASSIGN")
	case token.QUO_ASSIGN:
		w.writeSymbol("QUO_ASSIGN")
	case token.REM_ASSIGN:
		w.writeSymbol("REM_ASSIGN")
	case token.AND_ASSIGN:
		w.writeSymbol("AND_ASSIGN")
	case token.OR_ASSIGN:
		w.writeSymbol("OR_ASSIGN")
	case token.XOR_ASSIGN:
		w.writeSymbol("XOR_ASSIGN")
	case token.SHL_ASSIGN:
		w.writeSymbol("SHL_ASSIGN")
	case token.SHR_ASSIGN:
		w.writeSymbol("SHR_ASSIGN")
	case token.AND_NOT_ASSIGN:
		w.writeSymbol("AND_NOT_ASSIGN")

	default:
		w.writeSymbol("ILLEGAL")
	}
}

func (w *Writer) writeSpace() {
	w.buf.WriteString(" ")
}

func (w *Writer) openList() {
	w.buf.WriteString("(")
}

func (w *Writer) closeList() {
	w.buf.WriteString(")")
}

// WriteTo writes the result to an io.Writer
func (w *Writer) WriteTo(wr io.Writer) (int64, error) {
	n, err := wr.Write([]byte(w.buf.String()))
	return int64(n), err
}
