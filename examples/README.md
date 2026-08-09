# winsvcx 使用示例

## 库方式使用

```go
package main

import (
	"os"
	"sync"

	"github.com/lcylpzls/logx"
	"github.com/lcylpzls/winsvcx/lib/app"
	"github.com/lcylpzls/winsvcx/lib/config"
	"github.com/lcylpzls/winsvcx/lib/logger"
	"github.com/lcylpzls/winsvcx/lib/service"
	"golang.org/x/sys/windows/svc"
)

func main() {
	config.Log = logger.Init(logger.Options{
		LogDir:   "logs",
		Filename: "svc.log",
		Console:  true,
	})

	isWinServ, err := svc.IsWindowsService()
	if err != nil {
		config.Log.Error("服务检测失败", logx.Fields(logx.Any("error", err)))
		os.Exit(1)
	}
	if isWinServ {
		service.Run("MyService")
		return
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			// 自定义恢复动作与启动类型。
			err = service.InstallWithOptions("MyService", "我的服务", "示例服务",
				service.DefaultInstallOptions())
		case "uninstall":
			err = service.Uninstall("MyService")
		case "start":
			err = service.Start("MyService")
		case "stop":
			err = service.Stop("MyService")
		case "restart":
			err = service.Restart("MyService")
		}
		if err != nil {
			config.Log.Error("命令执行失败", logx.Fields(logx.Any("error", err)))
			os.Exit(1)
		}
		return
	}

	// 应用模式：业务主循环 + 系统信号优雅退出。
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	app.Run(stopCh, &wg, config.Log)
	wg.Wait()
}
```

## 可用命令

```text
MyService.exe install|uninstall|start|stop|restart
```

安装成功后服务为自动启动，崩溃后按 5s/10s/15s 自动重启，
运行失败会写入 Windows 事件日志。

静默部署：`MyService.exe -quiet install`（关闭弹窗与控制台输出，
退出码 0 成功 / 1 失败 / 2 无效命令）。

版本查询：`MyService.exe -V` 输出 `winsvcx <版本>`。

> 注意：install/uninstall/start/stop/restart 需要管理员权限。
