package app

import (
	"os"
	"os/signal"
	"github.com/lcylpzls/winsvcx/lib/config"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
)

func Run(stopCh chan struct{}, wg *sync.WaitGroup) {
	// 创建系统信号通道，仅在非服务模式下使用
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		defer wg.Done()

		// 创建定时器,每秒执行一次
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				// 收到停止信号
				config.Log.Info("正在停止应用")
				// 如果需要,在这里添加清理代码
				config.Log.Info("应用已停止")
				return
			case <-ticker.C:
				// 定时执行业务逻辑
				config.Log.Info("应用正在运行")
			}
		}
	}()

	// 检查是否为Windows服务
	isWinServ, err := svc.IsWindowsService()
	if err == nil && !isWinServ {
		// 仅在非服务模式下等待系统信号
		<-stopSignal
		// 发送停止信号
		close(stopCh)
	}
	// 在服务模式下，由service.Execute负责关闭stopCh
}
