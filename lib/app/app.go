// Package app 提供服务业务主循环与优雅退出。
package app

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lcylpzls/logx"
	"golang.org/x/sys/windows/svc"
)

// Run 启动应用主循环；服务模式下由服务框架关闭 stopCh，
// 应用模式下等待系统信号后关闭 stopCh。
func Run(stopCh chan struct{}, wg *sync.WaitGroup, logger logx.Logger) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		runLoop(stopCh, logger)
	}()

	isWinServ, err := svc.IsWindowsService()
	if err == nil && !isWinServ {
		stopSignal := make(chan os.Signal, 1)
		signal.Notify(stopSignal, syscall.SIGTERM, syscall.SIGINT)
		<-stopSignal
		close(stopCh)
	}
}

// runLoop 每秒输出运行日志，收到停止信号后优雅退出。
func runLoop(stopCh <-chan struct{}, logger logx.Logger) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stopCh:
			logger.Info("正在停止应用", logx.Fields())
			logger.Info("应用已停止", logx.Fields())
			return
		case <-ticker.C:
			logger.Info("应用正在运行", logx.Fields())
		}
	}
}
