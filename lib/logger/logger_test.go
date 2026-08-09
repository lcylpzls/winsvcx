package logger

import (
	"path/filepath"
	"testing"

	"github.com/lcylpzls/logx"
)

// TestInitFileConsole 覆盖文件与控制台初始化。
func TestInitFileConsole(t *testing.T) {
	Reset()
	defer Reset()
	dir := t.TempDir()
	l := Init(Options{
		LogDir:        dir,
		Filename:      "app.log",
		MaxSize:       1,
		MaxBackups:    2,
		MaxAge:        3,
		CompressAfter: 1,
		Level:         logx.DebugLevel,
		Console:       true,
	})
	if l == nil {
		t.Fatal("日志器为空")
	}
	l.Info("初始化测试", logx.Fields())
	if err := l.Close(); err != nil {
		t.Fatalf("关闭日志器失败：%v", err)
	}
	if Get() == nil {
		t.Fatal("全局日志器为空")
	}
	files, err := filepath.Glob(filepath.Join(dir, "app-*.log"))
	if err != nil || len(files) == 0 {
		t.Fatalf("日志文件未创建：%v", err)
	}
}

// TestInitDefaultLevel 覆盖默认级别与纯控制台模式。
func TestInitDefaultLevel(t *testing.T) {
	Reset()
	defer Reset()
	l := Init(Options{Console: true})
	l.Debug("不应输出", logx.Fields()) // 默认 InfoLevel，Debug 被过滤
	l.Info("应输出", logx.Fields())
}

// TestGetFallback 覆盖未初始化时的 stderr 降级。
func TestGetFallback(t *testing.T) {
	Reset()
	defer Reset()
	if Get() == nil {
		t.Fatal("降级日志器为空")
	}
}

// TestInitDefaults 覆盖文件配置默认值分支。
func TestInitDefaults(t *testing.T) {
	Reset()
	defer Reset()
	dir := t.TempDir()
	l := Init(Options{LogDir: dir, Filename: "default.log"})
	l.Info("默认配置测试", logx.Fields())
	if err := l.Close(); err != nil {
		t.Fatalf("关闭日志器失败：%v", err)
	}
	files, err := filepath.Glob(filepath.Join(dir, "default-*.log"))
	if err != nil || len(files) == 0 {
		t.Fatalf("日志文件未创建：%v", err)
	}
}
