package builder

import (
	"go/ast"

	"zylisp/zast/errors"
	"zylisp/zast/sexp"
)

// buildField parses a Field node
func (b *Builder) buildField(s sexp.SExp) (*ast.Field, error) {
	list, ok := b.expectList(s, "Field")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "Field") {
		return nil, errors.ErrExpectedNodeType("Field", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	typeVal, ok := b.requireKeyword(args, "type", "Field")
	if !ok {
		return nil, errors.ErrMissingField("type")
	}

	fieldType, err := b.buildExpr(typeVal)
	if err != nil {
		return nil, errors.ErrInvalidField("type", err)
	}

	// Build names list (optional)
	var names []*ast.Ident
	if namesVal, ok := args["names"]; ok {
		namesList, ok := b.expectList(namesVal, "Field names")
		if ok {
			for _, nameSexp := range namesList.Elements {
				ident, err := b.buildIdent(nameSexp)
				if err != nil {
					return nil, errors.ErrInvalidField("name", err)
				}
				names = append(names, ident)
			}
		}
	}

	return &ast.Field{
		Names: names,
		Type:  fieldType,
	}, nil
}

// buildFieldList parses a FieldList node
func (b *Builder) buildFieldList(s sexp.SExp) (*ast.FieldList, error) {
	list, ok := b.expectList(s, "FieldList")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FieldList") {
		return nil, errors.ErrExpectedNodeType("FieldList", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	openingVal, ok := b.requireKeyword(args, "opening", "FieldList")
	if !ok {
		return nil, errors.ErrMissingField("opening")
	}

	listVal, ok := b.requireKeyword(args, "list", "FieldList")
	if !ok {
		return nil, errors.ErrMissingField("list")
	}

	closingVal, ok := b.requireKeyword(args, "closing", "FieldList")
	if !ok {
		return nil, errors.ErrMissingField("closing")
	}

	// Build field list
	var fields []*ast.Field
	fieldsList, ok := b.expectList(listVal, "FieldList list")
	if ok {
		for _, fieldSexp := range fieldsList.Elements {
			field, err := b.buildField(fieldSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("field", err)
			}
			fields = append(fields, field)
		}
	}

	return &ast.FieldList{
		Opening: b.parsePos(openingVal),
		List:    fields,
		Closing: b.parsePos(closingVal),
	}, nil
}

// buildOptionalFieldList builds FieldList or returns nil
func (b *Builder) buildOptionalFieldList(s sexp.SExp) (*ast.FieldList, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildFieldList(s)
}

// buildFuncType parses a FuncType node
func (b *Builder) buildFuncType(s sexp.SExp) (*ast.FuncType, error) {
	list, ok := b.expectList(s, "FuncType")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FuncType") {
		return nil, errors.ErrExpectedNodeType("FuncType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	funcVal, ok := b.requireKeyword(args, "func", "FuncType")
	if !ok {
		return nil, errors.ErrMissingField("func")
	}

	paramsVal, ok := b.requireKeyword(args, "params", "FuncType")
	if !ok {
		return nil, errors.ErrMissingField("params")
	}

	resultsVal, _ := args["results"]

	params, err := b.buildFieldList(paramsVal)
	if err != nil {
		return nil, errors.ErrInvalidField("params", err)
	}

	results, err := b.buildOptionalFieldList(resultsVal)
	if err != nil {
		return nil, errors.ErrInvalidField("results", err)
	}

	return &ast.FuncType{
		Func:    b.parsePos(funcVal),
		Params:  params,
		Results: results,
	}, nil
}

// buildArrayType parses an ArrayType node
func (b *Builder) buildArrayType(s sexp.SExp) (*ast.ArrayType, error) {
	list, ok := b.expectList(s, "ArrayType")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ArrayType") {
		return nil, errors.ErrExpectedNodeType("ArrayType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lbrackVal, ok := b.requireKeyword(args, "lbrack", "ArrayType")
	if !ok {
		return nil, errors.ErrMissingField("lbrack")
	}

	lenVal, ok := b.requireKeyword(args, "len", "ArrayType")
	if !ok {
		return nil, errors.ErrMissingField("len")
	}

	eltVal, ok := b.requireKeyword(args, "elt", "ArrayType")
	if !ok {
		return nil, errors.ErrMissingField("elt")
	}

	len, err := b.buildOptionalExpr(lenVal)
	if err != nil {
		return nil, errors.ErrInvalidField("len", err)
	}

	elt, err := b.buildExpr(eltVal)
	if err != nil {
		return nil, errors.ErrInvalidField("elt", err)
	}

	return &ast.ArrayType{
		Lbrack: b.parsePos(lbrackVal),
		Len:    len,
		Elt:    elt,
	}, nil
}

// buildMapType parses a MapType node
func (b *Builder) buildMapType(s sexp.SExp) (*ast.MapType, error) {
	list, ok := b.expectList(s, "MapType")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "MapType") {
		return nil, errors.ErrExpectedNodeType("MapType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	mapVal, ok := b.requireKeyword(args, "map", "MapType")
	if !ok {
		return nil, errors.ErrMissingField("map")
	}

	keyVal, ok := b.requireKeyword(args, "key", "MapType")
	if !ok {
		return nil, errors.ErrMissingField("key")
	}

	valueVal, ok := b.requireKeyword(args, "value", "MapType")
	if !ok {
		return nil, errors.ErrMissingField("value")
	}

	key, err := b.buildExpr(keyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("key", err)
	}

	value, err := b.buildExpr(valueVal)
	if err != nil {
		return nil, errors.ErrInvalidField("value", err)
	}

	return &ast.MapType{
		Map:   b.parsePos(mapVal),
		Key:   key,
		Value: value,
	}, nil
}

// buildChanType parses a ChanType node
func (b *Builder) buildChanType(s sexp.SExp) (*ast.ChanType, error) {
	list, ok := b.expectList(s, "ChanType")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ChanType") {
		return nil, errors.ErrExpectedNodeType("ChanType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	beginVal, ok := b.requireKeyword(args, "begin", "ChanType")
	if !ok {
		return nil, errors.ErrMissingField("begin")
	}

	arrowVal, ok := b.requireKeyword(args, "arrow", "ChanType")
	if !ok {
		return nil, errors.ErrMissingField("arrow")
	}

	dirVal, ok := b.requireKeyword(args, "dir", "ChanType")
	if !ok {
		return nil, errors.ErrMissingField("dir")
	}

	valueVal, ok := b.requireKeyword(args, "value", "ChanType")
	if !ok {
		return nil, errors.ErrMissingField("value")
	}

	dir, err := b.parseChanDir(dirVal)
	if err != nil {
		return nil, errors.ErrInvalidField("dir", err)
	}

	value, err := b.buildExpr(valueVal)
	if err != nil {
		return nil, errors.ErrInvalidField("value", err)
	}

	return &ast.ChanType{
		Begin: b.parsePos(beginVal),
		Arrow: b.parsePos(arrowVal),
		Dir:   dir,
		Value: value,
	}, nil
}

// buildStructType parses a StructType node
func (b *Builder) buildStructType(s sexp.SExp) (*ast.StructType, error) {
	list, ok := b.expectList(s, "StructType")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "StructType") {
		return nil, errors.ErrExpectedNodeType("StructType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	structVal, ok := b.requireKeyword(args, "struct", "StructType")
	if !ok {
		return nil, errors.ErrMissingField("struct")
	}

	fieldsVal, ok := b.requireKeyword(args, "fields", "StructType")
	if !ok {
		return nil, errors.ErrMissingField("fields")
	}

	incompleteVal, ok := b.requireKeyword(args, "incomplete", "StructType")
	if !ok {
		return nil, errors.ErrMissingField("incomplete")
	}

	fields, err := b.buildFieldList(fieldsVal)
	if err != nil {
		return nil, errors.ErrInvalidField("fields", err)
	}

	incomplete, err := b.parseBool(incompleteVal)
	if err != nil {
		return nil, errors.ErrInvalidField("incomplete", err)
	}

	return &ast.StructType{
		Struct:     b.parsePos(structVal),
		Fields:     fields,
		Incomplete: incomplete,
	}, nil
}

// buildInterfaceType parses an InterfaceType node
func (b *Builder) buildInterfaceType(s sexp.SExp) (*ast.InterfaceType, error) {
	list, ok := b.expectList(s, "InterfaceType")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "InterfaceType") {
		return nil, errors.ErrExpectedNodeType("InterfaceType", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	interfaceVal, ok := b.requireKeyword(args, "interface", "InterfaceType")
	if !ok {
		return nil, errors.ErrMissingField("interface")
	}

	methodsVal, ok := b.requireKeyword(args, "methods", "InterfaceType")
	if !ok {
		return nil, errors.ErrMissingField("methods")
	}

	incompleteVal, ok := b.requireKeyword(args, "incomplete", "InterfaceType")
	if !ok {
		return nil, errors.ErrMissingField("incomplete")
	}

	methods, err := b.buildFieldList(methodsVal)
	if err != nil {
		return nil, errors.ErrInvalidField("methods", err)
	}

	incomplete, err := b.parseBool(incompleteVal)
	if err != nil {
		return nil, errors.ErrInvalidField("incomplete", err)
	}

	return &ast.InterfaceType{
		Interface:  b.parsePos(interfaceVal),
		Methods:    methods,
		Incomplete: incomplete,
	}, nil
}
