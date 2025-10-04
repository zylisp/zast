package zast

import (
	"os"
	"path/filepath"

	"zylisp/zast/builder"
	"zylisp/zast/sexp"
	"zylisp/zast/writer"
)

// OutputConfig holds configuration for output directories and files
type OutputConfig struct {
	// BuildDir is the directory for build artifacts (like JVM's target/ or Clojure's .cpcache/)
	// Default: "target"
	BuildDir string

	// ASTBuildDir is the directory for AST build artifacts
	// Default: "target/ast"
	ASTBuildDir string
	ASTFileExt  string // Default: ".zast"

	// UseTempDir forces use of system temp directory instead of BuildDir
	// Default: false
	UseTempDir bool

	// KeepFiles preserves build artifacts after execution
	// Default: true
	KeepFiles bool
}

// DefaultOutputConfig returns the default output configuration
func DefaultOutputConfig() *OutputConfig {
	return &OutputConfig{
		BuildDir:    "target",
		ASTBuildDir: filepath.Join("target", "ast"),
		ASTFileExt:  ".zast",
		UseTempDir:  false,
		KeepFiles:   true,
	}
}

// GetBuildDir returns the build directory path, creating it if necessary
func (c *OutputConfig) GetBuildDir() (string, error) {
	if c.UseTempDir {
		return os.MkdirTemp("", "zast-*")
	}

	// Expand to absolute path
	buildDir, err := filepath.Abs(c.BuildDir)
	if err != nil {
		return "", err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", err
	}

	return buildDir, nil
}

// GetASTBuildDir returns the AST build directory path, creating it if necessary
func (c *OutputConfig) GetASTBuildDir() (string, error) {
	if c.UseTempDir {
		return os.MkdirTemp("", "zast-ast-*")
	}

	// Expand to absolute path
	astBuildDir, err := filepath.Abs(c.ASTBuildDir)
	if err != nil {
		return "", err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(astBuildDir, 0755); err != nil {
		return "", err
	}

	return astBuildDir, nil
}

// Config holds configuration for all AST conversion components
type Config struct {
	Printer *sexp.PrettyPrintConfig
	Builder *builder.Config
	Writer  *writer.Config
	Output  *OutputConfig
}

// DefaultConfig returns the default configuration for all components
func DefaultConfig() *Config {
	return &Config{
		Printer: sexp.DefaultPrettyPrintConfig(),
		Builder: builder.DefaultConfig(),
		Writer:  writer.DefaultConfig(),
		Output:  DefaultOutputConfig(),
	}
}
