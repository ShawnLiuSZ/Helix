---
tags:
  - 核心模块
  - TUI
  - BubbleTea
created: 2026-07-14
updated: 2026-07-14
aliases:
  - UI
  - TUI
  - BubbleTea
---

# TUI 界面

> 🎨 Bubble Tea 交互式终端界面
> 📅 最后更新：2026-07-14

---

## 概述

TUI 基于 Bubble Tea + Lip Gloss 构建，提供交互式聊天界面、命令系统、模型切换、会话管理等功能。

**代码路径**：`internal/ui/`

## 关键文件

| 文件 | 职责 |
|------|------|
| `app.go` | TUI 主应用（Bubble Tea Model） |
| `remember.go` | `/remember` 长期记忆写入 |

## 交互操作

| 操作 | 说明 |
|------|------|
| 直接输入文字 | 发送任务给 AI |
| Tab | 切换 Agent 模式（build/plan/compose/max） |
| 输入 `/` 后 Tab | 命令自动补全 |
| ↑↓ / Enter / Esc | 交互式选择器（模型选择等） |
| Shift+Enter | 换行 |
| Ctrl+C 两次 | 退出（3 秒内二次确认） |

## TUI 命令

| 命令 | 说明 |
|------|------|
| `/help` | 显示帮助 |
| `/mode` | 显示当前模式和模型 |
| `/build` `/plan` `/compose` `/max` | 切换 Agent 模式 |
| `/model` | 交互式选择模型 |
| `/model <name>` | 直接切换模型 |
| `/rewind` | 列出/恢复编辑快照 |
| `/skills` | 显示内置工具和外部 skills |
| `/env` | 查看环境变量 |
| `/env set <KEY> <VALUE>` | 设置环境变量 |
| `/goal` | 设置/查看/清除停止条件 |
| `/sessions` | 查看会话列表 |
| `/sessions new <name>` | 创建新会话 |
| `/sessions switch <ID>` | 切换会话 |
| `/remember <text>` | 写入长期记忆 |
| `/cost` | 显示成本统计 |
| `/budget <amount>` | 设置会话预算上限 |
| `/compact` | 压缩上下文 |
| `/clear` | 清空聊天记录 |
| `/quit` | 退出 |

## 关键方法

| 方法 | 说明 |
|------|------|
| `NewApp(p, tools)` | 创建 TUI 应用 |
| `SetModel(id)` | 设置当前模型 |
| `SetProviders(providers)` | 设置可用 Provider 列表（跨 Provider 切换） |
| `AddApprovalGuard(perm)` | 注入权限守卫 |
| `SetCheckpointManager(cpMgr)` | 注入快照管理器 |
| `SetHooks(hm)` | 注入钩子管理器 |
| `SetSessionManager(mgr)` | 注入会话管理器 |
| `RestoreSession(sess)` | 恢复会话（消息历史） |
| `RestoreModelFromSession(sess)` | 仅恢复模型/Provider 选择 |

## 信号处理

捕获 `SIGTERM`/`SIGHUP`，终端关闭或 kill 时通知 TUI 保存退出，避免对话内容丢失。

## 相关文档

- [[agent-引擎|Agent 引擎]] — TUI 调用 Agent 执行
- [[session-管理|会话管理]] — 会话恢复
- [[control-层|控制层]] — 权限守卫
- [[../architecture/架构总览|架构总览]]
