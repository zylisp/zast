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

func (w *Writer) writeToken(tok token.Token) {
	switch tok {
	case token.IMPORT:
		w.writeSymbol("IMPORT")
	case token.CONST:
		w.writeSymbol("CONST")
	case token.TYPE:
		w.writeSymbol("TYPE")
	case token.VAR:
		w.writeSymbol("VAR")
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

// Spec dispatcher
func (w *Writer) writeSpec(spec ast.Spec) error {
	if spec == nil {
		w.writeSymbol("nil")
		return nil
	}

	switch s := spec.(type) {
	case *ast.ImportSpec:
		return w.writeImportSpec(s)
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
