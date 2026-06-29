package testutil

import "context"

// StubToolResult 通用工具 stub 结果
type StubToolResult struct {
	Content string
	Error   string
}

// StubTool 通用工具 stub（不依赖 tool 包，避免循环依赖）
type StubTool struct {
	NameVal        string
	DescriptionVal string
	IsReadOnlyVal  bool
	ExecuteFn      func(ctx context.Context, args map[string]any) (*StubToolResult, error)
	ExecuteCalls   int
}

func (s *StubTool) ExecuteCallCount() int { return s.ExecuteCalls }

// StubToolRegistry 预注册常用工具的 stub 实现
type StubToolRegistry struct {
	tools map[string]*StubTool
}

// NewStubToolRegistry 创建预注册的工具注册表
func NewStubToolRegistry() *StubToolRegistry {
	r := &StubToolRegistry{tools: make(map[string]*StubTool)}

	r.Register("read_file", &StubTool{
		NameVal: "read_file", IsReadOnlyVal: true,
		ExecuteFn: func(ctx context.Context, args map[string]any) (*StubToolResult, error) {
			return &StubToolResult{Content: "stub file content"}, nil
		},
	})

	r.Register("bash", &StubTool{
		NameVal: "bash", IsReadOnlyVal: false,
		ExecuteFn: func(ctx context.Context, args map[string]any) (*StubToolResult, error) {
			return &StubToolResult{Content: "stub command executed"}, nil
		},
	})

	r.Register("grep", &StubTool{
		NameVal: "grep", IsReadOnlyVal: true,
		ExecuteFn: func(ctx context.Context, args map[string]any) (*StubToolResult, error) {
			return &StubToolResult{Content: "stub grep results"}, nil
		},
	})

	return r
}

func (r *StubToolRegistry) Register(name string, t *StubTool) { r.tools[name] = t }
func (r *StubToolRegistry) Get(name string) (*StubTool, bool) { t, ok := r.tools[name]; return t, ok }
