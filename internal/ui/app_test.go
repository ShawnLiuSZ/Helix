package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ShawnLiuSZ/Helix/internal/testutil"
	"github.com/ShawnLiuSZ/Helix/internal/tool"
)

// newTestApp 构造一个用于测试的 App，并初始化窗口尺寸（viewport/glamour）。
func newTestApp(t *testing.T) *App {
	t.Helper()
	app := NewApp(testutil.NewStubProvider(nil), tool.NewRegistry())
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return app
}

// typeRunes 逐字符发送按键，模拟用户键入。
func typeRunes(m tea.Model, s string) tea.Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

// TestTextareaReceivesTyping 是核心回归测试：
// 防止「按键事件未转发给 textarea，导致用户打不了字」的 P0 再次发生。
func TestTextareaReceivesTyping(t *testing.T) {
	var m tea.Model = newTestApp(t)
	m = typeRunes(m, "hello")
	if got := m.(*App).textArea.Value(); got != "hello" {
		t.Fatalf("键入未到达 textarea: want %q got %q", "hello", got)
	}
}

// TestTextareaBackspace 校验退格能编辑输入（特殊键也需转发给 textarea）。
func TestTextareaBackspace(t *testing.T) {
	var m tea.Model = newTestApp(t)
	m = typeRunes(m, "hi")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if got := m.(*App).textArea.Value(); got != "h" {
		t.Fatalf("退格未生效: want %q got %q", "h", got)
	}
}

// TestEnterSubmitsPlainText 校验回车把普通文本作为用户消息提交，并清空输入框。
func TestEnterSubmitsPlainText(t *testing.T) {
	var m tea.Model = newTestApp(t)
	m = typeRunes(m, "hello")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // 不执行返回的 cmd，只校验同步状态

	app := m.(*App)
	if v := app.textArea.Value(); v != "" {
		t.Fatalf("提交后输入框未清空: got %q", v)
	}
	if !app.loading {
		t.Fatalf("提交后应进入 loading 状态")
	}
	if len(app.messages) == 0 {
		t.Fatalf("提交后应追加用户消息")
	}
	last := app.messages[len(app.messages)-1]
	if last.Role != "user" || last.Content != "hello" {
		t.Fatalf("最后一条消息应为 user/hello, got %s/%q", last.Role, last.Content)
	}
}

// TestSlashTriggersSuggestions 校验输入 "/" 前缀会触发命令联想。
func TestSlashTriggersSuggestions(t *testing.T) {
	var m tea.Model = newTestApp(t)
	m = typeRunes(m, "/cl")

	app := m.(*App)
	if !app.showSuggestions {
		t.Fatalf("输入 / 前缀应触发命令联想")
	}
	found := false
	for _, s := range app.suggestions {
		if s == "/clear" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("联想结果应包含 /clear, got %v", app.suggestions)
	}
}
