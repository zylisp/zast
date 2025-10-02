package zast

import "zylisp/zast/sexp"

// Config holds configuration for all AST conversion components
type Config struct {
	Printer *sexp.PrettyPrintConfig
	Builder *BuilderConfig
	Writer  *WriterConfig
}

// DefaultConfig returns the default configuration for all components
func DefaultConfig() *Config {
	return &Config{
		Printer: sexp.DefaultPrettyPrintConfig(),
		Builder: DefaultBuilderConfig(),
		Writer:  DefaultWriterConfig(),
	}
}
