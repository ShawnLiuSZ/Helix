---
tags:
  - 核心模块
  - Dashboard
  - Web
created: 2026-07-14
updated: 2026-07-14
aliases:
  - Dashboard
  - Web监控
---

# Dashboard

> 📊 Web Dashboard、HTTP + WebSocket 实时监控
> 📅 最后更新：2026-07-14

---

## 概述

Dashboard 是 LoomCode 的 Web 监控面板，通过 HTTP + WebSocket 提供实时状态查看。

**代码路径**：`internal/dashboard/`

## 关键文件

| 文件 | 职责 |
|------|------|
| `server.go` | HTTP 服务器 |
| `handlers.go` | 请求处理器 |
| `websocket.go` | WebSocket 实时通信 |
| `embed.go` | 静态资源嵌入 |
| `static/` | 前端静态文件（index.html/app.js/style.css） |

## 启动方式

```bash
# 默认端口 8080
loomcode dashboard

# 指定端口
loomcode dashboard :9090
```

## 技术要点

| 特性 | 说明 |
|------|------|
| 静态资源嵌入 | `embed.go` 将前端文件嵌入二进制 |
| WebSocket | 实时推送状态更新 |
| 信号处理 | 捕获 SIGTERM 优雅关闭 |

## 相关文档

- [[../architecture/架构总览|架构总览]]
