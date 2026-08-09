//go:build windows

package main

import (
	"errors"
	"testing"

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
	handleServiceCommand("install")
	restore()
	if len(boxes) != 2 || lastBox(boxes).caption != "启动服务成功" {
		t.Fatalf("安装成功流程消息框不符：%+v", boxes)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{installErr: errors.New("已存在"), boxes: &boxes})
	handleServiceCommand("install")
	restore()
	if got := lastBox(boxes); got.caption != "安装服务失败" {
		t.Fatalf("安装失败消息框不符：%+v", got)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{startErr: errors.New("启动失败"), boxes: &boxes})
	handleServiceCommand("install")
	restore()
	if got := lastBox(boxes); got.caption != "启动服务失败" {
		t.Fatalf("安装后启动失败消息框不符：%+v", got)
	}
}

func TestHandleUninstall(t *testing.T) {
	var boxes []boxCall
	restore := applyCmdStubs(cmdStubs{status: svc.Stopped, boxes: &boxes})
	handleServiceCommand("uninstall")
	restore()
	if got := lastBox(boxes); got.caption != "卸载服务成功" {
		t.Fatalf("卸载成功消息框不符：%+v", got)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{status: svc.Running, stopErr: errors.New("停止失败"), boxes: &boxes})
	handleServiceCommand("uninstall")
	restore()
	if got := lastBox(boxes); got.caption != "停止服务失败" {
		t.Fatalf("卸载前停止失败消息框不符：%+v", got)
	}

	boxes = nil
	restore = applyCmdStubs(cmdStubs{status: svc.Running, uninstallErr: errors.New("卸载失败"), boxes: &boxes})
	handleServiceCommand("uninstall")
	restore()
	if got := lastBox(boxes); got.caption != "卸载服务失败" {
		t.Fatalf("卸载失败消息框不符：%+v", got)
	}

	// 状态查询失败不阻塞卸载。
	boxes = nil
	restore = applyCmdStubs(cmdStubs{statusErr: errors.New("查询失败"), boxes: &boxes})
	handleServiceCommand("uninstall")
	restore()
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
		handleServiceCommand(tc.cmd)
		restore()
		if got := lastBox(boxes); got.caption != tc.okCap {
			t.Fatalf("%s 成功消息框不符：%+v", tc.cmd, got)
		}

		boxes = nil
		stub = cmdStubs{boxes: &boxes}
		tc.setErr(&stub, errors.New("失败"))
		restore = applyCmdStubs(stub)
		handleServiceCommand(tc.cmd)
		restore()
		if got := lastBox(boxes); got.caption != tc.failCap {
			t.Fatalf("%s 失败消息框不符：%+v", tc.cmd, got)
		}
	}
}

func TestHandleInvalidCommand(t *testing.T) {
	var boxes []boxCall
	restore := applyCmdStubs(cmdStubs{boxes: &boxes})
	handleServiceCommand("unknown")
	restore()
	if got := lastBox(boxes); got.caption != "无效命令" {
		t.Fatalf("无效命令消息框不符：%+v", got)
	}
}
