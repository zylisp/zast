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
// # Position Information
//
// All position information from the original source is preserved through
// round-trip conversion. Positions are represented as byte offsets.
//
// # Comments
//
// When parsing with parser.ParseComments, all comments are preserved and
// included in the S-expression representation.
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
// # Current Limitations
//
// The following features are not yet fully supported:
//   - Generic type parameters (Go 1.18+)
//   - Interface method types
//   - Comment position preservation (comments are preserved but may be repositioned)
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
