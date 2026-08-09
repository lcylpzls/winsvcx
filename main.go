//go:build windows

package main

import (
	"log"
	"os"
	"path/filepath"
	"github.com/lcylpzls/winsvcx/lib/app"
	"github.com/lcylpzls/winsvcx/lib/config"
	"github.com/lcylpzls/winsvcx/lib/logger"
	"github.com/lcylpzls/winsvcx/lib/service"
	"github.com/lcylpzls/winsvcx/lib/win32"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
)

var serviceName string = "MySampleService"
var serviceDisplayName string = "MySample测试服务"
var serviceDescription string = "这是一个基于go语言构建的具备Windows服务能力的测试服务，具备方便的命令行参数用于管理服务。"

func init() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取可执行文件路径: %v", err)
	}
	execDir := filepath.Dir(execPath)
	logger.InitLogger(filepath.Join(execDir, "logs", "center.log"), 20, 10, 60, true, logrus.DebugLevel)
	config.Log = logger.GetLogger()
}

func main() {
	isWinServ, err := svc.IsWindowsService()
	if err != nil {
		config.Log.Error("无法确定我们是否作为 Windows 服务运行: " + err.Error())
		return
	}

	// 如果是Windows服务，则运行服务
	if isWinServ {
		service.Run(serviceName)
		return
	}

	// 处理命令行参数
	if len(os.Args) > 1 {
		handleServiceCommand(os.Args[1])
		return
	}

	// 如果不是Windows服务，则运行应用程序
	var wg sync.WaitGroup
	stopCh := make(chan struct{})
	wg.Add(1)
	go func() {
		app.Run(stopCh, &wg)
	}()

	wg.Wait()
	config.Log.Info("退出")
}

// handleServiceCommand 处理服务控制命令
func handleServiceCommand(cmd string) {
	var err error
	cmd = strings.ToLower(cmd)

	switch cmd {
	case "install":
		err = service.Install(serviceName, serviceDisplayName, serviceDescription)
		if err != nil {
			win32.MessageBox("安装服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		win32.MessageBox("安装服务成功", "服务已成功安装，正在启动服务...", win32.MB_OK|win32.MB_ICONINFORMATION)

		// 安装成功后尝试启动服务
		err = service.Start(serviceName)
		if err != nil {
			win32.MessageBox("启动服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		win32.MessageBox("启动服务成功", "服务已成功启动", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "uninstall":
		// 卸载前尝试停止服务
		status, statusErr := service.GetServiceStatus(serviceName)
		if statusErr == nil && status == svc.Running {
			err = service.Stop(serviceName)
			if err != nil {
				win32.MessageBox("停止服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
				return
			}
		}

		err = service.Uninstall(serviceName)
		if err != nil {
			win32.MessageBox("卸载服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		win32.MessageBox("卸载服务成功", "服务已成功卸载", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "start":
		err = service.Start(serviceName)
		if err != nil {
			win32.MessageBox("启动服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		win32.MessageBox("启动服务成功", "服务已成功启动", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "stop":
		err = service.Stop(serviceName)
		if err != nil {
			win32.MessageBox("停止服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		win32.MessageBox("停止服务成功", "服务已成功停止", win32.MB_OK|win32.MB_ICONINFORMATION)

	case "restart":
		err = service.Restart(serviceName)
		if err != nil {
			win32.MessageBox("重启服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return
		}
		win32.MessageBox("重启服务成功", "服务已成功重启", win32.MB_OK|win32.MB_ICONINFORMATION)

	default:
		win32.MessageBox("无效命令", "不支持的命令: "+cmd+"\n\n支持的命令: install, uninstall, start, stop, restart", win32.MB_OK|win32.MB_ICONWARNING)
	}
}
