package provider

import (
	"testing"
	"time"
)

func TestOAuthManager_SetAndGet(t *testing.T) {
	m := NewOAuthManager("https://auth.example.com/token", "client-id", "secret")

	expiresAt := time.Now().Add(1 * time.Hour)
	m.SetToken(&OAuthToken{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	})

	token, err := m.GetToken(t.Context())
	if err != nil {
		t.Fatalf("GetToken error: %v", err)
	}
	if token.AccessToken != "access-token-123" {
		t.Errorf("AccessToken = %q", token.AccessToken)
	}
}

func TestOAuthManager_IsExpired(t *testing.T) {
	m := NewOAuthManager("url", "id", "secret")

	// 未设置 token
	if !m.IsExpired() {
		t.Error("should be expired when no token set")
	}

	// 有效 token
	m.SetToken(&OAuthToken{ExpiresAt: time.Now().Add(1 * time.Hour)})
	if m.IsExpired() {
		t.Error("should not be expired")
	}

	// 过期 token
	m.SetToken(&OAuthToken{ExpiresAt: time.Now().Add(-1 * time.Hour)})
	if !m.IsExpired() {
		t.Error("should be expired")
	}
}

func TestOAuthManager_NeedsRefresh(t *testing.T) {
	m := NewOAuthManager("url", "id", "secret")

	// 未设置
	if !m.NeedsRefresh() {
		t.Error("should need refresh when no token")
	}

	// 快过期（30 秒后）
	m.SetToken(&OAuthToken{ExpiresAt: time.Now().Add(30 * time.Second)})
	if !m.NeedsRefresh() {
		t.Error("should need refresh when expiring in 30s")
	}

	// 远未过期
	m.SetToken(&OAuthToken{ExpiresAt: time.Now().Add(2 * time.Hour)})
	if m.NeedsRefresh() {
		t.Error("should not need refresh when valid for 2h")
	}
}

func TestOAuthManager_AuthorizationHeader(t *testing.T) {
	m := NewOAuthManager("url", "id", "secret")
	m.SetToken(&OAuthToken{
		AccessToken: "my-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		TokenType:   "Bearer",
	})

	header, err := m.AuthorizationHeader(t.Context())
	if err != nil {
		t.Fatalf("AuthorizationHeader error: %v", err)
	}
	if header != "Bearer my-token" {
		t.Errorf("header = %q, want 'Bearer my-token'", header)
	}
}
