package dashboard

import "testing"

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
