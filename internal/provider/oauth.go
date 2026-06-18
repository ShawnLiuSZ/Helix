package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OAuthToken OAuth token
type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

// OAuthManager OAuth token 管理器（支持自动刷新）
type OAuthManager struct {
	mu           sync.Mutex
	token        *OAuthToken
	clientID     string
	clientSecret string
	tokenURL     string
	onRefresh    func(token *OAuthToken) error // 刷新回调
}

// NewOAuthManager 创建 OAuth 管理器
func NewOAuthManager(tokenURL, clientID, clientSecret string) *OAuthManager {
	return &OAuthManager{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// SetRefreshCallback 设置 token 刷新回调
func (m *OAuthManager) SetRefreshCallback(fn func(token *OAuthToken) error) {
	m.onRefresh = fn
}

// SetToken 设置初始 token
func (m *OAuthManager) SetToken(token *OAuthToken) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = token
}

// GetToken 获取有效 token（自动刷新）
func (m *OAuthManager) GetToken(ctx context.Context) (*OAuthToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.token == nil {
		return nil, fmt.Errorf("no token set")
	}

	// 检查是否过期（提前 60 秒刷新）
	if time.Now().Add(60 * time.Second).After(m.token.ExpiresAt) {
		if m.token.RefreshToken != "" {
			// TODO: 实现 OAuth refresh token 流程
			// 当前返回现有 token，由上层处理 401 错误
		}
	}

	return m.token, nil
}

// IsExpired 检查 token 是否过期
func (m *OAuthManager) IsExpired() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token == nil {
		return true
	}
	return time.Now().After(m.token.ExpiresAt)
}

// NeedsRefresh 检查是否需要刷新（提前 60 秒）
func (m *OAuthManager) NeedsRefresh() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token == nil {
		return true
	}
	return time.Now().Add(60 * time.Second).After(m.token.ExpiresAt)
}

// AuthorizationHeader 返回 Authorization header 值
func (m *OAuthManager) AuthorizationHeader(ctx context.Context) (string, error) {
	token, err := m.GetToken(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", token.TokenType, token.AccessToken), nil
}
