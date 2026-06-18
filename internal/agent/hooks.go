package agent

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// HookType 钩子类型
type HookType int

const (
	// PreToolUse 工具执行前
	HookPreToolUse HookType = iota
	// PostToolUse 工具执行后
	HookPostToolUse
	// UserPromptSubmit 用户输入后
	HookUserPromptSubmit
	// Stop Agent 停止时
	HookStop
)

// String 返回钩子类型字符串
func (h HookType) String() string {
	switch h {
	case HookPreToolUse:
		return "pre_tool_use"
	case HookPostToolUse:
		return "post_tool_use"
	case HookUserPromptSubmit:
		return "user_prompt_submit"
	case HookStop:
		return "stop"
	default:
		return "unknown"
	}
}

// ParseHookType 解析钩子类型
func ParseHookType(s string) HookType {
	switch strings.ToLower(s) {
	case "pre_tool_use":
		return HookPreToolUse
	case "post_tool_use":
		return HookPostToolUse
	case "user_prompt_submit":
		return HookUserPromptSubmit
	case "stop":
		return HookStop
	default:
		return -1
	}
}

// HookContext 钩子上下文
type HookContext struct {
	// Type 钩子类型
	Type HookType
	// ToolName 工具名称（仅工具相关钩子）
	ToolName string
	// ToolArgs 工具参数（仅工具相关钩子）
	ToolArgs map[string]any
	// ToolResult 工具结果（仅 PostToolUse）
	ToolResult string
	// ToolError 工具错误（仅 PostToolUse）
	ToolError error
	// Prompt 用户输入（仅 UserPromptSubmit）
	Prompt string
	// Env 环境变量
	Env []string
	// WorkingDir 工作目录
	WorkingDir string
}

// Hook 钩子接口
type Hook interface {
	// Execute 执行钩子
	Execute(ctx HookContext) error
	// Type 返回钩子类型
	Type() HookType
}

// ShellHook Shell 命令钩子
type ShellHook struct {
	hookType HookType
	command  string
	timeout  time.Duration
}

// NewShellHook 创建 Shell 钩子
func NewShellHook(hookType HookType, command string) *ShellHook {
	return &ShellHook{
		hookType: hookType,
		command:  command,
		timeout:  30 * time.Second, // 默认 30 秒超时
	}
}

// Execute 执行 Shell 钩子
func (h *ShellHook) Execute(ctx HookContext) error {
	// 替换变量
	command := h.command
	command = strings.ReplaceAll(command, "{{.ToolName}}", ctx.ToolName)
	command = strings.ReplaceAll(command, "{{.Prompt}}", ctx.Prompt)
	command = strings.ReplaceAll(command, "{{.ToolResult}}", ctx.ToolResult)

	// 创建带超时的上下文
	ctxTimeout, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(ctxTimeout, "sh", "-c", command)
	cmd.Env = ctx.Env
	if ctx.WorkingDir != "" {
		cmd.Dir = ctx.WorkingDir
	}

	// 执行
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook command failed: %s, output: %s", err, string(output))
	}

	return nil
}

// Type 返回钩子类型
func (h *ShellHook) Type() HookType {
	return h.hookType
}

// HookManager 钩子管理器
type HookManager struct {
	mu    sync.RWMutex
	hooks map[HookType][]Hook
}

// NewHookManager 创建钩子管理器
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[HookType][]Hook),
	}
}

// Register 注册钩子
func (m *HookManager) Register(hook Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hookType := hook.Type()
	m.hooks[hookType] = append(m.hooks[hookType], hook)
}

// Unregister 注销钩子
func (m *HookManager) Unregister(hookType HookType, index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hooks, ok := m.hooks[hookType]; ok && index < len(hooks) {
		m.hooks[hookType] = append(hooks[:index], hooks[index+1:]...)
	}
}

// Execute 执行指定类型的钩子
func (m *HookManager) Execute(hookType HookType, ctx HookContext) error {
	m.mu.RLock()
	hooks := m.hooks[hookType]
	m.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.Execute(ctx); err != nil {
			return fmt.Errorf("hook %v failed: %w", hookType, err)
		}
	}

	return nil
}

// ExecuteWithFallback 执行钩子，遇到错误时记录但继续
func (m *HookManager) ExecuteWithFallback(hookType HookType, ctx HookContext) []error {
	m.mu.RLock()
	hooks := m.hooks[hookType]
	m.mu.RUnlock()

	var errors []error
	for _, hook := range hooks {
		if err := hook.Execute(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// GetHooks 获取指定类型的所有钩子
func (m *HookManager) GetHooks(hookType HookType) []Hook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hooks[hookType]
}

// Clear 清除所有钩子
func (m *HookManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = make(map[HookType][]Hook)
}

// HookConfig 钩子配置
type HookConfig struct {
	// PreToolUse 工具执行前钩子
	PreToolUse []HookConfigItem `toml:"pre_tool_use"`
	// PostToolUse 工具执行后钩子
	PostToolUse []HookConfigItem `toml:"post_tool_use"`
	// UserPromptSubmit 用户输入后钩子
	UserPromptSubmit []HookConfigItem `toml:"user_prompt_submit"`
	// Stop Agent 停止时钩子
	Stop []HookConfigItem `toml:"stop"`
}

// HookConfigItem 钩子配置项
type HookConfigItem struct {
	// Command Shell 命令
	Command string `toml:"command"`
	// Timeout 超时时间（秒）
	Timeout int `toml:"timeout"`
}

// LoadHooksFromConfig 从配置加载钩子
func LoadHooksFromConfig(config HookConfig) *HookManager {
	manager := NewHookManager()

	// 加载 PreToolUse 钩子
	for _, item := range config.PreToolUse {
		hook := NewShellHook(HookPreToolUse, item.Command)
		if item.Timeout > 0 {
			hook.timeout = time.Duration(item.Timeout) * time.Second
		}
		manager.Register(hook)
	}

	// 加载 PostToolUse 钩子
	for _, item := range config.PostToolUse {
		hook := NewShellHook(HookPostToolUse, item.Command)
		if item.Timeout > 0 {
			hook.timeout = time.Duration(item.Timeout) * time.Second
		}
		manager.Register(hook)
	}

	// 加载 UserPromptSubmit 钩子
	for _, item := range config.UserPromptSubmit {
		hook := NewShellHook(HookUserPromptSubmit, item.Command)
		if item.Timeout > 0 {
			hook.timeout = time.Duration(item.Timeout) * time.Second
		}
		manager.Register(hook)
	}

	// 加载 Stop 钩子
	for _, item := range config.Stop {
		hook := NewShellHook(HookStop, item.Command)
		if item.Timeout > 0 {
			hook.timeout = time.Duration(item.Timeout) * time.Second
		}
		manager.Register(hook)
	}

	return manager
}

// HookError 钩子错误
type HookError struct {
	HookType HookType
	Command  string
	Err      error
}

func (e *HookError) Error() string {
	return fmt.Sprintf("hook %s failed for command '%s': %v", e.HookType, e.Command, e.Err)
}

func (e *HookError) Unwrap() error {
	return e.Err
}
