package agent

import (
	"context"
	"testing"

	"github.com/ShawnLiuSZ/Helix/internal/provider"
	"github.com/ShawnLiuSZ/Helix/internal/testutil"
	"github.com/ShawnLiuSZ/Helix/internal/tool"
)

func TestRun_SendsReasoningEffortWhenSupported(t *testing.T) {
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "done"}, nil
	})
	p.CapsVal = provider.Capabilities{SupportsToolCall: true, SupportsReasoning: true}

	a := New(p, tool.NewRegistry())
	a.GetEffort().SetLevel(EffortHigh)

	_, _ = a.Run(context.Background(), "task")

	req := p.LastChatRequest()
	if req == nil {
		t.Fatal("no chat request recorded")
	}
	if req.ReasoningEffort != "high" {
		t.Errorf("ReasoningEffort = %q, want \"high\"", req.ReasoningEffort)
	}
}

func TestRun_OmitsReasoningEffortWhenUnsupported(t *testing.T) {
	p := testutil.NewStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "done"}, nil
	})
	p.CapsVal = provider.Capabilities{SupportsToolCall: true} // no reasoning

	a := New(p, tool.NewRegistry())
	a.GetEffort().SetLevel(EffortHigh)

	_, _ = a.Run(context.Background(), "task")

	req := p.LastChatRequest()
	if req == nil {
		t.Fatal("no chat request recorded")
	}
	if req.ReasoningEffort != "" {
		t.Errorf("ReasoningEffort = %q, want empty for non-reasoning provider", req.ReasoningEffort)
	}
}
