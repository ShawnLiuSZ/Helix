package tool

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// Diagnoser 在文件写入/编辑后对其做快速静态检查，返回人类可读的问题摘要
// （无问题返回 ""）。这是"改完即得反馈"的扩展点：当前内置 Go（gofmt -e）后端，
// 未来可在此接入 LSP 推送式诊断以覆盖更多语言。
type Diagnoser interface {
	Diagnose(ctx context.Context, path string) string
}

// runDiagnoser 安全调用 diagnoser（nil 时直接返回空），供各文件工具复用。
func runDiagnoser(d Diagnoser, ctx context.Context, path string) string {
	if d == nil {
		return ""
	}
	return d.Diagnose(ctx, path)
}

// GoDiagnoser 用 gofmt -e 对 .go 文件做语法检查：确定性、无需常驻服务、
// 随 Go 工具链而来。语法正确仅格式不规范的文件不报错（gofmt -e 退出 0）。
type GoDiagnoser struct{}

// Diagnose 实现 Diagnoser。仅处理 .go 文件，其余返回空。
func (GoDiagnoser) Diagnose(ctx context.Context, path string) string {
	if filepath.Ext(path) != ".go" {
		return ""
	}
	bin, err := exec.LookPath("gofmt")
	if err != nil {
		return "" // 工具链不可用则跳过，不阻塞编辑
	}

	cmd := exec.CommandContext(ctx, bin, "-e", path)
	cmd.Stdout = io.Discard // 丢弃格式化后的源码，只关心错误
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil && stderr.Len() > 0 {
		return "\n⚠ Syntax issues detected (gofmt):\n" + strings.TrimSpace(stderr.String())
	}
	return ""
}
