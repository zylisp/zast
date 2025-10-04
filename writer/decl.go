package writer

import (
	"go/ast"

	"zylisp/zast/errors"
)

func (w *Writer) writeImportSpec(spec *ast.ImportSpec) error {
	w.openList()
	w.writeSymbol("ImportSpec")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	if err := w.writeCommentGroup(spec.Doc); err != nil {
		return err
	}
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
	if err := w.writeCommentGroup(spec.Comment); err != nil {
		return err
	}
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
	if err := w.writeCommentGroup(spec.Doc); err != nil {
		return err
	}
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
	if err := w.writeCommentGroup(spec.Comment); err != nil {
		return err
	}
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
	if err := w.writeCommentGroup(spec.Doc); err != nil {
		return err
	}
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
	if err := w.writeCommentGroup(spec.Comment); err != nil {
		return err
	}
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
		return errors.ErrUnknownSpecType(spec)
	}
}

func (w *Writer) writeGenDecl(decl *ast.GenDecl) error {
	w.openList()
	w.writeSymbol("GenDecl")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	if err := w.writeCommentGroup(decl.Doc); err != nil {
		return err
	}
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
	if err := w.writeCommentGroup(decl.Doc); err != nil {
		return err
	}
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
		return errors.ErrUnknownDeclType(decl)
	}
}
