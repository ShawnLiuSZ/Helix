package tool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestHookManager_AddRemove(t *testing.T) {
	m := NewHookManager()

	m.Add(Hook{Name: "h1", Type: HookPreExecute, ToolName: "*", Handler: func(ctx context.Context, call Call, result *Result) error {
		return nil
	}})
	m.Add(Hook{Name: "h2", Type: HookPostExecute, ToolName: "write_file", Handler: func(ctx context.Context, call Call, result *Result) error {
		return nil
	}})

	m.mu.RLock()
	if len(m.hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(m.hooks))
	}
	m.mu.RUnlock()

	m.Remove("h1")

	m.mu.RLock()
	if len(m.hooks) != 1 {
		t.Fatalf("expected 1 hook after remove, got %d", len(m.hooks))
	}
	if m.hooks[0].Name != "h2" {
		t.Errorf("expected remaining hook to be h2, got %s", m.hooks[0].Name)
	}
	m.mu.RUnlock()
}

func TestHookManager_RunPreHooks(t *testing.T) {
	var called int32
	m := NewHookManager()
	m.Add(Hook{
		Name:     "counter",
		Type:     HookPreExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			atomic.AddInt32(&called, 1)
			return nil
		},
	})

	err := m.RunPreHooks(context.Background(), Call{Name: "any_tool"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("expected hook called once, got %d", called)
	}
}

func TestHookManager_RunPreHooks_FilterByToolName(t *testing.T) {
	var called int32
	m := NewHookManager()
	m.Add(Hook{
		Name:     "specific",
		Type:     HookPreExecute,
		ToolName: "write_file",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			atomic.AddInt32(&called, 1)
			return nil
		},
	})

	m.RunPreHooks(context.Background(), Call{Name: "read_file"})
	if atomic.LoadInt32(&called) != 0 {
		t.Error("hook should not have been called for read_file")
	}

	m.RunPreHooks(context.Background(), Call{Name: "write_file"})
	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("expected hook called once for write_file, got %d", called)
	}
}

func TestHookManager_RunPreHooks_ErrorStopsExecution(t *testing.T) {
	var secondCalled int32
	m := NewHookManager()
	m.Add(Hook{
		Name:     "fail",
		Type:     HookPreExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			return errors.New("blocked")
		},
	})
	m.Add(Hook{
		Name:     "should-not-run",
		Type:     HookPreExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			atomic.AddInt32(&secondCalled, 1)
			return nil
		},
	})

	err := m.RunPreHooks(context.Background(), Call{Name: "anything"})
	if err == nil || err.Error() != "blocked" {
		t.Fatalf("expected 'blocked' error, got %v", err)
	}
	if atomic.LoadInt32(&secondCalled) != 0 {
		t.Error("second hook should not have been called")
	}
}

func TestHookManager_RunPostHooks(t *testing.T) {
	var capturedContent string
	m := NewHookManager()
	m.Add(Hook{
		Name:     "capture",
		Type:     HookPostExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			capturedContent = result.Content
			return nil
		},
	})

	result := &Result{Content: "hello"}
	err := m.RunPostHooks(context.Background(), Call{Name: "some_tool"}, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedContent != "hello" {
		t.Errorf("expected captured content 'hello', got %q", capturedContent)
	}
}

func TestHookManager_RunPostHooks_Error(t *testing.T) {
	m := NewHookManager()
	m.Add(Hook{
		Name:     "fail-post",
		Type:     HookPostExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			return errors.New("post failed")
		},
	})

	result := &Result{Content: "ok"}
	err := m.RunPostHooks(context.Background(), Call{Name: "tool"}, result)
	if err == nil || err.Error() != "post failed" {
		t.Fatalf("expected 'post failed' error, got %v", err)
	}
}

func TestHookManager_NoHooksNoop(t *testing.T) {
	m := NewHookManager()

	err := m.RunPreHooks(context.Background(), Call{Name: "tool"})
	if err != nil {
		t.Fatalf("expected nil error with no hooks, got %v", err)
	}

	err = m.RunPostHooks(context.Background(), Call{Name: "tool"}, &Result{})
	if err != nil {
		t.Fatalf("expected nil error with no hooks, got %v", err)
	}
}

func TestExecutor_WithHooks(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubReadTool{})

	hm := NewHookManager()

	var preCount, postCount int32
	hm.Add(Hook{
		Name:     "pre",
		Type:     HookPreExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			atomic.AddInt32(&preCount, 1)
			return nil
		},
	})
	hm.Add(Hook{
		Name:     "post",
		Type:     HookPostExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			atomic.AddInt32(&postCount, 1)
			return nil
		},
	})

	e := NewExecutor(r)
	e.SetHooks(hm)

	results := e.Execute(context.Background(), []Call{
		{Name: "read_file", Args: map[string]any{"path": "/tmp"}},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK() {
		t.Errorf("unexpected error: %s", results[0].Error)
	}
	if atomic.LoadInt32(&preCount) != 1 {
		t.Errorf("expected pre-hook called once, got %d", preCount)
	}
	if atomic.LoadInt32(&postCount) != 1 {
		t.Errorf("expected post-hook called once, got %d", postCount)
	}
}

func TestExecutor_PreHookBlocksExecution(t *testing.T) {
	var toolCalled int32
	r := NewRegistry()
	r.Register(&stubReadTool{fn: func() { atomic.AddInt32(&toolCalled, 1) }})

	hm := NewHookManager()
	hm.Add(Hook{
		Name:     "block",
		Type:     HookPreExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			return errors.New("denied")
		},
	})

	e := NewExecutor(r)
	e.SetHooks(hm)

	results := e.Execute(context.Background(), []Call{
		{Name: "read_file", Args: map[string]any{"path": "/tmp"}},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK() {
		t.Error("expected error from pre-hook")
	}
	if results[0].Error != "pre-hook: denied" {
		t.Errorf("unexpected error message: %s", results[0].Error)
	}
	if atomic.LoadInt32(&toolCalled) != 0 {
		t.Error("tool should not have been called when pre-hook blocks")
	}
}

func TestExecutor_PostHookAppendsError(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubReadTool{})

	hm := NewHookManager()
	hm.Add(Hook{
		Name:     "post-fail",
		Type:     HookPostExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			return errors.New("post-error")
		},
	})

	e := NewExecutor(r)
	e.SetHooks(hm)

	results := e.Execute(context.Background(), []Call{
		{Name: "read_file", Args: map[string]any{"path": "/tmp"}},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK() {
		t.Error("expected error from post-hook")
	}
	if results[0].Error != "post-hook: post-error" {
		t.Errorf("unexpected error message: %s", results[0].Error)
	}
}

func TestExecutor_NoHooks(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubReadTool{})

	e := NewExecutor(r)

	results := e.Execute(context.Background(), []Call{
		{Name: "read_file", Args: map[string]any{"path": "/tmp"}},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK() {
		t.Errorf("unexpected error: %s", results[0].Error)
	}
}

func TestExecutor_PostHookAppendsToExistingError(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubFailTool{})

	hm := NewHookManager()
	hm.Add(Hook{
		Name:     "post-fail",
		Type:     HookPostExecute,
		ToolName: "*",
		Handler: func(ctx context.Context, call Call, result *Result) error {
			return errors.New("hook-err")
		},
	})

	e := NewExecutor(r)
	e.SetHooks(hm)

	results := e.Execute(context.Background(), []Call{
		{Name: "fail_tool", Args: map[string]any{}},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != "tool-error; post-hook: hook-err" {
		t.Errorf("unexpected combined error: %s", results[0].Error)
	}
}

type stubFailTool struct{}

func (s *stubFailTool) Name() string        { return "fail_tool" }
func (s *stubFailTool) Description() string { return "stub" }
func (s *stubFailTool) Schema() Schema      { return Schema{Type: "object"} }
func (s *stubFailTool) IsReadOnly() bool    { return true }
func (s *stubFailTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	return &Result{Error: "tool-error"}, nil
}
