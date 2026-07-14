package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ShawnLiuSZ/loomcode/internal/provider"
	"github.com/ShawnLiuSZ/loomcode/internal/testutil"
	"github.com/ShawnLiuSZ/loomcode/internal/tool"
)

// --- System prompt environment grounding ---

func TestBuildSystemPrompt_IncludesEnvironment(t *testing.T) {
	p := testutil.NewStubProvider(nil)
	a := New(p, tool.NewRegistry())

	dir := t.TempDir()
	a.SetWorkDir(dir)
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/feature-x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	prompt := a.buildSystemPrompt()

	if !strings.Contains(prompt, runtime.GOOS) {
		t.Errorf("prompt missing OS %q", runtime.GOOS)
	}
	if !strings.Contains(prompt, dir) {
		t.Error("prompt missing working directory")
	}
	if !strings.Contains(prompt, "feature-x") {
		t.Error("prompt missing git branch")
	}
	if !strings.Contains(prompt, "main.go") {
		t.Error("prompt missing directory listing")
	}
}

func TestBuildSystemPrompt_NoWorkDirStillValid(t *testing.T) {
	p := testutil.NewStubProvider(nil)
	a := New(p, tool.NewRegistry())

	prompt := a.buildSystemPrompt()
	if !strings.Contains(prompt, "LoomCode") {
		t.Error("system prompt should mention LoomCode")
	}
	if !strings.Contains(prompt, "tools") {
		t.Error("system prompt should mention tools")
	}
}

// --- Cache-aware compaction ---

func bigMsgs() []provider.Message {
	return []provider.Message{
		{Role: "system", Content: strings.Repeat("s", 30)},
		{Role: "user", Content: strings.Repeat("u", 300)},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c1"}}},
		{Role: "tool", ToolCallID: "c1", Content: strings.Repeat("t", 300)},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c2"}}},
		{Role: "tool", ToolCallID: "c2", Content: strings.Repeat("t", 300)},
		{Role: "user", Content: "more"},
		{Role: "assistant", Content: "answer"},
		{Role: "user", Content: "again"},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c3"}}},
		{Role: "tool", ToolCallID: "c3", Content: strings.Repeat("t", 300)},
		{Role: "user", Content: "more2"},
		{Role: "assistant", Content: "answer2"},
	}
}

func TestCompactMessages_SummarizesOldRounds(t *testing.T) {
	var summarizeReq *provider.ChatRequest
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		summarizeReq = req
		return &provider.ChatResponse{Content: "summary of earlier work"}, nil
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")
	a.messages = bigMsgs()

	a.compactMessages(context.Background(), 100) // maxInput=80, tokens far exceed it

	if summarizeReq == nil {
		t.Fatal("expected a summarization LLM call")
	}
	// Prefix must stay: real system prompt first, then a single summary block.
	if a.messages[0].Role != "system" {
		t.Fatalf("messages[0] role = %q, want system", a.messages[0].Role)
	}
	if !strings.Contains(a.messages[1].Content, "summary of earlier work") {
		t.Errorf("messages[1] should hold the summary, got %q", a.messages[1].Content)
	}
	// Kept region must not start with an orphan tool result.
	if a.messages[2].Role == "tool" {
		t.Error("kept region starts with an orphan tool result")
	}
	// Recent tail preserved.
	last := a.messages[len(a.messages)-1]
	if last.Role != "assistant" || last.Content != "answer2" {
		t.Errorf("recent tail not preserved, last = %+v", last)
	}
	if len(a.messages) >= len(bigMsgs()) {
		t.Errorf("compaction did not shrink message list: %d", len(a.messages))
	}
}

func TestCompactMessages_KeepsDynamicSystemPrompt(t *testing.T) {
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "summary"}, nil
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")
	a.SetWorkDir("/tmp/project-x")
	a.messages = []provider.Message{
		{Role: "system", Content: "static system"},
		{Role: "system", Content: "dynamic env"},
		{Role: "user", Content: strings.Repeat("u", 300)},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c1"}}},
		{Role: "tool", ToolCallID: "c1", Content: strings.Repeat("t", 300)},
		{Role: "user", Content: "U6"},
		{Role: "assistant", Content: "A7"},
		{Role: "user", Content: "U8"},
		{Role: "assistant", Content: "A9"},
		{Role: "user", Content: "U10"},
		{Role: "assistant", Content: "A11"},
		{Role: "user", Content: "U12"},
	}

	a.compactMessages(context.Background(), 100)

	// 前两条 system 必须保留，顺序为 [static, dynamic, summary, ...]。
	if len(a.messages) < 3 {
		t.Fatalf("expected at least 3 messages after compaction, got %d", len(a.messages))
	}
	if a.messages[0].Content != "static system" {
		t.Errorf("messages[0] = %q, want static system", a.messages[0].Content)
	}
	if a.messages[1].Content != "dynamic env" {
		t.Errorf("messages[1] = %q, want dynamic env", a.messages[1].Content)
	}
	if !strings.Contains(a.messages[2].Content, "summary") {
		t.Errorf("messages[2] should hold the summary, got %q", a.messages[2].Content)
	}
}

func TestCompactMessages_CutAdjustsPastToolMessage(t *testing.T) {
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "S"}, nil
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")
	// len=11, keepRecent=8 -> naive cut=3 lands on a tool message; must advance to 4 (user).
	a.messages = []provider.Message{
		{Role: "system", Content: strings.Repeat("s", 30)},
		{Role: "user", Content: strings.Repeat("u", 300)},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c1"}}},
		{Role: "tool", ToolCallID: "c1", Content: strings.Repeat("t", 300)},
		{Role: "user", Content: "U4"},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c2"}}},
		{Role: "tool", ToolCallID: "c2", Content: "r2"},
		{Role: "user", Content: "U7"},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c3"}}},
		{Role: "tool", ToolCallID: "c3", Content: "r3"},
		{Role: "user", Content: "U10"},
	}

	a.compactMessages(context.Background(), 100)

	// [system, summary, user(U4), assistant(c2), tool(c2), user(U7), assistant(c3), tool(c3), user(U10)]
	if a.messages[2].Role != "user" || a.messages[2].Content != "U4" {
		t.Errorf("kept region should begin at the user message U4, got %+v", a.messages[2])
	}
	if a.messages[3].Role != "assistant" || a.messages[4].Role != "tool" {
		t.Error("assistant/tool pairing in kept region broken")
	}
}

func TestCompactMessages_FallbackOnSummarizeError(t *testing.T) {
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return nil, errors.New("summarize boom")
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")
	a.messages = bigMsgs()
	before := len(a.messages)

	a.compactMessages(context.Background(), 100)

	if len(a.messages) >= before {
		t.Errorf("fallback truncation should shrink list: before=%d after=%d", before, len(a.messages))
	}
	if a.messages[0].Role != "system" {
		t.Error("system prompt must be preserved by fallback")
	}
	// No summary block was inserted.
	if len(a.messages) > 1 && strings.Contains(a.messages[1].Content, "summary") {
		t.Error("fallback must not insert a summary block")
	}
}

func TestCompactMessages_NoOpUnderThreshold(t *testing.T) {
	called := false
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		called = true
		return &provider.ChatResponse{Content: "x"}, nil
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")
	a.messages = []provider.Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "ok"},
	}

	a.compactMessages(context.Background(), 100000) // huge window -> never triggers

	if called {
		t.Error("should not summarize when under threshold")
	}
	if len(a.messages) != 3 {
		t.Errorf("messages should be untouched, got %d", len(a.messages))
	}
}

func TestCompactMessages_UnknownWindowNoOp(t *testing.T) {
	called := false
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		called = true
		return &provider.ChatResponse{Content: "x"}, nil
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")
	a.messages = bigMsgs()

	a.compactMessages(context.Background(), 0) // unknown context window

	if called {
		t.Error("should not summarize when context window unknown")
	}
	if len(a.messages) != len(bigMsgs()) {
		t.Error("messages should be untouched when window unknown")
	}
}

// TestTruncateMessages_NoOrphanToolOnBoundary 验证 L6 修复：当 assistant tool_call
// 位于删除区、但其 tool 结果恰好落在保留区边界时，截断后不得留下 orphan tool 消息
// （即没有对应 assistant tool_call 的 tool 结果），否则 LLM API 返回 400。
func TestTruncateMessages_NoOrphanToolOnBoundary(t *testing.T) {
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "x"}, nil
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")

	// 6 条消息：[system, assistant(c1), tool(c1), user, assistant, user]
	// start=1, keepRecent=4 → 保留区 index 2-5
	// assistant(c1) 在 index 1（删除区），tool(c1) 在 index 2（保留区边界）
	// 截断后若 orphan 扫描条件用 < 而非 <=，tool(c1) 会被留下成 orphan
	pad := strings.Repeat("x", 200) // 每条消息足够长以触发截断
	a.messages = []provider.Message{
		{Role: "system", Content: "sys"},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c1"}}, Content: pad},
		{Role: "tool", ToolCallID: "c1", Content: pad},
		{Role: "user", Content: pad},
		{Role: "assistant", Content: pad},
		{Role: "user", Content: pad},
	}

	a.truncateMessages(100) // 小窗口强制截断

	// 断言：结果中不存在 orphan tool 消息（tool 消息必须有前置 assistant tool_call）
	for i, msg := range a.messages {
		if msg.Role != "tool" {
			continue
		}
		if i == 0 {
			t.Fatalf("tool message at index 0 has no preceding assistant (orphan)")
		}
		prev := a.messages[i-1]
		if prev.Role != "assistant" || len(prev.ToolCalls) == 0 {
			t.Fatalf("orphan tool message at index %d: preceding msg is %s, not assistant with tool_calls", i, prev.Role)
		}
	}
}

// TestTruncateMessages_PreservesRecentAndSystem 验证截断后仍保留前导 system 与最近消息。
func TestTruncateMessages_PreservesRecentAndSystem(t *testing.T) {
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "x"}, nil
	})
	a := New(p, tool.NewRegistry())
	a.SetModel("m")

	pad := strings.Repeat("x", 200)
	a.messages = []provider.Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: pad},
		{Role: "assistant", ToolCalls: []provider.ToolCall{{ID: "c1"}}, Content: pad},
		{Role: "tool", ToolCallID: "c1", Content: pad},
		{Role: "user", Content: "final-user"},
		{Role: "assistant", Content: "final-asst"},
		{Role: "user", Content: "tail1"},
		{Role: "assistant", Content: "tail2"},
	}

	a.truncateMessages(100)

	if a.messages[0].Role != "system" {
		t.Errorf("system prompt must be preserved, got %s at index 0", a.messages[0].Role)
	}
	// 最近 4 条应保留
	n := len(a.messages)
	if n < 4 {
		t.Fatalf("expected at least 4 messages, got %d", n)
	}
	last4 := a.messages[n-4:]
	expected := []string{"final-user", "final-asst", "tail1", "tail2"}
	for i, want := range expected {
		if last4[i].Content != want {
			t.Errorf("last4[%d] = %q, want %q", i, last4[i].Content, want)
		}
	}
}
