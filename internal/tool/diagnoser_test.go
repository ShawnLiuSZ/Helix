package tool

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoDiagnoser_ReportsSyntaxError(t *testing.T) {
	if _, err := exec.LookPath("gofmt"); err != nil {
		t.Skip("gofmt not available")
	}
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.go")
	os.WriteFile(bad, []byte("package x\nfunc {\n"), 0644)

	out := (GoDiagnoser{}).Diagnose(context.Background(), bad)
	if out == "" {
		t.Fatal("expected syntax issues to be reported")
	}
	if !strings.Contains(out, "bad.go") {
		t.Errorf("diagnostic should reference the file, got %q", out)
	}
}

func TestGoDiagnoser_CleanFile(t *testing.T) {
	if _, err := exec.LookPath("gofmt"); err != nil {
		t.Skip("gofmt not available")
	}
	dir := t.TempDir()
	good := filepath.Join(dir, "good.go")
	os.WriteFile(good, []byte("package x\n\nfunc F() int { return 1 }\n"), 0644)

	if out := (GoDiagnoser{}).Diagnose(context.Background(), good); out != "" {
		t.Errorf("clean file should produce no diagnostics, got %q", out)
	}
}

func TestGoDiagnoser_IgnoresNonGo(t *testing.T) {
	dir := t.TempDir()
	txt := filepath.Join(dir, "note.txt")
	os.WriteFile(txt, []byte("func {{{ not go"), 0644)

	if out := (GoDiagnoser{}).Diagnose(context.Background(), txt); out != "" {
		t.Errorf("non-go file should be ignored, got %q", out)
	}
}

func TestEditFileTool_AppendsDiagnostics(t *testing.T) {
	if _, err := exec.LookPath("gofmt"); err != nil {
		t.Skip("gofmt not available")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "x.go")
	os.WriteFile(path, []byte("package x\n\nfunc F() int { return 1 }\n"), 0644)

	tl := &EditFileTool{}
	tl.SetDiagnoser(GoDiagnoser{})
	// Drop the closing brace -> unbalanced block -> syntax error.
	res, err := tl.Execute(context.Background(), map[string]any{
		"path":     path,
		"old_text": "return 1 }",
		"new_text": "return 1",
	})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(res.Content, "Syntax issues") {
		t.Errorf("edit result should surface diagnostics, got %q", res.Content)
	}
}

func TestEditFileTool_NoDiagnoserNoNoise(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.go")
	os.WriteFile(path, []byte("package x\n\nfunc F() int { return 1 }\n"), 0644)

	tl := &EditFileTool{} // no diagnoser set
	res, err := tl.Execute(context.Background(), map[string]any{
		"path":     path,
		"old_text": "return 1",
		"new_text": "return 2",
	})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strings.Contains(res.Content, "Syntax issues") {
		t.Errorf("no diagnoser should mean no diagnostics, got %q", res.Content)
	}
}
