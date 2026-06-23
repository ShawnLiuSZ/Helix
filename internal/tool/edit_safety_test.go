package tool

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestEditFileTool_AmbiguousMatchErrors(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/edit.txt"
	original := "aaa bbb aaa bbb aaa"
	os.WriteFile(path, []byte(original), 0644)

	tl := &EditFileTool{}
	_, err := tl.Execute(context.Background(), map[string]any{
		"path":     path,
		"old_text": "aaa",
		"new_text": "zzz",
	})
	if err == nil {
		t.Fatal("expected error for ambiguous (multi-match) old_text")
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("error should report the match count, got: %v", err)
	}
	// File must be left untouched on an ambiguous edit.
	data, _ := os.ReadFile(path)
	if string(data) != original {
		t.Errorf("file was modified on ambiguous edit: %q", string(data))
	}
}

func TestEditFileTool_ReplaceAll(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/edit.txt"
	os.WriteFile(path, []byte("aaa bbb aaa bbb aaa"), 0644)

	tl := &EditFileTool{}
	res, err := tl.Execute(context.Background(), map[string]any{
		"path":        path,
		"old_text":    "aaa",
		"new_text":    "zzz",
		"replace_all": true,
	})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "zzz bbb zzz bbb zzz" {
		t.Errorf("file content = %q, want all occurrences replaced", string(data))
	}
	if !strings.Contains(res.Content, "3") {
		t.Errorf("result should report replacement count, got %q", res.Content)
	}
}
