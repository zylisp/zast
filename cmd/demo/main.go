package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"zylisp/zast"
	"zylisp/zast/builder"
	"zylisp/zast/sexp"
	"zylisp/zast/writer"
)

const (
	helloWorldSource = `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`

	expectedOutput = "Hello, world!\n"
)

var (
	keepFiles = flag.Bool("keep", false, "Keep intermediate files")
	verbose   = flag.Bool("verbose", false, "Verbose output")
	showSexp  = flag.Bool("show-sexp", false, "Show S-expression output")
)

func main() {
	flag.Parse()

	fmt.Print("=== Zylisp Bootstrap Demo ===\n\n")

	// Set up configuration
	config := zast.DefaultConfig()

	// Configure output based on flags
	if *keepFiles {
		config.Output.KeepFiles = true
	}

	// Get build directories
	buildDir, err := config.Output.GetBuildDir()
	if err != nil {
		log.Fatal(err)
	}

	astBuildDir, err := config.Output.GetASTBuildDir()
	if err != nil {
		log.Fatal(err)
	}

	// Clean up unless keeping files
	if !config.Output.KeepFiles {
		defer os.RemoveAll(buildDir)
	} else {
		fmt.Printf("Build directory: %s\n", buildDir)
		fmt.Printf("AST directory: %s\n", astBuildDir)
	}

	fmt.Printf("Working directory: %s\n", buildDir)
	fmt.Printf("AST directory: %s\n\n", astBuildDir)

	// Step 1: Parse Go source to AST
	fmt.Println("Step 1: Parsing Go source to AST...")
	fset, astFile := parseGoSource(helloWorldSource)
	fmt.Print("✓ Parsed successfully\n\n")

	// Step 2: Convert AST to S-expression
	fmt.Println("Step 2: Converting AST to S-expression...")
	sexpText := astToSexp(fset, astFile)

	// Parse and pretty print the S-expression
	parser := sexp.NewParser(sexpText)
	sexpTree, _ := parser.Parse()
	pp := sexp.NewPrettyPrinterWithConfig(config.Printer)
	pretty := pp.Format(sexpTree)
	fmt.Println("\n=== Pretty-Printed S-Expression ===")
	fmt.Println(pretty)
	fmt.Print("===================================\n\n")

	logVerbose("S-expression length: %d bytes", len(sexpText))
	fmt.Print("✓ Converted successfully\n\n")

	// Step 3: Write S-expression to file
	fmt.Println("Step 3: Writing S-expression to file...")
	sexpPath := writeSexpToFile(astBuildDir, sexpText, config.Output.ASTFileExt)
	fmt.Printf("✓ Written to %s\n\n", sexpPath)

	// Step 4: Read S-expression from file
	fmt.Println("Step 4: Reading S-expression from file...")
	sexpTextRead := readSexpFromFile(sexpPath)
	logVerbose("Read %d bytes", len(sexpTextRead))
	fmt.Print("✓ Read successfully\n\n")

	// Step 5: Parse S-expression to generic tree
	fmt.Println("Step 5: Parsing S-expression to generic tree...")
	sexpTree2 := parseSexp(sexpTextRead)
	logVerbose("Parsed to tree structure")
	fmt.Print("✓ Parsed successfully\n\n")

	// Step 6: Convert S-expression to AST
	fmt.Println("Step 6: Converting S-expression to AST...")
	fset2, astFile2 := sexpToAST(sexpTree2)
	logVerbose("Converted to AST with %d declarations", len(astFile2.Decls))
	fmt.Print("✓ Converted successfully\n\n")

	// Step 7: Generate Go source from AST
	fmt.Println("Step 7: Generating Go source from parsed AST...")
	goSource := astToGoSource(fset2, astFile2)
	logVerbose("Generated %d bytes of Go source", len(goSource))
	fmt.Print("✓ Generated successfully\n\n")

	// Step 8: Write Go source to file
	fmt.Println("Step 8: Writing Go source to file...")
	goPath := writeGoToFile(buildDir, goSource)
	fmt.Printf("✓ Written to %s\n\n", goPath)

	// Step 9: Compile Go source to binary
	fmt.Println("Step 9: Compiling Go source to binary...")
	binaryPath := compileGo(buildDir, goPath)
	fmt.Printf("✓ Compiled to %s\n\n", binaryPath)

	// Step 10: Execute binary
	fmt.Println("Step 10: Executing binary...")
	output := executeBinary(binaryPath)
	fmt.Printf("✓ Output: %q\n\n", output)

	// Step 11: Verify output
	fmt.Println("Step 11: Verifying output...")
	verifyOutput(output, expectedOutput)
	fmt.Print("✓ Output verified!\n\n")

	// Success!
	fmt.Println("=== SUCCESS! ===")
	fmt.Println("Complete round-trip successful:")
	fmt.Println("  Go → AST → S-expr → file → S-expr → AST → Go → binary → execution")
	fmt.Println("\nAll components working correctly! 🎉")

	// Optionally show the S-expression
	if *showSexp {
		fmt.Println("\n=== S-Expression Output ===")
		fmt.Println(sexpText)
	}

	// Compare sources if verbose
	if *verbose {
		compareSource(helloWorldSource, goSource)
	}
}

// Step 1: Parse Go Source
func parseGoSource(source string) (*token.FileSet, *ast.File) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "hello.go", source, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse Go source: %v", err)
	}
	return fset, file
}

// Step 2: AST to S-expression
func astToSexp(fset *token.FileSet, file *ast.File) string {
	wrtr := writer.New(fset)
	sexpText, err := wrtr.WriteProgram([]*ast.File{file})
	if err != nil {
		log.Fatalf("Failed to convert AST to S-expression: %v", err)
	}
	return sexpText
}

// Step 3: Write S-expression to File
func writeSexpToFile(dir string, sexpText string, ext string) string {
	path := filepath.Join(dir, "hello"+ext)
	err := os.WriteFile(path, []byte(sexpText), 0644)
	if err != nil {
		log.Fatalf("Failed to write S-expression to file: %v", err)
	}
	return path
}

// Step 4: Read S-expression from File
func readSexpFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read S-expression from file: %v", err)
	}
	return string(data)
}

// Step 5: Parse S-expression
func parseSexp(sexpText string) sexp.SExp {
	parser := sexp.NewParser(sexpText)
	tree, err := parser.Parse()
	if err != nil {
		log.Fatalf("Failed to parse S-expression: %v", err)
	}
	return tree
}

// Step 6: S-expression to AST
func sexpToAST(tree sexp.SExp) (*token.FileSet, *ast.File) {
	bldr := builder.New()
	fset, files, err := bldr.BuildProgram(tree)
	if err != nil {
		log.Fatalf("Failed to convert S-expression to AST: %v", err)
	}
	if len(files) != 1 {
		log.Fatalf("Expected 1 file, got %d", len(files))
	}
	return fset, files[0]
}

// Step 7: AST to Go Source
func astToGoSource(fset *token.FileSet, file *ast.File) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, fset, file)
	if err != nil {
		log.Fatalf("Failed to generate Go source: %v", err)
	}
	return buf.String()
}

// Step 8: Write Go Source to File
func writeGoToFile(dir string, source string) string {
	path := filepath.Join(dir, "hello_generated.go")
	err := os.WriteFile(path, []byte(source), 0644)
	if err != nil {
		log.Fatalf("Failed to write Go source to file: %v", err)
	}
	return path
}

// Step 9: Compile Go Source
func compileGo(dir string, goPath string) string {
	binaryPath := filepath.Join(dir, "hello")

	cmd := exec.Command("go", "build", "-o", binaryPath, goPath)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to compile Go source: %v\nOutput: %s", err, output)
	}

	return binaryPath
}

// Step 10: Execute Binary
func executeBinary(binaryPath string) string {
	cmd := exec.Command(binaryPath)
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to execute binary: %v", err)
	}
	return string(output)
}

// Step 11: Verify Output
func verifyOutput(actual, expected string) {
	if actual != expected {
		log.Fatalf("Output mismatch!\nExpected: %q\nActual: %q", expected, actual)
	}
}

// Helper: Verbose logging
func logVerbose(format string, args ...interface{}) {
	if *verbose {
		fmt.Printf("  → "+format+"\n", args...)
	}
}

// Helper: Compare sources
func compareSource(original, generated string) {
	fmt.Println("\n=== Source Comparison ===")

	if original == generated {
		fmt.Println("✓ Generated source is identical to original!")
		return
	}

	fmt.Println("⚠ Generated source differs from original (this is OK - formatting may differ)")

	fmt.Println("\n=== Original ===")
	fmt.Println(original)
	fmt.Println("\n=== Generated ===")
	fmt.Println(generated)
}
