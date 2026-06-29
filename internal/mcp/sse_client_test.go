package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSSEClient(t *testing.T) {
	t.Run("NewSSEClient", func(t *testing.T) {
		client := NewSSEClient("http://localhost:8080")
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.baseURL != "http://localhost:8080" {
			t.Errorf("expected baseURL http://localhost:8080, got %s", client.baseURL)
		}
	})

	t.Run("ParseSSEURL", func(t *testing.T) {
		tests := []struct {
			input    string
			baseURL  string
			endpoint string
		}{
			{"http://localhost:8080", "http://localhost:8080", "/sse"},
			{"http://localhost:8080/sse", "http://localhost:8080", "/sse"},
			{"https://example.com/mcp/sse", "https://example.com", "/mcp/sse"},
		}

		for _, tt := range tests {
			base, endpoint, err := ParseSSEURL(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if base != tt.baseURL {
				t.Errorf("expected baseURL %s, got %s", tt.baseURL, base)
			}
			if endpoint != tt.endpoint {
				t.Errorf("expected endpoint %s, got %s", tt.endpoint, endpoint)
			}
		}
	})
}

func TestSSEClientClose(t *testing.T) {
	client := NewSSEClient("http://localhost:8080")
	err := client.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSEClientServerInfo(t *testing.T) {
	client := NewSSEClient("http://localhost:8080")
	info := client.ServerInfo()
	if info.Name != "" {
		t.Errorf("expected empty server name, got %s", info.Name)
	}
}

func TestSSEEvent(t *testing.T) {
	event := SSEEvent{
		Event: "message",
		Data:  `{"test": "data"}`,
		ID:    "123",
	}

	if event.Event != "message" {
		t.Errorf("expected event message, got %s", event.Event)
	}
	if event.Data != `{"test": "data"}` {
		t.Errorf("expected data, got %s", event.Data)
	}
}

func TestSSEClientConnect(t *testing.T) {
	// 创建模拟 SSE 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sse":
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Mcp-Session-Id", "test-session")
			w.WriteHeader(http.StatusOK)
			// 模拟 SSE 响应
			w.Write([]byte("event: endpoint\ndata: /message\n\n"))
			w.(http.Flusher).Flush()
		case "/message":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewSSEClient(server.URL)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestSSEClientListToolsEmpty(t *testing.T) {
	client := NewSSEClient("http://localhost:8080")
	// 未连接时调用应返回错误
	_, err := client.ListTools(context.Background())
	if err == nil {
		t.Fatal("expected error for unconnected client")
	}
}

func TestSSEClientCallToolEmpty(t *testing.T) {
	client := NewSSEClient("http://localhost:8080")
	// 未连接时调用应返回错误
	_, err := client.CallTool(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error for unconnected client")
	}
}

func TestSSEClientJSONMarshal(t *testing.T) {
	// 测试 SSEEvent 可以正确序列化
	event := SSEEvent{
		Event: "message",
		Data:  `{"test": "data"}`,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(data), "message") {
		t.Error("expected data to contain message")
	}
}
