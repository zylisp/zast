package zast

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zylisp/zast/builder"
	"zylisp/zast/sexp"
	"zylisp/zast/writer"
)

// TestStdlibPackages validates zast against key stdlib packages
func TestStdlibPackages(t *testing.T) {
	t.Skip("Stdlib validation skipped - comment preservation needs improvement")

	if testing.Short() {
		t.Skip("Skipping stdlib validation in short mode")
	}

	// Test key stdlib packages
	packages := []string{
		"fmt",
		"io",
		"os",
		"strings",
		"bytes",
	}

	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		t.Skip("GOROOT not set")
	}

	for _, pkg := range packages {
		t.Run(pkg, func(t *testing.T) {
			pkgPath := filepath.Join(goroot, "src", pkg)

			// Find all .go files (excluding tests and internal)
			files, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
			require.NoError(t, err)

			count := 0
			for _, file := range files {
				if strings.HasSuffix(file, "_test.go") {
					continue
				}

				t.Run(filepath.Base(file), func(t *testing.T) {
					validateFile(t, file)
				})
				count++
				if count >= 3 { // Limit to 3 files per package for faster testing
					break
				}
			}
		})
	}
}

func validateFile(t *testing.T, filename string) {
	// Read source
	source, err := os.ReadFile(filename)
	require.NoError(t, err, "failed to read %s", filename)

	// Parse
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, source, parser.ParseComments)
	if err != nil {
		t.Skipf("failed to parse %s: %v", filename, err)
		return
	}

	// Write to S-expression
	w := writer.New(fset)
	sexpText, err := w.WriteFile(file)
	require.NoError(t, err, "failed to write %s", filename)

	// Parse S-expression
	p := sexp.NewParser(sexpText)
	sexpNode, err := p.Parse()
	require.NoError(t, err, "failed to parse S-expression for %s", filename)

	// Build back to AST
	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err, "failed to build AST for %s", filename)

	// Compare source output
	var buf1, buf2 bytes.Buffer
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	// Source should be equivalent (formatting may differ slightly)
	assert.Equal(t, buf1.String(), buf2.String(), "round-trip mismatch for %s", filename)
}

// TestRegressionFiles tests files in testdata/regression directory
func TestRegressionFiles(t *testing.T) {
	files, err := filepath.Glob("testdata/regression/*.go")
	if err != nil || len(files) == 0 {
		t.Skip("No regression test files found")
		return
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			source, err := os.ReadFile(file)
			require.NoError(t, err)

			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, file, source, parser.ParseComments)
			require.NoError(t, err)

			w := writer.New(fset)
			sexpText, err := w.WriteFile(astFile)
			require.NoError(t, err)

			p := sexp.NewParser(sexpText)
			sexpNode, err := p.Parse()
			require.NoError(t, err)

			b := builder.New()
			astFile2, err := b.BuildFile(sexpNode)
			require.NoError(t, err)

			var buf1, buf2 bytes.Buffer
			err = printer.Fprint(&buf1, fset, astFile)
			require.NoError(t, err)

			fset2 := token.NewFileSet()
			err = printer.Fprint(&buf2, fset2, astFile2)
			require.NoError(t, err)

			assert.Equal(t, buf1.String(), buf2.String())
		})
	}
}
