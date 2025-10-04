package builder

import (
	"go/ast"
	"go/token"

	"zylisp/zast/errors"
	"zylisp/zast/sexp"
)

// buildImportSpec parses an ImportSpec node
func (b *Builder) buildImportSpec(s sexp.SExp) (*ast.ImportSpec, error) {
	list, ok := b.expectList(s, "ImportSpec")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ImportSpec") {
		return nil, errors.ErrExpectedNodeType("ImportSpec", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	pathVal, ok := b.requireKeyword(args, "path", "ImportSpec")
	if !ok {
		return nil, errors.ErrMissingField("path")
	}

	path, err := b.buildBasicLit(pathVal)
	if err != nil {
		return nil, errors.ErrInvalidField("path", err)
	}

	// Optional name
	var name *ast.Ident
	if nameVal, ok := args["name"]; ok {
		name, err = b.buildOptionalIdent(nameVal)
		if err != nil {
			return nil, errors.ErrInvalidField("name", err)
		}
	}

	// Optional endpos
	var endPos token.Pos
	if endposVal, ok := args["endpos"]; ok {
		endPos = b.parsePos(endposVal)
	}

	return &ast.ImportSpec{
		Name:   name,
		Path:   path,
		EndPos: endPos,
	}, nil
}

// buildValueSpec parses a ValueSpec node
func (b *Builder) buildValueSpec(s sexp.SExp) (*ast.ValueSpec, error) {
	list, ok := b.expectList(s, "ValueSpec")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ValueSpec") {
		return nil, errors.ErrExpectedNodeType("ValueSpec", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	namesVal, ok := b.requireKeyword(args, "names", "ValueSpec")
	if !ok {
		return nil, errors.ErrMissingField("names")
	}

	typeVal, ok := b.requireKeyword(args, "type", "ValueSpec")
	if !ok {
		return nil, errors.ErrMissingField("type")
	}

	valuesVal, ok := b.requireKeyword(args, "values", "ValueSpec")
	if !ok {
		return nil, errors.ErrMissingField("values")
	}

	// Build names list
	var names []*ast.Ident
	namesList, ok := b.expectList(namesVal, "ValueSpec names")
	if ok {
		for _, nameSexp := range namesList.Elements {
			name, err := b.buildIdent(nameSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("name", err)
			}
			names = append(names, name)
		}
	}

	// Type is optional
	typ, err := b.buildOptionalExpr(typeVal)
	if err != nil {
		return nil, errors.ErrInvalidField("type", err)
	}

	// Build values list
	var values []ast.Expr
	valuesList, ok := b.expectList(valuesVal, "ValueSpec values")
	if ok {
		for _, valueSexp := range valuesList.Elements {
			value, err := b.buildExpr(valueSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("value", err)
			}
			values = append(values, value)
		}
	}

	return &ast.ValueSpec{
		Names:  names,
		Type:   typ,
		Values: values,
	}, nil
}

// buildTypeSpec parses a TypeSpec node
func (b *Builder) buildTypeSpec(s sexp.SExp) (*ast.TypeSpec, error) {
	list, ok := b.expectList(s, "TypeSpec")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "TypeSpec") {
		return nil, errors.ErrExpectedNodeType("TypeSpec", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "TypeSpec")
	if !ok {
		return nil, errors.ErrMissingField("name")
	}

	typeparamsVal, ok := b.requireKeyword(args, "typeparams", "TypeSpec")
	if !ok {
		return nil, errors.ErrMissingField("typeparams")
	}

	assignVal, ok := b.requireKeyword(args, "assign", "TypeSpec")
	if !ok {
		return nil, errors.ErrMissingField("assign")
	}

	typeVal, ok := b.requireKeyword(args, "type", "TypeSpec")
	if !ok {
		return nil, errors.ErrMissingField("type")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, errors.ErrInvalidField("name", err)
	}

	// Type params is optional (for generics)
	typeparams, err := b.buildOptionalFieldList(typeparamsVal)
	if err != nil {
		return nil, errors.ErrInvalidField("typeparams", err)
	}

	typ, err := b.buildExpr(typeVal)
	if err != nil {
		return nil, errors.ErrInvalidField("type", err)
	}

	return &ast.TypeSpec{
		Name:       name,
		TypeParams: typeparams,
		Assign:     b.parsePos(assignVal),
		Type:       typ,
	}, nil
}

// buildSpec dispatches to appropriate spec builder
func (b *Builder) buildSpec(s sexp.SExp) (ast.Spec, error) {
	list, ok := b.expectList(s, "spec")
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
	case "ImportSpec":
		return b.buildImportSpec(s)
	case "ValueSpec":
		return b.buildValueSpec(s)
	case "TypeSpec":
		return b.buildTypeSpec(s)
	default:
		return nil, errors.ErrUnknownNodeType(sym.Value, "spec")
	}
}

// buildGenDecl parses a GenDecl node
func (b *Builder) buildGenDecl(s sexp.SExp) (*ast.GenDecl, error) {
	list, ok := b.expectList(s, "GenDecl")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "GenDecl") {
		return nil, errors.ErrExpectedNodeType("GenDecl", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	tokVal, ok := b.requireKeyword(args, "tok", "GenDecl")
	if !ok {
		return nil, errors.ErrMissingField("tok")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "GenDecl")
	if !ok {
		return nil, errors.ErrMissingField("tokpos")
	}

	specsVal, ok := b.requireKeyword(args, "specs", "GenDecl")
	if !ok {
		return nil, errors.ErrMissingField("specs")
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, errors.ErrInvalidField("tok", err)
	}

	// Build specs list
	var specs []ast.Spec
	specsList, ok := b.expectList(specsVal, "GenDecl specs")
	if ok {
		for _, specSexp := range specsList.Elements {
			spec, err := b.buildSpec(specSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("spec", err)
			}
			specs = append(specs, spec)
		}
	}

	// Optional lparen/rparen
	var lparen, rparen token.Pos
	if lparenVal, ok := args["lparen"]; ok {
		lparen = b.parsePos(lparenVal)
	}
	if rparenVal, ok := args["rparen"]; ok {
		rparen = b.parsePos(rparenVal)
	}

	return &ast.GenDecl{
		Tok:    tok,
		TokPos: b.parsePos(tokposVal),
		Lparen: lparen,
		Specs:  specs,
		Rparen: rparen,
	}, nil
}

// buildFuncDecl parses a FuncDecl node
func (b *Builder) buildFuncDecl(s sexp.SExp) (*ast.FuncDecl, error) {
	list, ok := b.expectList(s, "FuncDecl")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FuncDecl") {
		return nil, errors.ErrExpectedNodeType("FuncDecl", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "FuncDecl")
	if !ok {
		return nil, errors.ErrMissingField("name")
	}

	typeVal, ok := b.requireKeyword(args, "type", "FuncDecl")
	if !ok {
		return nil, errors.ErrMissingField("type")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, errors.ErrInvalidField("name", err)
	}

	funcType, err := b.buildFuncType(typeVal)
	if err != nil {
		return nil, errors.ErrInvalidField("type", err)
	}

	// Optional recv
	var recv *ast.FieldList
	if recvVal, ok := args["recv"]; ok {
		recv, err = b.buildOptionalFieldList(recvVal)
		if err != nil {
			return nil, errors.ErrInvalidField("recv", err)
		}
	}

	// Optional body
	var body *ast.BlockStmt
	if bodyVal, ok := args["body"]; ok {
		body, err = b.buildOptionalBlockStmt(bodyVal)
		if err != nil {
			return nil, errors.ErrInvalidField("body", err)
		}
	}

	return &ast.FuncDecl{
		Recv: recv,
		Name: name,
		Type: funcType,
		Body: body,
	}, nil
}

// buildDecl dispatches to appropriate declaration builder
func (b *Builder) buildDecl(s sexp.SExp) (ast.Decl, error) {
	list, ok := b.expectList(s, "declaration")
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
	case "GenDecl":
		return b.buildGenDecl(s)
	case "FuncDecl":
		return b.buildFuncDecl(s)
	default:
		return nil, errors.ErrUnknownNodeType(sym.Value, "declaration")
	}
}
