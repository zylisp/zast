package writer

import (
	"go/ast"

	"zylisp/zast/errors"
)

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
		return errors.ErrUnknownStmtType(stmt)
	}
}
