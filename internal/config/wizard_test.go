package config

import (
	"bufio"
	"strings"
	"testing"
)

func TestSelectMiMoMode_PayAsYouGo(t *testing.T) {
	w := &Wizard{reader: bufio.NewReader(strings.NewReader("1\n"))}
	preset := providerPresets[2]
	got, err := w.selectMiMoMode(preset)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "mimo" {
		t.Errorf("name = %q, want mimo", got.Name)
	}
	if got.BaseURL != "https://api.xiaomimimo.com/v1" {
		t.Errorf("base_url = %q, want pay-as-you-go URL", got.BaseURL)
	}
	if got.APIKeyEnv != "MIMO_API_KEY" {
		t.Errorf("api_key_env = %q, want MIMO_API_KEY", got.APIKeyEnv)
	}
}

func TestSelectMiMoMode_TokenPlan(t *testing.T) {
	w := &Wizard{reader: bufio.NewReader(strings.NewReader("2\n"))}
	preset := providerPresets[2]
	got, err := w.selectMiMoMode(preset)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "mimo-token-plan" {
		t.Errorf("name = %q, want mimo-token-plan", got.Name)
	}
	if got.DisplayName != "MiMo Token Plan" {
		t.Errorf("display_name = %q, want MiMo Token Plan", got.DisplayName)
	}
	if got.BaseURL != "https://token-plan-cn.xiaomimimo.com/v1" {
		t.Errorf("base_url = %q, want token-plan URL", got.BaseURL)
	}
	if got.APIKeyEnv != "MIMO_TOKEN_PLAN_KEY" {
		t.Errorf("api_key_env = %q, want MIMO_TOKEN_PLAN_KEY", got.APIKeyEnv)
	}
}
