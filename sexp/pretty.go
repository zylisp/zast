package sexp

import (
	"io"
	"strings"
)

// PrettyPrintConfig holds configuration for the pretty printer
type PrettyPrintConfig struct {
	IndentWidth   int  // Spaces per indent level (default: 2)
	MaxLineWidth  int  // Target maximum line width (default: 80)
	AlignKeywords bool // Align keyword-value pairs (default: true)
	CompactSmall  bool // Keep small lists on one line (default: true)
	CompactLimit  int  // Max chars for compact lists (default: 60)

	// Blank lines between top-level forms
	BlankLinesBetweenForms int // default: 0

	// Use tabs instead of spaces for indentation
	UseTabs bool // default: false
}

// DefaultPrettyPrintConfig returns the default configuration
func DefaultPrettyPrintConfig() *PrettyPrintConfig {
	return &PrettyPrintConfig{
		IndentWidth:            2,
		MaxLineWidth:           80,
		AlignKeywords:          true,
		CompactSmall:           true,
		CompactLimit:           60,
		BlankLinesBetweenForms: 0,
		UseTabs:                false,
	}
}

// FormStyle defines how a form should be formatted
type FormStyle int

const (
	StyleDefault FormStyle = iota
	StyleKeywordPairs  // Format as :key value pairs with alignment
	StyleCompact       // Try to keep on one line
	StyleBody          // Special body formatting (BlockStmt, etc.)
	StyleList          // Simple list formatting
)

// Form style mappings for different node types
var formStyles = map[string]FormStyle{
	// Top-level structures
	"Program":  StyleKeywordPairs,
	"File":     StyleKeywordPairs,
	"FileSet":  StyleKeywordPairs,
	"FileInfo": StyleKeywordPairs,

	// Declarations
	"FuncDecl": StyleKeywordPairs,
	"GenDecl":  StyleKeywordPairs,

	// Types
	"FuncType":  StyleKeywordPairs,
	"FieldList": StyleKeywordPairs,
	"Field":     StyleKeywordPairs,

	// Statements
	"BlockStmt": StyleBody,
	"ExprStmt":  StyleKeywordPairs,

	// Expressions
	"CallExpr":     StyleKeywordPairs,
	"SelectorExpr": StyleCompact,
	"Ident":        StyleCompact,
	"BasicLit":     StyleCompact,

	// Specs
	"ImportSpec": StyleKeywordPairs,

	// Metadata
	"CommentGroup": StyleList,
	"Comment":      StyleCompact,
	"Scope":        StyleKeywordPairs,
	"Object":       StyleKeywordPairs,
}

func getFormStyle(nodeName string) FormStyle {
	if style, ok := formStyles[nodeName]; ok {
		return style
	}
	return StyleDefault
}

// PrettyPrinter formats S-expressions with intelligent indentation and alignment
type PrettyPrinter struct {
	buf           strings.Builder
	indentWidth   int
	maxLineWidth  int
	currentColumn int
	config        *PrettyPrintConfig
}

// NewPrettyPrinter creates a new pretty printer with default config
func NewPrettyPrinter() *PrettyPrinter {
	return NewPrettyPrinterWithConfig(DefaultPrettyPrintConfig())
}

// NewPrettyPrinterWithConfig creates a printer with custom config
func NewPrettyPrinterWithConfig(config *PrettyPrintConfig) *PrettyPrinter {
	return &PrettyPrinter{
		indentWidth:  config.IndentWidth,
		maxLineWidth: config.MaxLineWidth,
		config:       config,
	}
}

// Format formats an S-expression and returns the result
func (p *PrettyPrinter) Format(sexp SExp) string {
	p.buf.Reset()
	p.currentColumn = 0
	p.format(sexp, 0)
	return p.buf.String()
}

// FormatToWriter formats an S-expression to an io.Writer
func (p *PrettyPrinter) FormatToWriter(sexp SExp, w io.Writer) error {
	result := p.Format(sexp)
	_, err := w.Write([]byte(result))
	return err
}

// Main formatting dispatcher
func (p *PrettyPrinter) format(sexp SExp, depth int) {
	switch s := sexp.(type) {
	case *Symbol:
		p.formatSymbol(s)
	case *Keyword:
		p.formatKeyword(s)
	case *String:
		p.formatString(s)
	case *Number:
		p.formatNumber(s)
	case *Nil:
		p.formatNil(s)
	case *List:
		p.formatList(s, depth)
	}
}

// Atomic value formatting
func (p *PrettyPrinter) formatSymbol(s *Symbol) {
	p.write(s.Value)
}

func (p *PrettyPrinter) formatKeyword(k *Keyword) {
	p.write(":")
	p.write(k.Name)
}

func (p *PrettyPrinter) formatString(s *String) {
	p.write(`"`)
	p.write(s.Value) // Already escaped
	p.write(`"`)
}

func (p *PrettyPrinter) formatNumber(n *Number) {
	p.write(n.Value)
}

func (p *PrettyPrinter) formatNil(n *Nil) {
	p.write("nil")
}

// List formatting
func (p *PrettyPrinter) formatList(list *List, depth int) {
	if len(list.Elements) == 0 {
		p.write("()")
		return
	}

	// Get the node type (first element)
	var nodeName string
	if sym, ok := list.Elements[0].(*Symbol); ok {
		nodeName = sym.Value
	}

	// Determine formatting style
	style := getFormStyle(nodeName)

	// Try compact formatting first if enabled
	if p.config.CompactSmall && p.shouldBeCompact(list, style) {
		if p.tryCompactFormat(list, depth) {
			return
		}
	}

	// Otherwise use style-specific formatting
	switch style {
	case StyleKeywordPairs:
		p.formatKeywordPairs(list, depth)
	case StyleCompact:
		p.formatCompact(list, depth)
	case StyleBody:
		p.formatBody(list, depth)
	case StyleList:
		p.formatSimpleList(list, depth)
	default:
		p.formatDefault(list, depth)
	}
}

// Compact formatting helpers
func (p *PrettyPrinter) shouldBeCompact(list *List, style FormStyle) bool {
	// Only certain styles can be compact
	if style != StyleCompact && len(list.Elements) > 4 {
		return false
	}

	// Check estimated length
	est := p.estimateLength(list)
	return est < p.config.CompactLimit
}

func (p *PrettyPrinter) tryCompactFormat(list *List, depth int) bool {
	// Save current state
	savedBuf := p.buf.String()
	savedCol := p.currentColumn

	// Try compact format
	p.write("(")
	for i, elem := range list.Elements {
		if i > 0 {
			p.write(" ")
		}
		p.format(elem, depth)

		// If we exceeded line width, abort
		if p.currentColumn > p.config.MaxLineWidth {
			// Restore state
			p.buf.Reset()
			p.buf.WriteString(savedBuf)
			p.currentColumn = savedCol
			return false
		}
	}
	p.write(")")
	return true
}

func (p *PrettyPrinter) estimateLength(sexp SExp) int {
	switch s := sexp.(type) {
	case *Symbol:
		return len(s.Value)
	case *Keyword:
		return len(s.Name) + 1
	case *String:
		return len(s.Value) + 2
	case *Number:
		return len(s.Value)
	case *Nil:
		return 3
	case *List:
		total := 2 // for parens
		for i, elem := range s.Elements {
			if i > 0 {
				total += 1 // space
			}
			total += p.estimateLength(elem)
		}
		return total
	}
	return 0
}

// Compact style (force single line)
func (p *PrettyPrinter) formatCompact(list *List, depth int) {
	p.write("(")
	for i, elem := range list.Elements {
		if i > 0 {
			p.write(" ")
		}
		p.format(elem, depth)
	}
	p.write(")")
}

// Keyword-pair formatting with alignment
type keywordPair struct {
	keyword SExp
	value   SExp
}

func (p *PrettyPrinter) formatKeywordPairs(list *List, depth int) {
	if len(list.Elements) == 0 {
		p.write("()")
		return
	}

	p.write("(")

	// First element (node type) on same line
	p.format(list.Elements[0], depth)

	// Collect keyword-value pairs and find max keyword width
	pairs := p.extractKeywordPairs(list.Elements[1:])
	maxKeywordWidth := p.findMaxKeywordWidth(pairs)

	// Format each pair with alignment
	for _, pair := range pairs {
		p.newline()
		p.indent(depth + 1)

		// Write keyword
		p.format(pair.keyword, depth+1)

		// Pad to alignment column
		if p.config.AlignKeywords && maxKeywordWidth > 0 {
			keywordWidth := p.estimateLength(pair.keyword)
			padding := maxKeywordWidth - keywordWidth
			p.writeSpaces(padding + 1)
		} else {
			p.write(" ")
		}

		// Write value
		p.format(pair.value, depth+1)
	}

	p.write(")")
}

func (p *PrettyPrinter) extractKeywordPairs(elements []SExp) []keywordPair {
	var pairs []keywordPair

	for i := 0; i < len(elements); i += 2 {
		if i+1 >= len(elements) {
			break
		}

		pairs = append(pairs, keywordPair{
			keyword: elements[i],
			value:   elements[i+1],
		})
	}

	return pairs
}

func (p *PrettyPrinter) findMaxKeywordWidth(pairs []keywordPair) int {
	max := 0
	for _, pair := range pairs {
		width := p.estimateLength(pair.keyword)
		if width > max {
			max = width
		}
	}
	return max
}

// Body formatting (special for BlockStmt, etc.)
func (p *PrettyPrinter) formatBody(list *List, depth int) {
	if len(list.Elements) == 0 {
		p.write("()")
		return
	}

	p.write("(")

	// First element (node type)
	p.format(list.Elements[0], depth)

	// Process keyword-value pairs, but :list gets special treatment
	i := 1
	for i < len(list.Elements) {
		if i+1 >= len(list.Elements) {
			break
		}

		keyword := list.Elements[i]
		value := list.Elements[i+1]

		// Check if this is the :list field
		if kw, ok := keyword.(*Keyword); ok && kw.Name == "list" {
			// Format list body with extra indentation
			p.newline()
			p.indent(depth + 1)
			p.format(keyword, depth+1)
			p.write(" ")

			if valueList, ok := value.(*List); ok {
				p.formatListBody(valueList, depth+2)
			} else {
				p.format(value, depth+1)
			}
		} else {
			// Regular keyword-value pair
			p.newline()
			p.indent(depth + 1)
			p.format(keyword, depth+1)
			p.write(" ")
			p.format(value, depth+1)
		}

		i += 2
	}

	p.write(")")
}

func (p *PrettyPrinter) formatListBody(list *List, depth int) {
	if len(list.Elements) == 0 {
		p.write("()")
		return
	}

	p.write("(")
	for i, elem := range list.Elements {
		if i > 0 {
			// Add blank lines between forms if configured
			for j := 0; j < p.config.BlankLinesBetweenForms; j++ {
				p.newline()
			}
			p.newline()
			p.indent(depth)
		} else {
			p.newline()
			p.indent(depth)
		}
		p.format(elem, depth)
	}
	p.write(")")
}

// Simple list formatting
func (p *PrettyPrinter) formatSimpleList(list *List, depth int) {
	p.write("(")

	for i, elem := range list.Elements {
		if i > 0 {
			p.newline()
			p.indent(depth + 1)
		}
		p.format(elem, depth+1)
	}

	p.write(")")
}

// Default formatting (fallback)
func (p *PrettyPrinter) formatDefault(list *List, depth int) {
	p.write("(")

	for i, elem := range list.Elements {
		if i > 0 {
			if p.shouldBreakBefore(elem) {
				p.newline()
				p.indent(depth + 1)
			} else {
				p.write(" ")
			}
		}
		p.format(elem, depth+1)
	}

	p.write(")")
}

func (p *PrettyPrinter) shouldBreakBefore(sexp SExp) bool {
	// Break before lists
	_, isList := sexp.(*List)
	return isList
}

// Helper methods
func (p *PrettyPrinter) write(s string) {
	p.buf.WriteString(s)
	p.currentColumn += len(s)
}

func (p *PrettyPrinter) writeSpaces(n int) {
	for i := 0; i < n; i++ {
		p.buf.WriteString(" ")
		p.currentColumn++
	}
}

func (p *PrettyPrinter) newline() {
	p.buf.WriteString("\n")
	p.currentColumn = 0
}

func (p *PrettyPrinter) indent(depth int) {
	if p.config.UseTabs {
		for i := 0; i < depth; i++ {
			p.buf.WriteString("\t")
			p.currentColumn += p.config.IndentWidth // Approximate tab width
		}
	} else {
		spaces := depth * p.config.IndentWidth
		p.writeSpaces(spaces)
	}
}
