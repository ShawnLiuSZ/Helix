---
tags:
  - 核心模块
  - Agent引擎
  - 推理循环
created: 2026-07-14
updated: 2026-07-14
aliases:
  - Agent
  - 推理循环
---

# Agent 引擎

> 🧠 推理循环、模式切换、子Agent编排
> 📅 最后更新：2026-07-14

---

## 概述

Agent 引擎是 LoomCode 的应用层核心，负责推理循环、工具调用编排、多 Agent 协作。基于 Provider 的 `Capabilities` 声明自适应启用修复流水线、缓存调度等能力。

**代码路径**：`internal/agent/`

## 关键文件

| 文件 | 职责 |
|------|------|
| `loop.go` | 核心 Agent 结构与推理循环 |
| `modes.go` | Build/Plan/Compose/Max 模式 |
| `coordinator.go` | 多 Agent 协调器 |
| `scheduler.go` | 缓存调度器（CacheTTL-aware） |
| `subagent.go` | 子 Agent 管理 |
| `goal.go` | 目标停止条件评估 |
| `effort.go` | 思考强度管理器 |
| `context.go` | 上下文构建 |
| `fingerprint.go` | Prefix 指纹追踪（验证 cache 命中） |
| `eventlog.go` | 事件日志（缓存命中统计） |
| `message_bus.go` | 消息总线 |

## Agent 核心结构

```go
type Agent struct {
    provider    provider.Provider      // 模型提供者
    tools       *tool.Registry         // 工具注册中心
    executor    *tool.Executor         // 工具执行器
    messages    []provider.Message     // 消息历史
    maxSteps    int                    // 最大推理步数
    model       string                 // 模型 ID
    goal        *GoalStopCondition     // 停止条件
    skillsMgr   *skills.Manager        // Skills 管理器
    memory      MemoryProvider         // 记忆提供者
    effort      *EffortManager         // 思考强度
    eventLog    *EventLog              // 事件日志
    fingerprint *FingerprintTracker    // Prefix 指纹追踪
    cacheScheduler *CacheScheduler     // 缓存调度器（可选）
    repairPipeline *tool.RepairPipeline // 工具修复流水线（可选）
}
```

## 推理循环流程

```
用户输入
    │
    ▼
构建消息上下文（system prompt + 工具 schema + 历史）
    │
    ▼
Provider.Stream()  ← 流式推理
    │
    ├─ 有 tool_calls
    │     │
    │     ├─ RepairPipeline（如启用）← 修复 JSON 格式错误
    │     ├─ ToolExecutor             ← 并行/串行执行工具
    │     ├─ 结果追加到消息历史
    │     └─ 回到"构建消息上下文"
    │
    └─ 无 tool_calls
          │
          ├─ Goal 停止条件检查
          └─ 返回最终答案
```

## Agent 模式

| 模式 | 权限 | 适用场景 | 触发 |
|------|------|----------|------|
| **Build** | 完整工具权限 | 日常开发（默认） | 默认 |
| **Plan** | 只读分析 | 代码探索、方案设计 | `/plan` |
| **Compose** | 编排模式 | 规格驱动开发 | `/compose` |
| **Max** | 并行选优 | 高难度任务 | `experimental.maxMode` |

## 自适应能力启用

Agent 在创建时根据 Provider 的 `Capabilities` 自动启用可选能力：

| 能力 | 启用条件 | 对应组件 |
|------|----------|----------|
| 缓存调度 | `Capabilities.CacheTTL > 0` | `CacheScheduler` |
| 工具修复 | `Capabilities.NeedsToolRepair == true` | `RepairPipeline` |

## 关键方法

| 方法 | 说明 |
|------|------|
| `New(p, registry)` | 创建 Agent，自动启用可选能力 |
| `Run(ctx, task)` | 执行单次任务 |
| `SetMaxSteps(n)` | 设置最大推理步数 |
| `SetModel(id)` | 设置模型 |
| `SetMemory(m)` | 注入记忆提供者 |
| `SetHooks(hm)` | 注入钩子管理器 |
| `SetReadOnlyTools(b)` | 仅暴露只读工具（Plan 模式） |

## 相关文档

- [[provider-层|Provider 层]] — 提供模型推理能力
- [[tool-系统|工具系统]] — 提供工具执行能力
- [[../interfaces/provider-接口|Provider 接口]] — Capabilities 定义
- [[../architecture/架构总览|架构总览]]
