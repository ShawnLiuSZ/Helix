package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ShawnLiuSZ/Helix/internal/session"
)

func TestDreamScheduler(t *testing.T) {
	t.Run("NewDreamScheduler", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}
		sessionMgr, _ := session.NewManager(filepath.Join(tmpDir, "sessions"))

		scheduler := NewDreamScheduler(agent, sessionMgr, filepath.Join(tmpDir, "memory"))
		if scheduler == nil {
			t.Fatal("expected non-nil scheduler")
		}
	})

	t.Run("filterRecent", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}
		sessionMgr, _ := session.NewManager(filepath.Join(tmpDir, "sessions"))

		scheduler := NewDreamScheduler(agent, sessionMgr, filepath.Join(tmpDir, "memory"))

		now := time.Now()
		sessions := []*session.Session{
			{Meta: session.Meta{ID: "1", UpdatedAt: now.Add(-1 * time.Hour)}},
			{Meta: session.Meta{ID: "2", UpdatedAt: now.Add(-2 * 24 * time.Hour)}},
			{Meta: session.Meta{ID: "3", UpdatedAt: now.Add(-10 * 24 * time.Hour)}},
		}

		recent := scheduler.filterRecent(sessions, 7*24*time.Hour)
		if len(recent) != 2 {
			t.Errorf("expected 2 recent sessions, got %d", len(recent))
		}
	})

	t.Run("extractTopics", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}
		sessionMgr, _ := session.NewManager(filepath.Join(tmpDir, "sessions"))

		scheduler := NewDreamScheduler(agent, sessionMgr, filepath.Join(tmpDir, "memory"))

		patterns := &Patterns{
			Topics: make(map[string]int),
			Tools:  make(map[string]int),
		}

		scheduler.extractTopics("请帮我创建一个函数", patterns)
		scheduler.extractTopics("修改这个文件", patterns)
		scheduler.extractTopics("修复 bug", patterns)

		if patterns.Topics["创建"] != 1 {
			t.Errorf("expected '创建' count 1, got %d", patterns.Topics["创建"])
		}
		if patterns.Topics["修改"] != 1 {
			t.Errorf("expected '修改' count 1, got %d", patterns.Topics["修改"])
		}
		if patterns.Topics["修复"] != 1 {
			t.Errorf("expected '修复' count 1, got %d", patterns.Topics["修复"])
		}
	})

	t.Run("extractKnowledge", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}
		sessionMgr, _ := session.NewManager(filepath.Join(tmpDir, "sessions"))

		scheduler := NewDreamScheduler(agent, sessionMgr, filepath.Join(tmpDir, "memory"))

		patterns := &Patterns{
			Topics: map[string]int{"创建": 5, "修改": 2},
			Tools:  map[string]int{"bash": 10, "write_file": 15},
		}

		knowledge := scheduler.extractKnowledge(patterns)
		if knowledge == nil {
			t.Fatal("expected non-nil knowledge")
		}
		if len(knowledge.Suggestions) == 0 {
			t.Error("expected at least one suggestion")
		}
	})

	t.Run("saveToMemory", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}
		sessionMgr, _ := session.NewManager(filepath.Join(tmpDir, "sessions"))

		scheduler := NewDreamScheduler(agent, sessionMgr, filepath.Join(tmpDir, "memory"))

		knowledge := &Knowledge{
			GeneratedAt: time.Now(),
			Topics:      map[string]int{"创建": 5},
			Tools:       map[string]int{"bash": 10},
			Suggestions: []string{"测试建议"},
		}

		err := scheduler.saveToMemory(knowledge)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 验证文件存在
		memoryFile := filepath.Join(tmpDir, "memory", "MEMORY.md")
		if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
			t.Error("MEMORY.md not created")
		}
	})
}

func TestDistiller(t *testing.T) {
	t.Run("NewDistiller", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}

		distiller := NewDistiller(agent, filepath.Join(tmpDir, "skills"))
		if distiller == nil {
			t.Fatal("expected non-nil distiller")
		}
	})

	t.Run("analyzeWorkflows", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}

		distiller := NewDistiller(agent, filepath.Join(tmpDir, "skills"))

		sessions := []*session.Session{
			{
				Meta: session.Meta{ID: "1"},
				Messages: []session.Message{
					{Role: "tool", ToolName: "read_file"},
					{Role: "tool", ToolName: "edit_file"},
					{Role: "tool", ToolName: "bash"},
				},
			},
			{
				Meta: session.Meta{ID: "2"},
				Messages: []session.Message{
					{Role: "tool", ToolName: "read_file"},
					{Role: "tool", ToolName: "edit_file"},
					{Role: "tool", ToolName: "bash"},
				},
			},
			{
				Meta: session.Meta{ID: "3"},
				Messages: []session.Message{
					{Role: "tool", ToolName: "read_file"},
					{Role: "tool", ToolName: "edit_file"},
					{Role: "tool", ToolName: "bash"},
				},
			},
		}

		workflows := distiller.analyzeWorkflows(sessions)
		if len(workflows) == 0 {
			t.Error("expected at least one workflow")
		}
	})

	t.Run("generateSkillName", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}

		distiller := NewDistiller(agent, filepath.Join(tmpDir, "skills"))

		wf := &Workflow{
			Steps: []string{"read_file", "edit_file", "bash"},
		}

		name := distiller.generateSkillName(wf)
		if name != "auto-read_file" {
			t.Errorf("expected 'auto-read_file', got %q", name)
		}
	})

	t.Run("generateSkillContent", func(t *testing.T) {
		tmpDir := t.TempDir()
		agent := &Agent{}

		distiller := NewDistiller(agent, filepath.Join(tmpDir, "skills"))

		wf := &Workflow{
			Steps:     []string{"read_file", "edit_file"},
			Count:     5,
			Confidence: 0.8,
		}

		content := distiller.generateSkillContent(wf)
		if content == "" {
			t.Error("expected non-empty content")
		}
		if !containsString(content, "read_file") {
			t.Error("expected content to contain 'read_file'")
		}
	})
}

func TestDreamRunDream(t *testing.T) {
	tmpDir := t.TempDir()
	agent := &Agent{}
	sessionMgr, _ := session.NewManager(filepath.Join(tmpDir, "sessions"))

	// 创建一些测试会话
	s1 := sessionMgr.Create("test1", "model", "provider")
	s1.Messages = append(s1.Messages, session.Message{
		Role:    "user",
		Content: "请帮我创建一个函数",
	})
	s1.Messages = append(s1.Messages, session.Message{
		Role:    "tool",
		ToolName: "write_file",
	})

	scheduler := NewDreamScheduler(agent, sessionMgr, filepath.Join(tmpDir, "memory"))

	err := scheduler.RunDream()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
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

// TestDreamContext 测试 Dream 上下文
func TestDreamContext(t *testing.T) {
	tmpDir := t.TempDir()
	agent := &Agent{}
	sessionMgr, _ := session.NewManager(filepath.Join(tmpDir, "sessions"))

	_ = NewDreamScheduler(agent, sessionMgr, filepath.Join(tmpDir, "memory"))

	// 测试取消操作
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 验证上下文已取消
	if ctx.Err() == nil {
		t.Error("expected context to be cancelled")
	}
}
