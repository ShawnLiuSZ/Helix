package agent

import (
	"strings"
	"testing"

	"github.com/ShawnLiuSZ/Helix/internal/testutil"
	"github.com/ShawnLiuSZ/Helix/internal/tool"
)

type fakeMemory struct{ prompt string }

func (f fakeMemory) BuildContextPrompt() string { return f.prompt }

func TestBuildSystemPrompt_InjectsMemory(t *testing.T) {
	p := testutil.NewStubProvider(nil)
	a := New(p, tool.NewRegistry())
	a.SetMemory(fakeMemory{prompt: "## Project Knowledge\n- build: use make build\n"})

	prompt := a.buildSystemPrompt()
	if !strings.Contains(prompt, "use make build") {
		t.Errorf("system prompt should inject memory context, got:\n%s", prompt)
	}
}

func TestBuildSystemPrompt_NoMemoryNoError(t *testing.T) {
	p := testutil.NewStubProvider(nil)
	a := New(p, tool.NewRegistry())

	prompt := a.buildSystemPrompt() // no memory set
	if !strings.Contains(prompt, "Helix") {
		t.Error("prompt should still be built without memory")
	}
}

func TestBuildSystemPrompt_EmptyMemoryOmitted(t *testing.T) {
	p := testutil.NewStubProvider(nil)
	a := New(p, tool.NewRegistry())
	a.SetMemory(fakeMemory{prompt: ""}) // nothing remembered yet

	prompt := a.buildSystemPrompt()
	if strings.Contains(prompt, "Project Knowledge") {
		t.Error("empty memory should not add a heading")
	}
}
