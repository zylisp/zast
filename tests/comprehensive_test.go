package zast

import (
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

// testRoundTrip tests that Go source can be converted to S-expression and back
func testRoundTrip(t *testing.T, source string) {
	t.Helper()

	// Parse Go source
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err, "failed to parse Go source")

	// Convert to S-expression
	w := writer.New(fset)
	sexpStr, err := w.WriteFile(file)
	require.NoError(t, err, "failed to write S-expression")

	// Parse S-expression
	p := sexp.NewParser(sexpStr)
	sexpNode, err := p.Parse()
	require.NoError(t, err, "failed to parse S-expression")

	// Build back to AST
	b := builder.New()
	file2, err := b.BuildFile(sexpNode)
	require.NoError(t, err, "failed to build AST from S-expression")

	// Compare source output
	var buf1, buf2 strings.Builder
	err = printer.Fprint(&buf1, fset, file)
	require.NoError(t, err)

	fset2 := token.NewFileSet()
	err = printer.Fprint(&buf2, fset2, file2)
	require.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String(), "round-trip mismatch")
}

// TestBinaryExpressions tests all binary operators
func TestBinaryExpressions(t *testing.T) {
	source := `package main

func f() {
	_ = 1 + 2
	_ = 1 - 2
	_ = 1 * 2
	_ = 1 / 2
	_ = 1 % 2
	_ = 1 & 2
	_ = 1 | 2
	_ = 1 ^ 2
	_ = 1 &^ 2
	_ = 1 << 2
	_ = 1 >> 2
	_ = 1 == 2
	_ = 1 != 2
	_ = 1 < 2
	_ = 1 <= 2
	_ = 1 > 2
	_ = 1 >= 2
	_ = true && false
	_ = true || false
}
`
	testRoundTrip(t, source)
}

// TestUnaryExpressions tests all unary operators
func TestUnaryExpressions(t *testing.T) {
	source := `package main

func f() {
	x := 1
	_ = +x
	_ = -x
	_ = !true
	_ = ^x
	_ = &x
	_ = *(&x)
	_ = <-make(chan int)
}
`
	testRoundTrip(t, source)
}

// TestParenExpressions tests parenthesized expressions
func TestParenExpressions(t *testing.T) {
	source := `package main

func f() {
	_ = (1 + 2) * 3
	_ = ((1))
}
`
	testRoundTrip(t, source)
}

// TestIndexExpressions tests index and slice expressions
func TestIndexExpressions(t *testing.T) {
	source := `package main

func f() {
	a := []int{1, 2, 3}
	_ = a[0]
	_ = a[1:2]
	_ = a[:2]
	_ = a[1:]
	_ = a[:]
	_ = a[1:2:3]
}
`
	testRoundTrip(t, source)
}

// TestTypeAssertions tests type assertions
func TestTypeAssertions(t *testing.T) {
	source := `package main

func f() {
	var x interface{} = 1
	_ = x.(int)
	_, _ = x.(int)
}
`
	testRoundTrip(t, source)
}

// TestCompositeLiterals tests composite literals
func TestCompositeLiterals(t *testing.T) {
	source := `package main

func f() {
	_ = []int{1, 2, 3}
	_ = []int{1, 2, 3}
	_ = map[string]int{"a": 1, "b": 2}
	_ = struct{ x int }{x: 1}
}
`
	testRoundTrip(t, source)
}

// TestFunctionLiterals tests function literals
func TestFunctionLiterals(t *testing.T) {
	source := `package main

func f() {
	_ = func() {
	}
	_ = func(x int) int {
		return x
	}
}
`
	testRoundTrip(t, source)
}

// TestEllipsis tests ellipsis in function calls and types
func TestEllipsis(t *testing.T) {
	source := `package main

func variadic(args ...int) {
	slice := []int{1, 2, 3}
	variadic(slice...)
}
`
	testRoundTrip(t, source)
}

// TestStarExpressions tests pointer types
func TestStarExpressions(t *testing.T) {
	source := `package main

func f() {
	var x *int
	var y **int
	_ = x
	_ = y
}
`
	testRoundTrip(t, source)
}

// TestKeyValueExpressions tests key-value pairs in composite literals
func TestKeyValueExpressions(t *testing.T) {
	source := `package main

func f() {
	_ = map[string]int{"a": 1}
	_ = struct{ x int }{x: 1}
}
`
	testRoundTrip(t, source)
}

// TestReturnStatements tests return statements
func TestReturnStatements(t *testing.T) {
	source := `package main

func f() int {
	return 1
}
func g() (int, int) {
	return 1, 2
}
func h() {
	return
}
`
	testRoundTrip(t, source)
}

// TestAssignStatements tests assignment statements
func TestAssignStatements(t *testing.T) {
	source := `package main

func f() {
	x := 1
	y := 2
	x = y
	x, y = y, x
	x += 1
	x -= 1
	x *= 2
	x /= 2
	x %= 2
	x &= 1
	x |= 1
	x ^= 1
	x <<= 1
	x >>= 1
	x &^= 1
}
`
	testRoundTrip(t, source)
}

// TestIncDecStatements tests increment and decrement statements
func TestIncDecStatements(t *testing.T) {
	source := `package main

func f() {
	x := 1
	x++
	x--
}
`
	testRoundTrip(t, source)
}

// TestBranchStatements tests break, continue, goto, fallthrough
func TestBranchStatements(t *testing.T) {
	source := `package main

func f() {
	for {
		break
	}
	for {
		continue
	}
	goto label
label:
	switch {
	case true:
		fallthrough
	default:
	}
}
`
	testRoundTrip(t, source)
}

// TestDeferAndGoStatements tests defer and go statements
func TestDeferAndGoStatements(t *testing.T) {
	source := `package main

func f() {
	defer func() {
	}()
	go func() {
	}()
}
`
	testRoundTrip(t, source)
}

// TestSendStatements tests channel send statements
func TestSendStatements(t *testing.T) {
	source := `package main

func f() {
	ch := make(chan int)
	ch <- 1
}
`
	testRoundTrip(t, source)
}

// TestLabeledStatements tests labeled statements
func TestLabeledStatements(t *testing.T) {
	source := `package main

func f() {
label:
	_ = 1
	goto label
}
`
	testRoundTrip(t, source)
}

// TestIfStatements tests if statements
func TestIfStatements(t *testing.T) {
	source := `package main

func f() {
	if true {
	}
	if x := 1; x > 0 {
	}
	if true {
	} else {
	}
	if true {
	} else if false {
	} else {
	}
}
`
	testRoundTrip(t, source)
}

// TestForStatements tests for loops
func TestForStatements(t *testing.T) {
	source := `package main

func f() {
	for {
	}
	for i := 0; i < 10; i++ {
	}
	for i < 10 {
	}
}
`
	testRoundTrip(t, source)
}

// TestRangeStatements tests range loops
func TestRangeStatements(t *testing.T) {
	source := `package main

func f() {
	a := []int{1, 2, 3}
	for range a {
	}
	for i := range a {
		_ = i
	}
	for i, v := range a {
		_, _ = i, v
	}
}
`
	testRoundTrip(t, source)
}

// TestSwitchStatements tests switch statements
func TestSwitchStatements(t *testing.T) {
	source := `package main

func f() {
	x := 1
	switch {
	case x == 1:
	case x == 2:
	default:
	}
	switch x {
	case 1:
	case 2, 3:
	default:
	}
	switch y := x; y {
	case 1:
	}
}
`
	testRoundTrip(t, source)
}

// TestTypeSwitchStatements tests type switch statements
func TestTypeSwitchStatements(t *testing.T) {
	source := `package main

func f() {
	var x interface{} = 1
	switch x.(type) {
	case int:
	case string:
	default:
	}
	switch y := x.(type) {
	case int:
		_ = y
	}
}
`
	testRoundTrip(t, source)
}

// TestSelectStatements tests select statements
func TestSelectStatements(t *testing.T) {
	source := `package main

func f() {
	ch := make(chan int)
	select {
	case x := <-ch:
		_ = x
	case ch <- 1:
	default:
	}
}
`
	testRoundTrip(t, source)
}

// TestDeclStatements tests declaration statements in blocks
func TestDeclStatements(t *testing.T) {
	source := `package main

func f() {
	const x = 1
	var y int
	type T int
}
`
	testRoundTrip(t, source)
}

// TestEmptyStatements tests empty statements
func TestEmptyStatements(t *testing.T) {
	source := `package main

func f() {
}
`
	testRoundTrip(t, source)
}

// TestArrayTypes tests array types
func TestArrayTypes(t *testing.T) {
	source := `package main

var (
	x [10]int
	y [5][10]int
	z []int
)
`
	testRoundTrip(t, source)
}

// TestMapTypes tests map types
func TestMapTypes(t *testing.T) {
	source := `package main

var (
	x map[string]int
	y map[int]string
	z map[string]map[int]bool
)
`
	testRoundTrip(t, source)
}

// TestChannelTypes tests channel types
func TestChannelTypes(t *testing.T) {
	source := `package main

var (
	x chan int
	y <-chan int
	z chan<- int
)
`
	testRoundTrip(t, source)
}

// TestStructTypes tests struct types
func TestStructTypes(t *testing.T) {
	source := `package main

type S struct {
	x	int
	y	string
}
type Embedded struct {
	S
	z	bool
}
`
	testRoundTrip(t, source)
}

// TestInterfaceTypes tests interface types
func TestInterfaceTypes(t *testing.T) {
	t.Skip("Interface method types not yet supported - need to handle *ast.FuncType in interface Field.Type")
	source := `package main

type I interface {
	Method() int
}

type Empty interface{}
`
	testRoundTrip(t, source)
}

// TestValueSpecs tests value specifications
func TestValueSpecs(t *testing.T) {
	source := `package main

var (
	x int
	y int = 1
	z, w int
	a, b = 1, 2
)

const (
	c = 1
	d = 2
)
`
	testRoundTrip(t, source)
}

// TestTypeSpecs tests type specifications
func TestTypeSpecs(t *testing.T) {
	source := `package main

type (
	T1 int
	T2 string
)
`
	testRoundTrip(t, source)
}

// TestComments tests comment preservation
func TestComments(t *testing.T) {
	t.Skip("Comment preservation needs work - comments are moved to end of file")
	source := `package main

// This is a comment
func f() {
	// Another comment
	x := 1
	_ = x
}
`
	testRoundTrip(t, source)
}

// TestEmptyInterface tests empty interface type
func TestEmptyInterface(t *testing.T) {
	source := `package main

type Empty interface{}
`
	testRoundTrip(t, source)
}

// TestIndexListExpr tests generic index expressions (Go 1.18+)
func TestIndexListExpr(t *testing.T) {
	t.Skip("Generic type parameters not yet supported")
	source := `package main

func Generic[T any, U any](x T, y U) {
}

func f() {
	Generic[int, string](1, "hello")
}
`
	testRoundTrip(t, source)
}

// TestSliceExpressions tests all slice expression variants
func TestSliceExpressions(t *testing.T) {
	source := `package main

func f() {
	a := []int{1, 2, 3, 4, 5}
	_ = a[1:3]
	_ = a[:3]
	_ = a[2:]
	_ = a[:]
	_ = a[1:3:4]
}
`
	testRoundTrip(t, source)
}
