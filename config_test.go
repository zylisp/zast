package zast

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultOutputConfig(t *testing.T) {
	config := DefaultOutputConfig()
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.BuildDir != "target" {
		t.Fatalf("expected BuildDir 'target', got %q", config.BuildDir)
	}
	if config.ASTFileExt != ".zast" {
		t.Fatalf("expected ASTFileExt '.zast', got %q", config.ASTFileExt)
	}
	if config.UseTempDir {
		t.Fatal("expected UseTempDir false")
	}
	if !config.KeepFiles {
		t.Fatal("expected KeepFiles true")
	}
}

func TestGetBuildDir(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	config := &OutputConfig{
		BuildDir:   "test-build",
		KeepFiles:  true,
		UseTempDir: false,
	}

	buildDir, err := config.GetBuildDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !filepath.IsAbs(buildDir) {
		t.Fatalf("expected absolute path, got %q", buildDir)
	}

	// Verify directory was created
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		t.Fatalf("expected build dir to be created")
	}
}

func TestGetBuildDirTempDir(t *testing.T) {
	config := &OutputConfig{
		UseTempDir: true,
	}

	buildDir, err := config.GetBuildDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !filepath.IsAbs(buildDir) {
		t.Fatalf("expected absolute path, got %q", buildDir)
	}

	// Clean up temp dir
	os.RemoveAll(buildDir)
}

func TestGetASTBuildDir(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	config := &OutputConfig{
		ASTBuildDir: "test-ast",
		KeepFiles:   true,
		UseTempDir:  false,
	}

	astBuildDir, err := config.GetASTBuildDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !filepath.IsAbs(astBuildDir) {
		t.Fatalf("expected absolute path, got %q", astBuildDir)
	}

	// Verify directory was created
	if _, err := os.Stat(astBuildDir); os.IsNotExist(err) {
		t.Fatalf("expected AST build dir to be created")
	}
}

func TestGetASTBuildDirTempDir(t *testing.T) {
	config := &OutputConfig{
		UseTempDir: true,
	}

	astBuildDir, err := config.GetASTBuildDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !filepath.IsAbs(astBuildDir) {
		t.Fatalf("expected absolute path, got %q", astBuildDir)
	}

	// Clean up temp dir
	os.RemoveAll(astBuildDir)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Printer == nil {
		t.Fatal("expected non-nil Printer")
	}
	if config.Builder == nil {
		t.Fatal("expected non-nil Builder")
	}
	if config.Writer == nil {
		t.Fatal("expected non-nil Writer")
	}
	if config.Output == nil {
		t.Fatal("expected non-nil Output")
	}
}
