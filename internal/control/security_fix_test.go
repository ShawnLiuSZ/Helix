package control

import "testing"

// C1: 命令替换 / 反引号 / 重定向必须被分词并拒绝（白名单只含 git/go/echo 时）
func TestAllowlist_C1_SubstitutionAndRedirectionDenied(t *testing.T) {
	a := NewAllowlist()
	a.SetShellCommands([]string{"git", "go", "echo"})

	bypasses := []string{
		"git $(touch /tmp/pwned)",                 // 命令替换
		"git `touch /tmp/pwned`",                  // 反引号替换
		"echo hi>/home/user/.ssh/authorized_keys", // 重定向写敏感文件
		"echo hi >> /home/user/.bashrc",           // 追加重定向
		"git status\nrm file",                     // 换行夹带第二条命令
	}
	for _, cmd := range bypasses {
		if a.isShellAllowed(map[string]any{"command": cmd}) {
			t.Errorf("expected DENY for bypass %q, got ALLOW", cmd)
		}
	}
}

// C1: 合法命令在修复后仍应放行
func TestAllowlist_C1_LegitStillAllowed(t *testing.T) {
	a := NewAllowlist()
	a.SetShellCommands([]string{"git", "go", "echo", "ls"})

	ok := []string{"git status", "go build ./...", "echo hello", "ls -la"}
	for _, cmd := range ok {
		if !a.isShellAllowed(map[string]any{"command": cmd}) {
			t.Errorf("expected ALLOW for %q, got DENY", cmd)
		}
	}
}

// H6: Auto 模式必须拒绝未在白名单中的 bash（不再无条件放行）
func TestGate_AutoMode_DeniesNonAllowlistedBash(t *testing.T) {
	g := NewGate(ModeAuto, NewAllowlist())
	for _, cmd := range []string{"python3 -c x", "chmod +x /tmp/x && /tmp/x"} {
		if allowed, _ := g.Check("bash", map[string]any{"command": cmd}); allowed {
			t.Errorf("auto mode must deny non-allowlisted bash %q", cmd)
		}
	}
}

// H6: Auto 模式必须拒绝工作区外的写入
func TestGate_AutoMode_DeniesEscapingWrite(t *testing.T) {
	a := NewAllowlist()
	a.SetAllowedPaths([]string{"/project"})
	g := NewGate(ModeAuto, a)
	if allowed, _ := g.Check("write_file", map[string]any{"path": "/etc/passwd"}); allowed {
		t.Error("auto mode must deny writes outside workspace")
	}
}

// H6: Auto 模式应放行白名单内的 bash 与工作区内的写入
func TestGate_AutoMode_AllowsAllowlisted(t *testing.T) {
	a := NewAllowlist()
	a.SetShellCommands([]string{"git"})
	a.SetAllowedPaths([]string{"/project"})
	g := NewGate(ModeAuto, a)

	if allowed, _ := g.Check("bash", map[string]any{"command": "git status"}); !allowed {
		t.Error("auto mode should allow allowlisted bash")
	}
	if allowed, _ := g.Check("write_file", map[string]any{"path": "/project/main.go"}); !allowed {
		t.Error("auto mode should allow writes within workspace")
	}
}
