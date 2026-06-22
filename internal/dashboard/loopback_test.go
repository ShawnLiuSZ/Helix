package dashboard

import "testing"

// M9: 非回环地址必须被强制改回 127.0.0.1，避免监听全网卡（无鉴权）。
func TestLoopbackAddr(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "127.0.0.1:8080"},
		{":8080", "127.0.0.1:8080"},
		{":9090", "127.0.0.1:9090"},
		{"127.0.0.1:3000", "127.0.0.1:3000"},
		{"localhost:3000", "localhost:3000"},
		{"0.0.0.0:8080", "127.0.0.1:8080"},     // 全网卡 → 强制回环
		{"192.168.1.5:8080", "127.0.0.1:8080"}, // 局域网地址 → 强制回环
	}
	for _, c := range cases {
		if got := loopbackAddr(c.in); got != c.want {
			t.Errorf("loopbackAddr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
