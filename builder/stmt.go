package builder

import (
	"go/ast"

	"zylisp/zast/errors"
	"zylisp/zast/sexp"
)

// buildExprStmt parses an ExprStmt node
func (b *Builder) buildExprStmt(s sexp.SExp) (*ast.ExprStmt, error) {
	list, ok := b.expectList(s, "ExprStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ExprStmt") {
		return nil, errors.ErrExpectedNodeType("ExprStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "ExprStmt")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	return &ast.ExprStmt{X: x}, nil
}

// buildBlockStmt parses a BlockStmt node
func (b *Builder) buildBlockStmt(s sexp.SExp) (*ast.BlockStmt, error) {
	list, ok := b.expectList(s, "BlockStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BlockStmt") {
		return nil, errors.ErrExpectedNodeType("BlockStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lbraceVal, ok := b.requireKeyword(args, "lbrace", "BlockStmt")
	if !ok {
		return nil, errors.ErrMissingField("lbrace")
	}

	listVal, ok := b.requireKeyword(args, "list", "BlockStmt")
	if !ok {
		return nil, errors.ErrMissingField("list")
	}

	rbraceVal, ok := b.requireKeyword(args, "rbrace", "BlockStmt")
	if !ok {
		return nil, errors.ErrMissingField("rbrace")
	}

	// Build statement list
	var stmts []ast.Stmt
	stmtsList, ok := b.expectList(listVal, "BlockStmt list")
	if ok {
		for _, stmtSexp := range stmtsList.Elements {
			stmt, err := b.buildStmt(stmtSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("statement", err)
			}
			stmts = append(stmts, stmt)
		}
	}

	return &ast.BlockStmt{
		Lbrace: b.parsePos(lbraceVal),
		List:   stmts,
		Rbrace: b.parsePos(rbraceVal),
	}, nil
}

// buildOptionalBlockStmt builds BlockStmt or returns nil
func (b *Builder) buildOptionalBlockStmt(s sexp.SExp) (*ast.BlockStmt, error) {
	if b.parseNil(s) {
		return nil, nil
	}
	return b.buildBlockStmt(s)
}

// buildReturnStmt parses a ReturnStmt node
func (b *Builder) buildReturnStmt(s sexp.SExp) (*ast.ReturnStmt, error) {
	list, ok := b.expectList(s, "ReturnStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ReturnStmt") {
		return nil, errors.ErrExpectedNodeType("ReturnStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	returnVal, ok := b.requireKeyword(args, "return", "ReturnStmt")
	if !ok {
		return nil, errors.ErrMissingField("return")
	}

	resultsVal, ok := b.requireKeyword(args, "results", "ReturnStmt")
	if !ok {
		return nil, errors.ErrMissingField("results")
	}

	// Build results list
	var results []ast.Expr
	resultsList, ok := b.expectList(resultsVal, "ReturnStmt results")
	if ok {
		for _, resultSexp := range resultsList.Elements {
			result, err := b.buildExpr(resultSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("result", err)
			}
			results = append(results, result)
		}
	}

	return &ast.ReturnStmt{
		Return:  b.parsePos(returnVal),
		Results: results,
	}, nil
}

// buildAssignStmt parses an AssignStmt node
func (b *Builder) buildAssignStmt(s sexp.SExp) (*ast.AssignStmt, error) {
	list, ok := b.expectList(s, "AssignStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "AssignStmt") {
		return nil, errors.ErrExpectedNodeType("AssignStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	lhsVal, ok := b.requireKeyword(args, "lhs", "AssignStmt")
	if !ok {
		return nil, errors.ErrMissingField("lhs")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "AssignStmt")
	if !ok {
		return nil, errors.ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "AssignStmt")
	if !ok {
		return nil, errors.ErrMissingField("tok")
	}

	rhsVal, ok := b.requireKeyword(args, "rhs", "AssignStmt")
	if !ok {
		return nil, errors.ErrMissingField("rhs")
	}

	// Build lhs list
	var lhs []ast.Expr
	lhsList, ok := b.expectList(lhsVal, "AssignStmt lhs")
	if ok {
		for _, lhsSexp := range lhsList.Elements {
			lhsExpr, err := b.buildExpr(lhsSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("lhs", err)
			}
			lhs = append(lhs, lhsExpr)
		}
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, errors.ErrInvalidField("tok", err)
	}

	// Build rhs list
	var rhs []ast.Expr
	rhsList, ok := b.expectList(rhsVal, "AssignStmt rhs")
	if ok {
		for _, rhsSexp := range rhsList.Elements {
			rhsExpr, err := b.buildExpr(rhsSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("rhs", err)
			}
			rhs = append(rhs, rhsExpr)
		}
	}

	return &ast.AssignStmt{
		Lhs:    lhs,
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
		Rhs:    rhs,
	}, nil
}

// buildIncDecStmt parses an IncDecStmt node
func (b *Builder) buildIncDecStmt(s sexp.SExp) (*ast.IncDecStmt, error) {
	list, ok := b.expectList(s, "IncDecStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "IncDecStmt") {
		return nil, errors.ErrExpectedNodeType("IncDecStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "IncDecStmt")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "IncDecStmt")
	if !ok {
		return nil, errors.ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "IncDecStmt")
	if !ok {
		return nil, errors.ErrMissingField("tok")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, errors.ErrInvalidField("tok", err)
	}

	return &ast.IncDecStmt{
		X:      x,
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
	}, nil
}

// buildBranchStmt parses a BranchStmt node
func (b *Builder) buildBranchStmt(s sexp.SExp) (*ast.BranchStmt, error) {
	list, ok := b.expectList(s, "BranchStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "BranchStmt") {
		return nil, errors.ErrExpectedNodeType("BranchStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	tokposVal, ok := b.requireKeyword(args, "tokpos", "BranchStmt")
	if !ok {
		return nil, errors.ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "BranchStmt")
	if !ok {
		return nil, errors.ErrMissingField("tok")
	}

	labelVal, ok := b.requireKeyword(args, "label", "BranchStmt")
	if !ok {
		return nil, errors.ErrMissingField("label")
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, errors.ErrInvalidField("tok", err)
	}

	label, err := b.buildOptionalIdent(labelVal)
	if err != nil {
		return nil, errors.ErrInvalidField("label", err)
	}

	return &ast.BranchStmt{
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
		Label:  label,
	}, nil
}

// buildDeferStmt parses a DeferStmt node
func (b *Builder) buildDeferStmt(s sexp.SExp) (*ast.DeferStmt, error) {
	list, ok := b.expectList(s, "DeferStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "DeferStmt") {
		return nil, errors.ErrExpectedNodeType("DeferStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	deferVal, ok := b.requireKeyword(args, "defer", "DeferStmt")
	if !ok {
		return nil, errors.ErrMissingField("defer")
	}

	callVal, ok := b.requireKeyword(args, "call", "DeferStmt")
	if !ok {
		return nil, errors.ErrMissingField("call")
	}

	call, err := b.buildCallExpr(callVal)
	if err != nil {
		return nil, errors.ErrInvalidField("call", err)
	}

	return &ast.DeferStmt{
		Defer: b.parsePos(deferVal),
		Call:  call,
	}, nil
}

// buildGoStmt parses a GoStmt node
func (b *Builder) buildGoStmt(s sexp.SExp) (*ast.GoStmt, error) {
	list, ok := b.expectList(s, "GoStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "GoStmt") {
		return nil, errors.ErrExpectedNodeType("GoStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	goVal, ok := b.requireKeyword(args, "go", "GoStmt")
	if !ok {
		return nil, errors.ErrMissingField("go")
	}

	callVal, ok := b.requireKeyword(args, "call", "GoStmt")
	if !ok {
		return nil, errors.ErrMissingField("call")
	}

	call, err := b.buildCallExpr(callVal)
	if err != nil {
		return nil, errors.ErrInvalidField("call", err)
	}

	return &ast.GoStmt{
		Go:   b.parsePos(goVal),
		Call: call,
	}, nil
}

// buildSendStmt parses a SendStmt node
func (b *Builder) buildSendStmt(s sexp.SExp) (*ast.SendStmt, error) {
	list, ok := b.expectList(s, "SendStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SendStmt") {
		return nil, errors.ErrExpectedNodeType("SendStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	chanVal, ok := b.requireKeyword(args, "chan", "SendStmt")
	if !ok {
		return nil, errors.ErrMissingField("chan")
	}

	arrowVal, ok := b.requireKeyword(args, "arrow", "SendStmt")
	if !ok {
		return nil, errors.ErrMissingField("arrow")
	}

	valueVal, ok := b.requireKeyword(args, "value", "SendStmt")
	if !ok {
		return nil, errors.ErrMissingField("value")
	}

	chanExpr, err := b.buildExpr(chanVal)
	if err != nil {
		return nil, errors.ErrInvalidField("chan", err)
	}

	value, err := b.buildExpr(valueVal)
	if err != nil {
		return nil, errors.ErrInvalidField("value", err)
	}

	return &ast.SendStmt{
		Chan:  chanExpr,
		Arrow: b.parsePos(arrowVal),
		Value: value,
	}, nil
}

// buildEmptyStmt parses an EmptyStmt node
func (b *Builder) buildEmptyStmt(s sexp.SExp) (*ast.EmptyStmt, error) {
	list, ok := b.expectList(s, "EmptyStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "EmptyStmt") {
		return nil, errors.ErrExpectedNodeType("EmptyStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	semicolonVal, ok := b.requireKeyword(args, "semicolon", "EmptyStmt")
	if !ok {
		return nil, errors.ErrMissingField("semicolon")
	}

	implicitVal, ok := b.requireKeyword(args, "implicit", "EmptyStmt")
	if !ok {
		return nil, errors.ErrMissingField("implicit")
	}

	implicit, err := b.parseBool(implicitVal)
	if err != nil {
		return nil, errors.ErrInvalidField("implicit", err)
	}

	return &ast.EmptyStmt{
		Semicolon: b.parsePos(semicolonVal),
		Implicit:  implicit,
	}, nil
}

// buildLabeledStmt parses a LabeledStmt node
func (b *Builder) buildLabeledStmt(s sexp.SExp) (*ast.LabeledStmt, error) {
	list, ok := b.expectList(s, "LabeledStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "LabeledStmt") {
		return nil, errors.ErrExpectedNodeType("LabeledStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	labelVal, ok := b.requireKeyword(args, "label", "LabeledStmt")
	if !ok {
		return nil, errors.ErrMissingField("label")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "LabeledStmt")
	if !ok {
		return nil, errors.ErrMissingField("colon")
	}

	stmtVal, ok := b.requireKeyword(args, "stmt", "LabeledStmt")
	if !ok {
		return nil, errors.ErrMissingField("stmt")
	}

	label, err := b.buildIdent(labelVal)
	if err != nil {
		return nil, errors.ErrInvalidField("label", err)
	}

	stmt, err := b.buildStmt(stmtVal)
	if err != nil {
		return nil, errors.ErrInvalidField("stmt", err)
	}

	return &ast.LabeledStmt{
		Label: label,
		Colon: b.parsePos(colonVal),
		Stmt:  stmt,
	}, nil
}

// buildIfStmt parses an IfStmt node
func (b *Builder) buildIfStmt(s sexp.SExp) (*ast.IfStmt, error) {
	list, ok := b.expectList(s, "IfStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "IfStmt") {
		return nil, errors.ErrExpectedNodeType("IfStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	ifVal, ok := b.requireKeyword(args, "if", "IfStmt")
	if !ok {
		return nil, errors.ErrMissingField("if")
	}

	condVal, ok := b.requireKeyword(args, "cond", "IfStmt")
	if !ok {
		return nil, errors.ErrMissingField("cond")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "IfStmt")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	cond, err := b.buildExpr(condVal)
	if err != nil {
		return nil, errors.ErrInvalidField("cond", err)
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, errors.ErrInvalidField("init", err)
		}
	}

	// Optional else
	var els ast.Stmt
	if elseVal, ok := args["else"]; ok && !b.parseNil(elseVal) {
		els, err = b.buildStmt(elseVal)
		if err != nil {
			return nil, errors.ErrInvalidField("else", err)
		}
	}

	return &ast.IfStmt{
		If:   b.parsePos(ifVal),
		Init: init,
		Cond: cond,
		Body: body,
		Else: els,
	}, nil
}

// buildForStmt parses a ForStmt node
func (b *Builder) buildForStmt(s sexp.SExp) (*ast.ForStmt, error) {
	list, ok := b.expectList(s, "ForStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "ForStmt") {
		return nil, errors.ErrExpectedNodeType("ForStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	forVal, ok := b.requireKeyword(args, "for", "ForStmt")
	if !ok {
		return nil, errors.ErrMissingField("for")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "ForStmt")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, errors.ErrInvalidField("init", err)
		}
	}

	// Optional cond
	var cond ast.Expr
	if condVal, ok := args["cond"]; ok && !b.parseNil(condVal) {
		cond, err = b.buildExpr(condVal)
		if err != nil {
			return nil, errors.ErrInvalidField("cond", err)
		}
	}

	// Optional post
	var post ast.Stmt
	if postVal, ok := args["post"]; ok && !b.parseNil(postVal) {
		post, err = b.buildStmt(postVal)
		if err != nil {
			return nil, errors.ErrInvalidField("post", err)
		}
	}

	return &ast.ForStmt{
		For:  b.parsePos(forVal),
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}, nil
}

// buildRangeStmt parses a RangeStmt node
func (b *Builder) buildRangeStmt(s sexp.SExp) (*ast.RangeStmt, error) {
	list, ok := b.expectList(s, "RangeStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "RangeStmt") {
		return nil, errors.ErrExpectedNodeType("RangeStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	forVal, ok := b.requireKeyword(args, "for", "RangeStmt")
	if !ok {
		return nil, errors.ErrMissingField("for")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "RangeStmt")
	if !ok {
		return nil, errors.ErrMissingField("tokpos")
	}

	tokVal, ok := b.requireKeyword(args, "tok", "RangeStmt")
	if !ok {
		return nil, errors.ErrMissingField("tok")
	}

	xVal, ok := b.requireKeyword(args, "x", "RangeStmt")
	if !ok {
		return nil, errors.ErrMissingField("x")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "RangeStmt")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, errors.ErrInvalidField("tok", err)
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, errors.ErrInvalidField("x", err)
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("body", err)
	}

	// Optional key
	var key ast.Expr
	if keyVal, ok := args["key"]; ok && !b.parseNil(keyVal) {
		key, err = b.buildExpr(keyVal)
		if err != nil {
			return nil, errors.ErrInvalidField("key", err)
		}
	}

	// Optional value
	var value ast.Expr
	if valueVal, ok := args["value"]; ok && !b.parseNil(valueVal) {
		value, err = b.buildExpr(valueVal)
		if err != nil {
			return nil, errors.ErrInvalidField("value", err)
		}
	}

	return &ast.RangeStmt{
		For:    b.parsePos(forVal),
		Key:    key,
		Value:  value,
		TokPos: b.parsePos(tokposVal),
		Tok:    tok,
		X:      x,
		Body:   body,
	}, nil
}

// buildSwitchStmt parses a SwitchStmt node
func (b *Builder) buildSwitchStmt(s sexp.SExp) (*ast.SwitchStmt, error) {
	list, ok := b.expectList(s, "SwitchStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SwitchStmt") {
		return nil, errors.ErrExpectedNodeType("SwitchStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	switchVal, ok := b.requireKeyword(args, "switch", "SwitchStmt")
	if !ok {
		return nil, errors.ErrMissingField("switch")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "SwitchStmt")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, errors.ErrInvalidField("init", err)
		}
	}

	// Optional tag
	var tag ast.Expr
	if tagVal, ok := args["tag"]; ok && !b.parseNil(tagVal) {
		tag, err = b.buildExpr(tagVal)
		if err != nil {
			return nil, errors.ErrInvalidField("tag", err)
		}
	}

	return &ast.SwitchStmt{
		Switch: b.parsePos(switchVal),
		Init:   init,
		Tag:    tag,
		Body:   body,
	}, nil
}

// buildCaseClause parses a CaseClause node
func (b *Builder) buildCaseClause(s sexp.SExp) (*ast.CaseClause, error) {
	list, ok := b.expectList(s, "CaseClause")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "CaseClause") {
		return nil, errors.ErrExpectedNodeType("CaseClause", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	caseVal, ok := b.requireKeyword(args, "case", "CaseClause")
	if !ok {
		return nil, errors.ErrMissingField("case")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "CaseClause")
	if !ok {
		return nil, errors.ErrMissingField("colon")
	}

	listVal, ok := b.requireKeyword(args, "list", "CaseClause")
	if !ok {
		return nil, errors.ErrMissingField("list")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "CaseClause")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	// Build list of expressions
	var exprs []ast.Expr
	exprsList, ok := b.expectList(listVal, "CaseClause list")
	if ok {
		for _, exprSexp := range exprsList.Elements {
			expr, err := b.buildExpr(exprSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("list expr", err)
			}
			exprs = append(exprs, expr)
		}
	}

	// Build body statements
	var stmts []ast.Stmt
	stmtsList, ok := b.expectList(bodyVal, "CaseClause body")
	if ok {
		for _, stmtSexp := range stmtsList.Elements {
			stmt, err := b.buildStmt(stmtSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("body stmt", err)
			}
			stmts = append(stmts, stmt)
		}
	}

	return &ast.CaseClause{
		Case:  b.parsePos(caseVal),
		List:  exprs,
		Colon: b.parsePos(colonVal),
		Body:  stmts,
	}, nil
}

// buildTypeSwitchStmt parses a TypeSwitchStmt node
func (b *Builder) buildTypeSwitchStmt(s sexp.SExp) (*ast.TypeSwitchStmt, error) {
	list, ok := b.expectList(s, "TypeSwitchStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "TypeSwitchStmt") {
		return nil, errors.ErrExpectedNodeType("TypeSwitchStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	switchVal, ok := b.requireKeyword(args, "switch", "TypeSwitchStmt")
	if !ok {
		return nil, errors.ErrMissingField("switch")
	}

	assignVal, ok := b.requireKeyword(args, "assign", "TypeSwitchStmt")
	if !ok {
		return nil, errors.ErrMissingField("assign")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "TypeSwitchStmt")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	assign, err := b.buildStmt(assignVal)
	if err != nil {
		return nil, errors.ErrInvalidField("assign", err)
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("body", err)
	}

	// Optional init
	var init ast.Stmt
	if initVal, ok := args["init"]; ok && !b.parseNil(initVal) {
		init, err = b.buildStmt(initVal)
		if err != nil {
			return nil, errors.ErrInvalidField("init", err)
		}
	}

	return &ast.TypeSwitchStmt{
		Switch: b.parsePos(switchVal),
		Init:   init,
		Assign: assign,
		Body:   body,
	}, nil
}

// buildSelectStmt parses a SelectStmt node
func (b *Builder) buildSelectStmt(s sexp.SExp) (*ast.SelectStmt, error) {
	list, ok := b.expectList(s, "SelectStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "SelectStmt") {
		return nil, errors.ErrExpectedNodeType("SelectStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	selectVal, ok := b.requireKeyword(args, "select", "SelectStmt")
	if !ok {
		return nil, errors.ErrMissingField("select")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "SelectStmt")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	body, err := b.buildBlockStmt(bodyVal)
	if err != nil {
		return nil, errors.ErrInvalidField("body", err)
	}

	return &ast.SelectStmt{
		Select: b.parsePos(selectVal),
		Body:   body,
	}, nil
}

// buildCommClause parses a CommClause node
func (b *Builder) buildCommClause(s sexp.SExp) (*ast.CommClause, error) {
	list, ok := b.expectList(s, "CommClause")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "CommClause") {
		return nil, errors.ErrExpectedNodeType("CommClause", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	caseVal, ok := b.requireKeyword(args, "case", "CommClause")
	if !ok {
		return nil, errors.ErrMissingField("case")
	}

	colonVal, ok := b.requireKeyword(args, "colon", "CommClause")
	if !ok {
		return nil, errors.ErrMissingField("colon")
	}

	bodyVal, ok := b.requireKeyword(args, "body", "CommClause")
	if !ok {
		return nil, errors.ErrMissingField("body")
	}

	// Optional comm
	var comm ast.Stmt
	var err error
	if commVal, ok := args["comm"]; ok && !b.parseNil(commVal) {
		comm, err = b.buildStmt(commVal)
		if err != nil {
			return nil, errors.ErrInvalidField("comm", err)
		}
	}

	// Build body statements
	var stmts []ast.Stmt
	stmtsList, ok := b.expectList(bodyVal, "CommClause body")
	if ok {
		for _, stmtSexp := range stmtsList.Elements {
			stmt, err := b.buildStmt(stmtSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("body stmt", err)
			}
			stmts = append(stmts, stmt)
		}
	}

	return &ast.CommClause{
		Case:  b.parsePos(caseVal),
		Comm:  comm,
		Colon: b.parsePos(colonVal),
		Body:  stmts,
	}, nil
}

// buildDeclStmt parses a DeclStmt node
func (b *Builder) buildDeclStmt(s sexp.SExp) (*ast.DeclStmt, error) {
	list, ok := b.expectList(s, "DeclStmt")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "DeclStmt") {
		return nil, errors.ErrExpectedNodeType("DeclStmt", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	declVal, ok := b.requireKeyword(args, "decl", "DeclStmt")
	if !ok {
		return nil, errors.ErrMissingField("decl")
	}

	decl, err := b.buildDecl(declVal)
	if err != nil {
		return nil, errors.ErrInvalidField("decl", err)
	}

	return &ast.DeclStmt{
		Decl: decl,
	}, nil
}

// buildStmt dispatches to appropriate statement builder
func (b *Builder) buildStmt(s sexp.SExp) (ast.Stmt, error) {
	list, ok := b.expectList(s, "statement")
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
	case "ExprStmt":
		return b.buildExprStmt(s)
	case "BlockStmt":
		return b.buildBlockStmt(s)
	case "ReturnStmt":
		return b.buildReturnStmt(s)
	case "AssignStmt":
		return b.buildAssignStmt(s)
	case "IncDecStmt":
		return b.buildIncDecStmt(s)
	case "BranchStmt":
		return b.buildBranchStmt(s)
	case "DeferStmt":
		return b.buildDeferStmt(s)
	case "GoStmt":
		return b.buildGoStmt(s)
	case "SendStmt":
		return b.buildSendStmt(s)
	case "EmptyStmt":
		return b.buildEmptyStmt(s)
	case "LabeledStmt":
		return b.buildLabeledStmt(s)
	case "IfStmt":
		return b.buildIfStmt(s)
	case "ForStmt":
		return b.buildForStmt(s)
	case "RangeStmt":
		return b.buildRangeStmt(s)
	case "SwitchStmt":
		return b.buildSwitchStmt(s)
	case "CaseClause":
		return b.buildCaseClause(s)
	case "TypeSwitchStmt":
		return b.buildTypeSwitchStmt(s)
	case "SelectStmt":
		return b.buildSelectStmt(s)
	case "CommClause":
		return b.buildCommClause(s)
	case "DeclStmt":
		return b.buildDeclStmt(s)
	default:
		return nil, errors.ErrUnknownNodeType(sym.Value, "statement")
	}
}
