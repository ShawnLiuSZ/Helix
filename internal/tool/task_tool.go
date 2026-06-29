package tool

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type TaskItem struct {
	ID        string
	Summary   string
	Status    string
	ParentID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*TaskItem
	seq   int
}

var globalTaskStore *TaskStore

func SetTaskStore(s *TaskStore) {
	globalTaskStore = s
}

func GetTaskStore() *TaskStore {
	return globalTaskStore
}

func NewTaskStore() *TaskStore {
	return &TaskStore{tasks: make(map[string]*TaskItem)}
}

func (s *TaskStore) Create(summary, parentID string) *TaskItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	var id string
	if parentID == "" {
		id = fmt.Sprintf("T%d", s.seq)
	} else {
		childSeq := 0
		for _, t := range s.tasks {
			if t.ParentID == parentID {
				childSeq++
			}
		}
		childSeq++
		id = fmt.Sprintf("%s.%d", parentID, childSeq)
	}

	now := time.Now()
	item := &TaskItem{
		ID:        id,
		Summary:   summary,
		Status:    "open",
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.tasks[id] = item
	return item
}

func (s *TaskStore) Get(id string) (*TaskItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *TaskStore) UpdateStatus(id, status string) (*TaskItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %q not found", id)
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	return t, nil
}

func (s *TaskStore) Rename(id, summary string) (*TaskItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %q not found", id)
	}
	t.Summary = summary
	t.UpdatedAt = time.Now()
	return t, nil
}

func (s *TaskStore) List(statusFilter string) []*TaskItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*TaskItem
	for _, t := range s.tasks {
		if statusFilter != "" && t.Status != statusFilter {
			continue
		}
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

type TaskTool struct{}

func (t *TaskTool) Name() string        { return "task" }
func (t *TaskTool) Description() string { return "Persistent task management" }
func (t *TaskTool) IsReadOnly() bool    { return false }

func (t *TaskTool) Schema() Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Property{
			"action": {
				Type:        "string",
				Description: "Subcommand: create, list, get, start, done, block, unblock, abandon, rename",
			},
			"summary": {
				Type:        "string",
				Description: "Task summary (for create/rename)",
			},
			"parent_id": {
				Type:        "string",
				Description: "Parent task ID (for create)",
			},
			"status": {
				Type:        "string",
				Description: "Status filter (for list)",
			},
			"id": {
				Type:        "string",
				Description: "Task ID (for get/start/done/block/unblock/abandon/rename)",
			},
			"event_summary": {
				Type:        "string",
				Description: "Reason for block/abandon",
			},
		},
		Required: []string{"action"},
	}
}

func (t *TaskTool) Execute(_ context.Context, args map[string]any) (*Result, error) {
	store := GetTaskStore()
	if store == nil {
		store = NewTaskStore()
		SetTaskStore(store)
	}

	action, _ := args["action"].(string)
	switch action {
	case "create":
		return t.create(store, args)
	case "list":
		return t.list(store, args)
	case "get":
		return t.get(store, args)
	case "start":
		return t.transition(store, args, "in_progress")
	case "done":
		return t.transition(store, args, "done")
	case "block":
		return t.transitionWithReason(store, args, "blocked")
	case "unblock":
		return t.transition(store, args, "open")
	case "abandon":
		return t.transitionWithReason(store, args, "abandoned")
	case "rename":
		return t.rename(store, args)
	default:
		return &Result{Error: fmt.Sprintf("unknown action: %s", action)}, nil
	}
}

func (t *TaskTool) create(store *TaskStore, args map[string]any) (*Result, error) {
	summary, _ := args["summary"].(string)
	if summary == "" {
		return &Result{Error: "summary is required"}, nil
	}
	parentID, _ := args["parent_id"].(string)
	item := store.Create(summary, parentID)
	return &Result{Content: formatTask(item)}, nil
}

func (t *TaskTool) list(store *TaskStore, args map[string]any) (*Result, error) {
	status, _ := args["status"].(string)
	items := store.List(status)
	if len(items) == 0 {
		return &Result{Content: "No tasks found."}, nil
	}
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(formatTask(item))
		sb.WriteString("\n")
	}
	return &Result{Content: strings.TrimRight(sb.String(), "\n")}, nil
}

func (t *TaskTool) get(store *TaskStore, args map[string]any) (*Result, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return &Result{Error: "id is required"}, nil
	}
	item, ok := store.Get(id)
	if !ok {
		return &Result{Error: fmt.Sprintf("task %q not found", id)}, nil
	}
	return &Result{Content: formatTask(item)}, nil
}

func (t *TaskTool) transition(store *TaskStore, args map[string]any, status string) (*Result, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return &Result{Error: "id is required"}, nil
	}
	item, err := store.UpdateStatus(id, status)
	if err != nil {
		return &Result{Error: err.Error()}, nil
	}
	return &Result{Content: formatTask(item)}, nil
}

func (t *TaskTool) transitionWithReason(store *TaskStore, args map[string]any, status string) (*Result, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return &Result{Error: "id is required"}, nil
	}
	item, err := store.UpdateStatus(id, status)
	if err != nil {
		return &Result{Error: err.Error()}, nil
	}
	reason, _ := args["event_summary"].(string)
	if reason != "" {
		return &Result{Content: fmt.Sprintf("%s\nReason: %s", formatTask(item), reason)}, nil
	}
	return &Result{Content: formatTask(item)}, nil
}

func (t *TaskTool) rename(store *TaskStore, args map[string]any) (*Result, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return &Result{Error: "id is required"}, nil
	}
	summary, _ := args["summary"].(string)
	if summary == "" {
		return &Result{Error: "summary is required"}, nil
	}
	item, err := store.Rename(id, summary)
	if err != nil {
		return &Result{Error: err.Error()}, nil
	}
	return &Result{Content: formatTask(item)}, nil
}

func formatTask(t *TaskItem) string {
	if t.ParentID != "" {
		return fmt.Sprintf("[%s] (%s) %s [parent: %s]", t.ID, t.Status, t.Summary, t.ParentID)
	}
	return fmt.Sprintf("[%s] (%s) %s", t.ID, t.Status, t.Summary)
}
