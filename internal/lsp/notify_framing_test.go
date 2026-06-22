package lsp

import (
	"bytes"
	"strings"
	"testing"
)

type nopWriteCloser struct{ *bytes.Buffer }

func (nopWriteCloser) Close() error { return nil }

// R1: LSP 通知必须使用 Content-Length 分帧，而非裸换行（否则会打乱服务端解析）。
func TestSendNotification_UsesContentLengthFraming(t *testing.T) {
	buf := &bytes.Buffer{}
	c := &Client{stdin: nopWriteCloser{buf}}

	c.sendNotification("initialized", map[string]any{})

	out := buf.String()
	if !strings.HasPrefix(out, "Content-Length: ") {
		t.Fatalf("notification must use Content-Length framing, got: %q", out)
	}
	idx := strings.Index(out, "\r\n\r\n")
	if idx < 0 {
		t.Fatalf("missing CRLFCRLF header separator, got: %q", out)
	}
	body := out[idx+4:]
	if !strings.Contains(body, `"method":"initialized"`) {
		t.Errorf("body missing method, got: %q", body)
	}
	if strings.HasSuffix(out, "}\n") {
		t.Errorf("notification used bare-newline framing (regression), got: %q", out)
	}
}
