---
tags:
  - 接口协议
  - Tool接口
  - Schema
created: 2026-07-14
updated: 2026-07-14
aliases:
  - Tool Interface
  - 工具接口
---

# Tool 接口

> 🔧 工具抽象接口定义
> 📅 最后更新：2026-07-14

---

## 定义文件

`internal/tool/tool.go`

## Tool 接口

```go
type Tool interface {
    Name() string
    Description() string
    Schema() Schema           // OpenAI function 格式
    IsReadOnly() bool         // 用于并行调度
    Execute(ctx context.Context, args map[string]any) (*Result, error)
}
```

## Schema 结构

```go
type Schema struct {
    Type       string              `json:"type"`       // 固定 "object"
    Properties map[string]Property `json:"properties"` // 参数定义
    Required   []string            `json:"required,omitempty"` // 必填参数
}

type Property struct {
    Type        string `json:"type"`
    Description string `json:"description"`
}
```

## Result 结构

```go
type Result struct {
    Content string
    Error   string
}

func (r *Result) OK() bool { return r.Error == "" }
```

## Registry 注册中心

```go
type Registry struct {
    tools map[string]Tool
}

// 注册工具
func (r *Registry) Register(t Tool) error

// 获取工具
func (r *Registry) Get(name string) (Tool, bool)

// 列出所有工具（按名称升序，保证 prefix cache 稳定）
func (r *Registry) List() []Tool
```

**设计要点**：`List()` 按名称升序返回，保证 `tools` 数组顺序一致，从而命中 LLM prefix cache。

## RegisterDefaults 内置注册

```go
func (r *Registry) RegisterDefaults() {
    defaults := []Tool{
        &ReadFileTool{},
        &WriteFileTool{},
        &EditFileTool{},
        &BashTool{},
        &GrepTool{},
        &GlobTool{},
        &GitStatusTool{},
        &GitDiffTool{},
        &GitLogTool{},
        &GitCommitTool{},
        &RecallMemoryTool{},
        &ListSessionsTool{},
        &ReadSessionTool{},
    }
    // ...
}
```

## Executor 执行引擎

```go
type Executor struct {
    registry *Registry
}

func (e *Executor) Execute(ctx context.Context, calls []Call) []Result
```

执行流程：分区 → 守卫链 → 并行/串行调度 → 结果截断

## EnvProvider 接口

```go
type EnvProvider interface {
    EnvForSubprocess() []string
}
```

由 CLI 注入，工具执行子进程时过滤环境变量（排除 API Key）。

## 实现新工具

```go
type MyTool struct{}

func (t *MyTool) Name() string { return "my_tool" }
func (t *MyTool) Description() string { return "做某事" }
func (t *MyTool) IsReadOnly() bool { return false }
func (t *MyTool) Schema() Schema {
    return Schema{
        Type: "object",
        Properties: map[string]Property{
            "path": {Type: "string", Description: "文件路径"},
        },
        Required: []string{"path"},
    }
}
func (t *MyTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
    path, _ := args["path"].(string)
    // ... 执行逻辑
    return &Result{Content: "完成"}, nil
}

// 注册
tools.Register(&MyTool{})
```

## 内置工具实现

| 工具 | 文件 | 类型 |
|------|------|------|
| `ReadFileTool` | `file_tools.go` | 只读 |
| `WriteFileTool` | `file_tools.go` | 写入 |
| `EditFileTool` | `file_tools.go` | 写入 |
| `BashTool` | `command_tools.go` | 写入 |
| `GrepTool` | `command_tools.go` | 只读 |
| `GlobTool` | `command_tools.go` | 只读 |
| `GitStatusTool` | `git_tools.go` | 只读 |
| `GitDiffTool` | `git_tools.go` | 只读 |
| `GitLogTool` | `git_tools.go` | 只读 |
| `GitCommitTool` | `git_tools.go` | 写入 |
| `RecallMemoryTool` | `memory_tool.go` | 只读 |
| `ListSessionsTool` | `session_tools.go` | 只读 |
| `ReadSessionTool` | `session_tools.go` | 只读 |

## 相关文档

- [[../modules/tool-系统|工具系统]] — 模块概览
- [[../modules/agent-引擎|Agent 引擎]] — 调用工具执行
- [[../modules/control-层|控制层]] — 权限门控
- [[../architecture/架构总览|架构总览]]
