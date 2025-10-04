package writer

import "go/ast"

func (w *Writer) writeField(field *ast.Field) error {
	w.openList()
	w.writeSymbol("Field")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	if err := w.writeCommentGroup(field.Doc); err != nil {
		return err
	}
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
	if err := w.writeCommentGroup(field.Comment); err != nil {
		return err
	}
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

// writeStructType writes a StructType node
func (w *Writer) writeStructType(typ *ast.StructType) error {
	w.openList()
	w.writeSymbol("StructType")
	w.writeSpace()
	w.writeKeyword("struct")
	w.writeSpace()
	w.writePos(typ.Struct)
	w.writeSpace()
	w.writeKeyword("fields")
	w.writeSpace()
	if err := w.writeFieldList(typ.Fields); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("incomplete")
	w.writeSpace()
	w.writeBool(typ.Incomplete)
	w.closeList()
	return nil
}

// writeInterfaceType writes an InterfaceType node
func (w *Writer) writeInterfaceType(typ *ast.InterfaceType) error {
	w.openList()
	w.writeSymbol("InterfaceType")
	w.writeSpace()
	w.writeKeyword("interface")
	w.writeSpace()
	w.writePos(typ.Interface)
	w.writeSpace()
	w.writeKeyword("methods")
	w.writeSpace()
	if err := w.writeFieldList(typ.Methods); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("incomplete")
	w.writeSpace()
	w.writeBool(typ.Incomplete)
	w.closeList()
	return nil
}
