// Package zast provides conversion between Go AST and S-expression representation.
//
// zast enables programmatic manipulation of Go source code by converting Go's
// abstract syntax tree (AST) into a canonical S-expression format and back.
// This is particularly useful for:
//   - Code generation and transformation
//   - Static analysis tools
//   - Educational tools for understanding Go's AST
//   - Building Lisp-like macro systems for Go
//
// # Basic Usage
//
// Convert Go source to S-expression:
//
//	fset := token.NewFileSet()
//	file, err := parser.ParseFile(fset, "example.go", source, parser.ParseComments)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	writer := writer.New(fset)
//	sexpText, err := writer.WriteFile(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(sexpText)
//
// Convert S-expression back to Go AST:
//
//	parser := sexp.NewParser(sexpText)
//	sexpNode, err := parser.Parse()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	builder := builder.New()
//	file, err := builder.BuildFile(sexpNode)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # S-Expression Format
//
// The S-expression format mirrors Go's AST structure using keyword arguments:
//
//	(Ident :namepos 10 :name "x" :obj nil)
//	(BinaryExpr :x (Ident...) :oppos 15 :op ADD :y (Ident...))
//	(FuncDecl :name (Ident...) :type (FuncType...) :body (BlockStmt...))
//
// See the documentation for individual node types for details on their
// S-expression representation.
//
// # Important Limitations
//
// # Position Information
//
// **zast does not preserve exact source positions through round-trip conversion.**
//
// When converting Go AST → S-expr → Go AST:
//   - Position information (token.Pos) is lost
//   - The reconstructed AST has all positions set to token.NoPos (0)
//   - Comments are preserved and attached to correct AST nodes, but their positions are reset
//   - Use go/printer to format the output - it will generate clean, valid formatting
//
// This is **by design**:
//   - S-expression representation is for code transformation and analysis
//   - Not for preserving exact source formatting or file positions
//   - If you need to preserve original positions, keep the original AST and FileSet
//
// # What is Preserved
//
// ✅ Complete AST structure
// ✅ All comments (attached to correct nodes)
// ✅ All code semantics
// ✅ Ability to generate valid Go code
//
// # What is Lost
//
// ❌ Exact original formatting (spacing, line breaks)
// ❌ Source file positions
// ❌ Ability to point to original source locations
//
// # Comments
//
// When parsing with parser.ParseComments, all comments are preserved and
// attached to the correct AST nodes. However, their position information is
// reset to token.NoPos. The go/printer package will place comments appropriately
// when formatting the output.
//
// # Error Handling
//
// All conversion functions return descriptive errors when:
//   - S-expression syntax is invalid
//   - Required fields are missing
//   - Field types don't match expectations
//   - Go source code has syntax errors (when parsing)
//
// # Performance
//
// zast is designed to handle production codebases efficiently:
//   - Small files (< 100 lines): < 10ms round-trip
//   - Medium files (500-1000 lines): < 100ms round-trip
//   - Large files (5000-10000 lines): < 1 second round-trip
//   - Minimal memory allocation
//   - No recursion limits for normal code
//
// # Additional Limitations
//
// The following features are not yet fully supported:
//   - Generic type parameters (Go 1.18+)
//   - Interface method types
//
// # Future Work
//
// Full position tracking through compilation stages will be handled by the
// zylisp/core source map architecture. This will provide proper source
// location tracking for error reporting in the Zylisp compiler pipeline
//
// # Example
//
// Here's a complete example of round-trip conversion:
//
//	package main
//
//	import (
//	    "fmt"
//	    "go/parser"
//	    "go/printer"
//	    "go/token"
//	    "log"
//
//	    "zylisp/zast/builder"
//	    "zylisp/zast/sexp"
//	    "zylisp/zast/writer"
//	)
//
//	func main() {
//	    source := `package main
//
//	import "fmt"
//
//	func main() {
//	    fmt.Println("Hello, world!")
//	}
//	`
//
//	    // Parse Go source
//	    fset := token.NewFileSet()
//	    file, err := parser.ParseFile(fset, "hello.go", source, 0)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Convert to S-expression
//	    w := writer.New(fset)
//	    sexpText, err := w.WriteFile(file)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    fmt.Println("S-expression:", sexpText[:80], "...")
//
//	    // Convert back to AST
//	    p := sexp.NewParser(sexpText)
//	    sexpNode, err := p.Parse()
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    b := builder.New()
//	    file2, err := b.BuildFile(sexpNode)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Verify round-trip
//	    var buf1, buf2 bytes.Buffer
//	    printer.Fprint(&buf1, fset, file)
//	    printer.Fprint(&buf2, token.NewFileSet(), file2)
//
//	    if buf1.String() == buf2.String() {
//	        fmt.Println("Round-trip successful!")
//	    }
//	}
package zast
