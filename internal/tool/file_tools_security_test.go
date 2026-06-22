package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// C2: 文件工具必须把路径限制在 workspace root 之内
func TestFileTools_C2_Containment(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "a.txt")
	if err := os.WriteFile(inside, []byte("hi"), 0600); err != nil {
		t.Fatal(err)
	}

	rt := &ReadFileTool{}
	rt.SetRoot(root)

	// 区内读取应成功
	res, err := rt.Execute(context.Background(), map[string]any{"path": inside})
	if err != nil {
		t.Fatalf("read inside root should succeed: %v", err)
	}
	if res.Content != "hi" {
		t.Errorf("content = %q, want %q", res.Content, "hi")
	}

	// 父目录穿越应被拒绝
	if _, err := rt.Execute(context.Background(), map[string]any{"path": filepath.Join(root, "../../etc/passwd")}); err == nil {
		t.Error("read with ../ traversal must be denied")
	}
	// 绝对路径逃逸应被拒绝
	if _, err := rt.Execute(context.Background(), map[string]any{"path": "/etc/passwd"}); err == nil {
		t.Error("read /etc/passwd must be denied")
	}

	wt := &WriteFileTool{}
	wt.SetRoot(root)

	// 区外写入应被拒绝且不创建文件
	escape := filepath.Join(filepath.Dir(root), "escape.txt")
	if _, err := wt.Execute(context.Background(), map[string]any{"path": escape, "content": "x"}); err == nil {
		t.Error("write outside root must be denied")
	}
	if _, statErr := os.Stat(escape); statErr == nil {
		t.Error("escape file must not be created")
		os.Remove(escape)
	}

	// 区内写入应成功
	if _, err := wt.Execute(context.Background(), map[string]any{"path": filepath.Join(root, "out.txt"), "content": "ok"}); err != nil {
		t.Fatalf("write inside root should succeed: %v", err)
	}
}

// C2: 符号链接逃逸应被拒绝
func TestFileTools_C2_SymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("topsecret"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(secret, link); err != nil {
		t.Skip("symlink unsupported on this platform")
	}

	rt := &ReadFileTool{}
	rt.SetRoot(root)
	if _, err := rt.Execute(context.Background(), map[string]any{"path": link}); err == nil {
		t.Error("reading symlink that escapes root must be denied")
	}
}

// C2: root 为空时保持向后兼容（不限制），但仍可挂权限检查器
func TestFileTools_C2_NoRootBackwardCompat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "x.txt")
	os.WriteFile(f, []byte("data"), 0600)

	rt := &ReadFileTool{} // 未 SetRoot
	res, err := rt.Execute(context.Background(), map[string]any{"path": f})
	if err != nil || res.Content != "data" {
		t.Fatalf("no-root read should work: err=%v content=%q", err, res.Content)
	}
}
