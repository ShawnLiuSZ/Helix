package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReviewTool(t *testing.T) {
	t.Run("NewReviewTool", func(t *testing.T) {
		review := NewReviewTool()
		if review == nil {
			t.Fatal("expected non-nil review tool")
		}
		if !review.IsEnabled() {
			t.Error("expected review tool to be enabled by default")
		}
	})

	t.Run("AddPending", func(t *testing.T) {
		review := NewReviewTool()

		id := review.AddPending(PendingEdit{
			File:    "test.go",
			OldText: "old",
			NewText: "new",
		})

		if id != 1 {
			t.Errorf("expected id 1, got %d", id)
		}

		pending := review.GetPending()
		if len(pending) != 1 {
			t.Errorf("expected 1 pending edit, got %d", len(pending))
		}
	})

	t.Run("ClearPending", func(t *testing.T) {
		review := NewReviewTool()

		review.AddPending(PendingEdit{File: "test.go"})
		review.ClearPending()

		pending := review.GetPending()
		if len(pending) != 0 {
			t.Errorf("expected 0 pending edits, got %d", len(pending))
		}
	})

	t.Run("Preview", func(t *testing.T) {
		review := NewReviewTool()

		edit := PendingEdit{
			File:    "test.go",
			OldText: "func old() {}",
			NewText: "func new() {}",
		}

		preview := review.Preview(edit)
		if preview == "" {
			t.Error("expected non-empty preview")
		}
		if !containsStr(preview, "test.go") {
			t.Error("expected preview to contain file name")
		}
	})

	t.Run("PreviewAll_Empty", func(t *testing.T) {
		review := NewReviewTool()

		preview := review.PreviewAll()
		if !containsStr(preview, "没有待确认的编辑") {
			t.Error("expected 'no pending edits' message")
		}
	})

	t.Run("PreviewAll_WithEdits", func(t *testing.T) {
		review := NewReviewTool()

		review.AddPending(PendingEdit{File: "test.go"})
		review.AddPending(PendingEdit{File: "main.go"})

		preview := review.PreviewAll()
		if !containsStr(preview, "2 个") {
			t.Error("expected preview to show 2 edits")
		}
	})
}

func TestReviewToolApply(t *testing.T) {
	t.Run("Apply", func(t *testing.T) {
		// 创建临时文件
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
			t.Fatal(err)
		}

		review := NewReviewTool()
		id := review.AddPending(PendingEdit{
			File:    testFile,
			OldText: "world",
			NewText: "Go",
		})

		if err := review.Apply(id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 验证文件内容
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != "hello Go" {
			t.Errorf("expected 'hello Go', got %q", string(content))
		}
	})

	t.Run("Apply_AlreadyApplied", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("hello world"), 0644)

		review := NewReviewTool()
		id := review.AddPending(PendingEdit{
			File:    testFile,
			OldText: "world",
			NewText: "Go",
		})

		review.Apply(id)
		err := review.Apply(id)
		if err == nil {
			t.Error("expected error for already applied edit")
		}
	})

	t.Run("Apply_NotFound", func(t *testing.T) {
		review := NewReviewTool()
		err := review.Apply(999)
		if err == nil {
			t.Error("expected error for not found edit")
		}
	})
}

func TestReviewToolReject(t *testing.T) {
	t.Run("Reject", func(t *testing.T) {
		review := NewReviewTool()
		id := review.AddPending(PendingEdit{File: "test.go"})

		if err := review.Reject(id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pending := review.GetPending()
		if len(pending) != 0 {
			t.Errorf("expected 0 pending edits, got %d", len(pending))
		}
	})

	t.Run("Reject_NotFound", func(t *testing.T) {
		review := NewReviewTool()
		err := review.Reject(999)
		if err == nil {
			t.Error("expected error for not found edit")
		}
	})

	t.Run("RejectAll", func(t *testing.T) {
		review := NewReviewTool()
		review.AddPending(PendingEdit{File: "test.go"})
		review.AddPending(PendingEdit{File: "main.go"})

		review.RejectAll()

		pending := review.GetPending()
		if len(pending) != 0 {
			t.Errorf("expected 0 pending edits, got %d", len(pending))
		}
	})
}

func TestReviewEditTool(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		review := NewReviewTool()
		tool := NewReviewEditTool(review)
		if tool.Name() != "edit_file" {
			t.Errorf("expected 'edit_file', got %q", tool.Name())
		}
	})

	t.Run("IsReadOnly", func(t *testing.T) {
		review := NewReviewTool()
		tool := NewReviewEditTool(review)
		if tool.IsReadOnly() {
			t.Error("expected edit_file to not be read-only")
		}
	})

	t.Run("Execute_WithPreview", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("hello world"), 0644)

		review := NewReviewTool()
		tool := NewReviewEditTool(review)

		result, err := tool.Execute(context.Background(), map[string]any{
			"path":     testFile,
			"old_text": "world",
			"new_text": "Go",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !containsStr(result.Content, "待确认列表") {
			t.Error("expected result to mention pending list")
		}

		// 验证文件未被修改
		content, _ := os.ReadFile(testFile)
		if string(content) != "hello world" {
			t.Error("expected file to remain unchanged")
		}
	})

	t.Run("Execute_WithoutPreview", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("hello world"), 0644)

		review := NewReviewTool()
		review.SetEnabled(false)
		tool := NewReviewEditTool(review)

		result, err := tool.Execute(context.Background(), map[string]any{
			"path":     testFile,
			"old_text": "world",
			"new_text": "Go",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !containsStr(result.Content, "File edited") {
			t.Error("expected result to mention file edited")
		}

		// 验证文件已被修改
		content, _ := os.ReadFile(testFile)
		if string(content) != "hello Go" {
			t.Errorf("expected 'hello Go', got %q", string(content))
		}
	})
}

func TestReviewToolApplyAll(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试文件
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	os.WriteFile(testFile1, []byte("file1 old"), 0644)
	os.WriteFile(testFile2, []byte("file2 old"), 0644)

	review := NewReviewTool()
	review.AddPending(PendingEdit{File: testFile1, OldText: "old", NewText: "new"})
	review.AddPending(PendingEdit{File: testFile2, OldText: "old", NewText: "new"})

	if err := review.ApplyAll(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证文件内容
	content1, _ := os.ReadFile(testFile1)
	content2, _ := os.ReadFile(testFile2)

	if string(content1) != "file1 new" {
		t.Errorf("expected 'file1 new', got %q", string(content1))
	}
	if string(content2) != "file2 new" {
		t.Errorf("expected 'file2 new', got %q", string(content2))
	}
}

// containsStr 检查字符串是否包含子串
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
