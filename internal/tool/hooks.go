package tool

import (
	"context"
	"sync"
)

type HookType int

const (
	HookPreExecute HookType = iota
	HookPostExecute
)

type Hook struct {
	Name     string
	Type     HookType
	ToolName string
	Handler  func(ctx context.Context, call Call, result *Result) error
}

type HookManager struct {
	hooks []Hook
	mu    sync.RWMutex
}

func NewHookManager() *HookManager {
	return &HookManager{}
}

func (m *HookManager) Add(hook Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

func (m *HookManager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	filtered := m.hooks[:0]
	for _, h := range m.hooks {
		if h.Name != name {
			filtered = append(filtered, h)
		}
	}
	m.hooks = filtered
}

func (m *HookManager) RunPreHooks(ctx context.Context, call Call) error {
	m.mu.RLock()
	hooks := make([]Hook, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.RUnlock()

	for _, h := range hooks {
		if h.Type != HookPreExecute {
			continue
		}
		if h.ToolName != "*" && h.ToolName != call.Name {
			continue
		}
		if err := h.Handler(ctx, call, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *HookManager) RunPostHooks(ctx context.Context, call Call, result *Result) error {
	m.mu.RLock()
	hooks := make([]Hook, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.RUnlock()

	for _, h := range hooks {
		if h.Type != HookPostExecute {
			continue
		}
		if h.ToolName != "*" && h.ToolName != call.Name {
			continue
		}
		if err := h.Handler(ctx, call, result); err != nil {
			return err
		}
	}
	return nil
}
