package testutil

import (
	"context"

	"github.com/ShawnLiuSZ/Helix/internal/provider"
)

// StubProvider 测试用 Provider stub（返回预设数据）
type StubProvider struct {
	NameVal       string
	ModelsVal     []provider.ModelInfo
	CapsVal       provider.Capabilities
	ChatFn        func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error)
	chatCalls     []*provider.ChatRequest
}

func (s *StubProvider) Name() string                        { return s.NameVal }
func (s *StubProvider) Models() []provider.ModelInfo        { return s.ModelsVal }
func (s *StubProvider) Capabilities() provider.Capabilities { return s.CapsVal }
func (s *StubProvider) Cost(modelID string, usage provider.Usage) provider.Cost {
	return provider.Cost{Currency: "USD"}
}

func (s *StubProvider) Chat(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
	s.chatCalls = append(s.chatCalls, req)
	if s.ChatFn != nil {
		return s.ChatFn(ctx, req)
	}
	return &provider.ChatResponse{Content: "stub response"}, nil
}

func (s *StubProvider) Stream(ctx context.Context, req *provider.ChatRequest) (<-chan provider.StreamEvent, error) {
	ch := make(chan provider.StreamEvent, 1)
	ch <- provider.StreamEvent{Type: provider.EventText, Content: "stub stream"}
	close(ch)
	return ch, nil
}

// ChatCallCount 返回 Chat 调用次数
func (s *StubProvider) ChatCallCount() int {
	return len(s.chatCalls)
}

// NewStubProvider 创建 StubProvider
func NewStubProvider(chatFn func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error)) *StubProvider {
	return &StubProvider{
		NameVal: "stub",
		CapsVal: provider.Capabilities{SupportsToolCall: true},
		ChatFn:  chatFn,
	}
}

// CallRecorder 通用调用记录器
type CallRecorder struct {
	Calls [][]any
}

func (r *CallRecorder) Record(args ...any) {
	r.Calls = append(r.Calls, args)
}

func (r *CallRecorder) Count() int {
	return len(r.Calls)
}
