package provider

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}
	if cfg.BaseDelay != 1*time.Second {
		t.Errorf("BaseDelay = %v", cfg.BaseDelay)
	}
}

func TestBackoff(t *testing.T) {
	c := NewRetryableClient(10 * time.Second)

	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{1, 1 * time.Second, 1 * time.Second},
		{2, 2 * time.Second, 2 * time.Second},
		{3, 4 * time.Second, 4 * time.Second},
		{10, 30 * time.Second, 30 * time.Second},
	}

	for _, tt := range tests {
		delay := c.backoff(tt.attempt)
		if delay < tt.min || delay > tt.max {
			t.Errorf("backoff(%d) = %v, want between %v and %v", tt.attempt, delay, tt.min, tt.max)
		}
	}
}

func TestShouldRetry(t *testing.T) {
	c := NewRetryableClient(10 * time.Second)

	if !c.shouldRetry(429) {
		t.Error("429 should be retryable")
	}
	if !c.shouldRetry(502) {
		t.Error("502 should be retryable")
	}
	if c.shouldRetry(200) {
		t.Error("200 should not be retryable")
	}
	if c.shouldRetry(400) {
		t.Error("400 should not be retryable")
	}
}

func TestRetryableHTTPClient_Do(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		RetryOn:    []int{503},
	}

	c := NewRetryableClientWithConfig(5*time.Second, cfg)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestRetryableHTTPClient_MaxRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := RetryConfig{
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		RetryOn:    []int{500},
	}

	c := NewRetryableClientWithConfig(5*time.Second, cfg)

	req, _ := http.NewRequest("GET", server.URL, nil)
	_, err := c.Do(req)
	if err == nil {
		t.Error("expected error after max retries")
	}
}

func TestRetryableHTTPClient_BodyReset(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		RetryOn:    []int{503},
	}

	c := NewRetryableClientWithConfig(5*time.Second, cfg)

	body := []byte(`{"test":true}`)
	req, _ := http.NewRequest("POST", server.URL, bytes.NewReader(body))
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}
