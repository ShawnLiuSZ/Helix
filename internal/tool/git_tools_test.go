package tool

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestGitStatusTool_Clean(t *testing.T) {
	dir := setupGitRepo(t)
	tl := &GitStatusTool{}
	tl.SetRoot(dir)

	result, err := tl.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if result.Content != "Working tree clean" {
		t.Errorf("Content = %q, want %q", result.Content, "Working tree clean")
	}
}

func TestGitStatusTool_Modified(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("changed"), 0644)

	tl := &GitStatusTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(result.Content, "Modified") {
		t.Errorf("expected Modified in output, got: %s", result.Content)
	}
}

func TestGitStatusTool_Untracked(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644)

	tl := &GitStatusTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(result.Content, "Untracked") {
		t.Errorf("expected Untracked in output, got: %s", result.Content)
	}
}

func TestGitStatusTool_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	tl := &GitStatusTool{}
	tl.SetRoot(dir)
	_, err := tl.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error = %v, want 'not a git repository'", err)
	}
}

func TestGitStatusTool_Schema(t *testing.T) {
	tl := &GitStatusTool{}
	if tl.Name() != "git_status" {
		t.Errorf("Name() = %q", tl.Name())
	}
	if !tl.IsReadOnly() {
		t.Error("expected IsReadOnly() = true")
	}
	s := tl.Schema()
	if s.Type != "object" {
		t.Errorf("Schema.Type = %q", s.Type)
	}
}

func TestGitDiffTool_NoChanges(t *testing.T) {
	dir := setupGitRepo(t)
	tl := &GitDiffTool{}
	tl.SetRoot(dir)

	result, err := tl.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if result.Content != "No changes" {
		t.Errorf("Content = %q, want %q", result.Content, "No changes")
	}
}

func TestGitDiffTool_Unstaged(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("changed"), 0644)

	tl := &GitDiffTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(result.Content, "changed") {
		t.Errorf("expected diff content, got: %s", result.Content)
	}
}

func TestGitDiffTool_Staged(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("changed"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()

	tl := &GitDiffTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{"staged": true})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(result.Content, "changed") {
		t.Errorf("expected staged diff content, got: %s", result.Content)
	}
}

func TestGitDiffTool_PathFilter(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("aa"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("bb"), 0644)

	tl := &GitDiffTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{"path": "a.txt"})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strings.Contains(result.Content, "bb") {
		t.Error("expected only a.txt diff, but got b.txt content")
	}
}

func TestGitDiffTool_MaxLines(t *testing.T) {
	dir := setupGitRepo(t)
	var committed strings.Builder
	for i := 0; i < 200; i++ {
		committed.WriteString("original\n")
	}
	os.WriteFile(filepath.Join(dir, "big.txt"), []byte(committed.String()), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	var modified strings.Builder
	for i := 0; i < 200; i++ {
		modified.WriteString("changed\n")
	}
	os.WriteFile(filepath.Join(dir, "big.txt"), []byte(modified.String()), 0644)

	tl := &GitDiffTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{"max_lines": float64(10)})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(result.Content, "truncated at 10 lines") {
		t.Errorf("expected truncation message, got: %s", result.Content)
	}
}

func TestGitDiffTool_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	tl := &GitDiffTool{}
	tl.SetRoot(dir)
	_, err := tl.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGitDiffTool_Schema(t *testing.T) {
	tl := &GitDiffTool{}
	if tl.Name() != "git_diff" {
		t.Errorf("Name() = %q", tl.Name())
	}
	if !tl.IsReadOnly() {
		t.Error("expected IsReadOnly() = true")
	}
}

func TestGitLogTool_Empty(t *testing.T) {
	dir := setupGitRepo(t)
	tl := &GitLogTool{}
	tl.SetRoot(dir)

	result, err := tl.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if result.Content != "No commits found" {
		t.Errorf("Content = %q, want %q", result.Content, "No commits found")
	}
}

func TestGitLogTool_WithCommits(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "first commit").Run()

	tl := &GitLogTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(result.Content, "first commit") {
		t.Errorf("expected commit message in output, got: %s", result.Content)
	}
}

func TestGitLogTool_Count(t *testing.T) {
	dir := setupGitRepo(t)
	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(dir, "a.txt"), []byte(strings.Repeat("x", i+1)), 0644)
		exec.Command("git", "-C", dir, "add", "-A").Run()
		exec.Command("git", "-C", dir, "commit", "-m", "commit").Run()
	}

	tl := &GitLogTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{"count": float64(3)})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %s", len(lines), result.Content)
	}
}

func TestGitLogTool_PathFilter(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "add a").Run()
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)
	exec.Command("git", "-C", dir, "add", "-A").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "add b").Run()

	tl := &GitLogTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{"path": "a.txt"})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strings.Contains(result.Content, "add b") {
		t.Error("expected only a.txt history, but got b.txt commit")
	}
}

func TestGitLogTool_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	tl := &GitLogTool{}
	tl.SetRoot(dir)
	_, err := tl.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGitLogTool_Schema(t *testing.T) {
	tl := &GitLogTool{}
	if tl.Name() != "git_log" {
		t.Errorf("Name() = %q", tl.Name())
	}
	if !tl.IsReadOnly() {
		t.Error("expected IsReadOnly() = true")
	}
}

func TestGitCommitTool_Basic(t *testing.T) {
	dir := setupGitRepo(t)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)

	tl := &GitCommitTool{}
	tl.SetRoot(dir)
	result, err := tl.Execute(context.Background(), map[string]any{"message": "initial commit"})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(result.Content, "initial commit") {
		t.Errorf("expected commit message in output, got: %s", result.Content)
	}

	logOut, _ := exec.Command("git", "-C", dir, "log", "--oneline").Output()
	if !strings.Contains(string(logOut), "initial commit") {
		t.Error("commit not found in git log")
	}
}

func TestGitCommitTool_MissingMessage(t *testing.T) {
	dir := setupGitRepo(t)
	tl := &GitCommitTool{}
	tl.SetRoot(dir)
	_, err := tl.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for missing message")
	}
}

func TestGitCommitTool_EmptyMessage(t *testing.T) {
	dir := setupGitRepo(t)
	tl := &GitCommitTool{}
	tl.SetRoot(dir)
	_, err := tl.Execute(context.Background(), map[string]any{"message": ""})
	if err == nil {
		t.Error("expected error for empty message")
	}
}

func TestGitCommitTool_NothingToCommit(t *testing.T) {
	dir := setupGitRepo(t)
	tl := &GitCommitTool{}
	tl.SetRoot(dir)
	_, err := tl.Execute(context.Background(), map[string]any{"message": "empty"})
	if err == nil {
		t.Error("expected error when nothing to commit")
	}
}

func TestGitCommitTool_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	tl := &GitCommitTool{}
	tl.SetRoot(dir)
	_, err := tl.Execute(context.Background(), map[string]any{"message": "test"})
	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error = %v, want 'not a git repository'", err)
	}
}

func TestGitCommitTool_IsReadOnly(t *testing.T) {
	tl := &GitCommitTool{}
	if tl.IsReadOnly() {
		t.Error("expected IsReadOnly() = false")
	}
}

func TestGitCommitTool_Schema(t *testing.T) {
	tl := &GitCommitTool{}
	if tl.Name() != "git_commit" {
		t.Errorf("Name() = %q", tl.Name())
	}
	s := tl.Schema()
	if len(s.Required) != 1 || s.Required[0] != "message" {
		t.Errorf("Required = %v, want [message]", s.Required)
	}
}

func TestGitTools_RegisterDefaults(t *testing.T) {
	r := NewRegistry()
	r.RegisterDefaults()

	for _, name := range []string{"git_status", "git_diff", "git_log", "git_commit"} {
		if _, ok := r.Get(name); !ok {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}
