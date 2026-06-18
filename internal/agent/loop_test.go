package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ShawnLiuSZ/Helix/internal/provider"
	"github.com/ShawnLiuSZ/Helix/internal/tool"
)

// stubProvider 测试用 Provider
type stubProvider struct {
	name     string
	models   []provider.ModelInfo
	caps     provider.Capabilities
	chatFn   func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error)
}

func (s *stubProvider) Name() string                          { return s.name }
func (s *stubProvider) Models() []provider.ModelInfo          { return s.models }
func (s *stubProvider) Capabilities() provider.Capabilities   { return s.caps }
func (s *stubProvider) Cost(modelID string, usage provider.Usage) provider.Cost {
	return provider.Cost{Currency: "USD"}
}
func (s *stubProvider) Chat(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
	if s.chatFn != nil {
		return s.chatFn(ctx, req)
	}
	return &provider.ChatResponse{Content: "default response"}, nil
}
func (s *stubProvider) Stream(ctx context.Context, req *provider.ChatRequest) (<-chan provider.StreamEvent, error) {
	ch := make(chan provider.StreamEvent, 1)
	ch <- provider.StreamEvent{Type: provider.EventText, Content: "stream response"}
	close(ch)
	return ch, nil
}

func newStubProvider(chatFn func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error)) *stubProvider {
	return &stubProvider{
		name: "stub",
		caps: provider.Capabilities{SupportsToolCall: true},
		chatFn: chatFn,
	}
}

func TestAgent_SingleTurn(t *testing.T) {
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "Hello, I can help!"}, nil
	})

	r := tool.NewRegistry()
	agent := New(p, r)

	result, err := agent.Run(context.Background(), "say hello")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result != "Hello, I can help!" {
		t.Errorf("result = %q", result)
	}
}

func TestAgent_MultiTurnToolCall(t *testing.T) {
	callCount := 0
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		callCount++
		if callCount == 1 {
			// 第一轮：返回工具调用
			return &provider.ChatResponse{
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "read_file", Args: map[string]any{"path": "/tmp/test.txt"}},
				},
			}, nil
		}
		// 第二轮：返回最终答案
		return &provider.ChatResponse{Content: "File analysis complete"}, nil
	})

	r := tool.NewRegistry()
	r.Register(&tool.ReadFileTool{})

	agent := New(p, r)
	agent.SetMaxSteps(5)

	result, err := agent.Run(context.Background(), "analyze /tmp/test.txt")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result != "File analysis complete" {
		t.Errorf("result = %q", result)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

func TestAgent_ToolCallWithError(t *testing.T) {
	callCount := 0
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		callCount++
		if callCount == 1 {
			return &provider.ChatResponse{
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "read_file", Args: map[string]any{"path": "/nonexistent"}},
				},
			}, nil
		}
		return &provider.ChatResponse{Content: "File not found, but I'll continue"}, nil
	})

	r := tool.NewRegistry()
	r.Register(&tool.ReadFileTool{})

	agent := New(p, r)
	agent.SetMaxSteps(5)

	result, err := agent.Run(context.Background(), "read nonexistent file")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	// Agent 应该继续处理并返回最终答案
	if !strings.Contains(result, "continue") {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestAgent_MaxSteps(t *testing.T) {
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{
			ToolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "read_file", Args: map[string]any{"path": "/tmp/test"}},
			},
		}, nil
	})

	r := tool.NewRegistry()
	r.Register(&tool.ReadFileTool{})

	agent := New(p, r)
	agent.SetMaxSteps(3)

	_, err := agent.Run(context.Background(), "do something")
	if err == nil {
		t.Fatal("expected error for max steps")
	}
	if !strings.Contains(err.Error(), "max steps") {
		t.Errorf("error = %v", err)
	}
}

func TestAgent_ChatError(t *testing.T) {
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return nil, errors.New("api unavailable")
	})

	r := tool.NewRegistry()
	agent := New(p, r)

	_, err := agent.Run(context.Background(), "do something")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "api unavailable") {
		t.Errorf("error = %v", err)
	}
}

func TestAgent_ContextCancellation(t *testing.T) {
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "ok"}, nil
	})

	r := tool.NewRegistry()
	agent := New(p, r)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	_, err := agent.Run(ctx, "do something")
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestAgent_GuardChain(t *testing.T) {
	guardCalled := false
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{
			ToolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "read_file", Args: map[string]any{"path": "/tmp/test"}},
			},
		}, nil
	})

	r := tool.NewRegistry()
	r.Register(&tool.ReadFileTool{})

	agent := New(p, r)
	agent.AddGuard(func(c tool.Call) error {
		guardCalled = true
		return nil
	})

	// 需要文件存在才能通过 guard 后的执行
	t.Setenv("HELIX_TEST", "1")
	_, _ = agent.Run(context.Background(), "read file")

	if !guardCalled {
		t.Error("guard was not called")
	}
}

func TestAgent_BuildToolDefs(t *testing.T) {
	r := tool.NewRegistry()
	r.Register(&tool.ReadFileTool{})
	r.Register(&tool.GrepTool{})

	p := newStubProvider(nil)
	agent := New(p, r)

	defs := agent.buildToolDefs()
	if len(defs) != 2 {
		t.Errorf("buildToolDefs() count = %d, want 2", len(defs))
	}

	names := map[string]bool{}
	for _, d := range defs {
		names[d.Function.Name] = true
	}
	if !names["read_file"] || !names["grep"] {
		t.Errorf("missing tool in defs: %v", names)
	}
}

func TestAgent_BuildSystemPrompt(t *testing.T) {
	p := newStubProvider(nil)
	r := tool.NewRegistry()
	agent := New(p, r)

	prompt := agent.buildSystemPrompt()
	if !strings.Contains(prompt, "Helix") {
		t.Error("system prompt should mention Helix")
	}
	if !strings.Contains(prompt, "tools") {
		t.Error("system prompt should mention tools")
	}
}
