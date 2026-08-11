package config

import "testing"

// TestLogGlobal 覆盖全局日志器变量（默认为 nil，由入口初始化）。
func TestLogGlobal(t *testing.T) {
	Log = nil
	if Log != nil {
		t.Fatal("Log 应为 nil")
	}
}
