package zast

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"zylisp/zast/sexp"
)

// Builder builds Go AST nodes from S-expressions
type Builder struct {
	fset   *token.FileSet
	errors []string
}

// FileSetInfo stores the parsed FileSet information
type FileSetInfo struct {
	Base  int
	Files []FileInfo
}

// FileInfo stores information about a source file
type FileInfo struct {
	Name  string
	Base  int
	Size  int
	Lines []int // byte offsets of line starts
}

// NewBuilder creates a new AST builder
func NewBuilder() *Builder {
	return &Builder{
		errors: []string{},
	}
}

// Errors returns accumulated errors
func (b *Builder) Errors() []string {
	return b.errors
}

// addError records an error
func (b *Builder) addError(format string, args ...interface{}) {
	b.errors = append(b.errors, fmt.Sprintf(format, args...))
}

// expectList verifies sexp is a List and returns it
func (b *Builder) expectList(s sexp.SExp, context string) (*sexp.List, bool) {
	list, ok := s.(*sexp.List)
	if !ok {
		b.addError("%s: expected list, got %T at line %d, column %d",
			context, s, s.Pos().Line, s.Pos().Column)
		return nil, false
	}
	return list, true
}

// expectSymbol verifies sexp is a Symbol with expected value
func (b *Builder) expectSymbol(s sexp.SExp, expected string) bool {
	sym, ok := s.(*sexp.Symbol)
	if !ok {
		b.addError("expected symbol %q, got %T", expected, s)
		return false
	}
	if sym.Value != expected {
		b.addError("expected symbol %q, got %q", expected, sym.Value)
		return false
	}
	return true
}

// parseKeywordArgs converts a list of alternating keywords and values into a map
func (b *Builder) parseKeywordArgs(elements []sexp.SExp) map[string]sexp.SExp {
	args := make(map[string]sexp.SExp)

	// Start at index 1 (skip the node type symbol)
	for i := 1; i < len(elements); i += 2 {
		if i+1 >= len(elements) {
			b.addError("keyword argument missing value at index %d", i)
			break
		}

		keyword, ok := elements[i].(*sexp.Keyword)
		if !ok {
			b.addError("expected keyword at index %d, got %T", i, elements[i])
			continue
		}

		args[keyword.Name] = elements[i+1]
	}

	return args
}

// getKeyword retrieves a keyword value from the args map
func (b *Builder) getKeyword(args map[string]sexp.SExp, name string) (sexp.SExp, bool) {
	val, ok := args[name]
	return val, ok
}

// requireKeyword gets a keyword value or adds an error if missing
func (b *Builder) requireKeyword(args map[string]sexp.SExp, name string, context string) (sexp.SExp, bool) {
	val, ok := args[name]
	if !ok {
		b.addError("%s: missing required field :%s", context, name)
		return nil, false
	}
	return val, true
}

// parseInt converts a Number or Symbol to int
func (b *Builder) parseInt(s sexp.SExp) (int, error) {
	switch v := s.(type) {
	case *sexp.Number:
		return strconv.Atoi(v.Value)
	case *sexp.Symbol:
		return strconv.Atoi(v.Value)
	default:
		return 0, fmt.Errorf("expected number, got %T", s)
	}
}

// parsePos converts a Number to token.Pos
func (b *Builder) parsePos(s sexp.SExp) token.Pos {
	n, err := b.parseInt(s)
	if err != nil {
		b.addError("invalid position: %v", err)
		return token.NoPos
	}
	return token.Pos(n)
}

// parseString extracts string value from String node
func (b *Builder) parseString(s sexp.SExp) (string, error) {
	str, ok := s.(*sexp.String)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", s)
	}
	return str.Value, nil
}

// parseNil checks if value is nil
func (b *Builder) parseNil(s sexp.SExp) bool {
	_, ok := s.(*sexp.Nil)
	return ok
}

// parseToken converts a symbol to a token type
func (b *Builder) parseToken(s sexp.SExp) (token.Token, error) {
	sym, ok := s.(*sexp.Symbol)
	if !ok {
		return token.ILLEGAL, fmt.Errorf("expected symbol for token, got %T", s)
	}

	switch sym.Value {
	case "IMPORT":
		return token.IMPORT, nil
	case "CONST":
		return token.CONST, nil
	case "TYPE":
		return token.TYPE, nil
	case "VAR":
		return token.VAR, nil
	case "INT":
		return token.INT, nil
	case "FLOAT":
		return token.FLOAT, nil
	case "IMAG":
		return token.IMAG, nil
	case "CHAR":
		return token.CHAR, nil
	case "STRING":
		return token.STRING, nil
	default:
		return token.ILLEGAL, fmt.Errorf("unknown token: %s", sym.Value)
	}
}

// buildIdent parses an Ident node
func (b *Builder) buildIdent(s sexp.SExp) (*ast.Ident, error) {
	list, ok := b.expectList(s, "Ident")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if len(list.Elements) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	if !b.expectSymbol(list.Elements[0], "Ident") {
		return nil, fmt.Errorf("not an Ident node")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameposVal, ok := b.requireKeyword(args, "namepos", "Ident")
	if !ok {
		return nil, fmt.Errorf("missing namepos")
	}

	nameVal, ok := b.requireKeyword(args, "name", "Ident")
	if !ok {
		return nil, fmt.Errorf("missing name")
	}

	name, err := b.parseString(nameVal)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %v", err)
	}

	ident := &ast.Ident{
		NamePos: b.parsePos(nameposVal),
		Name:    name,
		Obj:     nil, // Objects handled separately if needed
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
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "BasicLit") {
		return nil, fmt.Errorf("not a BasicLit node")
	}

	args := b.parseKeywordArgs(list.Elements)

	valueposVal, ok := b.requireKeyword(args, "valuepos", "BasicLit")
	if !ok {
		return nil, fmt.Errorf("missing valuepos")
	}

	kindVal, ok := b.requireKeyword(args, "kind", "BasicLit")
	if !ok {
		return nil, fmt.Errorf("missing kind")
	}

	valueVal, ok := b.requireKeyword(args, "value", "BasicLit")
	if !ok {
		return nil, fmt.Errorf("missing value")
	}

	kind, err := b.parseToken(kindVal)
	if err != nil {
		return nil, fmt.Errorf("invalid kind: %v", err)
	}

	value, err := b.parseString(valueVal)
	if err != nil {
		return nil, fmt.Errorf("invalid value: %v", err)
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
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "SelectorExpr") {
		return nil, fmt.Errorf("not a SelectorExpr node")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "SelectorExpr")
	if !ok {
		return nil, fmt.Errorf("missing x")
	}

	selVal, ok := b.requireKeyword(args, "sel", "SelectorExpr")
	if !ok {
		return nil, fmt.Errorf("missing sel")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, fmt.Errorf("invalid x: %v", err)
	}

	sel, err := b.buildIdent(selVal)
	if err != nil {
		return nil, fmt.Errorf("invalid sel: %v", err)
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
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "CallExpr") {
		return nil, fmt.Errorf("not a CallExpr node")
	}

	args := b.parseKeywordArgs(list.Elements)

	funVal, ok := b.requireKeyword(args, "fun", "CallExpr")
	if !ok {
		return nil, fmt.Errorf("missing fun")
	}

	lparenVal, ok := b.requireKeyword(args, "lparen", "CallExpr")
	if !ok {
		return nil, fmt.Errorf("missing lparen")
	}

	argsVal, ok := b.requireKeyword(args, "args", "CallExpr")
	if !ok {
		return nil, fmt.Errorf("missing args")
	}

	ellipsisVal, ok := b.requireKeyword(args, "ellipsis", "CallExpr")
	if !ok {
		return nil, fmt.Errorf("missing ellipsis")
	}

	rparenVal, ok := b.requireKeyword(args, "rparen", "CallExpr")
	if !ok {
		return nil, fmt.Errorf("missing rparen")
	}

	fun, err := b.buildExpr(funVal)
	if err != nil {
		return nil, fmt.Errorf("invalid fun: %v", err)
	}

	// Build args list
	var callArgs []ast.Expr
	argsList, ok := b.expectList(argsVal, "CallExpr args")
	if ok {
		for _, argSexp := range argsList.Elements {
			arg, err := b.buildExpr(argSexp)
			if err != nil {
				return nil, fmt.Errorf("invalid arg: %v", err)
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

// buildExpr dispatches to appropriate expression builder
func (b *Builder) buildExpr(s sexp.SExp) (ast.Expr, error) {
	list, ok := b.expectList(s, "expression")
	if !ok {
		return nil, fmt.Errorf("expected list")
	}

	if len(list.Elements) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, fmt.Errorf("expected symbol as first element")
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
	default:
		return nil, fmt.Errorf("unknown expression type: %s", sym.Value)
	}
}

// buildExprStmt parses an ExprStmt node
func (b *Builder) buildExprStmt(s sexp.SExp) (*ast.ExprStmt, error) {
	list, ok := b.expectList(s, "ExprStmt")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "ExprStmt") {
		return nil, fmt.Errorf("not an ExprStmt node")
	}

	args := b.parseKeywordArgs(list.Elements)

	xVal, ok := b.requireKeyword(args, "x", "ExprStmt")
	if !ok {
		return nil, fmt.Errorf("missing x")
	}

	x, err := b.buildExpr(xVal)
	if err != nil {
		return nil, fmt.Errorf("invalid x: %v", err)
	}

	return &ast.ExprStmt{X: x}, nil
}

// buildBlockStmt parses a BlockStmt node
func (b *Builder) buildBlockStmt(s sexp.SExp) (*ast.BlockStmt, error) {
	list, ok := b.expectList(s, "BlockStmt")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "BlockStmt") {
		return nil, fmt.Errorf("not a BlockStmt node")
	}

	args := b.parseKeywordArgs(list.Elements)

	lbraceVal, ok := b.requireKeyword(args, "lbrace", "BlockStmt")
	if !ok {
		return nil, fmt.Errorf("missing lbrace")
	}

	listVal, ok := b.requireKeyword(args, "list", "BlockStmt")
	if !ok {
		return nil, fmt.Errorf("missing list")
	}

	rbraceVal, ok := b.requireKeyword(args, "rbrace", "BlockStmt")
	if !ok {
		return nil, fmt.Errorf("missing rbrace")
	}

	// Build statement list
	var stmts []ast.Stmt
	stmtsList, ok := b.expectList(listVal, "BlockStmt list")
	if ok {
		for _, stmtSexp := range stmtsList.Elements {
			stmt, err := b.buildStmt(stmtSexp)
			if err != nil {
				return nil, fmt.Errorf("invalid statement: %v", err)
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

// buildStmt dispatches to appropriate statement builder
func (b *Builder) buildStmt(s sexp.SExp) (ast.Stmt, error) {
	list, ok := b.expectList(s, "statement")
	if !ok {
		return nil, fmt.Errorf("expected list")
	}

	if len(list.Elements) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, fmt.Errorf("expected symbol as first element")
	}

	switch sym.Value {
	case "ExprStmt":
		return b.buildExprStmt(s)
	case "BlockStmt":
		return b.buildBlockStmt(s)
	default:
		return nil, fmt.Errorf("unknown statement type: %s", sym.Value)
	}
}

// buildField parses a Field node
func (b *Builder) buildField(s sexp.SExp) (*ast.Field, error) {
	list, ok := b.expectList(s, "Field")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "Field") {
		return nil, fmt.Errorf("not a Field node")
	}

	args := b.parseKeywordArgs(list.Elements)

	typeVal, ok := b.requireKeyword(args, "type", "Field")
	if !ok {
		return nil, fmt.Errorf("missing type")
	}

	fieldType, err := b.buildExpr(typeVal)
	if err != nil {
		return nil, fmt.Errorf("invalid type: %v", err)
	}

	// Build names list (optional)
	var names []*ast.Ident
	if namesVal, ok := args["names"]; ok {
		namesList, ok := b.expectList(namesVal, "Field names")
		if ok {
			for _, nameSexp := range namesList.Elements {
				ident, err := b.buildIdent(nameSexp)
				if err != nil {
					return nil, fmt.Errorf("invalid name: %v", err)
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
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "FieldList") {
		return nil, fmt.Errorf("not a FieldList node")
	}

	args := b.parseKeywordArgs(list.Elements)

	openingVal, ok := b.requireKeyword(args, "opening", "FieldList")
	if !ok {
		return nil, fmt.Errorf("missing opening")
	}

	listVal, ok := b.requireKeyword(args, "list", "FieldList")
	if !ok {
		return nil, fmt.Errorf("missing list")
	}

	closingVal, ok := b.requireKeyword(args, "closing", "FieldList")
	if !ok {
		return nil, fmt.Errorf("missing closing")
	}

	// Build field list
	var fields []*ast.Field
	fieldsList, ok := b.expectList(listVal, "FieldList list")
	if ok {
		for _, fieldSexp := range fieldsList.Elements {
			field, err := b.buildField(fieldSexp)
			if err != nil {
				return nil, fmt.Errorf("invalid field: %v", err)
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
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "FuncType") {
		return nil, fmt.Errorf("not a FuncType node")
	}

	args := b.parseKeywordArgs(list.Elements)

	funcVal, ok := b.requireKeyword(args, "func", "FuncType")
	if !ok {
		return nil, fmt.Errorf("missing func")
	}

	paramsVal, ok := b.requireKeyword(args, "params", "FuncType")
	if !ok {
		return nil, fmt.Errorf("missing params")
	}

	resultsVal, _ := args["results"]

	params, err := b.buildFieldList(paramsVal)
	if err != nil {
		return nil, fmt.Errorf("invalid params: %v", err)
	}

	results, err := b.buildOptionalFieldList(resultsVal)
	if err != nil {
		return nil, fmt.Errorf("invalid results: %v", err)
	}

	return &ast.FuncType{
		Func:    b.parsePos(funcVal),
		Params:  params,
		Results: results,
	}, nil
}

// buildImportSpec parses an ImportSpec node
func (b *Builder) buildImportSpec(s sexp.SExp) (*ast.ImportSpec, error) {
	list, ok := b.expectList(s, "ImportSpec")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "ImportSpec") {
		return nil, fmt.Errorf("not an ImportSpec node")
	}

	args := b.parseKeywordArgs(list.Elements)

	pathVal, ok := b.requireKeyword(args, "path", "ImportSpec")
	if !ok {
		return nil, fmt.Errorf("missing path")
	}

	path, err := b.buildBasicLit(pathVal)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Optional name
	var name *ast.Ident
	if nameVal, ok := args["name"]; ok {
		name, err = b.buildOptionalIdent(nameVal)
		if err != nil {
			return nil, fmt.Errorf("invalid name: %v", err)
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

// buildSpec dispatches to appropriate spec builder
func (b *Builder) buildSpec(s sexp.SExp) (ast.Spec, error) {
	list, ok := b.expectList(s, "spec")
	if !ok {
		return nil, fmt.Errorf("expected list")
	}

	if len(list.Elements) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, fmt.Errorf("expected symbol as first element")
	}

	switch sym.Value {
	case "ImportSpec":
		return b.buildImportSpec(s)
	default:
		return nil, fmt.Errorf("unknown spec type: %s", sym.Value)
	}
}

// buildGenDecl parses a GenDecl node
func (b *Builder) buildGenDecl(s sexp.SExp) (*ast.GenDecl, error) {
	list, ok := b.expectList(s, "GenDecl")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "GenDecl") {
		return nil, fmt.Errorf("not a GenDecl node")
	}

	args := b.parseKeywordArgs(list.Elements)

	tokVal, ok := b.requireKeyword(args, "tok", "GenDecl")
	if !ok {
		return nil, fmt.Errorf("missing tok")
	}

	tokposVal, ok := b.requireKeyword(args, "tokpos", "GenDecl")
	if !ok {
		return nil, fmt.Errorf("missing tokpos")
	}

	specsVal, ok := b.requireKeyword(args, "specs", "GenDecl")
	if !ok {
		return nil, fmt.Errorf("missing specs")
	}

	tok, err := b.parseToken(tokVal)
	if err != nil {
		return nil, fmt.Errorf("invalid tok: %v", err)
	}

	// Build specs list
	var specs []ast.Spec
	specsList, ok := b.expectList(specsVal, "GenDecl specs")
	if ok {
		for _, specSexp := range specsList.Elements {
			spec, err := b.buildSpec(specSexp)
			if err != nil {
				return nil, fmt.Errorf("invalid spec: %v", err)
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
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "FuncDecl") {
		return nil, fmt.Errorf("not a FuncDecl node")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "FuncDecl")
	if !ok {
		return nil, fmt.Errorf("missing name")
	}

	typeVal, ok := b.requireKeyword(args, "type", "FuncDecl")
	if !ok {
		return nil, fmt.Errorf("missing type")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %v", err)
	}

	funcType, err := b.buildFuncType(typeVal)
	if err != nil {
		return nil, fmt.Errorf("invalid type: %v", err)
	}

	// Optional recv
	var recv *ast.FieldList
	if recvVal, ok := args["recv"]; ok {
		recv, err = b.buildOptionalFieldList(recvVal)
		if err != nil {
			return nil, fmt.Errorf("invalid recv: %v", err)
		}
	}

	// Optional body
	var body *ast.BlockStmt
	if bodyVal, ok := args["body"]; ok {
		body, err = b.buildOptionalBlockStmt(bodyVal)
		if err != nil {
			return nil, fmt.Errorf("invalid body: %v", err)
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
		return nil, fmt.Errorf("expected list")
	}

	if len(list.Elements) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	sym, ok := list.Elements[0].(*sexp.Symbol)
	if !ok {
		return nil, fmt.Errorf("expected symbol as first element")
	}

	switch sym.Value {
	case "GenDecl":
		return b.buildGenDecl(s)
	case "FuncDecl":
		return b.buildFuncDecl(s)
	default:
		return nil, fmt.Errorf("unknown declaration type: %s", sym.Value)
	}
}

// buildFileInfo parses a FileInfo node
func (b *Builder) buildFileInfo(s sexp.SExp) (*FileInfo, error) {
	list, ok := b.expectList(s, "FileInfo")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "FileInfo") {
		return nil, fmt.Errorf("not a FileInfo node")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "FileInfo")
	if !ok {
		return nil, fmt.Errorf("missing name")
	}

	baseVal, ok := b.requireKeyword(args, "base", "FileInfo")
	if !ok {
		return nil, fmt.Errorf("missing base")
	}

	sizeVal, ok := b.requireKeyword(args, "size", "FileInfo")
	if !ok {
		return nil, fmt.Errorf("missing size")
	}

	linesVal, ok := b.requireKeyword(args, "lines", "FileInfo")
	if !ok {
		return nil, fmt.Errorf("missing lines")
	}

	name, err := b.parseString(nameVal)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %v", err)
	}

	base, err := b.parseInt(baseVal)
	if err != nil {
		return nil, fmt.Errorf("invalid base: %v", err)
	}

	size, err := b.parseInt(sizeVal)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %v", err)
	}

	// Parse lines list
	linesList, ok := b.expectList(linesVal, "FileInfo lines")
	if !ok {
		return nil, fmt.Errorf("invalid lines")
	}

	var lines []int
	for _, lineSexp := range linesList.Elements {
		line, err := b.parseInt(lineSexp)
		if err != nil {
			return nil, fmt.Errorf("invalid line offset: %v", err)
		}
		lines = append(lines, line)
	}

	return &FileInfo{
		Name:  name,
		Base:  base,
		Size:  size,
		Lines: lines,
	}, nil
}

// buildFileSet parses a FileSet node
func (b *Builder) buildFileSet(s sexp.SExp) (*FileSetInfo, error) {
	list, ok := b.expectList(s, "FileSet")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "FileSet") {
		return nil, fmt.Errorf("not a FileSet node")
	}

	args := b.parseKeywordArgs(list.Elements)

	baseVal, ok := b.requireKeyword(args, "base", "FileSet")
	if !ok {
		return nil, fmt.Errorf("missing base")
	}

	filesVal, ok := b.requireKeyword(args, "files", "FileSet")
	if !ok {
		return nil, fmt.Errorf("missing files")
	}

	base, err := b.parseInt(baseVal)
	if err != nil {
		return nil, fmt.Errorf("invalid base: %v", err)
	}

	// Parse files list
	filesList, ok := b.expectList(filesVal, "FileSet files")
	if !ok {
		return nil, fmt.Errorf("invalid files")
	}

	var files []FileInfo
	for _, fileSexp := range filesList.Elements {
		fileInfo, err := b.buildFileInfo(fileSexp)
		if err != nil {
			return nil, fmt.Errorf("invalid file info: %v", err)
		}
		files = append(files, *fileInfo)
	}

	return &FileSetInfo{
		Base:  base,
		Files: files,
	}, nil
}

// BuildFile converts a File s-expression to *ast.File
func (b *Builder) BuildFile(s sexp.SExp) (*ast.File, error) {
	list, ok := b.expectList(s, "File")
	if !ok {
		return nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "File") {
		return nil, fmt.Errorf("not a File node")
	}

	args := b.parseKeywordArgs(list.Elements)

	packageVal, ok := b.requireKeyword(args, "package", "File")
	if !ok {
		return nil, fmt.Errorf("missing package")
	}

	nameVal, ok := b.requireKeyword(args, "name", "File")
	if !ok {
		return nil, fmt.Errorf("missing name")
	}

	declsVal, ok := b.requireKeyword(args, "decls", "File")
	if !ok {
		return nil, fmt.Errorf("missing decls")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %v", err)
	}

	// Build declarations list
	var decls []ast.Decl
	declsList, ok := b.expectList(declsVal, "File decls")
	if ok {
		for _, declSexp := range declsList.Elements {
			decl, err := b.buildDecl(declSexp)
			if err != nil {
				return nil, fmt.Errorf("invalid declaration: %v", err)
			}
			decls = append(decls, decl)
		}
	}

	// Optional imports, unresolved, comments - ignore for now

	file := &ast.File{
		Package: b.parsePos(packageVal),
		Name:    name,
		Decls:   decls,
	}

	return file, nil
}

// BuildProgram parses a Program s-expression and returns FileSet and Files
func (b *Builder) BuildProgram(s sexp.SExp) (*token.FileSet, []*ast.File, error) {
	list, ok := b.expectList(s, "Program")
	if !ok {
		return nil, nil, fmt.Errorf("not a list")
	}

	if !b.expectSymbol(list.Elements[0], "Program") {
		return nil, nil, fmt.Errorf("not a Program node")
	}

	args := b.parseKeywordArgs(list.Elements)

	filesetVal, ok := b.requireKeyword(args, "fileset", "Program")
	if !ok {
		return nil, nil, fmt.Errorf("missing fileset")
	}

	filesVal, ok := b.requireKeyword(args, "files", "Program")
	if !ok {
		return nil, nil, fmt.Errorf("missing files")
	}

	// Build FileSet
	fileSetInfo, err := b.buildFileSet(filesetVal)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid fileset: %v", err)
	}

	// Create token.FileSet from FileSetInfo
	fset := token.NewFileSet()
	for _, fi := range fileSetInfo.Files {
		fset.AddFile(fi.Name, fi.Base, fi.Size)
	}
	b.fset = fset

	// Build files list
	var files []*ast.File
	filesList, ok := b.expectList(filesVal, "Program files")
	if ok {
		for _, fileSexp := range filesList.Elements {
			file, err := b.BuildFile(fileSexp)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid file: %v", err)
			}
			files = append(files, file)
		}
	}

	if len(b.errors) > 0 {
		return nil, nil, fmt.Errorf("build errors: %s", strings.Join(b.errors, "; "))
	}

	return fset, files, nil
}
