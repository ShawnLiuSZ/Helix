package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_Create(t *testing.T) {
	dir := t.TempDir()
	m, err := NewManager(dir)
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	s := m.Create("test session", "deepseek-v4-flash", "deepseek")
	if s.ID == "" {
		t.Error("session ID should not be empty")
	}
	if s.Name != "test session" {
		t.Errorf("Name = %q", s.Name)
	}

	if m.Count() != 1 {
		t.Errorf("Count() = %d, want 1", m.Count())
	}
}

func TestManager_Get(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir)
	s := m.Create("test", "flash", "deepseek")

	got, ok := m.Get(s.ID)
	if !ok {
		t.Fatal("session not found")
	}
	if got.Name != "test" {
		t.Errorf("Name = %q", got.Name)
	}

	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("should not find nonexistent session")
	}
}

func TestManager_Active(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir)

	s1 := m.Create("session 1", "flash", "deepseek")
	if m.Active().ID != s1.ID {
		t.Error("newly created session should be active")
	}

	s2 := m.Create("session 2", "pro", "deepseek")
	if m.Active().ID != s2.ID {
		t.Error("latest created session should be active")
	}

	// 切换回 s1
	if err := m.SetActive(s1.ID); err != nil {
		t.Fatalf("SetActive error: %v", err)
	}
	if m.Active().ID != s1.ID {
		t.Error("active session should be s1")
	}
}

func TestManager_SetActive_NotFound(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir)

	err := m.SetActive("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestManager_List(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir)

	m.Create("a", "flash", "deepseek")
	time.Sleep(2 * time.Millisecond)
	m.Create("b", "pro", "deepseek")
	time.Sleep(2 * time.Millisecond)
	m.Create("c", "flash", "mimo")

	list := m.List()
	if len(list) != 3 {
		t.Errorf("List() count = %d, want 3", len(list))
	}
}

func TestManager_AddMessage(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir)
	m.Create("test", "flash", "deepseek")

	m.AddMessage(Message{Role: "user", Content: "hello"})
	m.AddMessage(Message{Role: "assistant", Content: "hi there"})

	s := m.Active()
	if len(s.Messages) != 2 {
		t.Errorf("Messages count = %d, want 2", len(s.Messages))
	}
	if s.Messages[0].Role != "user" {
		t.Errorf("msg[0].Role = %q", s.Messages[0].Role)
	}
}

func TestManager_Save(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir)
	s := m.Create("test", "flash", "deepseek")

	m.AddMessage(Message{Role: "user", Content: "hello"})
	m.AddMessage(Message{Role: "assistant", Content: "world"})

	if err := m.Save(s.ID); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// 验证文件存在
	path := filepath.Join(dir, s.ID+".jsonl")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("session file should exist")
	}
}

func TestManager_Delete(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir)
	s := m.Create("test", "flash", "deepseek")

	m.AddMessage(Message{Role: "user", Content: "hello"})
	m.Save(s.ID)

	if err := m.Delete(s.ID); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	if m.Count() != 0 {
		t.Error("session should be deleted")
	}

	// 文件应被删除
	path := filepath.Join(dir, s.ID+".jsonl")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("session file should be deleted")
	}
}

func TestManager_Persistence(t *testing.T) {
	dir := t.TempDir()

	// 创建并保存
	m1, _ := NewManager(dir)
	s := m1.Create("persist", "flash", "deepseek")
	m1.AddMessage(Message{Role: "user", Content: "hello"})
	m1.Save(s.ID)

	// 重新加载
	m2, err := NewManager(dir)
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	loaded, ok := m2.Get(s.ID)
	if !ok {
		t.Fatal("session should persist across manager instances")
	}
	if len(loaded.Messages) != 1 {
		t.Errorf("loaded Messages count = %d, want 1", len(loaded.Messages))
	}
	if loaded.Messages[0].Content != "hello" {
		t.Errorf("loaded msg content = %q", loaded.Messages[0].Content)
	}
}

func TestMessage_Timestamp(t *testing.T) {
	msg := Message{
		Role:      "user",
		Content:   "test",
		Timestamp: time.Now(),
	}

	if msg.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}
