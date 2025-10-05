package zast

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"zylisp/zast/builder"
	"zylisp/zast/sexp"
	"zylisp/zast/writer"
)

// BenchmarkWriteSmallFile benchmarks writing a small file
func BenchmarkWriteSmallFile(b *testing.B) {
	source := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}
`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)
	w := writer.New(fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = w.WriteFile(file)
	}
}

// BenchmarkWriteMediumFile benchmarks writing a medium file (~500 lines)
func BenchmarkWriteMediumFile(b *testing.B) {
	source := generateMediumFile()
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)
	w := writer.New(fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = w.WriteFile(file)
	}
}

// BenchmarkWriteLargeFile benchmarks writing a large file (~5000 lines)
func BenchmarkWriteLargeFile(b *testing.B) {
	source := generateLargeFile()
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)
	w := writer.New(fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = w.WriteFile(file)
	}
}

// BenchmarkBuildSmallFile benchmarks building a small file from S-expression
func BenchmarkBuildSmallFile(b *testing.B) {
	source := `package main

func main() {
}
`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)
	w := writer.New(fset)
	sexpText, _ := w.WriteFile(file)

	p := sexp.NewParser(sexpText)
	sexpNode, _ := p.Parse()
	bldr := builder.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bldr.BuildFile(sexpNode)
	}
}

// BenchmarkBuildMediumFile benchmarks building a medium file from S-expression
func BenchmarkBuildMediumFile(b *testing.B) {
	source := generateMediumFile()
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)
	w := writer.New(fset)
	sexpText, _ := w.WriteFile(file)

	p := sexp.NewParser(sexpText)
	sexpNode, _ := p.Parse()
	bldr := builder.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bldr.BuildFile(sexpNode)
	}
}

// BenchmarkRoundTripSmall benchmarks a complete round-trip for a small file
func BenchmarkRoundTripSmall(b *testing.B) {
	source := `package main

func main() {
}
`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := writer.New(fset)
		sexpText, _ := w.WriteFile(file)

		p := sexp.NewParser(sexpText)
		sexpNode, _ := p.Parse()

		bldr := builder.New()
		_, _ = bldr.BuildFile(sexpNode)
	}
}

// BenchmarkRoundTripMedium benchmarks a complete round-trip for a medium file
func BenchmarkRoundTripMedium(b *testing.B) {
	source := generateMediumFile()
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := writer.New(fset)
		sexpText, _ := w.WriteFile(file)

		p := sexp.NewParser(sexpText)
		sexpNode, _ := p.Parse()

		bldr := builder.New()
		_, _ = bldr.BuildFile(sexpNode)
	}
}

// BenchmarkWriteAllocations benchmarks memory allocations during writing
func BenchmarkWriteAllocations(b *testing.B) {
	source := `package main

func main() {
	x := 1 + 2
}
`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)
	w := writer.New(fset)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = w.WriteFile(file)
	}
}

// BenchmarkBuildAllocations benchmarks memory allocations during building
func BenchmarkBuildAllocations(b *testing.B) {
	source := `package main

func main() {
	x := 1 + 2
}
`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", source, 0)
	w := writer.New(fset)
	sexpText, _ := w.WriteFile(file)

	p := sexp.NewParser(sexpText)
	sexpNode, _ := p.Parse()
	bldr := builder.New()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bldr.BuildFile(sexpNode)
	}
}

// Helper functions

func generateMediumFile() string {
	var buf bytes.Buffer
	buf.WriteString("package main\n\n")
	for i := 0; i < 50; i++ {
		buf.WriteString(fmt.Sprintf(`func function%d(x, y int) int {
	if x > y {
		return x
	}
	return y
}
`, i))
	}
	return buf.String()
}

func generateLargeFile() string {
	var buf bytes.Buffer
	buf.WriteString("package main\n\n")
	for i := 0; i < 500; i++ {
		buf.WriteString(fmt.Sprintf(`func function%d(x, y int) int {
	result := x + y
	for j := 0; j < 10; j++ {
		result *= 2
	}
	return result
}
`, i))
	}
	return buf.String()
}
