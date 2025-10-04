package writer

import (
	"go/ast"

	"zylisp/zast/errors"
)

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

// writeCompositeLit writes a CompositeLit node
func (w *Writer) writeCompositeLit(expr *ast.CompositeLit) error {
	w.openList()
	w.writeSymbol("CompositeLit")
	w.writeSpace()
	w.writeKeyword("type")
	w.writeSpace()
	if err := w.writeExpr(expr.Type); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("lbrace")
	w.writeSpace()
	w.writePos(expr.Lbrace)
	w.writeSpace()
	w.writeKeyword("elts")
	w.writeSpace()
	if err := w.writeExprList(expr.Elts); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("rbrace")
	w.writeSpace()
	w.writePos(expr.Rbrace)
	w.writeSpace()
	w.writeKeyword("incomplete")
	w.writeSpace()
	w.writeBool(expr.Incomplete)
	w.closeList()
	return nil
}

// writeFuncLit writes a FuncLit node
func (w *Writer) writeFuncLit(expr *ast.FuncLit) error {
	w.openList()
	w.writeSymbol("FuncLit")
	w.writeSpace()
	w.writeKeyword("type")
	w.writeSpace()
	if err := w.writeFuncType(expr.Type); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("body")
	w.writeSpace()
	if err := w.writeBlockStmt(expr.Body); err != nil {
		return err
	}
	w.closeList()
	return nil
}

// writeEllipsis writes an Ellipsis node
func (w *Writer) writeEllipsis(expr *ast.Ellipsis) error {
	w.openList()
	w.writeSymbol("Ellipsis")
	w.writeSpace()
	w.writeKeyword("ellipsis")
	w.writeSpace()
	w.writePos(expr.Ellipsis)
	w.writeSpace()
	w.writeKeyword("elt")
	w.writeSpace()
	if err := w.writeExpr(expr.Elt); err != nil {
		return err
	}
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
	case *ast.StructType:
		return w.writeStructType(e)
	case *ast.InterfaceType:
		return w.writeInterfaceType(e)
	case *ast.CompositeLit:
		return w.writeCompositeLit(e)
	case *ast.FuncLit:
		return w.writeFuncLit(e)
	case *ast.Ellipsis:
		return w.writeEllipsis(e)
	default:
		return errors.ErrUnknownExprType(expr)
	}
}
