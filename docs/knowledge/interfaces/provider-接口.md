---
tags:
  - 接口协议
  - Provider接口
  - Adapter
created: 2026-07-14
updated: 2026-07-14
aliases:
  - Provider Interface
  - Adapter Interface
---

# Provider 接口

> 🔌 模型提供者抽象接口定义
> 📅 最后更新：2026-07-14

---

## 定义文件

`internal/provider/provider.go`

## Provider 接口

```go
type Provider interface {
    // Chat 非流式对话
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // Stream 流式对话
    Stream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error)

    // 元信息
    Name() string
    Models() []ModelInfo
    Capabilities() Capabilities

    // Cost 计算成本
    Cost(modelID string, usage Usage) Cost
}
```

## Adapter 工厂接口

```go
type Adapter interface {
    // Kind 返回适配器类型标识（对应配置中的 kind 字段）
    Kind() string

    // Create 根据配置创建 Provider 实例
    Create(cfg Config) (Provider, error)

    // ValidateConfig 验证配置是否合法
    ValidateConfig(cfg Config) error
}
```

## Capabilities 能力声明

```go
type Capabilities struct {
    SupportsReasoning    bool          // 支持 reasoning_content
    SupportsToolCall     bool          // 支持原生工具调用
    SupportsPrefixCache  bool          // 支持前缀缓存
    SupportsStreaming    bool          // 支持流式输出
    SupportsVision       bool          // 支持图片输入
    SupportsVoice        bool          // 支持语音输入
    SupportsOAuth        bool          // 支持 OAuth 认证
    NeedsToolRepair      bool          // 需要工具调用修复
    CacheTTL             time.Duration // 缓存有效期
    MaxToolCallsPerRound int           // 单轮最大工具调用数
}
```

## 核心数据结构

### ChatRequest

| 字段 | 类型 | 说明 |
|------|------|------|
| `Model` | `string` | 模型 ID |
| `Messages` | `[]Message` | 消息列表 |
| `Tools` | `[]ToolSchema` | 工具 Schema 列表 |
| `Stream` | `bool` | 是否流式 |
| `Temperature` | `float64` | 温度参数 |
| `MaxTokens` | `int` | 最大 token 数 |

### ChatResponse

| 字段 | 类型 | 说明 |
|------|------|------|
| `Content` | `string` | 响应内容 |
| `ToolCalls` | `[]ToolCall` | 工具调用列表 |
| `Usage` | `Usage` | Token 使用量 |
| `ReasoningContent` | `string` | 推理过程内容 |

### StreamEvent

流式事件通过 channel 传递，包含增量内容、工具调用片段等。

### Cost

| 字段 | 类型 | 说明 |
|------|------|------|
| `Input` | `float64` | 输入成本 |
| `CachedInput` | `float64` | 缓存输入成本 |
| `Output` | `float64` | 输出成本 |
| `Total` | `float64` | 总成本 |

## 内置适配器实现

| 适配器 | Kind | 文件 | 特性 |
|--------|------|------|------|
| `openai.Adapter` | `"openai"` | `openai/provider.go` | 通用 OpenAI 兼容 |
| `deepseek.Adapter` | `"deepseek"` | `deepseek/provider.go` | Prefix Cache + 工具修复 |
| `mimo.Adapter` | `"mimo"` | `mimo/provider.go` | OAuth + 语音 |

## Registry 注册中心

```go
reg := provider.NewRegistry()
reg.Register(&openai.Adapter{})
reg.Register(&deepseek.Adapter{})
reg.Register(&mimo.Adapter{})

p, err := reg.Create(kind, config)
```

## 自适应行为

Agent 根据 `Capabilities` 自动调整策略：

| 能力 | 启用条件 | Agent 行为 |
|------|----------|------------|
| 缓存调度 | `CacheTTL > 0` | 启用 `CacheScheduler` |
| 工具修复 | `NeedsToolRepair == true` | 启用 `RepairPipeline` |
| 前缀缓存 | `SupportsPrefixCache` | 构建缓存稳定的上下文分区 |

## 扩展新适配器

如需非 OpenAI 兼容的厂商，实现 `Adapter` 接口并注册：

```go
type MyAdapter struct{}

func (a *MyAdapter) Kind() string { return "my-vendor" }
func (a *MyAdapter) Create(cfg provider.Config) (provider.Provider, error) { ... }
func (a *MyAdapter) ValidateConfig(cfg provider.Config) error { ... }

// 注册
reg.Register(&MyAdapter{})
```

## 相关文档

- [[../modules/provider-层|Provider 层]] — 模块概览
- [[../modules/agent-引擎|Agent 引擎]] — 消费 Provider 能力
- [[../architecture/架构总览|架构总览]]
