---
tags:
  - 核心模块
  - 工具系统
  - 工具执行
created: 2026-07-14
updated: 2026-07-14
aliases:
  - Tool
  - 工具注册
  - 工具执行
---

# 工具系统

> 🔧 工具注册、执行引擎、修复流水线、并行调度
> 📅 最后更新：2026-07-14

---

## 概述

工具系统是 Agent 与外部世界交互的桥梁，负责工具注册、执行、修复和并行调度。包含文件操作、Git、命令执行、搜索、会话、记忆等 12+ 内置工具。

**代码路径**：`internal/tool/`

## 关键文件

| 文件 | 职责 |
|------|------|
| `tool.go` | `Tool` 接口定义 |
| `registry.go` | 工具注册中心 |
| `executor.go` | 执行引擎（分区+守卫链+并行） |
| `repair.go` | 工具调用修复流水线 |
| `file_tools.go` | Read/Write/Edit 文件工具 |
| `command_tools.go` | Bash/Grep/Glob 工具 |
| `git_tools.go` | Git 工具（status/diff/log/commit） |
| `session_tools.go` | 跨会话上下文工具 |
| `memory_tool.go` | 记忆读取工具 |
| `task_tool.go` | 任务管理工具 |
| `skill_tool.go` | Skill 调用工具 |
| `websearch.go` | Web 搜索/抓取工具 |
| `checkpoint.go` | 编辑快照管理器 |
| `hooks.go` | 钩子管理器 |
| `diagnoser.go` | Go 诊断器 |
| `review.go` | 代码审查工具 |
| `repair.go` | RepairPipeline |
| `platform_*.go` | 跨平台进程管理（Unix/Windows） |

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

详见：[[../interfaces/tool-接口|Tool 接口]]

## 内置工具清单

| 工具名 | 类型 | 说明 |
|--------|------|------|
| `read_file` | 只读 | 读取文件内容 |
| `write_file` | 写入 | 创建或覆盖文件 |
| `edit_file` | 写入 | 精确字符串替换 |
| `bash` | 写入 | 执行 Shell 命令 |
| `grep` | 只读 | 内容搜索（Go 原生 ripgrep） |
| `glob` | 只读 | 文件模式匹配 |
| `git_status` | 只读 | Git 状态查询 |
| `git_diff` | 只读 | Git 差异查看 |
| `git_log` | 只读 | Git 日志查询 |
| `git_commit` | 写入 | Git 提交 |
| `recall_memory` | 只读 | 读取记忆（占位，需注入 MemoryProvider） |
| `list_sessions` | 只读 | 列出历史会话（需注入 SessionManager） |
| `read_session` | 只读 | 读取会话消息（需注入 SessionManager） |

## 执行引擎流程

```
ToolExecutor.Execute(toolCalls)
  │
  ├── 1. 分区 (Partition)
  │     ├── 连续只读工具 → ReadGroup（并行）
  │     └── 写工具        → WriteSeq（串行）
  │
  ├── 2. 执行守卫链（每个工具独立）
  │     ├── 工具存在检查
  │     ├── 重复成功阻断（同一写操作 ≥2 次）
  │     ├── Plan 模式阻断（非只读工具）
  │     ├── 权限门控（用户确认）
  │     └── 沙箱执行
  │
  ├── 3. 并行调度
  │     ├── ReadGroup 内工具并行执行
  │     └── WriteSeq 间串行
  │
  └── 4. 结果收集与截断（单结果最大 32KB）
```

## 工具调用修复流水线

专门处理 DeepSeek/MiMo 已知的工具调用 JSON 格式问题：

| 阶段 | 说明 |
|------|------|
| `flatten` | 深层嵌套参数扁平化（参数>10 或 深度>2） |
| `scavenge` | 从 reasoning_content 回收遗漏的工具调用 |
| `truncation` | 检测截断 JSON，补全括号 |
| `storm` | 滑动窗口检测重复调用，抑制风暴 |

## 编辑快照安全网

写文件/编辑文件前自动创建快照，存储在 `~/.loomcode/checkpoints/`：

- 最多保留 100 个快照
- `meta.json` 记录原路径、时间、触发工具、原文件信息
- `/rewind` 命令恢复：`/rewind`（列表）/ `/rewind last` / `/rewind <id>`

## 钩子系统

| 钩子类型 | 触发时机 |
|----------|----------|
| `HookPreExecute` | 工具执行前 |
| `HookPostExecute` | 工具执行后 |

内置钩子：`auto-gofmt`（write_file/edit_file 后自动格式化 .go 文件）

## 相关文档

- [[../interfaces/tool-接口|Tool 接口]] — 完整接口定义
- [[agent-引擎|Agent 引擎]] — 调用工具执行
- [[control-层|控制层]] — 权限门控
- [[session-管理|会话管理]] — 跨会话工具
- [[../architecture/架构总览|架构总览]]
