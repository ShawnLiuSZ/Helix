package tool

import (
	"context"
	"strings"
	"testing"
)

func setupTaskTool() (*TaskTool, *TaskStore) {
	store := NewTaskStore()
	SetTaskStore(store)
	return &TaskTool{}, store
}

func TestTaskCreate(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	res, err := tt.Execute(ctx, map[string]any{"action": "create", "summary": "First task"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.OK() {
		t.Fatalf("expected OK, got error: %s", res.Error)
	}
	if !strings.Contains(res.Content, "T1") || !strings.Contains(res.Content, "First task") {
		t.Fatalf("unexpected content: %s", res.Content)
	}

	res, err = tt.Execute(ctx, map[string]any{"action": "create", "summary": "Second task"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Content, "T2") {
		t.Fatalf("expected T2, got: %s", res.Content)
	}
}

func TestTaskCreateWithParent(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Parent task", "")
	res, err := tt.Execute(ctx, map[string]any{"action": "create", "summary": "Child 1", "parent_id": "T1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Content, "T1.1") {
		t.Fatalf("expected T1.1, got: %s", res.Content)
	}

	res, err = tt.Execute(ctx, map[string]any{"action": "create", "summary": "Child 2", "parent_id": "T1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Content, "T1.2") {
		t.Fatalf("expected T1.2, got: %s", res.Content)
	}
}

func TestTaskCreateMissingSummary(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	res, _ := tt.Execute(ctx, map[string]any{"action": "create"})
	if res.OK() {
		t.Fatal("expected error for missing summary")
	}
}

func TestTaskList(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Task A", "")
	store.Create("Task B", "")

	res, err := tt.Execute(ctx, map[string]any{"action": "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Content, "Task A") || !strings.Contains(res.Content, "Task B") {
		t.Fatalf("expected both tasks, got: %s", res.Content)
	}
}

func TestTaskListWithStatusFilter(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Open task", "")
	item := store.Create("Done task", "")
	store.UpdateStatus(item.ID, "done")

	res, _ := tt.Execute(ctx, map[string]any{"action": "list", "status": "done"})
	if strings.Contains(res.Content, "Open task") {
		t.Fatalf("should not contain open task: %s", res.Content)
	}
	if !strings.Contains(res.Content, "Done task") {
		t.Fatalf("expected done task: %s", res.Content)
	}
}

func TestTaskListEmpty(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	res, _ := tt.Execute(ctx, map[string]any{"action": "list"})
	if !strings.Contains(res.Content, "No tasks") {
		t.Fatalf("expected no tasks message: %s", res.Content)
	}
}

func TestTaskGet(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Lookup me", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "get", "id": "T1"})
	if !strings.Contains(res.Content, "Lookup me") {
		t.Fatalf("expected task content: %s", res.Content)
	}
}

func TestTaskGetNotFound(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	res, _ := tt.Execute(ctx, map[string]any{"action": "get", "id": "T999"})
	if res.OK() {
		t.Fatal("expected error for missing task")
	}
}

func TestTaskGetMissingID(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	res, _ := tt.Execute(ctx, map[string]any{"action": "get"})
	if res.OK() {
		t.Fatal("expected error for missing id")
	}
}

func TestTaskStart(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Start me", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "start", "id": "T1"})
	if !strings.Contains(res.Content, "in_progress") {
		t.Fatalf("expected in_progress: %s", res.Content)
	}
}

func TestTaskDone(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Finish me", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "done", "id": "T1"})
	if !strings.Contains(res.Content, "done") {
		t.Fatalf("expected done: %s", res.Content)
	}
}

func TestTaskBlock(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Block me", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "block", "id": "T1", "event_summary": "waiting on dependency"})
	if !strings.Contains(res.Content, "blocked") || !strings.Contains(res.Content, "waiting on dependency") {
		t.Fatalf("expected blocked with reason: %s", res.Content)
	}
}

func TestTaskBlockWithoutReason(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Block me", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "block", "id": "T1"})
	if !strings.Contains(res.Content, "blocked") {
		t.Fatalf("expected blocked: %s", res.Content)
	}
}

func TestTaskUnblock(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Unblock me", "")
	store.UpdateStatus("T1", "blocked")

	res, _ := tt.Execute(ctx, map[string]any{"action": "unblock", "id": "T1"})
	if !strings.Contains(res.Content, "open") {
		t.Fatalf("expected open: %s", res.Content)
	}
}

func TestTaskAbandon(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Abandon me", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "abandon", "id": "T1", "event_summary": "out of scope"})
	if !strings.Contains(res.Content, "abandoned") || !strings.Contains(res.Content, "out of scope") {
		t.Fatalf("expected abandoned with reason: %s", res.Content)
	}
}

func TestTaskRename(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Old name", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "rename", "id": "T1", "summary": "New name"})
	if !strings.Contains(res.Content, "New name") {
		t.Fatalf("expected new name: %s", res.Content)
	}
}

func TestTaskRenameMissingSummary(t *testing.T) {
	tt, store := setupTaskTool()
	ctx := context.Background()

	store.Create("Some task", "")

	res, _ := tt.Execute(ctx, map[string]any{"action": "rename", "id": "T1"})
	if res.OK() {
		t.Fatal("expected error for missing summary")
	}
}

func TestTaskRenameNotFound(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	res, _ := tt.Execute(ctx, map[string]any{"action": "rename", "id": "T999", "summary": "x"})
	if res.OK() {
		t.Fatal("expected error for missing task")
	}
}

func TestTaskUnknownAction(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	res, _ := tt.Execute(ctx, map[string]any{"action": "invalid"})
	if res.OK() {
		t.Fatal("expected error for unknown action")
	}
}

func TestTaskTransitionNotFound(t *testing.T) {
	tt, _ := setupTaskTool()
	ctx := context.Background()

	for _, action := range []string{"start", "done", "block", "unblock", "abandon"} {
		res, _ := tt.Execute(ctx, map[string]any{"action": action, "id": "T999"})
		if res.OK() {
			t.Fatalf("expected error for %s on missing task", action)
		}
	}
}

func TestTaskAutoCreateStore(t *testing.T) {
	globalTaskStore = nil
	tt := &TaskTool{}
	ctx := context.Background()

	res, _ := tt.Execute(ctx, map[string]any{"action": "create", "summary": "Auto store"})
	if !res.OK() {
		t.Fatalf("expected auto-creation to work: %s", res.Error)
	}
	if GetTaskStore() == nil {
		t.Fatal("expected global store to be created")
	}
}
