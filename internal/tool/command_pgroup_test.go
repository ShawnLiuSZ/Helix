package tool

import (
	"context"
	"testing"
	"time"
)

// M3: 输出超限时必须立即杀掉进程组，避免子进程写满管道导致 Wait 挂到 60s 超时。
func TestBashTool_RunawayOutputReturnsPromptly(t *testing.T) {
	bt := &BashTool{}

	done := make(chan struct{})
	var contentLen int
	var execErr error
	go func() {
		res, err := bt.Execute(context.Background(), map[string]any{"command": "cat /dev/zero"})
		execErr = err
		if res != nil {
			contentLen = len(res.Content)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("runaway producer not killed promptly — process-group / truncation-hang bug (M3)")
	}

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if contentLen < maxOutputSize {
		t.Errorf("expected output truncated near %d bytes, got %d", maxOutputSize, contentLen)
	}
}
