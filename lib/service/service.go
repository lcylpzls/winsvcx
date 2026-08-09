package service

import (
	"fmt"
	"github.com/lcylpzls/winsvcx/lib/app"
	"github.com/lcylpzls/winsvcx/lib/config"
	"sync"

	"golang.org/x/sys/windows/svc"
)

type Service struct {
	StopFlag bool
}

func (s *Service) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	var wg sync.WaitGroup
	stopCh := make(chan struct{})
	wg.Add(1)
	go func() {
		app.Run(stopCh, &wg)
	}()

	config.Log.Info("服务已启动")

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	for c := range r {
		switch c.Cmd {
		case svc.Stop, svc.Shutdown:
			s.StopFlag = true
			config.Log.Info("收到停止服务命令")
			changes <- svc.Status{State: svc.StopPending}
			// 关闭stopCh通道，通知应用程序停止
			close(stopCh)
			// 等待应用程序退出
			config.Log.Info("等待应用程序退出...")
			wg.Wait()
			config.Log.Info("应用程序已退出")
			return
		default:
			config.Log.Warning("无法识别命令: " + fmt.Sprint(c.Cmd))
		}
	}
	return
}

func Run(name string) {
	config.Log.Info("服务启动中...")
	err := svc.Run(name, &Service{})
	if err != nil {
		config.Log.Error("启动服务失败: " + err.Error())
	}
	config.Log.Info("服务已退出。")
}
