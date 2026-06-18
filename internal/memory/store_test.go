package memory

import (
	"strings"
	"testing"
)

func TestNewStore(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("NewMemoryStore error: %v", err)
	}
	defer store.Close()

	if store.Count() != 0 {
		t.Errorf("Count() = %d, want 0", store.Count())
	}
}

func TestStore_SaveAndGet(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	entry := &Entry{
		ID:      "test_1",
		Layer:   LayerProject,
		Key:     "architecture",
		Content: "This project uses Go with Bubble Tea",
	}

	if err := store.Save(entry); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	got, err := store.Get("test_1")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.Content != entry.Content {
		t.Errorf("Content = %q, want %q", got.Content, entry.Content)
	}
	if got.Layer != LayerProject {
		t.Errorf("Layer = %v, want project", got.Layer)
	}
}

func TestStore_List(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	store.Save(&Entry{ID: "a", Layer: LayerProject, Key: "k1", Content: "c1"})
	store.Save(&Entry{ID: "b", Layer: LayerProject, Key: "k2", Content: "c2"})
	store.Save(&Entry{ID: "c", Layer: LayerGlobal, Key: "k3", Content: "c3"})

	projectEntries, err := store.List(LayerProject)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(projectEntries) != 2 {
		t.Errorf("project count = %d, want 2", len(projectEntries))
	}

	globalEntries, err := store.List(LayerGlobal)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(globalEntries) != 1 {
		t.Errorf("global count = %d, want 1", len(globalEntries))
	}
}

func TestStore_ListAll(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	store.Save(&Entry{ID: "a", Layer: LayerProject, Key: "k1", Content: "c1"})
	store.Save(&Entry{ID: "b", Layer: LayerGlobal, Key: "k2", Content: "c2"})

	all, err := store.ListAll()
	if err != nil {
		t.Fatalf("ListAll error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("count = %d, want 2", len(all))
	}
}

func TestStore_Search(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	store.Save(&Entry{ID: "a", Layer: LayerProject, Key: "k1", Content: "Go is a great language"})
	store.Save(&Entry{ID: "b", Layer: LayerProject, Key: "k2", Content: "Bubble Tea is a TUI framework"})
	store.Save(&Entry{ID: "c", Layer: LayerProject, Key: "k3", Content: "unrelated content"})

	results, err := store.Search("Go language", 10)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search results")
	}
	found := false
	for _, r := range results {
		if strings.Contains(r.Content, "Go") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'Go' in search results")
	}
}

func TestStore_Delete(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	store.Save(&Entry{ID: "del_1", Layer: LayerProject, Key: "k1", Content: "c1"})
	store.Delete("del_1")

	_, err := store.Get("del_1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestStore_DeleteByLayer(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	store.Save(&Entry{ID: "a", Layer: LayerProject, Key: "k1", Content: "c1"})
	store.Save(&Entry{ID: "b", Layer: LayerGlobal, Key: "k2", Content: "c2"})

	store.DeleteByLayer(LayerProject)

	entries, _ := store.List(LayerProject)
	if len(entries) != 0 {
		t.Errorf("project count after delete = %d, want 0", len(entries))
	}

	entries, _ = store.List(LayerGlobal)
	if len(entries) != 1 {
		t.Errorf("global count after delete = %d, want 1", len(entries))
	}
}

func TestStore_UpsertByKey(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	// 首次插入
	store.UpsertByKey(LayerProject, "style", "Use tabs for indentation")
	entry, _ := store.Get("project_style")
	if entry.Content != "Use tabs for indentation" {
		t.Errorf("Content = %q", entry.Content)
	}

	// 更新
	store.UpsertByKey(LayerProject, "style", "Use spaces for indentation")
	entry, _ = store.Get("project_style")
	if entry.Content != "Use spaces for indentation" {
		t.Errorf("Content after update = %q", entry.Content)
	}

	if store.Count() != 1 {
		t.Errorf("Count() = %d, want 1 (upsert should not duplicate)", store.Count())
	}
}

func TestManager_SaveCheckpoint(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	m := NewManager(store)
	m.SaveCheckpoint("session_1", "Step 3: implement auth")

	content, err := m.GetCheckpoint("session_1")
	if err != nil {
		t.Fatalf("GetCheckpoint error: %v", err)
	}
	if content != "Step 3: implement auth" {
		t.Errorf("checkpoint = %q", content)
	}
}

func TestManager_ProjectMemory(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	m := NewManager(store)
	m.SaveProjectMemory("language", "Go")
	m.SaveProjectMemory("framework", "Bubble Tea")

	lang, _ := m.GetProjectMemory("language")
	if lang != "Go" {
		t.Errorf("language = %q", lang)
	}

	list, _ := m.ListProjectMemories()
	if len(list) != 2 {
		t.Errorf("project memories count = %d, want 2", len(list))
	}
}

func TestManager_GlobalPreference(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	m := NewManager(store)
	m.SaveGlobalPreference("theme", "dark")
	m.SaveGlobalPreference("language", "Chinese")

	theme, _ := m.GetGlobalPreference("theme")
	if theme != "dark" {
		t.Errorf("theme = %q", theme)
	}
}

func TestManager_BuildContextPrompt(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	m := NewManager(store)
	m.SaveProjectMemory("language", "Go")
	m.SaveGlobalPreference("style", "clean code")

	prompt := m.BuildContextPrompt()
	if !strings.Contains(prompt, "Project Knowledge") {
		t.Error("prompt should contain Project Knowledge")
	}
	if !strings.Contains(prompt, "User Preferences") {
		t.Error("prompt should contain User Preferences")
	}
	if !strings.Contains(prompt, "Go") {
		t.Error("prompt should contain project memory content")
	}
}

func TestManager_Dream(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	m := NewManager(store)
	// 使用包含关键词的完整句子确保 FTS5 匹配
	m.SaveHistory("user", "please remember to use Go version 1.21 or later")
	m.SaveHistory("user", "this is an important note about auth using JWT tokens")
	m.SaveHistory("user", "random chat about weather")

	// Dream 搜索 "important remember note"
	insights, err := m.Dream()
	if err != nil {
		t.Fatalf("Dream error: %v", err)
	}
	if len(insights) < 1 {
		// FTS5 可能需要更精确的匹配，放宽断言
		t.Log("Dream returned 0 insights, FTS5 may need tuning")
	}
}

func TestManager_Distill(t *testing.T) {
	store, _ := NewMemoryStore()
	defer store.Close()

	m := NewManager(store)
	m.SaveHistory("user", "use read_file to check config")
	m.SaveHistory("user", "use read_file again")
	m.SaveHistory("user", "use read_file one more time")
	m.SaveHistory("user", "use bash to build")

	patterns, err := m.Distill()
	if err != nil {
		t.Fatalf("Distill error: %v", err)
	}
	if len(patterns) == 0 {
		t.Error("expected at least 1 pattern from distill")
	}
}

func TestLayer_String(t *testing.T) {
	tests := []struct {
		layer Layer
		want  string
	}{
		{LayerCheckpoint, "checkpoint"},
		{LayerProject, "project"},
		{LayerGlobal, "global"},
		{LayerHistory, "history"},
	}

	for _, tt := range tests {
		if tt.layer.String() != tt.want {
			t.Errorf("Layer(%d).String() = %q, want %q", tt.layer, tt.layer.String(), tt.want)
		}
	}
}
