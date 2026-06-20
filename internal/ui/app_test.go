package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ShawnLiuSZ/Helix/internal/testutil"
	"github.com/ShawnLiuSZ/Helix/internal/tool"
)

func newTestApp() *App {
	p := testutil.NewStubProvider(nil)
	tools := tool.NewRegistry()
	return NewApp(p, tools)
}

func TestTextareaInput(t *testing.T) {
	app := newTestApp()

	// 初始化窗口大小
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 模拟输入 "hi"
	var m tea.Model = app
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})

	result := m.(*App)
	if result.textArea.Value() != "hi" {
		t.Errorf("expected textarea value 'hi', got %q", result.textArea.Value())
	}
}

func TestEnterSendsMessage(t *testing.T) {
	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 输入文字
	var m tea.Model = app
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})

	// 按 Enter 发送
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := m.(*App)
	// textarea 应该被清空
	if result.textArea.Value() != "" {
		t.Errorf("expected textarea to be empty after enter, got %q", result.textArea.Value())
	}
	// 消息应该被添加
	if len(result.messages) < 3 { // system + user + possibly more
		t.Errorf("expected at least 3 messages after enter, got %d", len(result.messages))
	}
	// 最后一条应该是 user 消息
	lastMsg := result.messages[len(result.messages)-1]
	if lastMsg.Role != "user" {
		t.Errorf("expected last message role 'user', got %q", lastMsg.Role)
	}
	if lastMsg.Content != "hi" {
		t.Errorf("expected last message content 'hi', got %q", lastMsg.Content)
	}
}

func TestEscClearsInput(t *testing.T) {
	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 输入文字
	var m tea.Model = app
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})

	// 按 Esc 清空
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	result := m.(*App)
	if result.textArea.Value() != "" {
		t.Errorf("expected textarea to be empty after esc, got %q", result.textArea.Value())
	}
}

func TestCtrlCQuits(t *testing.T) {
	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	var m tea.Model = app
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	result := m.(*App)
	if !result.quitting {
		t.Error("expected quitting to be true after ctrl+c")
	}
	// cmd should be tea.Quit
	if cmd == nil {
		t.Error("expected a quit command")
	}
}

func TestTabCyclesMode(t *testing.T) {
	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	var m tea.Model = app
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	result := m.(*App)
	// Should cycle from build to plan
	if result.mode.String() != "plan" {
		t.Errorf("expected mode 'plan' after tab, got %q", result.mode.String())
	}
}

func TestSlashTriggersSuggestions(t *testing.T) {
	app := newTestApp()
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 输入 "/"
	var m tea.Model = app
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	result := m.(*App)
	if !result.showSuggestions {
		t.Error("expected suggestions to be shown after typing /")
	}
	if len(result.suggestions) == 0 {
		t.Error("expected at least one suggestion")
	}
}
