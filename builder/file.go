package builder

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"zylisp/zast/errors"
	"zylisp/zast/sexp"
)

// buildFileInfo parses a FileInfo node
func (b *Builder) buildFileInfo(s sexp.SExp) (*FileInfo, error) {
	list, ok := b.expectList(s, "FileInfo")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FileInfo") {
		return nil, errors.ErrExpectedNodeType("FileInfo", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	nameVal, ok := b.requireKeyword(args, "name", "FileInfo")
	if !ok {
		return nil, errors.ErrMissingField("name")
	}

	baseVal, ok := b.requireKeyword(args, "base", "FileInfo")
	if !ok {
		return nil, errors.ErrMissingField("base")
	}

	sizeVal, ok := b.requireKeyword(args, "size", "FileInfo")
	if !ok {
		return nil, errors.ErrMissingField("size")
	}

	linesVal, ok := b.requireKeyword(args, "lines", "FileInfo")
	if !ok {
		return nil, errors.ErrMissingField("lines")
	}

	name, err := b.parseString(nameVal)
	if err != nil {
		return nil, errors.ErrInvalidField("name", err)
	}

	base, err := b.parseInt(baseVal)
	if err != nil {
		return nil, errors.ErrInvalidField("base", err)
	}

	size, err := b.parseInt(sizeVal)
	if err != nil {
		return nil, errors.ErrInvalidField("size", err)
	}

	// Parse lines list
	linesList, ok := b.expectList(linesVal, "FileInfo lines")
	if !ok {
		return nil, errors.ErrInvalidField("lines", errors.ErrNotList)
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

// buildFileSet parses a FileSet node (simplified - positions are not preserved)
func (b *Builder) buildFileSet(s sexp.SExp) (*FileSetInfo, error) {
	list, ok := b.expectList(s, "FileSet")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "FileSet") {
		return nil, errors.ErrExpectedNodeType("FileSet", "unknown")
	}

	// We parse the FileSet structure but don't use it for position reconstruction
	// Positions cannot be meaningfully preserved through S-expression serialization
	return &FileSetInfo{
		Base:  1,
		Files: nil,
	}, nil
}

// BuildFile converts a File s-expression to *ast.File
func (b *Builder) BuildFile(s sexp.SExp) (*ast.File, error) {
	list, ok := b.expectList(s, "File")
	if !ok {
		return nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "File") {
		return nil, errors.ErrExpectedNodeType("File", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	packageVal, ok := b.requireKeyword(args, "package", "File")
	if !ok {
		return nil, errors.ErrMissingField("package")
	}

	nameVal, ok := b.requireKeyword(args, "name", "File")
	if !ok {
		return nil, errors.ErrMissingField("name")
	}

	declsVal, ok := b.requireKeyword(args, "decls", "File")
	if !ok {
		return nil, errors.ErrMissingField("decls")
	}

	name, err := b.buildIdent(nameVal)
	if err != nil {
		return nil, errors.ErrInvalidField("name", err)
	}

	// Build declarations list
	var decls []ast.Decl
	declsList, ok := b.expectList(declsVal, "File decls")
	if ok {
		for _, declSexp := range declsList.Elements {
			decl, err := b.buildDecl(declSexp)
			if err != nil {
				return nil, errors.ErrInvalidField("declaration", err)
			}
			decls = append(decls, decl)
		}
	}

	// Optional doc
	var doc *ast.CommentGroup
	if docVal, ok := args["doc"]; ok {
		doc, err = b.buildCommentGroup(docVal)
		if err != nil {
			return nil, errors.ErrInvalidField("doc", err)
		}
	}

	// Optional scope
	var scope *ast.Scope
	if scopeVal, ok := args["scope"]; ok {
		scope, err = b.buildScope(scopeVal)
		if err != nil {
			return nil, errors.ErrInvalidField("scope", err)
		}
	}

	// Optional imports
	var imports []*ast.ImportSpec
	if importsVal, ok := args["imports"]; ok {
		importsList, ok := b.expectList(importsVal, "File imports")
		if ok {
			for _, importSexp := range importsList.Elements {
				importSpec, err := b.buildImportSpec(importSexp)
				if err != nil {
					return nil, errors.ErrInvalidField("import", err)
				}
				imports = append(imports, importSpec)
			}
		}
	}

	// Optional unresolved
	var unresolved []*ast.Ident
	if unresolvedVal, ok := args["unresolved"]; ok {
		unresolvedList, ok := b.expectList(unresolvedVal, "File unresolved")
		if ok {
			for _, identSexp := range unresolvedList.Elements {
				ident, err := b.buildIdent(identSexp)
				if err != nil {
					return nil, errors.ErrInvalidField("unresolved ident", err)
				}
				unresolved = append(unresolved, ident)
			}
		}
	}

	// Optional comments
	var comments []*ast.CommentGroup
	if commentsVal, ok := args["comments"]; ok {
		commentsList, ok := b.expectList(commentsVal, "File comments")
		if ok {
			for _, cgSexp := range commentsList.Elements {
				cg, err := b.buildCommentGroup(cgSexp)
				if err != nil {
					return nil, errors.ErrInvalidField("comment group", err)
				}
				comments = append(comments, cg)
			}
		}
	}

	file := &ast.File{
		Doc:        doc,
		Package:    b.parsePos(packageVal),
		Name:       name,
		Decls:      decls,
		Scope:      scope,
		Imports:    imports,
		Unresolved: unresolved,
		Comments:   comments,
	}

	return file, nil
}

// BuildProgram parses a Program s-expression and returns FileSet and Files
func (b *Builder) BuildProgram(s sexp.SExp) (*token.FileSet, []*ast.File, error) {
	list, ok := b.expectList(s, "Program")
	if !ok {
		return nil, nil, errors.ErrNotList
	}

	if !b.expectSymbol(list.Elements[0], "Program") {
		return nil, nil, errors.ErrExpectedNodeType("Program", "unknown")
	}

	args := b.parseKeywordArgs(list.Elements)

	filesetVal, ok := b.requireKeyword(args, "fileset", "Program")
	if !ok {
		return nil, nil, errors.ErrMissingField("fileset")
	}

	filesVal, ok := b.requireKeyword(args, "files", "Program")
	if !ok {
		return nil, nil, errors.ErrMissingField("files")
	}

	// Parse FileSet (but don't use it - positions can't be preserved)
	_, err := b.buildFileSet(filesetVal)
	if err != nil {
		return nil, nil, errors.ErrInvalidField("fileset", err)
	}

	// Create a simple FileSet for the reconstructed AST
	// Positions will all be token.NoPos (0) - this is intentional
	fset := token.NewFileSet()
	b.fset = fset

	// Build files list
	var files []*ast.File
	filesList, ok := b.expectList(filesVal, "Program files")
	if ok {
		for _, fileSexp := range filesList.Elements {
			file, err := b.BuildFile(fileSexp)
			if err != nil {
				return nil, nil, errors.ErrInvalidField("file", err)
			}
			files = append(files, file)

			// Add each file to the FileSet with a generous size estimate
			// Actual positions don't matter since they're all NoPos anyway
			fset.AddFile(file.Name.Name, fset.Base(), 1000000)
		}
	}

	if len(b.errors) > 0 {
		return nil, nil, fmt.Errorf("build errors: %s", strings.Join(b.errors, "; "))
	}

	return fset, files, nil
}
