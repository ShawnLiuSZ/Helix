---
tags:
  - 核心模块
  - LSP
  - 语言服务器
created: 2026-07-14
updated: 2026-07-14
aliases:
  - LSP
  - Language Server Protocol
---

# LSP 集成

> 🔍 LSP 客户端、JSON-RPC 2.0、诊断信息
> 📅 最后更新：2026-07-14

---

## 概述

LSP（Language Server Protocol）集成让 LoomCode 能与语言服务器通信，获取代码诊断、定义跳转等信息。

**代码路径**：`internal/lsp/`

## 关键文件

| 文件 | 职责 |
|------|------|
| `client.go` | LSP 客户端 |
| `discovery.go` | 语言服务器发现 |
| `protocol.go` | LSP 协议定义（JSON-RPC 2.0） |

## 技术要点

| 特性 | 说明 |
|------|------|
| 协议 | JSON-RPC 2.0 |
| 传输 | stdio（与语言服务器子进程通信） |
| 发现 | 自动发现系统已安装的语言服务器 |

## 相关文档

- [[../architecture/架构总览|架构总览]]
