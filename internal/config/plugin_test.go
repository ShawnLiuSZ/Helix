package config

import "testing"

func TestPluginConfig_Kind(t *testing.T) {
	cases := []struct {
		name string
		pc   PluginConfig
		want string
	}{
		{"url means sse", PluginConfig{Name: "a", URL: "http://localhost:9000"}, "sse"},
		{"command means stdio", PluginConfig{Name: "b", Command: "my-server"}, "stdio"},
		{"url wins over command", PluginConfig{Name: "c", URL: "http://x", Command: "y"}, "sse"},
		{"empty is none", PluginConfig{Name: "d"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pc.Kind(); got != tc.want {
				t.Errorf("Kind() = %q, want %q", got, tc.want)
			}
		})
	}
}
