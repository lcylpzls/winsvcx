//go:build windows

package main

import (
	"errors"
	"sync"
	"testing"

	"github.com/lcylpzls/logx"
	"github.com/lcylpzls/winsvcx/lib/cli"
	"github.com/lcylpzls/winsvcx/lib/win32"
	"golang.org/x/sys/windows/svc"
)

// boxCall 记录消息框调用。
type boxCall struct {
	caption string
	text    string
	style   uint32
}

// cmdStubs 命令依赖测试桩。
type cmdStubs struct {
	installErr   error
	uninstallErr error
	startErr     error
	stopErr      error
	restartErr   error
	status       svc.State
	statusErr    error
	boxes        *[]boxCall
}

// applyCmdStubs 注入命令依赖并返回恢复函数。
func applyCmdStubs(s cmdStubs) func() {
	origInstall, origUninstall, origStart, origStop, origRestart :=
		installSvc, uninstallSvc, startSvc, stopSvc, restartSvc
	origStatus, origBox := getSvcStatus, messageBox
	installSvc = func(string, string, string) error { return s.installErr }
	uninstallSvc = func(string) error { return s.uninstallErr }
	startSvc = func(string) error { return s.startErr }
	stopSvc = func(string) error { return s.stopErr }
	restartSvc = func(string) error { return s.restartErr }
	getSvcStatus = func(string) (svc.State, error) { return s.status, s.statusErr }
	messageBox = func(caption, text string, style uint32) int {
		*s.boxes = append(*s.boxes, boxCall{caption: caption, text: text, style: style})
		return win32.IDOK
	}
	return func() {
		installSvc, uninstallSvc, startSvc, stopSvc, restartSvc =
			origInstall, origUninstall, origStart, origStop, origRestart
		getSvcStatus, messageBox = origStatus, origBox
	}
}

// lastBox 返回最后一条消息框记录。
func lastBox(boxes []boxCall) boxCall {
	if len(boxes) == 0 {
		return boxCall{}
	}
	return boxes[len(boxes)-1]
}

func TestHandleInstall(t *testing.T) {
	var boxes []boxCall
	restore := applyCmdStubs(cmdStubs{boxes: &boxes})
	code := handleServiceCommand("install")
	restore()
	if code != 0 || len(boxes) != 2 || lastBox(boxes).caption != "启动服务成功" {
		t.Fatalf("安装成功流程消息框不符：%+v", boxes)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{installErr: errors.New("已存在"), boxes: &boxes})
	code = handleServiceCommand("install")
	restore()
	if code != 1 {
		t.Fatalf("安装失败应返回 1，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "安装服务失败" {
		t.Fatalf("安装失败消息框不符：%+v", got)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{startErr: errors.New("启动失败"), boxes: &boxes})
	code = handleServiceCommand("install")
	restore()
	if code != 1 {
		t.Fatalf("安装后启动失败应返回 1，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "启动服务失败" {
		t.Fatalf("安装后启动失败消息框不符：%+v", got)
	}
}

func TestHandleUninstall(t *testing.T) {
	var boxes []boxCall
	restore := applyCmdStubs(cmdStubs{status: svc.Stopped, boxes: &boxes})
	code := handleServiceCommand("uninstall")
	restore()
	if code != 0 {
		t.Fatalf("卸载成功应返回 0，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "卸载服务成功" {
		t.Fatalf("卸载成功消息框不符：%+v", got)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{status: svc.Running, stopErr: errors.New("停止失败"), boxes: &boxes})
	code = handleServiceCommand("uninstall")
	restore()
	if code != 1 {
		t.Fatalf("卸载前停止失败应返回 1，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "停止服务失败" {
		t.Fatalf("卸载前停止失败消息框不符：%+v", got)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{status: svc.Running, uninstallErr: errors.New("卸载失败"), boxes: &boxes})
	code = handleServiceCommand("uninstall")
	restore()
	if code != 1 {
		t.Fatalf("卸载失败应返回 1，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "卸载服务失败" {
		t.Fatalf("卸载失败消息框不符：%+v", got)
	}

	// 状态查询失败不阻塞卸载。
	boxes = nil
	restore = applyCmdStubs(cmdStubs{statusErr: errors.New("查询失败"), boxes: &boxes})
	code = handleServiceCommand("uninstall")
	restore()
	if code != 0 {
		t.Fatalf("状态查询失败应继续卸载，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "卸载服务成功" {
		t.Fatalf("状态查询失败应继续卸载：%+v", got)
	}
}

func TestHandleStartStopRestart(t *testing.T) {
	for _, tc := range []struct {
		cmd     string
		err     error
		okCap   string
		failCap string
		setErr  func(*cmdStubs, error)
	}{
		{cmd: "start", okCap: "启动服务成功", failCap: "启动服务失败", setErr: func(s *cmdStubs, e error) { s.startErr = e }},
		{cmd: "stop", okCap: "停止服务成功", failCap: "停止服务失败", setErr: func(s *cmdStubs, e error) { s.stopErr = e }},
		{cmd: "restart", okCap: "重启服务成功", failCap: "重启服务失败", setErr: func(s *cmdStubs, e error) { s.restartErr = e }},
	} {
		var boxes []boxCall
		stub := cmdStubs{boxes: &boxes}
		restore := applyCmdStubs(stub)
		code := handleServiceCommand(tc.cmd)
		restore()
		if code != 0 {
			t.Fatalf("%s 成功应返回 0，实际：%d", tc.cmd, code)
		}
		if got := lastBox(boxes); got.caption != tc.okCap {
			t.Fatalf("%s 成功消息框不符：%+v", tc.cmd, got)
		}

		boxes = nil
		stub = cmdStubs{boxes: &boxes}
		tc.setErr(&stub, errors.New("失败"))
		restore = applyCmdStubs(stub)
		code = handleServiceCommand(tc.cmd)
		restore()
		if code != 1 {
			t.Fatalf("%s 失败应返回 1，实际：%d", tc.cmd, code)
		}
		if got := lastBox(boxes); got.caption != tc.failCap {
			t.Fatalf("%s 失败消息框不符：%+v", tc.cmd, got)
		}
	}
}

func TestHandleInvalidCommand(t *testing.T) {
	var boxes []boxCall
	restore := applyCmdStubs(cmdStubs{boxes: &boxes})
	code := handleServiceCommand("unknown")
	restore()
	if code != 2 {
		t.Fatalf("无效命令应返回 2，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "无效命令" {
		t.Fatalf("无效命令消息框不符：%+v", got)
	}
}

// TestHandleQuietMode 覆盖安静模式：不弹窗、退出码正常。
func TestHandleQuietMode(t *testing.T) {
	origQuiet := quietMode
	quietMode = true
	defer func() { quietMode = origQuiet }()

	var boxes []boxCall
	restore := applyCmdStubs(cmdStubs{boxes: &boxes})
	if code := handleServiceCommand("install"); code != 0 {
		t.Fatalf("安静安装应返回 0，实际：%d", code)
	}
	restore()
	if len(boxes) != 0 {
		t.Fatalf("安静模式不应弹窗：%+v", boxes)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{startErr: errors.New("失败"), boxes: &boxes})
	if code := handleServiceCommand("start"); code != 1 {
		t.Fatalf("安静失败应返回 1，实际：%d", code)
	}
	restore()
	if len(boxes) != 0 {
		t.Fatalf("安静模式失败也不应弹窗：%+v", boxes)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{boxes: &boxes})
	if code := handleServiceCommand("bad"); code != 2 {
		t.Fatalf("安静无效命令应返回 2，实际：%d", code)
	}
	restore()
	if len(boxes) != 0 {
		t.Fatalf("安静无效命令不应弹窗：%+v", boxes)
	}
}

// mainStubs 主流程依赖测试桩。
type mainStubs struct {
	exeErr   error
	svcErr   error
	isSvc    bool
	args     []string
	runSvc   func(string)
	appRun   func(chan struct{}, *sync.WaitGroup, logx.Logger)
	boxes    *[]boxCall
	startErr error
	quiet    bool
}

// applyMainStubs 注入主流程依赖并返回恢复函数。
func applyMainStubs(s mainStubs) func() {
	origExe, origSvc, origRunSvc, origApp, origArgs, origBox, origStart :=
		executablePath, isWindowsService, runService, runApp, getArgs, messageBox, startSvc
	origQuiet := quietMode
	quietMode = s.quiet
	executablePath = func() (string, error) {
		if s.exeErr != nil {
			return "", s.exeErr
		}
		return `C:\test\app.exe`, nil
	}
	isWindowsService = func() (bool, error) { return s.isSvc, s.svcErr }
	runService = s.runSvc
	runApp = s.appRun
	getArgs = func() []string { return s.args }
	messageBox = func(caption, text string, style uint32) int {
		*s.boxes = append(*s.boxes, boxCall{caption: caption, text: text, style: style})
		return win32.IDOK
	}
	startSvc = func(string) error { return s.startErr }
	return func() {
		executablePath, isWindowsService, runService, runApp, getArgs, messageBox, startSvc =
			origExe, origSvc, origRunSvc, origApp, origArgs, origBox, origStart
		quietMode = origQuiet
	}
}

func TestRunMainExecutableError(t *testing.T) {
	restore := applyMainStubs(mainStubs{exeErr: errors.New("路径失败")})
	code := runMain()
	restore()
	if code != 1 {
		t.Fatalf("可执行文件路径失败应返回 1，实际：%d", code)
	}
}

func TestRunMainServiceCheckError(t *testing.T) {
	restore := applyMainStubs(mainStubs{svcErr: errors.New("检测失败")})
	code := runMain()
	restore()
	if code != 1 {
		t.Fatalf("服务检测失败应返回 1，实际：%d", code)
	}
}

func TestRunMainServiceMode(t *testing.T) {
	called := false
	restore := applyMainStubs(mainStubs{
		isSvc:  true,
		runSvc: func(string) { called = true },
	})
	code := runMain()
	restore()
	if code != 0 || !called {
		t.Fatalf("服务模式应调用 runService：code=%d called=%v", code, called)
	}
}

func TestRunMainCommandMode(t *testing.T) {
	var boxes []boxCall
	restore := applyMainStubs(mainStubs{
		args:  []string{"app.exe", "start"},
		boxes: &boxes,
	})
	code := runMain()
	restore()
	if code != 0 || lastBox(boxes).caption != "启动服务成功" {
		t.Fatalf("命令模式应分发命令：code=%d boxes=%+v", code, boxes)
	}
}

// TestRunMainQuietCommand 覆盖安静模式下命令分发不弹窗。
func TestRunMainQuietCommand(t *testing.T) {
	var boxes []boxCall
	restore := applyMainStubs(mainStubs{
		args:  []string{"app.exe", "-quiet", "start"},
		boxes: &boxes,
		quiet: true,
	})
	code := runMain()
	restore()
	if code != 0 {
		t.Fatalf("安静命令应返回 0，实际：%d", code)
	}
	if len(boxes) != 0 {
		t.Fatalf("安静模式不应弹窗：%+v", boxes)
	}
}

// TestRunMainInvalidCommand 覆盖无效命令与未知开关。
func TestRunMainInvalidCommand(t *testing.T) {
	var boxes []boxCall
	restore := applyMainStubs(mainStubs{args: []string{"app.exe", "bad"}, boxes: &boxes})
	code := runMain()
	restore()
	if code != 2 {
		t.Fatalf("无效命令应返回 2，实际：%d", code)
	}
	if got := lastBox(boxes); got.caption != "无效命令" {
		t.Fatalf("无效命令消息框不符：%+v", got)
	}

	boxes = nil
	restore = applyMainStubs(mainStubs{args: []string{"app.exe", "-quiet", "-verbose"}, boxes: &boxes})
	code = runMain()
	restore()
	if code != 2 {
		t.Fatalf("未知开关应返回 2，实际：%d", code)
	}
	if len(boxes) != 0 {
		t.Fatalf("安静无效参数不应弹窗：%+v", boxes)
	}
}

// TestRunMainVersion 覆盖版本号输出。
func TestRunMainVersion(t *testing.T) {
	origPrint := printVersion
	var printed string
	printVersion = func() { printed = "winsvcx " + cli.Version }
	defer func() { printVersion = origPrint }()

	restore := applyMainStubs(mainStubs{args: []string{"app.exe", "-V"}})
	code := runMain()
	restore()
	if code != 0 || printed == "" {
		t.Fatalf("版本参数应输出并返回 0：code=%d printed=%q", code, printed)
	}
}

func TestRunMainAppMode(t *testing.T) {
	restore := applyMainStubs(mainStubs{
		appRun: func(stopCh chan struct{}, wg *sync.WaitGroup, _ logx.Logger) {
			wg.Add(1)
			go func() {
				close(stopCh)
				wg.Done()
			}()
		},
	})
	code := runMain()
	restore()
	if code != 0 {
		t.Fatalf("应用模式应返回 0，实际：%d", code)
	}
}
