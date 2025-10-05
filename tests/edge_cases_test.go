package zast

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zylisp/zast/builder"
	"zylisp/zast/sexp"
	"zylisp/zast/writer"
)

// TestEmptyFile tests an empty file
func TestEmptyFile(t *testing.T) {
	source := `package main
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String())
}

// TestDeeplyNestedExpressions tests deeply nested expressions
func TestDeeplyNestedExpressions(t *testing.T) {
	source := `package main

func f() {
	return ((((((1 + 2) * 3) - 4) / 5) % 6) & 7)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String())
}

// TestUnicodeIdentifiers tests unicode in identifiers
func TestUnicodeIdentifiers(t *testing.T) {
	source := `package main

var 変数 = 42
var переменная string = "test"
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String())
}

// TestAllOperators tests all Go operators
func TestAllOperators(t *testing.T) {
	source := `package main

func allOps() {
	_ = 1 + 2 - 3 * 4 / 5 % 6
	_ = 1 & 2 | 3 ^ 4 &^ 5
	_ = 1 << 2 >> 3
	_ = 1 == 2 != 3 < 4 <= 5 > 6 >= 7
	_ = true && false || !true
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String())
}

// TestEmptyInterfacesAndStructs tests empty interfaces and structs
func TestEmptyInterfacesAndStructs(t *testing.T) {
	source := `package main

type Empty interface{}
type EmptyStruct struct{}

var _ Empty = EmptyStruct{}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String())
}

// TestVariadicFunctions tests variadic functions
func TestVariadicFunctions(t *testing.T) {
	source := `package main

func variadic(args ...interface{}) {
}
func main() {
	variadic()
	variadic(1)
	variadic(1, 2, 3)
	slice := []interface{}{1, 2}
	variadic(slice...)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String())
}

// TestLargeFile tests a large file with many functions
func TestLargeFile(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("package main\n\n")

	// Generate 100 functions
	for i := 0; i < 100; i++ {
		buf.WriteString(fmt.Sprintf("func function%d() {\n", i))
		buf.WriteString("}\n")
	}

	source := buf.String()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String())
}

// TestDeepNesting tests deeply nested control structures
func TestDeepNesting(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("package main\n\nfunc f() {\n")

	// Create 50 levels of nesting
	depth := 50
	for i := 0; i < depth; i++ {
		buf.WriteString("\tif true {\n")
	}
	buf.WriteString("\t\t_ = 1\n")
	for i := 0; i < depth; i++ {
		buf.WriteString("\t}\n")
	}
	buf.WriteString("}\n")

	source := buf.String()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, 0)
	require.NoError(t, err)

	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err)

	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err)

	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err)

	// Just verify it doesn't crash
	require.NotNil(t, file2)
}
