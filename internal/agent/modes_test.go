package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/ShawnLiuSZ/Helix/internal/provider"
	"github.com/ShawnLiuSZ/Helix/internal/tool"
)

func TestMode_String(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeBuild, "build"},
		{ModePlan, "plan"},
		{ModeCompose, "compose"},
		{ModeMax, "max"},
	}

	for _, tt := range tests {
		if tt.mode.String() != tt.want {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, tt.mode.String(), tt.want)
		}
	}
}

func TestModeFromString(t *testing.T) {
	tests := []struct {
		s    string
		want Mode
	}{
		{"build", ModeBuild},
		{"plan", ModePlan},
		{"compose", ModeCompose},
		{"max", ModeMax},
		{"unknown", ModeBuild}, // 默认 build
	}

	for _, tt := range tests {
		got := ModeFromString(tt.s)
		if got != tt.want {
			t.Errorf("ModeFromString(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestMultiAgent_SetMode(t *testing.T) {
	p := newStubProvider(nil)
	r := tool.NewRegistry()
	a := NewMultiAgent(p, r)

	a.SetMode(ModePlan)
	if a.Mode() != ModePlan {
		t.Errorf("Mode() = %v", a.Mode())
	}
}

func TestMultiAgent_PlanMode(t *testing.T) {
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "Plan: 1. Analyze code\n2. Implement\n3. Test"}, nil
	})

	r := tool.NewRegistry()
	r.Register(&tool.ReadFileTool{})
	r.Register(&tool.GrepTool{})
	r.Register(&tool.WriteFileTool{}) // Plan 模式不应使用

	a := NewMultiAgent(p, r)
	a.SetMode(ModePlan)

	result, err := a.Run(context.Background(), "plan this feature")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !strings.Contains(result, "Plan") {
		t.Errorf("result should contain plan: %q", result)
	}
}

func TestMultiAgent_PlanMode_ReadOnlyTools(t *testing.T) {
	p := newStubProvider(nil)
	r := tool.NewRegistry()
	r.Register(&tool.ReadFileTool{})
	r.Register(&tool.WriteFileTool{})

	a := NewMultiAgent(p, r)
	defs := a.buildReadOnlyToolDefs()

	// 只应包含只读工具
	if len(defs) != 1 {
		t.Errorf("readOnly defs count = %d, want 1", len(defs))
	}
	if defs[0].Function.Name != "read_file" {
		t.Errorf("def name = %q", defs[0].Function.Name)
	}
}

func TestMultiAgent_ComposeMode(t *testing.T) {
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "Feature implemented according to spec"}, nil
	})

	r := tool.NewRegistry()
	a := NewMultiAgent(p, r)
	a.SetMode(ModeCompose)
	a.SetSpec("Implement user login with JWT")

	result, err := a.Run(context.Background(), "implement login")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !strings.Contains(result, "implement") {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestMultiAgent_MaxMode(t *testing.T) {
	p := newStubProvider(func(ctx context.Context, req *provider.ChatRequest) (*provider.ChatResponse, error) {
		return &provider.ChatResponse{Content: "best answer from parallel candidates"}, nil
	})

	r := tool.NewRegistry()
	a := NewMultiAgent(p, r)
	a.SetMode(ModeMax)
	a.SetMaxSteps(3)

	result, err := a.Run(context.Background(), "solve complex problem")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result == "" {
		t.Error("result should not be empty")
	}
}

func TestMultiAgent_BuildPrompt(t *testing.T) {
	p := newStubProvider(nil)
	r := tool.NewRegistry()
	a := NewMultiAgent(p, r)

	planPrompt := a.buildPlanPrompt()
	if !strings.Contains(planPrompt, "Plan mode") {
		t.Error("plan prompt should mention Plan mode")
	}
	if !strings.Contains(planPrompt, "READ-ONLY") {
		t.Error("plan prompt should mention READ-ONLY")
	}

	composePrompt := a.buildComposePrompt()
	if !strings.Contains(composePrompt, "Compose mode") {
		t.Error("compose prompt should mention Compose mode")
	}
}
