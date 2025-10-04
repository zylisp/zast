// This file contains functions to write lists of various AST node types.
//
// Note that a similar file doesn't exist for the builder package because the
// builder builds lists inline (see the example above where it iterates over
// argsList.Elements and builds each arg directly), whereas the writer needs
// separate list-writing methods for formatting.
package writer

import "go/ast"

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
