//go:build windows

package win32

import (
	"syscall"
	"unsafe"
)

// 可替换的 Win32 调用（测试注入用）。
var (
	loadUser32 = func() *syscall.LazyDLL { return syscall.NewLazyDLL("user32.dll") }
	loadProc   = func(dll *syscall.LazyDLL, name string) *syscall.LazyProc {
		return dll.NewProc(name)
	}
	callMessageBox = func(proc *syscall.LazyProc, hwnd uintptr, text, caption *uint16, style uint32) uintptr {
		ret, _, _ := proc.Call(hwnd, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(caption)), uintptr(style))
		return ret
	}
)

// 按钮类型常量
const (
	MB_OK               = 0x00000000
	MB_OKCANCEL         = 0x00000001
	MB_ABORTRETRYIGNORE = 0x00000002
	MB_YESNOCANCEL      = 0x00000003
	MB_YESNO            = 0x00000004
	MB_RETRYCANCEL      = 0x00000005
)

// 图标类型常量
const (
	MB_ICONERROR       = 0x00000010
	MB_ICONQUESTION    = 0x00000020
	MB_ICONWARNING     = 0x00000030
	MB_ICONINFORMATION = 0x00000040
)

// 其他特殊属性常量
const (
	MB_SYSTEMMODAL          = 0x00001000 // 系统模态
	MB_TOPMOST              = 0x00040000 // 置顶显示
	MB_SETFOREGROUND        = 0x00010000 // 设置前台
	MB_DEFAULT_DESKTOP_ONLY = 0x00020000 // 仅在默认桌面显示
)

// MessageBox 显示一个消息对话框
// caption: 对话框标题
// text: 对话框内容
// style: 对话框样式（按钮类型|图标类型|其他属性）
// 返回值: 用户点击的按钮ID
func MessageBox(caption, text string, style uint32) int {
	user32 := loadUser32()
	messageBox := loadProc(user32, "MessageBoxW")

	title, _ := syscall.UTF16PtrFromString(caption)
	message, _ := syscall.UTF16PtrFromString(text)

	// 添加 MB_TOPMOST 和 MB_SETFOREGROUND 确保窗口置顶
	style |= MB_TOPMOST | MB_SETFOREGROUND

	return int(callMessageBox(messageBox, 0, message, title, style))
}

// 返回值常量
const (
	IDOK     = 1
	IDCANCEL = 2
	IDABORT  = 3
	IDRETRY  = 4
	IDIGNORE = 5
	IDYES    = 6
	IDNO     = 7
)
