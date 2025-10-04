package writer

import (
	"go/ast"
	"go/token"
)

// FileSet writers

func (w *Writer) writeFileInfo(file *token.File) error {
	w.openList()
	w.writeSymbol("FileInfo")
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	w.writeString(file.Name())
	w.writeSpace()
	w.writeKeyword("base")
	w.writeSpace()
	w.writePos(token.Pos(file.Base()))
	w.writeSpace()
	w.writeKeyword("size")
	w.writeSpace()
	w.writePos(token.Pos(file.Size()))
	w.writeSpace()
	w.writeKeyword("lines")
	w.writeSpace()

	// Write line offsets
	w.openList()
	// Get line count
	lineCount := file.LineCount()
	for i := 1; i <= lineCount; i++ {
		if i > 1 {
			w.writeSpace()
		}
		offset := file.LineStart(i) - token.Pos(file.Base())
		w.writePos(token.Pos(int(offset) + 1))
	}
	w.closeList()

	w.closeList()
	return nil
}

func (w *Writer) writeFileSet() error {
	w.openList()
	w.writeSymbol("FileSet")
	w.writeSpace()
	w.writeKeyword("base")
	w.writeSpace()
	w.writePos(token.Pos(w.fset.Base()))
	w.writeSpace()
	w.writeKeyword("files")
	w.writeSpace()

	w.openList()
	// Collect all files from the FileSet
	// This is a simplified approach - iterate through positions
	files := w.collectFiles()
	for i, file := range files {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeFileInfo(file); err != nil {
			return err
		}
	}
	w.closeList()

	w.closeList()
	return nil
}

func (w *Writer) collectFiles() []*token.File {
	var files []*token.File

	// Iterate through all positions to find files
	// Start from base position and try to find files
	base := w.fset.Base()
	maxPos := base + w.config.MaxFileSetScan

	seen := make(map[*token.File]bool)
	for i := base; i < maxPos; i++ {
		pos := token.Pos(i)
		if file := w.fset.File(pos); file != nil && !seen[file] {
			files = append(files, file)
			seen[file] = true
		}

		// Stop if we haven't found a new file in a while
		if len(files) > 0 && i > base+w.config.FileSearchWindow {
			break
		}
	}

	return files
}

// WriteFile writes a single File node
func (w *Writer) WriteFile(file *ast.File) (string, error) {
	w.buf.Reset()

	w.openList()
	w.writeSymbol("File")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	if err := w.writeCommentGroup(file.Doc); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("package")
	w.writeSpace()
	w.writePos(file.Package)
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	if err := w.writeIdent(file.Name); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("decls")
	w.writeSpace()
	if err := w.writeDeclList(file.Decls); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("scope")
	w.writeSpace()
	if err := w.writeScope(file.Scope); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("imports")
	w.writeSpace()
	if err := w.writeImportSpecList(file.Imports); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("unresolved")
	w.writeSpace()
	if err := w.writeIdentList(file.Unresolved); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("comments")
	w.writeSpace()
	if err := w.writeCommentGroupList(file.Comments); err != nil {
		return "", err
	}
	w.closeList()

	return w.buf.String(), nil
}

// WriteProgram writes a complete Program with FileSet and Files
func (w *Writer) WriteProgram(files []*ast.File) (string, error) {
	w.buf.Reset()

	w.openList()
	w.writeSymbol("Program")
	w.writeSpace()
	w.writeKeyword("fileset")
	w.writeSpace()
	if err := w.writeFileSet(); err != nil {
		return "", err
	}
	w.writeSpace()
	w.writeKeyword("files")
	w.writeSpace()

	w.openList()
	for i, file := range files {
		if i > 0 {
			w.writeSpace()
		}
		// Write file inline
		if err := w.writeFileNode(file); err != nil {
			return "", err
		}
	}
	w.closeList()

	w.closeList()

	return w.buf.String(), nil
}

func (w *Writer) writeFileNode(file *ast.File) error {
	w.openList()
	w.writeSymbol("File")
	w.writeSpace()
	w.writeKeyword("doc")
	w.writeSpace()
	if err := w.writeCommentGroup(file.Doc); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("package")
	w.writeSpace()
	w.writePos(file.Package)
	w.writeSpace()
	w.writeKeyword("name")
	w.writeSpace()
	if err := w.writeIdent(file.Name); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("decls")
	w.writeSpace()
	if err := w.writeDeclList(file.Decls); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("scope")
	w.writeSpace()
	if err := w.writeScope(file.Scope); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("imports")
	w.writeSpace()
	if err := w.writeImportSpecList(file.Imports); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("unresolved")
	w.writeSpace()
	if err := w.writeIdentList(file.Unresolved); err != nil {
		return err
	}
	w.writeSpace()
	w.writeKeyword("comments")
	w.writeSpace()
	if err := w.writeCommentGroupList(file.Comments); err != nil {
		return err
	}
	w.closeList()

	return nil
}

// writeImportSpecList writes a list of ImportSpec nodes
func (w *Writer) writeImportSpecList(specs []*ast.ImportSpec) error {
	w.openList()
	for i, spec := range specs {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeImportSpec(spec); err != nil {
			return err
		}
	}
	w.closeList()
	return nil
}

// writeCommentGroupList writes a list of CommentGroup nodes
func (w *Writer) writeCommentGroupList(groups []*ast.CommentGroup) error {
	w.openList()
	for i, group := range groups {
		if i > 0 {
			w.writeSpace()
		}
		if err := w.writeCommentGroup(group); err != nil {
			return err
		}
	}
	w.closeList()
	return nil
}
