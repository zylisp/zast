package builder

import (
	"go/ast"

	"zylisp/zast/errors"
	"zylisp/zast/sexp"
)

// buildIdent parses an Ident node
func (b *Builder) buildIdent(s sexp.SExp) (*ast.Ident, error) {
	list, ok := b.expectList(s, "Ident")
	if !ok {
		return nil, errors.ErrNotList
	}

	if len(list.Elements) == 0 {
		return nil, errors.ErrEmptyList
	}

	if !b.expectSymbol(list.Elements[0], "Ident") {
		return nil, errors.ErrExpectedNodeType("Ident", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameposVal, ok := b.requireKeyword(args, "namepos", "Ident")
	if !ok {
		return nil, errors.ErrMissingField("namepos")
	}

	nameVal, ok := b.requireKeyword(args, "name", "Ident")
	if !ok {
		return nil, errors.ErrMissingField("name")
	}

	name, err := b.parseString(nameVal)
	if err != nil {
		return nil, errors.ErrInvalidField("name", err)
	}

	// Optional obj field
	var obj *ast.Object
	if objVal, ok := args["obj"]; ok && !b.parseNil(objVal) {
		obj, err = b.buildObject(objVal)
		if err != nil {
			return nil, errors.ErrInvalidField("obj", err)
		}
	}

	ident := &ast.Ident{
		NamePos: b.parsePos(nameposVal),
		Name:    name,
		Obj:     obj,
	}

	return ident, nil
}

// buildOptionalIdent builds Ident or returns nil
func (b *Builder) buildOptionalIdent(s sexp.SExp) (*ast.Ident, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildIdent(s)
}

// buildBasicLit parses a BasicLit node
func (b *Builder) buildBasicLit(s sexp.SExp) (*ast.BasicLit, error) {
	list, ok := b.expectList(s, "BasicLit")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BasicLit") {
		return nil, errors.ErrExpectedNodeType("BasicLit", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	valueposVal, ok := b.requireKeyword(args, "valuepos", "BasicLit")
	if !ok {
		return nil, errors.ErrMissingField("valuepos")
	}

	kindVal, ok := b.requireKeyword(args, "kind", "BasicLit")
	if !ok {
		return nil, errors.ErrMissingField("kind")
	}

	valueVal, ok := b.requireKeyword(args, "value", "BasicLit")
	if !ok {
		return nil, errors.ErrMissingField("value")
	}

	kind, err := b.parseToken(kindVal)
	if err != nil {
		return nil, errors.ErrInvalidField("kind", err)
	}

	value, err := b.parseString(valueVal)
	if err != nil {
		return nil, errors.ErrInvalidField("value", err)
	}

	return &ast.BasicLit{
		ValuePos: b.parsePos(valueposVal),
		Kind:     kind,
		Value:    value,
	}, nil
}

// buildSelectorExpr parses a SelectorExpr node
func (b *Builder) buildSelectorExpr(s sexp.SExp) (*ast.SelectorExpr, error) {
	list, ok := b.expectList(s, "SelectorExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SelectorExpr") {
		return nil, errors.ErrExpectedNodeType("SelectorExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "SelectorExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	selVal, ok := b.requireKeyword(args, "sel", "SelectorExpr")
	if !ok {
		return nil, errors.ErrMissingField("sel")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	sel, err := b.buildIdent(selVal)
	if err != nil {
		return nil, errors.ErrInvalidField("sel", err)
	}

	return &ast.SelectorExpr{
		X:   x,
		Sel: sel,
	}, nil
}

// buildCallExpr parses a CallExpr node
func (b *Builder) buildCallExpr(s sexp.SExp) (*ast.CallExpr, error) {
	list, ok := b.expectList(s, "CallExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "CallExpr") {
		return nil, errors.ErrExpectedNodeType("CallExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	funVal, ok := b.requireKeyword(args, "fun", "CallExpr")
	if !ok {
		return nil, errors.ErrMissingField("fun")
	}

	lparenVal, ok := b.requireKeyword(args, "lparen", "CallExpr")
	if !ok {
		return nil, errors.ErrMissingField("lparen")
	}

	argsVal, ok := b.requireKeyword(args, "args", "CallExpr")
	if !ok {
		return nil, errors.ErrMissingField("args")
	}

	ellipsisVal, ok := b.requireKeyword(args, "ellipsis", "CallExpr")
	if !ok {
		return nil, errors.ErrMissingField("ellipsis")
	}

	rparenVal, ok := b.requireKeyword(args, "rparen", "CallExpr")
	if !ok {
		return nil, errors.ErrMissingField("rparen")
	}

	fun, err := b.buildExpr(funVal)
	if err != nil {
		return nil, errors.ErrInvalidField("fun", err)
	}

	// Build args list
	var callArgs []ast.Expr
	argsList, ok := b.expectList(argsVal, "CallExpr args")
	if ok {
		for _, argSexp := range argsList.Elements {
			arg, err := b.buildExpr(argSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("arg", err)
			}
			callArgs = append(callArgs, arg)
		}
	}

	return &ast.CallExpr{
		Fun:      fun,
		Lparen:   b.parsePos(lparenVal),
		Args:     callArgs,
		Ellipsis: b.parsePos(ellipsisVal),
		Rparen:   b.parsePos(rparenVal),
	}, nil
}

// buildBinaryExpr parses a BinaryExpr node
func (b *Builder) buildBinaryExpr(s sexp.SExp) (*ast.BinaryExpr, error) {
	list, ok := b.expectList(s, "BinaryExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BinaryExpr") {
		return nil, errors.ErrExpectedNodeType("BinaryExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "BinaryExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	opposVal, ok := b.requireKeyword(args, "oppos", "BinaryExpr")
	if !ok {
		return nil, errors.ErrMissingField("oppos")
	}

	opVal, ok := b.requireKeyword(args, "op", "BinaryExpr")
	if !ok {
		return nil, errors.ErrMissingField("op")
	}

	yVal, ok := b.requireKeyword(args, "y", "BinaryExpr")
	if !ok {
		return nil, errors.ErrMissingField("y")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	op, err := b.parseToken(opVal)
	if err != nil {
		return nil, errors.ErrInvalidField("op", err)
	}

	y, err := b.buildExpr(yVal)
	if err != nil {
		return nil, errors.ErrInvalidField("y", err)
	}

	return &ast.BinaryExpr{
		X:     x,
		OpPos: b.parsePos(opposVal),
		Op:    op,
		Y:     y,
	}, nil
}

// buildParenExpr parses a ParenExpr node
func (b *Builder) buildParenExpr(s sexp.SExp) (*ast.ParenExpr, error) {
	list, ok := b.expectList(s, "ParenExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ParenExpr") {
		return nil, errors.ErrExpectedNodeType("ParenExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lparenVal, ok := b.requireKeyword(args, "lparen", "ParenExpr")
	if !ok {
		return nil, errors.ErrMissingField("lparen")
	}

	xVal, ok := b.requireKeyword(args, "x", "ParenExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	rparenVal, ok := b.requireKeyword(args, "rparen", "ParenExpr")
	if !ok {
		return nil, errors.ErrMissingField("rparen")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	return &ast.ParenExpr{
		Lparen: b.parsePos(lparenVal),
		X:      x,
		Rparen: b.parsePos(rparenVal),
	}, nil
}

// buildStarExpr parses a StarExpr node
func (b *Builder) buildStarExpr(s sexp.SExp) (*ast.StarExpr, error) {
	list, ok := b.expectList(s, "StarExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "StarExpr") {
		return nil, errors.ErrExpectedNodeType("StarExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	starVal, ok := b.requireKeyword(args, "star", "StarExpr")
	if !ok {
		return nil, errors.ErrMissingField("star")
	}

	xVal, ok := b.requireKeyword(args, "x", "StarExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	return &ast.StarExpr{
		Star: b.parsePos(starVal),
		X:    x,
	}, nil
}

// buildIndexExpr parses an IndexExpr node
func (b *Builder) buildIndexExpr(s sexp.SExp) (*ast.IndexExpr, error) {
	list, ok := b.expectList(s, "IndexExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "IndexExpr") {
		return nil, errors.ErrExpectedNodeType("IndexExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "IndexExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	lbrackVal, ok := b.requireKeyword(args, "lbrack", "IndexExpr")
	if !ok {
		return nil, errors.ErrMissingField("lbrack")
	}

	indexVal, ok := b.requireKeyword(args, "index", "IndexExpr")
	if !ok {
		return nil, errors.ErrMissingField("index")
	}

	rbrackVal, ok := b.requireKeyword(args, "rbrack", "IndexExpr")
	if !ok {
		return nil, errors.ErrMissingField("rbrack")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	index, err := b.buildExpr(indexVal)
	if err != nil {
		return nil, errors.ErrInvalidField("index", err)
	}

	return &ast.IndexExpr{
		X:      x,
		Lbrack: b.parsePos(lbrackVal),
		Index:  index,
		Rbrack: b.parsePos(rbrackVal),
	}, nil
}

// buildIndexListExpr parses an IndexListExpr node (Go 1.18+ generics)
func (b *Builder) buildIndexListExpr(s sexp.SExp) (*ast.IndexListExpr, error) {
	list, ok := b.expectList(s, "IndexListExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "IndexListExpr") {
		return nil, errors.ErrExpectedNodeType("IndexListExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "IndexListExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	lbrackVal, ok := b.requireKeyword(args, "lbrack", "IndexListExpr")
	if !ok {
		return nil, errors.ErrMissingField("lbrack")
	}

	indicesVal, ok := b.requireKeyword(args, "indices", "IndexListExpr")
	if !ok {
		return nil, errors.ErrMissingField("indices")
	}

	rbrackVal, ok := b.requireKeyword(args, "rbrack", "IndexListExpr")
	if !ok {
		return nil, errors.ErrMissingField("rbrack")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	// Build indices list
	var indices []ast.Expr
	indicesList, ok := b.expectList(indicesVal, "IndexListExpr indices")
	if ok {
		for _, indexSexp := range indicesList.Elements {
			index, err := b.buildExpr(indexSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("index", err)
			}
			indices = append(indices, index)
		}
	}

	return &ast.IndexListExpr{
		X:       x,
		Lbrack:  b.parsePos(lbrackVal),
		Indices: indices,
		Rbrack:  b.parsePos(rbrackVal),
	}, nil
}

// buildBadExpr parses a BadExpr node (for syntax errors)
func (b *Builder) buildBadExpr(s sexp.SExp) (*ast.BadExpr, error) {
	list, ok := b.expectList(s, "BadExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BadExpr") {
		return nil, errors.ErrExpectedNodeType("BadExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	fromVal, ok := b.requireKeyword(args, "from", "BadExpr")
	if !ok {
		return nil, errors.ErrMissingField("from")
	}

	toVal, ok := b.requireKeyword(args, "to", "BadExpr")
	if !ok {
		return nil, errors.ErrMissingField("to")
	}

	return &ast.BadExpr{
		From: b.parsePos(fromVal),
		To:   b.parsePos(toVal),
	}, nil
}

// buildOptionalExpr builds an Expr or returns nil
func (b *Builder) buildOptionalExpr(s sexp.SExp) (ast.Expr, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildExpr(s)
}

// buildSliceExpr parses a SliceExpr node
func (b *Builder) buildSliceExpr(s sexp.SExp) (*ast.SliceExpr, error) {
	list, ok := b.expectList(s, "SliceExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SliceExpr") {
		return nil, errors.ErrExpectedNodeType("SliceExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "SliceExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	lbrackVal, ok := b.requireKeyword(args, "lbrack", "SliceExpr")
	if !ok {
		return nil, errors.ErrMissingField("lbrack")
	}

	lowVal, ok := b.requireKeyword(args, "low", "SliceExpr")
	if !ok {
		return nil, errors.ErrMissingField("low")
	}

	highVal, ok := b.requireKeyword(args, "high", "SliceExpr")
	if !ok {
		return nil, errors.ErrMissingField("high")
	}

	maxVal, ok := b.requireKeyword(args, "max", "SliceExpr")
	if !ok {
		return nil, errors.ErrMissingField("max")
	}

	slice3Val, ok := b.requireKeyword(args, "slice3", "SliceExpr")
	if !ok {
		return nil, errors.ErrMissingField("slice3")
	}

	rbrackVal, ok := b.requireKeyword(args, "rbrack", "SliceExpr")
	if !ok {
		return nil, errors.ErrMissingField("rbrack")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	low, err := b.buildOptionalExpr(lowVal)
	if err != nil {
		return nil, errors.ErrInvalidField("low", err)
	}

	high, err := b.buildOptionalExpr(highVal)
	if err != nil {
		return nil, errors.ErrInvalidField("high", err)
	}

	max, err := b.buildOptionalExpr(maxVal)
	if err != nil {
		return nil, errors.ErrInvalidField("max", err)
	}

	slice3, err := b.parseBool(slice3Val)
	if err != nil {
		return nil, errors.ErrInvalidField("slice3", err)
	}

	return &ast.SliceExpr{
		X:      x,
		Lbrack: b.parsePos(lbrackVal),
		Low:    low,
		High:   high,
		Max:    max,
		Slice3: slice3,
		Rbrack: b.parsePos(rbrackVal),
	}, nil
}

// buildKeyValueExpr parses a KeyValueExpr node
func (b *Builder) buildKeyValueExpr(s sexp.SExp) (*ast.KeyValueExpr, error) {
	list, ok := b.expectList(s, "KeyValueExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "KeyValueExpr") {
		return nil, errors.ErrExpectedNodeType("KeyValueExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	keyVal, ok := b.requireKeyword(args, "key", "KeyValueExpr")
	if !ok {
		return nil, errors.ErrMissingField("key")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "KeyValueExpr")
	if !ok {
		return nil, errors.ErrMissingField("colon")
	}

	valueVal, ok := b.requireKeyword(args, "value", "KeyValueExpr")
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

	return &ast.KeyValueExpr{
		Key:   key,
		Colon: b.parsePos(colonVal),
		Value: value,
	}, nil
}

// buildUnaryExpr parses a UnaryExpr node
func (b *Builder) buildUnaryExpr(s sexp.SExp) (*ast.UnaryExpr, error) {
	list, ok := b.expectList(s, "UnaryExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "UnaryExpr") {
		return nil, errors.ErrExpectedNodeType("UnaryExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	opposVal, ok := b.requireKeyword(args, "oppos", "UnaryExpr")
	if !ok {
		return nil, errors.ErrMissingField("oppos")
	}

	opVal, ok := b.requireKeyword(args, "op", "UnaryExpr")
	if !ok {
		return nil, errors.ErrMissingField("op")
	}

	xVal, ok := b.requireKeyword(args, "x", "UnaryExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	op, err := b.parseToken(opVal)
	if err != nil {
		return nil, errors.ErrInvalidField("op", err)
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	return &ast.UnaryExpr{
		OpPos: b.parsePos(opposVal),
		Op:    op,
		X:     x,
	}, nil
}

// buildTypeAssertExpr parses a TypeAssertExpr node
func (b *Builder) buildTypeAssertExpr(s sexp.SExp) (*ast.TypeAssertExpr, error) {
	list, ok := b.expectList(s, "TypeAssertExpr")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "TypeAssertExpr") {
		return nil, errors.ErrExpectedNodeType("TypeAssertExpr", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "TypeAssertExpr")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	lparenVal, ok := b.requireKeyword(args, "lparen", "TypeAssertExpr")
	if !ok {
		return nil, errors.ErrMissingField("lparen")
	}

	typeVal, ok := b.requireKeyword(args, "type", "TypeAssertExpr")
	if !ok {
		return nil, errors.ErrMissingField("type")
	}

	rparenVal, ok := b.requireKeyword(args, "rparen", "TypeAssertExpr")
	if !ok {
		return nil, errors.ErrMissingField("rparen")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	// Type can be nil for x.(type) in type switches
	typ, err := b.buildOptionalExpr(typeVal)
	if err != nil {
		return nil, errors.ErrInvalidField("type", err)
	}

	return &ast.TypeAssertExpr{
		X:      x,
		Lparen: b.parsePos(lparenVal),
		Type:   typ,
		Rparen: b.parsePos(rparenVal),
	}, nil
}

// buildCompositeLit parses a CompositeLit node
func (b *Builder) buildCompositeLit(s sexp.SExp) (*ast.CompositeLit, error) {
	list, ok := b.expectList(s, "CompositeLit")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "CompositeLit") {
		return nil, errors.ErrExpectedNodeType("CompositeLit", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lbraceVal, ok := b.requireKeyword(args, "lbrace", "CompositeLit")
	if !ok {
		return nil, errors.ErrMissingField("lbrace")
	}

	eltsVal, ok := b.requireKeyword(args, "elts", "CompositeLit")
	if !ok {
		return nil, errors.ErrMissingField("elts")
	}

	rbraceVal, ok := b.requireKeyword(args, "rbrace", "CompositeLit")
	if !ok {
		return nil, errors.ErrMissingField("rbrace")
	}

	incompleteVal, ok := b.requireKeyword(args, "incomplete", "CompositeLit")
	if !ok {
		return nil, errors.ErrMissingField("incomplete")
	}

	// Optional type
	var typ ast.Expr
	var err error
	if typeVal, ok := args["type"]; ok && !b.parseNil(typeVal) {
		typ, err = b.buildExpr(typeVal)
		if err != nil {
			return nil, errors.ErrInvalidField("type", err)
		}
	}

	// Build elements list
	var elts []ast.Expr
	eltsList, ok := b.expectList(eltsVal, "CompositeLit elts")
	if ok {
		for _, eltSexp := range eltsList.Elements {
			elt, err := b.buildExpr(eltSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("element", err)
			}
			elts = append(elts, elt)
		}
	}

	incomplete, err := b.parseBool(incompleteVal)
	if err != nil {
		return nil, errors.ErrInvalidField("incomplete", err)
	}

	return &ast.CompositeLit{
		Type:       typ,
		Lbrace:     b.parsePos(lbraceVal),
		Elts:       elts,
		Rbrace:     b.parsePos(rbraceVal),
		Incomplete: incomplete,
	}, nil
}

// buildFuncLit parses a FuncLit node
func (b *Builder) buildFuncLit(s sexp.SExp) (*ast.FuncLit, error) {
	list, ok := b.expectList(s, "FuncLit")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FuncLit") {
		return nil, errors.ErrExpectedNodeType("FuncLit", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	typeVal, ok := b.requireKeyword(args, "type", "FuncLit")
	if !ok {
		return nil, errors.ErrMissingField("type")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "FuncLit")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	funcType, err := b.buildFuncType(typeVal)
	if err != nil {
		return nil, errors.ErrInvalidField("type", err)
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("body", err)
	}

	return &ast.FuncLit{
		Type: funcType,
		Body: body,
	}, nil
}

// buildEllipsis parses an Ellipsis node
func (b *Builder) buildEllipsis(s sexp.SExp) (*ast.Ellipsis, error) {
	list, ok := b.expectList(s, "Ellipsis")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "Ellipsis") {
		return nil, errors.ErrExpectedNodeType("Ellipsis", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	ellipsisVal, ok := b.requireKeyword(args, "ellipsis", "Ellipsis")
	if !ok {
		return nil, errors.ErrMissingField("ellipsis")
	}

	// Optional elt
	var elt ast.Expr
	var err error
	if eltVal, ok := args["elt"]; ok && !b.parseNil(eltVal) {
		elt, err = b.buildExpr(eltVal)
		if err != nil {
			return nil, errors.ErrInvalidField("elt", err)
		}
	}

	return &ast.Ellipsis{
		Ellipsis: b.parsePos(ellipsisVal),
		Elt:      elt,
	}, nil
}

// buildExpr dispatches to appropriate expression builder
func (b *Builder) buildExpr(s sexp.SExp) (ast.Expr, error) {
	if err := b.enterDepth(); err != nil {
		return nil, err
	}
	defer b.exitDepth()

	list, ok := b.expectList(s, "expression")
	if !ok {
		return nil, errors.ErrNotList
	}

	if len(list.Elements) == 0 {
		return nil, errors.ErrEmptyList
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, errors.ErrExpectedSymbol
	}

	switch sym.Value {
	case "Ident":
		return b.buildIdent(s)
	case "BasicLit":
		return b.buildBasicLit(s)
	case "CallExpr":
		return b.buildCallExpr(s)
	case "SelectorExpr":
		return b.buildSelectorExpr(s)
	case "UnaryExpr":
		return b.buildUnaryExpr(s)
	case "BinaryExpr":
		return b.buildBinaryExpr(s)
	case "ParenExpr":
		return b.buildParenExpr(s)
	case "StarExpr":
		return b.buildStarExpr(s)
	case "IndexExpr":
		return b.buildIndexExpr(s)
	case "SliceExpr":
		return b.buildSliceExpr(s)
	case "KeyValueExpr":
		return b.buildKeyValueExpr(s)
	case "ArrayType":
		return b.buildArrayType(s)
	case "MapType":
		return b.buildMapType(s)
	case "ChanType":
		return b.buildChanType(s)
	case "TypeAssertExpr":
		return b.buildTypeAssertExpr(s)
	case "StructType":
		return b.buildStructType(s)
	case "InterfaceType":
		return b.buildInterfaceType(s)
	case "CompositeLit":
		return b.buildCompositeLit(s)
	case "FuncLit":
		return b.buildFuncLit(s)
	case "Ellipsis":
		return b.buildEllipsis(s)
	case "IndexListExpr":
		return b.buildIndexListExpr(s)
	case "BadExpr":
		return b.buildBadExpr(s)
	default:
		return nil, errors.ErrUnknownNodeType(sym.Value, "expression")
	}
}
