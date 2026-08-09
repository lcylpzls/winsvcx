package win32

import (
	"syscall"
	"testing"
)

// TestMessageBox 覆盖消息框调用与置顶标志注入。
func TestMessageBox(t *testing.T) {
	origDLL, origProc, origCall := loadUser32, loadProc, callMessageBox
	defer func() { loadUser32, loadProc, callMessageBox = origDLL, origProc, origCall }()

	var gotStyle uint32
	var gotProcName string
	loadUser32 = func() *syscall.LazyDLL { return &syscall.LazyDLL{} }
	loadProc = func(_ *syscall.LazyDLL, name string) *syscall.LazyProc {
		gotProcName = name
		return &syscall.LazyProc{}
	}
	callMessageBox = func(_ *syscall.LazyProc, _ uintptr, _, _ *uint16, style uint32) uintptr {
		gotStyle = style
		return IDYES
	}

	if got := MessageBox("标题", "内容", MB_OK|MB_ICONINFORMATION); got != IDYES {
		t.Fatalf("返回值不符：%d", got)
	}
	if gotProcName != "MessageBoxW" {
		t.Fatalf("过程名不符：%s", gotProcName)
	}
	if gotStyle&(MB_TOPMOST|MB_SETFOREGROUND) == 0 {
		t.Fatalf("置顶标志未附加：%#x", gotStyle)
	}
}

// TestRealProcLookup 覆盖真实 user32.dll 过程查询（不弹窗）。
func TestRealProcLookup(t *testing.T) {
	dll := loadUser32()
	if dll == nil {
		t.Fatal("user32.dll 句柄为空")
	}
	proc := loadProc(dll, "MessageBoxW")
	if proc == nil {
		t.Fatal("MessageBoxW 过程为空")
	}
}
