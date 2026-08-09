// Package service 提供 Windows 服务生命周期实现与控制命令。
package service

import (
	"sync"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/logx"
	"github.com/lcylpzls/winsvcx/lib/app"
	"github.com/lcylpzls/winsvcx/lib/config"
	"github.com/lcylpzls/winsvcx/lib/errors"
	"github.com/lcylpzls/winsvcx/lib/logger"
	"golang.org/x/sys/windows/svc"
)

// runApp 可替换的应用启动函数（测试注入用）。
var runApp = app.Run

// runSvc 可替换的服务运行函数（测试注入用）。
var runSvc = svc.Run

// Service 实现 svc.Handler 接口。
type Service struct {
	// StopFlag 标记服务是否收到停止命令。
	StopFlag bool
}

// currentLogger 返回当前日志器；未注入时降级为全局日志器。
func currentLogger() logx.Logger {
	if config.Log != nil {
		return config.Log
	}
	return logger.Get()
}

// Execute 处理服务状态变化与业务启动/停止。
func (s *Service) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	l := currentLogger()
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	var wg sync.WaitGroup
	stopCh := make(chan struct{})
	runApp(stopCh, &wg, l)

	l.Info("服务已启动", logx.Fields())
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	for c := range r {
		switch c.Cmd {
		case svc.Stop, svc.Shutdown:
			s.StopFlag = true
			l.Info("收到停止服务命令", logx.Fields())
			changes <- svc.Status{State: svc.StopPending}
			close(stopCh)
			l.Info("等待应用程序退出", logx.Fields())
			wg.Wait()
			l.Info("应用程序已退出", logx.Fields())
			return
		default:
			l.Warn("无法识别的服务控制命令", logx.Fields(logx.Any("命令", c.Cmd)))
		}
	}
	return
}

// Run 以 Windows 服务模式运行。
func Run(name string) {
	l := currentLogger()
	l.Info("服务启动中", logx.Fields())
	if err := runSvc(name, &Service{}); err != nil {
		l.Error("服务运行失败", logx.Fields(logx.Any("error",
			errx.WrapCode(err, errors.CodeServiceRunFailed, "服务运行失败"))))
	}
	l.Info("服务已退出", logx.Fields())
}
