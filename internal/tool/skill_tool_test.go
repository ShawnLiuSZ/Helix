package tool

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func newTestSkillTool() *SkillTool {
	return NewSkillTool(
		func() []SkillInfo {
			return []SkillInfo{
				{Name: "commit", Description: "make a git commit"},
				{Name: "review", Description: "review a diff"},
			}
		},
		func(name string) (string, error) {
			if name == "commit" {
				return "# Commit skill\nFull instructions here.", nil
			}
			return "", fmt.Errorf("skill %q not found", name)
		},
	)
}

func TestSkillTool_ListsWhenNoName(t *testing.T) {
	res, err := newTestSkillTool().Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(res.Content, "commit") || !strings.Contains(res.Content, "review") {
		t.Errorf("listing should include all skill names, got %q", res.Content)
	}
}

func TestSkillTool_LoadsContentByName(t *testing.T) {
	res, err := newTestSkillTool().Execute(context.Background(), map[string]any{"name": "commit"})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !strings.Contains(res.Content, "Full instructions here") {
		t.Errorf("should return skill content, got %q", res.Content)
	}
}

func TestSkillTool_UnknownNameErrors(t *testing.T) {
	_, err := newTestSkillTool().Execute(context.Background(), map[string]any{"name": "nope"})
	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestSkillTool_IsReadOnly(t *testing.T) {
	if !newTestSkillTool().IsReadOnly() {
		t.Error("skill tool should be read-only")
	}
}
