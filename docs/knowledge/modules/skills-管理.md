---
tags:
  - 核心模块
  - Skills
  - 自动加载
created: 2026-07-14
updated: 2026-07-14
aliases:
  - Skills
  - 技能管理
---

# Skills 管理

> 🎯 Skill 自动加载与管理
> 📅 最后更新：2026-07-14

---

## 概述

Skills 管理器负责从文件系统自动加载 Skill 定义，扩展 Agent 的能力。每个 Skill 是一个包含 `SKILL.md` 的目录。

**代码路径**：`internal/skills/`

## 关键文件

| 文件 | 职责 |
|------|------|
| `manager.go` | `Manager` Skill 管理器 |

## 加载路径

```
~/.agents/skills/*/SKILL.md       ← 低优先级
~/.loomcode/skills/*/SKILL.md     ← 高优先级（同名覆盖）
```

## 使用方式

```bash
# 查看已加载的 skills
/skills
```

## 技术要点

| 特性 | 说明 |
|------|------|
| 自动加载 | 启动时扫描目录 |
| 优先级覆盖 | `~/.loomcode/skills/` 覆盖 `~/.agents/skills/` |
| Skill 结构 | 每个 Skill 是独立目录 + `SKILL.md` |

## 相关文档

- [[agent-引擎|Agent 引擎]] — Agent 持有 SkillsManager
- [[tool-系统|工具系统]] — Skill 调用工具
- [[../architecture/架构总览|架构总览]]
