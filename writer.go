package zast

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strings"
)

// WriterConfig holds configuration for the AST writer
type WriterConfig struct {
	// Maximum position to scan when collecting files from FileSet
	MaxFileSetScan int // default: 1000000

	// Early termination: stop scanning after this many positions with no new files
	FileSearchWindow int // default: 10000
}

// DefaultWriterConfig returns the default writer configuration
func DefaultWriterConfig() *WriterConfig {
	return &WriterConfig{
		MaxFileSetScan:   1000000,
		FileSearchWindow: 10000,
	}
}

// Writer writes Go AST nodes as S-expressions
type Writer struct {
	fset   *token.FileSet
	buf    strings.Builder
	config *WriterConfig
}

// NewWriter creates a new S-expression writer with default configuration
func NewWriter(fset *token.FileSet) *Writer {
	return NewWriterWithConfig(fset, DefaultWriterConfig())
}

// NewWriterWithConfig creates a new S-expression writer with custom configuration
func NewWriterWithConfig(fset *token.FileSet, config *WriterConfig) *Writer {
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

// Node writers

func (w *Writer) writeIdent(ident *ast.Ident) error {
	if ident == nil {
		w.writeSymbol("nil")
		return nil
	}

	w.openList()
	w.writeSymbol("Ident")
	w.writeSpace()
	w.writeKeyword("namepos")
	w.writeSpace()
	w.writePos(ident.NamePos)
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	w.writeString(ident.Name)
	w.writeSpace()
	w.writeKeyword("obj")
	w.writeSpace()
	w.writeSymbol("nil") // Objects handled separately if needed
	w.closeList()
	return nil
}

func (w *Writer) writeBasicLit(lit *ast.BasicLit) error {
	if lit == nil {
		w.writeSymbol("nil")
		return nil
	}

	w.openList()
	w.writeSymbol("BasicLit")
	w.writeSpace()
	w.writeKeyword("valuepos")
	w.writeSpace()
	w.writePos(lit.ValuePos)
	w.writeSpace()
	w.writeKeyword("kind")
	w.writeSpace()
	w.writeToken(lit.Kind)
	w.writeSpace()
	w.writeKeyword("value")
	w.writeSpace()
	w.writeString(lit.Value)
	w.closeList()
	return nil
}

func (w *Writer) writeSelectorExpr(expr *ast.SelectorExpr) error {
	w.openList()
	w.writeSymbol("SelectorExpr")
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("sel")
	w.writeSpace()
	if err := w.writeIdent(expr.Sel); err != nil {
		return err
	}
	w.closeList()
	return nil
}

func (w *Writer) writeCallExpr(expr *ast.CallExpr) error {
	w.openList()
	w.writeSymbol("CallExpr")
	w.writeSpace()
	w.writeKeyword("fun")
	w.writeSpace()
	if err := w.writeExpr(expr.Fun); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("lparen")
	w.writeSpace()
	w.writePos(expr.Lparen)
	w.writeSpace()
	w.writeKeyword("args")
	w.writeSpace()
	if err := w.writeExprList(expr.Args); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("ellipsis")
	w.writeSpace()
	w.writePos(expr.Ellipsis)
	w.writeSpace()
	w.writeKeyword("rparen")
	w.writeSpace()
	w.writePos(expr.Rparen)
	w.closeList()
	return nil
}

func (w *Writer) writeBinaryExpr(expr *ast.BinaryExpr) error {
	w.openList()
	w.writeSymbol("BinaryExpr")
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("oppos")
	w.writeSpace()
	w.writePos(expr.OpPos)
	w.writeSpace()
	w.writeKeyword("op")
	w.writeSpace()
	w.writeToken(expr.Op)
	w.writeSpace()
	w.writeKeyword("y")
	w.writeSpace()
	if err := w.writeExpr(expr.Y); err != nil {
		return err
	}
	w.closeList()
	return nil
}

func (w *Writer) writeParenExpr(expr *ast.ParenExpr) error {
	w.openList()
	w.writeSymbol("ParenExpr")
	w.writeSpace()
	w.writeKeyword("lparen")
	w.writeSpace()
	w.writePos(expr.Lparen)
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("rparen")
	w.writeSpace()
	w.writePos(expr.Rparen)
	w.closeList()
	return nil
}

func (w *Writer) writeStarExpr(expr *ast.StarExpr) error {
	w.openList()
	w.writeSymbol("StarExpr")
	w.writeSpace()
	w.writeKeyword("star")
	w.writeSpace()
	w.writePos(expr.Star)
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.closeList()
	return nil
}

func (w *Writer) writeIndexExpr(expr *ast.IndexExpr) error {
	w.openList()
	w.writeSymbol("IndexExpr")
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("lbrack")
	w.writeSpace()
	w.writePos(expr.Lbrack)
	w.writeSpace()
	w.writeKeyword("index")
	w.writeSpace()
	if err := w.writeExpr(expr.Index); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("rbrack")
	w.writeSpace()
	w.writePos(expr.Rbrack)
	w.closeList()
	return nil
}

func (w *Writer) writeSliceExpr(expr *ast.SliceExpr) error {
	w.openList()
	w.writeSymbol("SliceExpr")
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("lbrack")
	w.writeSpace()
	w.writePos(expr.Lbrack)
	w.writeSpace()
	w.writeKeyword("low")
	w.writeSpace()
	if err := w.writeExpr(expr.Low); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("high")
	w.writeSpace()
	if err := w.writeExpr(expr.High); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("max")
	w.writeSpace()
	if err := w.writeExpr(expr.Max); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("slice3")
	w.writeSpace()
	w.writeBool(expr.Slice3)
	w.writeSpace()
	w.writeKeyword("rbrack")
	w.writeSpace()
	w.writePos(expr.Rbrack)
	w.closeList()
	return nil
}

func (w *Writer) writeKeyValueExpr(expr *ast.KeyValueExpr) error {
	w.openList()
	w.writeSymbol("KeyValueExpr")
	w.writeSpace()
	w.writeKeyword("key")
	w.writeSpace()
	if err := w.writeExpr(expr.Key); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("colon")
	w.writeSpace()
	w.writePos(expr.Colon)
	w.writeSpace()
	w.writeKeyword("value")
	w.writeSpace()
	if err := w.writeExpr(expr.Value); err != nil {
		return err
	}
	w.closeList()
	return nil
}

func (w *Writer) writeUnaryExpr(expr *ast.UnaryExpr) error {
	w.openList()
	w.writeSymbol("UnaryExpr")
	w.writeSpace()
	w.writeKeyword("oppos")
	w.writeSpace()
	w.writePos(expr.OpPos)
	w.writeSpace()
	w.writeKeyword("op")
	w.writeSpace()
	w.writeToken(expr.Op)
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeTypeAssertExpr writes a TypeAssertExpr node
func (w *Writer) writeTypeAssertExpr(expr *ast.TypeAssertExpr) error {
	w.openList()
	w.writeSymbol("TypeAssertExpr")
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(expr.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("lparen")
	w.writeSpace()
	w.writePos(expr.Lparen)
	w.writeSpace()
	w.writeKeyword("type")
	w.writeSpace()
	if err := w.writeExpr(expr.Type); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("rparen")
	w.writeSpace()
	w.writePos(expr.Rparen)
	w.closeList()
	return nil
}

// Expression dispatcher
func (w *Writer) writeExpr(expr ast.Expr) error {
	if expr == nil {
		w.writeSymbol("nil")
		return nil
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return w.writeIdent(e)
	case *ast.BasicLit:
		return w.writeBasicLit(e)
	case *ast.CallExpr:
		return w.writeCallExpr(e)
	case *ast.SelectorExpr:
		return w.writeSelectorExpr(e)
	case *ast.UnaryExpr:
		return w.writeUnaryExpr(e)
	case *ast.BinaryExpr:
		return w.writeBinaryExpr(e)
	case *ast.ParenExpr:
		return w.writeParenExpr(e)
	case *ast.StarExpr:
		return w.writeStarExpr(e)
	case *ast.IndexExpr:
		return w.writeIndexExpr(e)
	case *ast.SliceExpr:
		return w.writeSliceExpr(e)
	case *ast.KeyValueExpr:
		return w.writeKeyValueExpr(e)
	case *ast.ArrayType:
		return w.writeArrayType(e)
	case *ast.MapType:
		return w.writeMapType(e)
	case *ast.ChanType:
		return w.writeChanType(e)
	case *ast.TypeAssertExpr:
		return w.writeTypeAssertExpr(e)
	default:
		return ErrUnknownExprType(expr)
	}
}

func (w *Writer) writeExprStmt(stmt *ast.ExprStmt) error {
	w.openList()
	w.writeSymbol("ExprStmt")
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(stmt.X); err != nil {
		return err
	}
	w.closeList()
	return nil
}

func (w *Writer) writeBlockStmt(stmt *ast.BlockStmt) error {
	if stmt == nil {
		w.writeSymbol("nil")
		return nil
	}

	w.openList()
	w.writeSymbol("BlockStmt")
	w.writeSpace()
	w.writeKeyword("lbrace")
	w.writeSpace()
	w.writePos(stmt.Lbrace)
	w.writeSpace()
	w.writeKeyword("list")
	w.writeSpace()
	if err := w.writeStmtList(stmt.List); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("rbrace")
	w.writeSpace()
	w.writePos(stmt.Rbrace)
	w.closeList()
	return nil
}

// writeReturnStmt writes a ReturnStmt node
func (w *Writer) writeReturnStmt(stmt *ast.ReturnStmt) error {
	w.openList()
	w.writeSymbol("ReturnStmt")
	w.writeSpace()
	w.writeKeyword("return")
	w.writeSpace()
	w.writePos(stmt.Return)
	w.writeSpace()
	w.writeKeyword("results")
	w.writeSpace()
	if err := w.writeExprList(stmt.Results); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeAssignStmt writes an AssignStmt node
func (w *Writer) writeAssignStmt(stmt *ast.AssignStmt) error {
	w.openList()
	w.writeSymbol("AssignStmt")
	w.writeSpace()
	w.writeKeyword("lhs")
	w.writeSpace()
	if err := w.writeExprList(stmt.Lhs); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("tokpos")
	w.writeSpace()
	w.writePos(stmt.TokPos)
	w.writeSpace()
	w.writeKeyword("tok")
	w.writeSpace()
	w.writeToken(stmt.Tok)
	w.writeSpace()
	w.writeKeyword("rhs")
	w.writeSpace()
	if err := w.writeExprList(stmt.Rhs); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeIncDecStmt writes an IncDecStmt node
func (w *Writer) writeIncDecStmt(stmt *ast.IncDecStmt) error {
	w.openList()
	w.writeSymbol("IncDecStmt")
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(stmt.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("tokpos")
	w.writeSpace()
	w.writePos(stmt.TokPos)
	w.writeSpace()
	w.writeKeyword("tok")
	w.writeSpace()
	w.writeToken(stmt.Tok)
	w.closeList()
	return nil
}

// writeBranchStmt writes a BranchStmt node
func (w *Writer) writeBranchStmt(stmt *ast.BranchStmt) error {
	w.openList()
	w.writeSymbol("BranchStmt")
	w.writeSpace()
	w.writeKeyword("tokpos")
	w.writeSpace()
	w.writePos(stmt.TokPos)
	w.writeSpace()
	w.writeKeyword("tok")
	w.writeSpace()
	w.writeToken(stmt.Tok)
	w.writeSpace()
	w.writeKeyword("label")
	w.writeSpace()
	if err := w.writeIdent(stmt.Label); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeDeferStmt writes a DeferStmt node
func (w *Writer) writeDeferStmt(stmt *ast.DeferStmt) error {
	w.openList()
	w.writeSymbol("DeferStmt")
	w.writeSpace()
	w.writeKeyword("defer")
	w.writeSpace()
	w.writePos(stmt.Defer)
	w.writeSpace()
	w.writeKeyword("call")
	w.writeSpace()
	if err := w.writeCallExpr(stmt.Call); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeGoStmt writes a GoStmt node
func (w *Writer) writeGoStmt(stmt *ast.GoStmt) error {
	w.openList()
	w.writeSymbol("GoStmt")
	w.writeSpace()
	w.writeKeyword("go")
	w.writeSpace()
	w.writePos(stmt.Go)
	w.writeSpace()
	w.writeKeyword("call")
	w.writeSpace()
	if err := w.writeCallExpr(stmt.Call); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeSendStmt writes a SendStmt node
func (w *Writer) writeSendStmt(stmt *ast.SendStmt) error {
	w.openList()
	w.writeSymbol("SendStmt")
	w.writeSpace()
	w.writeKeyword("chan")
	w.writeSpace()
	if err := w.writeExpr(stmt.Chan); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("arrow")
	w.writeSpace()
	w.writePos(stmt.Arrow)
	w.writeSpace()
	w.writeKeyword("value")
	w.writeSpace()
	if err := w.writeExpr(stmt.Value); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeEmptyStmt writes an EmptyStmt node
func (w *Writer) writeEmptyStmt(stmt *ast.EmptyStmt) error {
	w.openList()
	w.writeSymbol("EmptyStmt")
	w.writeSpace()
	w.writeKeyword("semicolon")
	w.writeSpace()
	w.writePos(stmt.Semicolon)
	w.writeSpace()
	w.writeKeyword("implicit")
	w.writeSpace()
	w.writeBool(stmt.Implicit)
	w.closeList()
	return nil
}

// writeLabeledStmt writes a LabeledStmt node
func (w *Writer) writeLabeledStmt(stmt *ast.LabeledStmt) error {
	w.openList()
	w.writeSymbol("LabeledStmt")
	w.writeSpace()
	w.writeKeyword("label")
	w.writeSpace()
	if err := w.writeIdent(stmt.Label); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("colon")
	w.writeSpace()
	w.writePos(stmt.Colon)
	w.writeSpace()
	w.writeKeyword("stmt")
	w.writeSpace()
	if err := w.writeStmt(stmt.Stmt); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeIfStmt writes an IfStmt node
func (w *Writer) writeIfStmt(stmt *ast.IfStmt) error {
	w.openList()
	w.writeSymbol("IfStmt")
	w.writeSpace()
	w.writeKeyword("if")
	w.writeSpace()
	w.writePos(stmt.If)
	w.writeSpace()
	w.writeKeyword("init")
	w.writeSpace()
	if err := w.writeStmt(stmt.Init); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("cond")
	w.writeSpace()
	if err := w.writeExpr(stmt.Cond); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(stmt.Body); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("else")
	w.writeSpace()
	if err := w.writeStmt(stmt.Else); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeForStmt writes a ForStmt node
func (w *Writer) writeForStmt(stmt *ast.ForStmt) error {
	w.openList()
	w.writeSymbol("ForStmt")
	w.writeSpace()
	w.writeKeyword("for")
	w.writeSpace()
	w.writePos(stmt.For)
	w.writeSpace()
	w.writeKeyword("init")
	w.writeSpace()
	if err := w.writeStmt(stmt.Init); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("cond")
	w.writeSpace()
	if err := w.writeExpr(stmt.Cond); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("post")
	w.writeSpace()
	if err := w.writeStmt(stmt.Post); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(stmt.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeRangeStmt writes a RangeStmt node
func (w *Writer) writeRangeStmt(stmt *ast.RangeStmt) error {
	w.openList()
	w.writeSymbol("RangeStmt")
	w.writeSpace()
	w.writeKeyword("for")
	w.writeSpace()
	w.writePos(stmt.For)
	w.writeSpace()
	w.writeKeyword("key")
	w.writeSpace()
	if err := w.writeExpr(stmt.Key); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("value")
	w.writeSpace()
	if err := w.writeExpr(stmt.Value); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("tokpos")
	w.writeSpace()
	w.writePos(stmt.TokPos)
	w.writeSpace()
	w.writeKeyword("tok")
	w.writeSpace()
	w.writeToken(stmt.Tok)
	w.writeSpace()
	w.writeKeyword("x")
	w.writeSpace()
	if err := w.writeExpr(stmt.X); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(stmt.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeSwitchStmt writes a SwitchStmt node
func (w *Writer) writeSwitchStmt(stmt *ast.SwitchStmt) error {
	w.openList()
	w.writeSymbol("SwitchStmt")
	w.writeSpace()
	w.writeKeyword("switch")
	w.writeSpace()
	w.writePos(stmt.Switch)
	w.writeSpace()
	w.writeKeyword("init")
	w.writeSpace()
	if err := w.writeStmt(stmt.Init); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("tag")
	w.writeSpace()
	if err := w.writeExpr(stmt.Tag); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(stmt.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeCaseClause writes a CaseClause node
func (w *Writer) writeCaseClause(stmt *ast.CaseClause) error {
	w.openList()
	w.writeSymbol("CaseClause")
	w.writeSpace()
	w.writeKeyword("case")
	w.writeSpace()
	w.writePos(stmt.Case)
	w.writeSpace()
	w.writeKeyword("list")
	w.writeSpace()
	if err := w.writeExprList(stmt.List); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("colon")
	w.writeSpace()
	w.writePos(stmt.Colon)
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeStmtList(stmt.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeTypeSwitchStmt writes a TypeSwitchStmt node
func (w *Writer) writeTypeSwitchStmt(stmt *ast.TypeSwitchStmt) error {
	w.openList()
	w.writeSymbol("TypeSwitchStmt")
	w.writeSpace()
	w.writeKeyword("switch")
	w.writeSpace()
	w.writePos(stmt.Switch)
	w.writeSpace()
	w.writeKeyword("init")
	w.writeSpace()
	if err := w.writeStmt(stmt.Init); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("assign")
	w.writeSpace()
	if err := w.writeStmt(stmt.Assign); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(stmt.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeSelectStmt writes a SelectStmt node
func (w *Writer) writeSelectStmt(stmt *ast.SelectStmt) error {
	w.openList()
	w.writeSymbol("SelectStmt")
	w.writeSpace()
	w.writeKeyword("select")
	w.writeSpace()
	w.writePos(stmt.Select)
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(stmt.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeCommClause writes a CommClause node
func (w *Writer) writeCommClause(stmt *ast.CommClause) error {
	w.openList()
	w.writeSymbol("CommClause")
	w.writeSpace()
	w.writeKeyword("case")
	w.writeSpace()
	w.writePos(stmt.Case)
	w.writeSpace()
	w.writeKeyword("comm")
	w.writeSpace()
	if err := w.writeStmt(stmt.Comm); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("colon")
	w.writeSpace()
	w.writePos(stmt.Colon)
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeStmtList(stmt.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeDeclStmt writes a DeclStmt node
func (w *Writer) writeDeclStmt(stmt *ast.DeclStmt) error {
	w.openList()
	w.writeSymbol("DeclStmt")
	w.writeSpace()
	w.writeKeyword("decl")
	w.writeSpace()
	if err := w.writeDecl(stmt.Decl); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// Statement dispatcher
func (w *Writer) writeStmt(stmt ast.Stmt) error {
	if stmt == nil {
		w.writeSymbol("nil")
		return nil
	}

	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return w.writeExprStmt(s)
	case *ast.BlockStmt:
		return w.writeBlockStmt(s)
	case *ast.ReturnStmt:
		return w.writeReturnStmt(s)
	case *ast.AssignStmt:
		return w.writeAssignStmt(s)
	case *ast.IncDecStmt:
		return w.writeIncDecStmt(s)
	case *ast.BranchStmt:
		return w.writeBranchStmt(s)
	case *ast.DeferStmt:
		return w.writeDeferStmt(s)
	case *ast.GoStmt:
		return w.writeGoStmt(s)
	case *ast.SendStmt:
		return w.writeSendStmt(s)
	case *ast.EmptyStmt:
		return w.writeEmptyStmt(s)
	case *ast.LabeledStmt:
		return w.writeLabeledStmt(s)
	case *ast.IfStmt:
		return w.writeIfStmt(s)
	case *ast.ForStmt:
		return w.writeForStmt(s)
	case *ast.RangeStmt:
		return w.writeRangeStmt(s)
	case *ast.SwitchStmt:
		return w.writeSwitchStmt(s)
	case *ast.CaseClause:
		return w.writeCaseClause(s)
	case *ast.TypeSwitchStmt:
		return w.writeTypeSwitchStmt(s)
	case *ast.SelectStmt:
		return w.writeSelectStmt(s)
	case *ast.CommClause:
		return w.writeCommClause(s)
	case *ast.DeclStmt:
		return w.writeDeclStmt(s)
	default:
		return ErrUnknownStmtType(stmt)
	}
}

func (w *Writer) writeField(field *ast.Field) error {
	w.openList()
	w.writeSymbol("Field")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	w.writeSymbol("nil") // CommentGroup - simplified for Phase 1
	w.writeSpace()
	w.writeKeyword("names")
	w.writeSpace()
	if err := w.writeIdentList(field.Names); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("type")
	w.writeSpace()
	if err := w.writeExpr(field.Type); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("tag")
	w.writeSpace()
	if field.Tag != nil {
		if err := w.writeBasicLit(field.Tag); err != nil {
			return err
		}
	} else {
		w.writeSymbol("nil")
	}
	w.writeSpace()
	w.writeKeyword("comment")
	w.writeSpace()
	w.writeSymbol("nil") // CommentGroup - simplified for Phase 1
	w.closeList()
	return nil
}

func (w *Writer) writeFieldList(fields *ast.FieldList) error {
	if fields == nil {
		w.writeSymbol("nil")
		return nil
	}

	w.openList()
	w.writeSymbol("FieldList")
	w.writeSpace()
	w.writeKeyword("opening")
	w.writeSpace()
	w.writePos(fields.Opening)
	w.writeSpace()
	w.writeKeyword("list")
	w.writeSpace()
	w.openList()
	for i, field := range fields.List {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeField(field); err != nil {
			return err
		}
	}
	w.closeList()
	w.writeSpace()
	w.writeKeyword("closing")
	w.writeSpace()
	w.writePos(fields.Closing)
	w.closeList()
	return nil
}

func (w *Writer) writeFuncType(typ *ast.FuncType) error {
	w.openList()
	w.writeSymbol("FuncType")
	w.writeSpace()
	w.writeKeyword("func")
	w.writeSpace()
	w.writePos(typ.Func)
	w.writeSpace()
	w.writeKeyword("params")
	w.writeSpace()
	if err := w.writeFieldList(typ.Params); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("results")
	w.writeSpace()
	if err := w.writeFieldList(typ.Results); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeArrayType writes an ArrayType node
func (w *Writer) writeArrayType(typ *ast.ArrayType) error {
	w.openList()
	w.writeSymbol("ArrayType")
	w.writeSpace()
	w.writeKeyword("lbrack")
	w.writeSpace()
	w.writePos(typ.Lbrack)
	w.writeSpace()
	w.writeKeyword("len")
	w.writeSpace()
	if err := w.writeExpr(typ.Len); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("elt")
	w.writeSpace()
	if err := w.writeExpr(typ.Elt); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeMapType writes a MapType node
func (w *Writer) writeMapType(typ *ast.MapType) error {
	w.openList()
	w.writeSymbol("MapType")
	w.writeSpace()
	w.writeKeyword("map")
	w.writeSpace()
	w.writePos(typ.Map)
	w.writeSpace()
	w.writeKeyword("key")
	w.writeSpace()
	if err := w.writeExpr(typ.Key); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("value")
	w.writeSpace()
	if err := w.writeExpr(typ.Value); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeChanType writes a ChanType node
func (w *Writer) writeChanType(typ *ast.ChanType) error {
	w.openList()
	w.writeSymbol("ChanType")
	w.writeSpace()
	w.writeKeyword("begin")
	w.writeSpace()
	w.writePos(typ.Begin)
	w.writeSpace()
	w.writeKeyword("arrow")
	w.writeSpace()
	w.writePos(typ.Arrow)
	w.writeSpace()
	w.writeKeyword("dir")
	w.writeSpace()
	w.writeChanDir(typ.Dir)
	w.writeSpace()
	w.writeKeyword("value")
	w.writeSpace()
	if err := w.writeExpr(typ.Value); err != nil {
		return err
	}
	w.closeList()
	return nil
}

func (w *Writer) writeImportSpec(spec *ast.ImportSpec) error {
	w.openList()
	w.writeSymbol("ImportSpec")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	w.writeSymbol("nil") // CommentGroup - simplified for Phase 1
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	if err := w.writeIdent(spec.Name); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("path")
	w.writeSpace()
	if err := w.writeBasicLit(spec.Path); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("comment")
	w.writeSpace()
	w.writeSymbol("nil") // CommentGroup - simplified for Phase 1
	w.writeSpace()
	w.writeKeyword("endpos")
	w.writeSpace()
	w.writePos(spec.EndPos)
	w.closeList()
	return nil
}

// writeValueSpec writes a ValueSpec node
func (w *Writer) writeValueSpec(spec *ast.ValueSpec) error {
	w.openList()
	w.writeSymbol("ValueSpec")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	w.writeSymbol("nil")
	w.writeSpace()
	w.writeKeyword("names")
	w.writeSpace()
	if err := w.writeIdentList(spec.Names); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("type")
	w.writeSpace()
	if err := w.writeExpr(spec.Type); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("values")
	w.writeSpace()
	if err := w.writeExprList(spec.Values); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("comment")
	w.writeSpace()
	w.writeSymbol("nil")
	w.closeList()
	return nil
}

// writeTypeSpec writes a TypeSpec node
func (w *Writer) writeTypeSpec(spec *ast.TypeSpec) error {
	w.openList()
	w.writeSymbol("TypeSpec")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	w.writeSymbol("nil")
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	if err := w.writeIdent(spec.Name); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("typeparams")
	w.writeSpace()
	w.writeSymbol("nil")
	w.writeSpace()
	w.writeKeyword("assign")
	w.writeSpace()
	w.writePos(spec.Assign)
	w.writeSpace()
	w.writeKeyword("type")
	w.writeSpace()
	if err := w.writeExpr(spec.Type); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("comment")
	w.writeSpace()
	w.writeSymbol("nil")
	w.closeList()
	return nil
}

// Spec dispatcher
func (w *Writer) writeSpec(spec ast.Spec) error {
	if spec == nil {
		w.writeSymbol("nil")
		return nil
	}

	switch s := spec.(type) {
	case *ast.ImportSpec:
		return w.writeImportSpec(s)
	case *ast.ValueSpec:
		return w.writeValueSpec(s)
	case *ast.TypeSpec:
		return w.writeTypeSpec(s)
	default:
		return ErrUnknownSpecType(spec)
	}
}

func (w *Writer) writeGenDecl(decl *ast.GenDecl) error {
	w.openList()
	w.writeSymbol("GenDecl")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	w.writeSymbol("nil") // CommentGroup - simplified for Phase 1
	w.writeSpace()
	w.writeKeyword("tok")
	w.writeSpace()
	w.writeToken(decl.Tok)
	w.writeSpace()
	w.writeKeyword("tokpos")
	w.writeSpace()
	w.writePos(decl.TokPos)
	w.writeSpace()
	w.writeKeyword("lparen")
	w.writeSpace()
	w.writePos(decl.Lparen)
	w.writeSpace()
	w.writeKeyword("specs")
	w.writeSpace()
	if err := w.writeSpecList(decl.Specs); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("rparen")
	w.writeSpace()
	w.writePos(decl.Rparen)
	w.closeList()
	return nil
}

func (w *Writer) writeFuncDecl(decl *ast.FuncDecl) error {
	w.openList()
	w.writeSymbol("FuncDecl")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	w.writeSymbol("nil") // CommentGroup - simplified for Phase 1
	w.writeSpace()
	w.writeKeyword("recv")
	w.writeSpace()
	if err := w.writeFieldList(decl.Recv); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	if err := w.writeIdent(decl.Name); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("type")
	w.writeSpace()
	if err := w.writeFuncType(decl.Type); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(decl.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// Declaration dispatcher
func (w *Writer) writeDecl(decl ast.Decl) error {
	if decl == nil {
		w.writeSymbol("nil")
		return nil
	}

	switch d := decl.(type) {
	case *ast.GenDecl:
		return w.writeGenDecl(d)
	case *ast.FuncDecl:
		return w.writeFuncDecl(d)
	default:
		return ErrUnknownDeclType(decl)
	}
}

// List writers

func (w *Writer) writeExprList(exprs []ast.Expr) error {
	w.openList()
	for i, expr := range exprs {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeExpr(expr); err != nil {
			return err
		}
	}
	w.closeList()
	return nil
}

func (w *Writer) writeStmtList(stmts []ast.Stmt) error {
	w.openList()
	for i, stmt := range stmts {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeStmt(stmt); err != nil {
			return err
		}
	}
	w.closeList()
	return nil
}

func (w *Writer) writeDeclList(decls []ast.Decl) error {
	w.openList()
	for i, decl := range decls {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeDecl(decl); err != nil {
			return err
		}
	}
	w.closeList()
	return nil
}

func (w *Writer) writeSpecList(specs []ast.Spec) error {
	w.openList()
	for i, spec := range specs {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeSpec(spec); err != nil {
			return err
		}
	}
	w.closeList()
	return nil
}

func (w *Writer) writeIdentList(idents []*ast.Ident) error {
	w.openList()
	for i, ident := range idents {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeIdent(ident); err != nil {
			return err
		}
	}
	w.closeList()
	return nil
}

// FileSet writers

func (w *Writer) writeFileInfo(file *token.File) error {
	w.openList()
	w.writeSymbol("FileInfo")
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	w.writeString(file.Name())
	w.writeSpace()
	w.writeKeyword("base")
	w.writeSpace()
	w.writePos(token.Pos(file.Base()))
	w.writeSpace()
	w.writeKeyword("size")
	w.writeSpace()
	w.writePos(token.Pos(file.Size()))
	w.writeSpace()
	w.writeKeyword("lines")
	w.writeSpace()

	// Write line offsets
	w.openList()
	// Get line count
	lineCount := file.LineCount()
	for i := 1; i <= lineCount; i++ {
		if i > 1 {
			w.writeSpace()
		}
		offset := file.LineStart(i) - token.Pos(file.Base())
		w.writePos(token.Pos(int(offset) + 1))
	}
	w.closeList()

	w.closeList()
	return nil
}

func (w *Writer) writeFileSet() error {
	w.openList()
	w.writeSymbol("FileSet")
	w.writeSpace()
	w.writeKeyword("base")
	w.writeSpace()
	w.writePos(token.Pos(w.fset.Base()))
	w.writeSpace()
	w.writeKeyword("files")
	w.writeSpace()

	w.openList()
	// Collect all files from the FileSet
	// This is a simplified approach - iterate through positions
	files := w.collectFiles()
	for i, file := range files {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeFileInfo(file); err != nil {
			return err
		}
	}
	w.closeList()

	w.closeList()
	return nil
}

func (w *Writer) collectFiles() []*token.File {
	var files []*token.File

	// Iterate through all positions to find files
	// Start from base position and try to find files
	base := w.fset.Base()
	maxPos := base + w.config.MaxFileSetScan

	seen := make(map[*token.File]bool)
	for i := base; i < maxPos; i++ {
		pos := token.Pos(i)
		if file := w.fset.File(pos); file != nil && !seen[file] {
			files = append(files, file)
			seen[file] = true
		}

		// Stop if we haven't found a new file in a while
		if len(files) > 0 && i > base+w.config.FileSearchWindow {
			break
		}
	}

	return files
}

// WriteFile writes a single File node
func (w *Writer) WriteFile(file *ast.File) (string, error) {
	w.buf.Reset()

	w.openList()
	w.writeSymbol("File")
	w.writeSpace()
	w.writeKeyword("package")
	w.writeSpace()
	w.writePos(file.Package)
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	if err := w.writeIdent(file.Name); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("decls")
	w.writeSpace()
	if err := w.writeDeclList(file.Decls); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("scope")
	w.writeSpace()
	w.writeSymbol("nil") // Scope - simplified for Phase 1
	w.writeSpace()
	w.writeKeyword("imports")
	w.writeSpace()
	w.openList()
	w.closeList()
	w.writeSpace()
	w.writeKeyword("unresolved")
	w.writeSpace()
	w.openList()
	w.closeList()
	w.writeSpace()
	w.writeKeyword("comments")
	w.writeSpace()
	w.openList()
	w.closeList()
	w.closeList()

	return w.buf.String(), nil
}

// WriteProgram writes a complete Program with FileSet and Files
func (w *Writer) WriteProgram(files []*ast.File) (string, error) {
	w.buf.Reset()

	w.openList()
	w.writeSymbol("Program")
	w.writeSpace()
	w.writeKeyword("fileset")
	w.writeSpace()
	if err := w.writeFileSet(); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("files")
	w.writeSpace()

	w.openList()
	for i, file := range files {
		if i > 0 {
			w.writeSpace()
		}
		// Write file inline
		if err := w.writeFileNode(file); err != nil {
			return "", err
		}
	}
	w.closeList()

	w.closeList()

	return w.buf.String(), nil
}

func (w *Writer) writeFileNode(file *ast.File) error {
	w.openList()
	w.writeSymbol("File")
	w.writeSpace()
	w.writeKeyword("package")
	w.writeSpace()
	w.writePos(file.Package)
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	if err := w.writeIdent(file.Name); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("decls")
	w.writeSpace()
	if err := w.writeDeclList(file.Decls); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("scope")
	w.writeSpace()
	w.writeSymbol("nil") // Scope - simplified for Phase 1
	w.writeSpace()
	w.writeKeyword("imports")
	w.writeSpace()
	w.openList()
	w.closeList()
	w.writeSpace()
	w.writeKeyword("unresolved")
	w.writeSpace()
	w.openList()
	w.closeList()
	w.writeSpace()
	w.writeKeyword("comments")
	w.writeSpace()
	w.openList()
	w.closeList()
	w.closeList()

	return nil
}

// WriteTo writes the result to an io.Writer
func (w *Writer) WriteTo(wr io.Writer) (int64, error) {
	n, err := wr.Write([]byte(w.buf.String()))
	return int64(n), err
}
