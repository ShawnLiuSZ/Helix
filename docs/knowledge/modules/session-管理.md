---
tags:
  - 核心模块
  - 会话管理
  - JSONL
created: 2026-07-14
updated: 2026-07-14
aliases:
  - Session
  - 会话持久化
---

# 会话管理

> 💾 会话生命周期、JSONL 持久化、加密、重放
> 📅 最后更新：2026-07-14

---

## 概述

会话管理负责对话生命周期的管理，包括创建、切换、恢复、持久化。采用 JSONL 格式存储，支持会话加密和重放。

**代码路径**：`internal/session/`

## 关键文件

| 文件 | 职责 |
|------|------|
| `session.go` | `Session` 结构体、`Manager` 管理器 |
| `crypto.go` | 会话加密 |
| `replay.go` | 会话重放 |

## 存储位置

```
~/.loomcode/sessions/
├── <session_id>.jsonl    ← JSONL 格式消息历史
└── ...
```

## 会话生命周期

```
创建会话
    │
    ├── 每轮追加: [user, assistant, tool_call, tool_result]
    │
    ├── 切换会话（--session <id>）
    │     └── Manager.Get(id) → App.RestoreSession()
    │
    └── 恢复最近会话
          └── Manager.MostRecent() → RestoreModelFromSession()
```

## 跨会话上下文工具

Agent 可通过以下工具访问历史会话：

| 工具 | 说明 |
|------|------|
| `list_sessions` | 列出最近会话（可按 limit/role 过滤） |
| `read_session` | 读取指定会话的完整消息历史 |

工具默认注册为占位（返回 "No session configured"），需通过 `SetSessionManagerForTools()` 注入真正的 SessionManager。

## 关键方法

| 方法 | 说明 |
|------|------|
| `NewManager(dir)` | 创建会话管理器 |
| `Manager.Get(id)` | 获取指定会话 |
| `Manager.MostRecent()` | 获取最近会话 |
| `Manager.SetActive(id)` | 设置活跃会话 |
| `SetSessionManagerForTools(tools, mgr)` | 注入到工具系统 |

## CLI 用法

```bash
# 恢复指定会话
loomcode --session session_1234567890

# TUI 中管理会话
/sessions              # 查看会话列表
/sessions new <name>   # 创建新会话
/sessions switch <ID>  # 切换会话
```

## 相关文档

- [[tool-系统|工具系统]] — 跨会话工具注册
- [[ui-TUI|TUI 界面]] — 会话恢复交互
- [[../architecture/架构总览|架构总览]]
