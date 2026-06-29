package control

import "testing"

// C1 (residual): 分词器必须是 quote-aware 的。
// 引号内的 shell 元字符（; && || | < > ( ) =）是字面量，不是命令分隔符，
// 因此带标点的合法命令（如 git commit -m "..."）必须放行；
// 但引号外的分隔符、以及双引号内仍然生效的命令替换（$()/反引号）必须照旧拒绝。
func TestAllowlist_C1_QuoteAware_LegitAllowed(t *testing.T) {
	a := NewAllowlist()
	a.SetShellCommands([]string{"git", "go", "echo", "ls"})

	allow := []string{
		`git commit -m "fix: a; b"`,         // 引号内的分号
		`git commit -m "a && b"`,            // 引号内的 &&
		`git commit -m 'single; quoted'`,    // 单引号内的分号
		`git commit -m "refactor (parser)"`, // 引号内的括号
		`git commit -m "key=val pairs"`,     // 引号内含 = （argv0 不应被误判为 env 赋值）
		`git commit -m "use | pipe char"`,   // 引号内的管道符
		`echo "redirect > char"`,            // 引号内的重定向符
	}
	for _, cmd := range allow {
		if !a.isShellAllowed(map[string]any{"command": cmd}) {
			t.Errorf("expected ALLOW for legit quoted command %q, got DENY", cmd)
		}
	}
}

func TestAllowlist_C1_QuoteAware_BypassStillDenied(t *testing.T) {
	a := NewAllowlist()
	a.SetShellCommands([]string{"git", "go", "echo", "ls"})

	deny := []string{
		`git commit -m "ok"; rm file`,  // 引号外的分号夹带 rm
		"git commit -m \"$(rm file)\"", // 双引号内的命令替换仍生效
		"git commit -m \"`rm file`\"",  // 双引号内的反引号替换仍生效
		`echo hi && rm file`,           // 引号外的 &&
		`git log | tee out`,            // 引号外的管道接非白名单命令
		`git $(touch evil)`,            // 命令替换
	}
	for _, cmd := range deny {
		if a.isShellAllowed(map[string]any{"command": cmd}) {
			t.Errorf("expected DENY for bypass %q, got ALLOW", cmd)
		}
	}
}

// 单引号内的命令替换不应被展开（bash 语义），但 argv0 仍是 git → 放行。
func TestAllowlist_C1_QuoteAware_SingleQuoteNoSubstitution(t *testing.T) {
	a := NewAllowlist()
	a.SetShellCommands([]string{"git"})
	if !a.isShellAllowed(map[string]any{"command": `git commit -m '$(rm file)'`}) {
		t.Error(`single-quoted $(...) is literal; git commit should be ALLOWED`)
	}
}
