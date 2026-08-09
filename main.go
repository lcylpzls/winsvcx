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
	installSvc       = service.Install
	uninstallSvc     = service.Uninstall
	startSvc         = service.Start
	stopSvc          = service.Stop
	restartSvc       = service.Restart
	getSvcStatus     = service.GetServiceStatus
	messageBox       = win32.MessageBox
	executablePath   = os.Executable
	isWindowsService = svc.IsWindowsService
	runService       = service.Run
	runApp           = app.Run
	getArgs          = func() []string { return os.Args }
)

// quietMode 安静模式：关闭消息框与控制台输出，仅保留文件日志与退出码。
var quietMode bool

func main() {
	os.Exit(runMain())
}

// parseQuietFlag 解析安静模式参数并返回剩余参数。
// 支持 -quiet / --quiet / /quiet / -q（大小写不敏感，位置任意）。
func parseQuietFlag(args []string) ([]string, bool) {
	quiet := false
	rest := make([]string, 0, len(args))
	for _, a := range args {
		switch strings.ToLower(a) {
		case "-quiet", "--quiet", "/quiet", "-q":
			quiet = true
		default:
			rest = append(rest, a)
		}
	}
	return rest, quiet
}

// runMain 主流程（可测试入口）。
func runMain() int {
	args := getArgs()
	cleanArgs, quiet := parseQuietFlag(args)
	quietMode = quiet
	execPath, err := executablePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法获取可执行文件路径：%v\n", err)
		return 1
	}
	log := logger.Init(logger.Options{
		LogDir:        filepath.Join(filepath.Dir(execPath), "logs"),
		Filename:      "center.log",
		MaxSize:       20,
		MaxBackups:    10,
		MaxAge:        60,
		CompressAfter: 1,
		Level:         logx.DebugLevel,
		Console:       !quietMode,
	})
	config.Log = log

	isWinServ, err := isWindowsService()
	if err != nil {
		config.Log.Error("无法确定是否作为 Windows 服务运行", logx.Fields(logx.Any("error", err)))
		return 1
	}
	if isWinServ {
		runService(serviceName)
		return 0
	}
	if len(cleanArgs) > 1 {
		return handleServiceCommand(cleanArgs[1])
	}

	var wg sync.WaitGroup
	stopCh := make(chan struct{})
	runApp(stopCh, &wg, config.Log)
	wg.Wait()
	config.Log.Info("退出", logx.Fields())
	return 0
}

// notify 展示结果消息框；安静模式下跳过。
func notify(caption, text string, style uint32) {
	if quietMode {
		return
	}
	messageBox(caption, text, style)
}

// handleServiceCommand 处理服务控制命令，返回进程退出码：
// 0 成功，1 操作失败，2 无效命令。
func handleServiceCommand(cmd string) int {
	cmd = strings.ToLower(cmd)
	switch cmd {
	case "install":
		if err := installSvc(serviceName, serviceDisplayName, serviceDescription); err != nil {
			notify("安装服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return 1
		}
		notify("安装服务成功", "服务已成功安装，正在启动服务...", win32.MB_OK|win32.MB_ICONINFORMATION)
		if err := startSvc(serviceName); err != nil {
			notify("启动服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return 1
		}
		notify("启动服务成功", "服务已成功启动", win32.MB_OK|win32.MB_ICONINFORMATION)
		return 0

	case "uninstall":
		if status, statusErr := getSvcStatus(serviceName); statusErr == nil && status == svc.Running {
			if err := stopSvc(serviceName); err != nil {
				notify("停止服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
				return 1
			}
		}
		if err := uninstallSvc(serviceName); err != nil {
			notify("卸载服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return 1
		}
		notify("卸载服务成功", "服务已成功卸载", win32.MB_OK|win32.MB_ICONINFORMATION)
		return 0

	case "start":
		if err := startSvc(serviceName); err != nil {
			notify("启动服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return 1
		}
		notify("启动服务成功", "服务已成功启动", win32.MB_OK|win32.MB_ICONINFORMATION)
		return 0

	case "stop":
		if err := stopSvc(serviceName); err != nil {
			notify("停止服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return 1
		}
		notify("停止服务成功", "服务已成功停止", win32.MB_OK|win32.MB_ICONINFORMATION)
		return 0

	case "restart":
		if err := restartSvc(serviceName); err != nil {
			notify("重启服务失败", err.Error(), win32.MB_OK|win32.MB_ICONERROR)
			return 1
		}
		notify("重启服务成功", "服务已成功重启", win32.MB_OK|win32.MB_ICONINFORMATION)
		return 0

	default:
		notify("无效命令", "不支持的命令: "+cmd+"\n\n支持的命令: install, uninstall, start, stop, restart",
			win32.MB_OK|win32.MB_ICONWARNING)
		return 2
	}
}
