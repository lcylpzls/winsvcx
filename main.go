//go:build windows

// winsvcx 示例服务：基于 x/sys/windows/svc 的 Windows 服务框架入口。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lcylpzls/logx"
	"github.com/lcylpzls/winsvcx/lib/app"
	"github.com/lcylpzls/winsvcx/lib/config"
	"github.com/lcylpzls/winsvcx/lib/logger"
	"github.com/lcylpzls/winsvcx/lib/service"
	"github.com/lcylpzls/winsvcx/lib/win32"
	"golang.org/x/sys/windows/svc"
)

var (
	serviceName        = "MySampleService"
	serviceDisplayName = "MySample测试服务"
	serviceDescription = "这是一个基于 go 语言构建的具备 Windows 服务能力的测试服务，具备方便的命令行参数用于管理服务。"
)

// 可替换的服务命令与消息框（测试注入用）。
var (
	installSvc   = service.Install
	uninstallSvc = service.Uninstall
	startSvc     = service.Start
	stopSvc      = service.Stop
	restartSvc   = service.Restart
	getSvcStatus = service.GetServiceStatus
	messageBox   = win32.MessageBox
)

func main() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法获取可执行文件路径：%v\n", err)
		os.Exit(1)
	}
	log := logger.Init(logger.Options{
		LogDir:        filepath.Join(filepath.Dir(execPath), "logs"),
		Filename:      "center.log",
		MaxSize:       20,
		MaxBackups:    10,
		MaxAge:        60,
		CompressAfter: 1,
		Level:         logx.DebugLevel,
		Console:       true,
	})
	config.Log = log

	isWinServ, err := svc.IsWindowsService()
	if err != nil {
		config.Log.Error("无法确定是否作为 Windows 服务运行", logx.Fields(logx.Any("error", err)))
		return
	}
	if isWinServ {
		service.Run(serviceName)
		return
	}
	if len(os.Args) > 1 {
		handleServiceCommand(os.Args[1])
		return
	}

	var wg sync.WaitGroup
	stopCh := make(chan struct{})
	app.Run(stopCh, &wg, config.Log)
	wg.Wait()
	config.Log.Info("退出", logx.Fields())
}

// handleServiceCommand 处理服务控制命令。
func handleServiceCommand(cmd string) {
	cmd = strings.ToLower(cmd)
	switch cmd {
	case "install":
		if err := installSvc(serviceName, serviceDisplayName, serviceDescription); err != nil {
			messageBox("安装服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		messageBox("安装服务成功", "服务已成功安装，正在启动服务...", win32.MB_OK|win32.MB_ICONINFORMATION)
		if err := startSvc(serviceName); err != nil {
			messageBox("启动服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		messageBox("启动服务成功", "服务已成功启动", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "uninstall":
		if status, statusErr := getSvcStatus(serviceName); statusErr == nil && status == svc.Running {
			if err := stopSvc(serviceName); err != nil {
				messageBox("停止服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
				return
			}
		}
		if err := uninstallSvc(serviceName); err != nil {
			messageBox("卸载服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		messageBox("卸载服务成功", "服务已成功卸载", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "start":
		if err := startSvc(serviceName); err != nil {
			messageBox("启动服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		messageBox("启动服务成功", "服务已成功启动", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "stop":
		if err := stopSvc(serviceName); err != nil {
			messageBox("停止服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		messageBox("停止服务成功", "服务已成功停止", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "restart":
		if err := restartSvc(serviceName); err != nil {
			messageBox("重启服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		messageBox("重启服务成功", "服务已成功重启", win32.MB_OK|win32.MB_ICONINFORMATION)

	default:
		messageBox("无效命令", "不支持的命令: "+cmd+"\n\n支持的命令: install, uninstall, start, stop, restart",
			win32.MB_OK|win32.MB_ICONWARNING)
	}
}
