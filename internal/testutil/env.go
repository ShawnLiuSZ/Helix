package testutil

// TestEnvProvider 测试用环境变量提供者（用于 tool 包测试）
// 注意：此类型不依赖 tool.EnvProvider 接口，避免循环依赖
// 使用时通过类型断言适配
type TestEnvProvider struct {
	Env []string
}

func (p *TestEnvProvider) EnvForSubprocess() []string {
	return p.Env
}
