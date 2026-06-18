package agent

import (
	"testing"
	"time"
)

func TestHookType(t *testing.T) {
	tests := []struct {
		hookType HookType
		expected string
	}{
		{HookPreToolUse, "pre_tool_use"},
		{HookPostToolUse, "post_tool_use"},
		{HookUserPromptSubmit, "user_prompt_submit"},
		{HookStop, "stop"},
		{HookType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.hookType.String(); got != tt.expected {
				t.Errorf("HookType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseHookType(t *testing.T) {
	tests := []struct {
		input    string
		expected HookType
	}{
		{"pre_tool_use", HookPreToolUse},
		{"post_tool_use", HookPostToolUse},
		{"user_prompt_submit", HookUserPromptSubmit},
		{"stop", HookStop},
		{"unknown", HookType(-1)},
		{"PRE_TOOL_USE", HookPreToolUse},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseHookType(tt.input); got != tt.expected {
				t.Errorf("ParseHookType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHookManager(t *testing.T) {
	t.Run("NewHookManager", func(t *testing.T) {
		manager := NewHookManager()
		if manager == nil {
			t.Fatal("expected non-nil manager")
		}
	})

	t.Run("Register", func(t *testing.T) {
		manager := NewHookManager()
		hook := NewShellHook(HookPreToolUse, "echo test")
		manager.Register(hook)

		hooks := manager.GetHooks(HookPreToolUse)
		if len(hooks) != 1 {
			t.Errorf("expected 1 hook, got %d", len(hooks))
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		manager := NewHookManager()
		manager.Register(NewShellHook(HookPreToolUse, "echo 1"))
		manager.Register(NewShellHook(HookPreToolUse, "echo 2"))

		manager.Unregister(HookPreToolUse, 0)

		hooks := manager.GetHooks(HookPreToolUse)
		if len(hooks) != 1 {
			t.Errorf("expected 1 hook, got %d", len(hooks))
		}
	})

	t.Run("Clear", func(t *testing.T) {
		manager := NewHookManager()
		manager.Register(NewShellHook(HookPreToolUse, "echo 1"))
		manager.Register(NewShellHook(HookPostToolUse, "echo 2"))

		manager.Clear()

		if len(manager.GetHooks(HookPreToolUse)) != 0 {
			t.Error("expected 0 pre_tool_use hooks")
		}
		if len(manager.GetHooks(HookPostToolUse)) != 0 {
			t.Error("expected 0 post_tool_use hooks")
		}
	})
}

func TestShellHook(t *testing.T) {
	t.Run("NewShellHook", func(t *testing.T) {
		hook := NewShellHook(HookPreToolUse, "echo test")
		if hook == nil {
			t.Fatal("expected non-nil hook")
		}
		if hook.Type() != HookPreToolUse {
			t.Errorf("expected HookPreToolUse, got %v", hook.Type())
		}
	})

	t.Run("Execute_Success", func(t *testing.T) {
		hook := NewShellHook(HookPreToolUse, "echo hello")
		ctx := HookContext{
			Type: HookPreToolUse,
			Env:  []string{},
		}

		err := hook.Execute(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Execute_WithToolName", func(t *testing.T) {
		hook := NewShellHook(HookPreToolUse, "echo {{.ToolName}}")
		ctx := HookContext{
			Type:     HookPreToolUse,
			ToolName: "bash",
			Env:      []string{},
		}

		err := hook.Execute(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Execute_WithPrompt", func(t *testing.T) {
		hook := NewShellHook(HookUserPromptSubmit, "echo {{.Prompt}}")
		ctx := HookContext{
			Type:   HookUserPromptSubmit,
			Prompt: "test prompt",
			Env:    []string{},
		}

		err := hook.Execute(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Execute_Failure", func(t *testing.T) {
		hook := NewShellHook(HookPreToolUse, "exit 1")
		ctx := HookContext{
			Type: HookPreToolUse,
			Env:  []string{},
		}

		err := hook.Execute(ctx)
		if err == nil {
			t.Error("expected error for failed command")
		}
	})

	t.Run("Execute_Timeout", func(t *testing.T) {
		hook := NewShellHook(HookPreToolUse, "sleep 10")
		hook.timeout = 100 * time.Millisecond
		ctx := HookContext{
			Type: HookPreToolUse,
			Env:  []string{},
		}

		err := hook.Execute(ctx)
		if err == nil {
			t.Error("expected timeout error")
		}
	})
}

func TestLoadHooksFromConfig(t *testing.T) {
	config := HookConfig{
		PreToolUse: []HookConfigItem{
			{Command: "echo pre", Timeout: 10},
		},
		PostToolUse: []HookConfigItem{
			{Command: "echo post", Timeout: 20},
		},
		UserPromptSubmit: []HookConfigItem{
			{Command: "echo prompt"},
		},
		Stop: []HookConfigItem{
			{Command: "echo stop", Timeout: 5},
		},
	}

	manager := LoadHooksFromConfig(config)

	if len(manager.GetHooks(HookPreToolUse)) != 1 {
		t.Error("expected 1 pre_tool_use hook")
	}
	if len(manager.GetHooks(HookPostToolUse)) != 1 {
		t.Error("expected 1 post_tool_use hook")
	}
	if len(manager.GetHooks(HookUserPromptSubmit)) != 1 {
		t.Error("expected 1 user_prompt_submit hook")
	}
	if len(manager.GetHooks(HookStop)) != 1 {
		t.Error("expected 1 stop hook")
	}
}

func TestHookError(t *testing.T) {
	err := &HookError{
		HookType: HookPreToolUse,
		Command:  "echo test",
		Err:      &testError{msg: "test error"},
	}

	if err.Error() != "hook pre_tool_use failed for command 'echo test': test error" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if err.Unwrap().Error() != "test error" {
		t.Errorf("unexpected unwrapped error: %s", err.Unwrap().Error())
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
