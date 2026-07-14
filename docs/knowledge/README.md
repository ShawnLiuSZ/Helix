---
tags:
  - 知识库说明
  - README
created: 2026-07-14
updated: 2026-07-14
---

# LoomCode 知识库

> 📚 面向智能体快速参考的项目知识库
> 📅 最后更新：2026-07-14

---

## 说明

本知识库为 LoomCode 项目的结构化文档体系，面向智能体（ClaudeCode、Codex、LoomCode 等）快速参考，重点在索引与关键路径。

## 目录结构

```
docs/knowledge/
├── AGENTS.md              ← 🤖 智能体专属索引（优先阅读）
├── 00-Index.md            ← 知识库首页
├── README.md              ← 本说明文件
├── MOC/                   ← Map of Content 索引层
│   ├── MOC-全部.md
│   ├── MOC-技术架构.md
│   ├── MOC-核心模块.md
│   └── MOC-接口协议.md
├── architecture/          ← 技术架构
│   └── 架构总览.md
├── modules/               ← 核心模块文档
│   ├── agent-引擎.md
│   ├── provider-层.md
│   ├── tool-系统.md
│   ├── config-系统.md
│   ├── control-层.md
│   ├── session-管理.md
│   ├── mcp-插件.md
│   ├── ui-TUI.md
│   ├── dashboard.md
│   ├── lsp-集成.md
│   └── skills-管理.md
└── interfaces/            ← 接口协议
    ├── provider-接口.md
    ├── tool-接口.md
    └── mcp-协议.md
```

## 使用方式

- **智能体**：优先阅读 [AGENTS.md](AGENTS.md)
- **新人**：从 [00-Index.md](00-Index.md) 开始浏览
- **查找特定模块**：直接访问 `modules/` 下对应文档

## 规范

- 文档格式：Markdown + YAML frontmatter（Obsidian 规范）
- 内部链接：使用 `[[Wikilink]]` 双向链接
- 时间：北京时间（UTC+8）
- 代码引用：以实际代码为准（架构设计文档中的部分模块已调整）
