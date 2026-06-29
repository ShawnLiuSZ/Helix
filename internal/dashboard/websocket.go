package dashboard

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// WSClient WebSocket 客户端
type WSClient struct {
	conn *websocket.Conn
	hub  *WSHub
	send chan []byte
}

// WSHub WebSocket 管理器
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

// NewWSHub 创建 WebSocket 管理器
func NewWSHub() *WSHub {
	hub := &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
	go hub.run()
	return hub
}

// run WebSocket 管理器主循环
func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected: %d total", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected: %d total", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 发送缓冲区满，断开连接
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					close(client.send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 广播消息到所有客户端
func (h *WSHub) Broadcast(message []byte) {
	h.broadcast <- message
}

// ClientCount 返回当前连接数
func (h *WSHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// handleWebSocket 处理 WebSocket 连接（带安全检查）
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// CheckOrigin 白名单：只允许同源连接。
	// 默认拒绝：缺失/空 Origin 头也必须拒绝（否则非浏览器客户端可绕过 CSWSH 防护）。
	origin := r.Header.Get("Origin")
	if !s.isAllowedOrigin(origin) {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	hub := s.wsHub
	if hub == nil {
		http.Error(w, "WebSocket not initialized", http.StatusServiceUnavailable)
		return
	}

	// 检查是否支持 WebSocket（httptest.ResponseRecorder 不支持）
	if _, ok := w.(http.Hijacker); !ok {
		// 回退到普通 HTTP 响应（用于测试）
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "WebSocket endpoint (test mode - no upgrade)",
		})
		return
	}

	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		defer func() { _ = ws.Close() }()

		_ = ws.SetWriteDeadline(time.Now().Add(10 * time.Second))

		client := &WSClient{
			conn: ws,
			hub:  hub,
			send: make(chan []byte, 256),
		}

		hub.register <- client

		// 读循环（仅用于检测断开）。
		// 每次收到消息后刷新读 deadline，避免健康连接被一次性 60s deadline 误杀（R2）。
		go func() {
			defer func() {
				hub.unregister <- client
			}()
			for {
				_ = ws.SetReadDeadline(time.Now().Add(120 * time.Second))
				var msg []byte
				err := websocket.Message.Receive(ws, &msg)
				if err != nil {
					return
				}
				// 客户端消息暂不处理（可扩展为命令）
			}
		}()

		// 写循环
		for msg := range client.send {
			if err := websocket.Message.Send(ws, msg); err != nil {
				return
			}
			// 重置写 deadline
			_ = ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		}
	})

	wsHandler.ServeHTTP(w, r)
}

// isAllowedOrigin 检查 Origin 是否允许（按精确 host 匹配，防止前缀绕过 CSWSH）。
func (s *Server) isAllowedOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return false
}

// WSMessage WebSocket 消息格式
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SendJSON 发送 JSON 消息到所有客户端
func (h *WSHub) SendJSON(msgType string, payload interface{}) {
	msg := WSMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.Broadcast(data)
}
