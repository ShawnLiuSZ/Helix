package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ShawnLiuSZ/Helix/internal/session"
)

// DreamScheduler Dream 定时任务管理器
type DreamScheduler struct {
	agent      *Agent
	sessionMgr *session.Manager
	interval   time.Duration
	stopCh     chan struct{}
	memoryDir  string
}

// NewDreamScheduler 创建 Dream 调度器
func NewDreamScheduler(agent *Agent, sessionMgr *session.Manager, memoryDir string) *DreamScheduler {
	return &DreamScheduler{
		agent:      agent,
		sessionMgr: sessionMgr,
		interval:   24 * time.Hour, // 默认每天执行一次
		stopCh:     make(chan struct{}),
		memoryDir:  memoryDir,
	}
}

// Start 启动 Dream 调度器
func (d *DreamScheduler) Start() {
	go d.run()
}

// Stop 停止 Dream 调度器
func (d *DreamScheduler) Stop() {
	close(d.stopCh)
}

// run 运行 Dream 循环
func (d *DreamScheduler) run() {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.RunDream()
		case <-d.stopCh:
			return
		}
	}
}

// RunDream 执行 Dream 任务
func (d *DreamScheduler) RunDream() error {
	// 1. 收集最近会话
	sessions := d.sessionMgr.List()
	if len(sessions) == 0 {
		return nil
	}

	// 2. 提取最近 7 天的会话
	recentSessions := d.filterRecent(sessions, 7*24*time.Hour)
	if len(recentSessions) == 0 {
		return nil
	}

	// 3. 分析会话内容
	patterns := d.analyzePatterns(recentSessions)

	// 4. 提取知识
	knowledge := d.extractKnowledge(patterns)

	// 5. 保存到 MEMORY.md
	return d.saveToMemory(knowledge)
}

// filterRecent 过滤最近的会话
func (d *DreamScheduler) filterRecent(sessions []*session.Session, duration time.Duration) []*session.Session {
	cutoff := time.Now().Add(-duration)
	var recent []*session.Session
	for _, s := range sessions {
		if s.UpdatedAt.After(cutoff) {
			recent = append(recent, s)
		}
	}
	return recent
}

// analyzePatterns 分析会话模式
func (d *DreamScheduler) analyzePatterns(sessions []*session.Session) *Patterns {
	patterns := &Patterns{
		Topics:    make(map[string]int),
		Tools:     make(map[string]int),
		Successes: 0,
		Failures:  0,
	}

	for _, s := range sessions {
		for _, msg := range s.Messages {
			if msg.Role == "user" {
				// 提取用户意图
				d.extractTopics(msg.Content, patterns)
			}
			if msg.Role == "tool" {
				// 统计工具使用
				patterns.Tools[msg.ToolName]++
			}
		}
	}

	return patterns
}

// extractTopics 提取主题
func (d *DreamScheduler) extractTopics(content string, patterns *Patterns) {
	// 简单的关键词提取
	keywords := []string{
		"创建", "修改", "删除", "修复", "优化", "重构", "测试", "部署",
		"create", "modify", "delete", "fix", "optimize", "refactor", "test", "deploy",
	}

	content = strings.ToLower(content)
	for _, kw := range keywords {
		if strings.Contains(content, kw) {
			patterns.Topics[kw]++
		}
	}
}

// extractKnowledge 提取知识
func (d *DreamScheduler) extractKnowledge(patterns *Patterns) *Knowledge {
	knowledge := &Knowledge{
		GeneratedAt: time.Now(),
		Topics:      patterns.Topics,
		Tools:       patterns.Tools,
		Suggestions: make([]string, 0),
	}

	// 生成建议
	if patterns.Tools["bash"] > 5 {
		knowledge.Suggestions = append(knowledge.Suggestions,
			"用户频繁使用 bash 工具，建议考虑创建常用命令别名")
	}
	if patterns.Tools["write_file"] > 10 {
		knowledge.Suggestions = append(knowledge.Suggestions,
			"用户频繁写入文件，建议考虑批量写入或模板生成")
	}

	return knowledge
}

// saveToMemory 保存到 MEMORY.md
func (d *DreamScheduler) saveToMemory(knowledge *Knowledge) error {
	if d.memoryDir == "" {
		return nil
	}

	// 确保目录存在
	if err := os.MkdirAll(d.memoryDir, 0755); err != nil {
		return err
	}

	memoryFile := filepath.Join(d.memoryDir, "MEMORY.md")

	// 读取现有内容
	existing := ""
	if data, err := os.ReadFile(memoryFile); err == nil {
		existing = string(data)
	}

	// 生成新的知识条目
	entry := d.formatKnowledgeEntry(knowledge)

	// 追加到文件
	content := existing + "\n" + entry
	return os.WriteFile(memoryFile, []byte(content), 0644)
}

// formatKnowledgeEntry 格式化知识条目
func (d *DreamScheduler) formatKnowledgeEntry(k *Knowledge) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n## Dream Entry - %s\n\n", k.GeneratedAt.Format("2006-01-02 15:04:05")))

	sb.WriteString("### 常见主题\n")
	for topic, count := range k.Topics {
		if count > 2 {
			sb.WriteString(fmt.Sprintf("- %s (出现 %d 次)\n", topic, count))
		}
	}

	sb.WriteString("\n### 工具使用统计\n")
	for tool, count := range k.Tools {
		sb.WriteString(fmt.Sprintf("- %s: %d 次\n", tool, count))
	}

	if len(k.Suggestions) > 0 {
		sb.WriteString("\n### 建议\n")
		for _, s := range k.Suggestions {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
	}

	return sb.String()
}

// Patterns 会话模式
type Patterns struct {
	Topics    map[string]int
	Tools     map[string]int
	Successes int
	Failures  int
}

// Knowledge 提取的知识
type Knowledge struct {
	GeneratedAt time.Time
	Topics      map[string]int
	Tools       map[string]int
	Suggestions []string
}

// Distiller Distill 管理器
type Distiller struct {
	agent     *Agent
	skillsDir string
}

// NewDistiller 创建 Distiller
func NewDistiller(agent *Agent, skillsDir string) *Distiller {
	return &Distiller{
		agent:     agent,
		skillsDir: skillsDir,
	}
}

// RunDistill 执行 Distill 任务
func (d *Distiller) RunDistill(ctx context.Context, sessions []*session.Session) error {
	// 1. 分析重复工作流
	workflows := d.analyzeWorkflows(sessions)

	// 2. 生成 skills
	for _, wf := range workflows {
		if wf.Confidence > 0.7 {
			if err := d.createSkill(ctx, wf); err != nil {
				// 记录错误但继续
				fmt.Printf("Failed to create skill: %v\n", err)
			}
		}
	}

	return nil
}

// analyzeWorkflows 分析重复工作流
func (d *Distiller) analyzeWorkflows(sessions []*session.Session) []*Workflow {
	workflowMap := make(map[string]*Workflow)

	for _, s := range sessions {
		var currentSteps []string
		for _, msg := range s.Messages {
			if msg.Role == "tool" && msg.ToolName != "" {
				currentSteps = append(currentSteps, msg.ToolName)
			}
		}

		if len(currentSteps) >= 3 {
			key := strings.Join(currentSteps, "->")
			if wf, ok := workflowMap[key]; ok {
				wf.Count++
			} else {
				workflowMap[key] = &Workflow{
					Steps:  currentSteps,
					Count:  1,
					Sessions: []string{s.ID},
				}
			}
		}
	}

	// 转换为切片并计算置信度
	var workflows []*Workflow
	for _, wf := range workflowMap {
		wf.Confidence = float64(wf.Count) / float64(len(sessions))
		if wf.Count >= 3 {
			workflows = append(workflows, wf)
		}
	}

	return workflows
}

// createSkill 创建 skill 文件
func (d *Distiller) createSkill(ctx context.Context, wf *Workflow) error {
	if d.skillsDir == "" {
		return nil
	}

	// 确保目录存在
	if err := os.MkdirAll(d.skillsDir, 0755); err != nil {
		return err
	}

	// 生成 skill 名称
	skillName := d.generateSkillName(wf)
	skillPath := filepath.Join(d.skillsDir, skillName+".md")

	// 生成 skill 内容
	content := d.generateSkillContent(wf)

	return os.WriteFile(skillPath, []byte(content), 0644)
}

// generateSkillName 生成 skill 名称
func (d *Distiller) generateSkillName(wf *Workflow) string {
	if len(wf.Steps) > 0 {
		return "auto-" + wf.Steps[0]
	}
	return "auto-workflow"
}

// generateSkillContent 生成 skill 内容
func (d *Distiller) generateSkillContent(wf *Workflow) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", d.generateSkillName(wf)))
	sb.WriteString("> 自动生成的工作流 skill\n\n")
	sb.WriteString("## 步骤\n\n")

	for i, step := range wf.Steps {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}

	sb.WriteString(fmt.Sprintf("\n## 统计\n\n"))
	sb.WriteString(fmt.Sprintf("- 识别次数: %d\n", wf.Count))
	sb.WriteString(fmt.Sprintf("- 置信度: %.0f%%\n", wf.Confidence*100))

	return sb.String()
}

// Workflow 工作流
type Workflow struct {
	Steps     []string
	Count     int
	Confidence float64
	Sessions  []string
}
