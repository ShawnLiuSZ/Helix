package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// R2: Origin 校验必须按精确 host 匹配，前缀匹配会被 http://localhost.evil.com 绕过（CSWSH）。
func TestIsAllowedOrigin_ExactHost(t *testing.T) {
	s := &Server{}
	cases := []struct {
		origin string
		want   bool
	}{
		{"http://localhost", true},
		{"http://localhost:8080", true},
		{"http://127.0.0.1", true},
		{"http://127.0.0.1:3000", true},
		{"https://localhost", true},
		// 以下为前缀绕过攻击，必须拒绝：
		{"http://localhost.evil.com", false},
		{"http://127.0.0.1.attacker.com", false},
		{"http://localhostXSS", false},
		{"http://evil.com", false},
		{"http://notlocalhost", false},
		{"ftp://localhost", false},
		{"garbage", false},
	}
	for _, c := range cases {
		if got := s.isAllowedOrigin(c.origin); got != c.want {
			t.Errorf("isAllowedOrigin(%q) = %v, want %v", c.origin, got, c.want)
		}
	}
}

// R2 (residual): 缺失/空 Origin 头必须默认拒绝。原实现 `origin != "" && ...`
// 让无 Origin 头的非浏览器客户端（curl/脚本）无条件放行，是 CSWSH 防护的缺口。
func TestHandleWebSocket_RejectsMissingOrigin(t *testing.T) {
	s := NewServer("127.0.0.1:0")
	cases := []struct {
		name       string
		setOrigin  bool
		origin     string
		wantReject bool
	}{
		{"missing Origin header", false, "", true},
		{"empty Origin header", true, "", true},
		{"disallowed Origin", true, "http://evil.com", true},
		{"prefix-bypass Origin", true, "http://localhost.evil.com", true},
		{"allowed Origin", true, "http://localhost:8080", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if c.setOrigin {
				req.Header.Set("Origin", c.origin)
			}
			rec := httptest.NewRecorder()
			s.handleWebSocket(rec, req)
			rejected := rec.Code == http.StatusForbidden
			if rejected != c.wantReject {
				t.Errorf("status=%d rejected=%v, wantReject=%v", rec.Code, rejected, c.wantReject)
			}
		})
	}
}
