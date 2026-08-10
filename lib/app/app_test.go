package app

import (
	"errors"
	testx "github.com/lcylpzls/testx"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/lcylpzls/logx"
)

// testLogger 构造写入丢弃目标的日志器。
func testLogger() logx.Logger {
	l, err := logx.NewBuilder().EnableWriter(io.Discard, logx.InfoLevel).Build()
	if err != nil {
		panic(err)
	}
	return l
}

// TestRunLoopStop 覆盖收到停止信号后优雅退出。
func TestRunLoopStop(t *testing.T) {
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runLoop(stopCh, testLogger())
	}()
	close(stopCh)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runLoop 未在停止后退出")
	}
}

// TestRunLoopTick 覆盖定时运行日志路径。
func TestRunLoopTick(t *testing.T) {
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runLoop(stopCh, testLogger())
	}()
	time.Sleep(1100 * time.Millisecond)
	close(stopCh)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runLoop 未在停止后退出")
	}
}

// TestRunServiceMode 覆盖服务模式（不等待系统信号）。
func TestRunServiceMode(t *testing.T) {
	origSvc, origWait := isWindowsService, waitSignalAndStop
	isWindowsService = func() (bool, error) { return true, nil }
	waitSignalAndStop = func(chan struct{}) { t.Fatal("服务模式不应等待系统信号") }
	defer func() { isWindowsService, waitSignalAndStop = origSvc, origWait }()

	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	Run(stopCh, &wg, testLogger())
	close(stopCh)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("服务模式 Run 未在停止后完成")
	}
}

// TestRunAppMode 覆盖应用模式（等待系统信号后停止）。
func TestRunAppMode(t *testing.T) {
	origSvc, origWait := isWindowsService, waitSignalAndStop
	isWindowsService = func() (bool, error) { return false, nil }
	closed := false
	waitSignalAndStop = func(ch chan struct{}) {
		closed = true
		close(ch)
	}
	defer func() { isWindowsService, waitSignalAndStop = origSvc, origWait }()

	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	Run(stopCh, &wg, testLogger())
	testx.RequireTrue(t, closed)

	wg.Wait()
}

// TestRunCheckError 覆盖服务检测失败（不阻塞直接返回）。
func TestRunCheckError(t *testing.T) {
	origSvc, origWait := isWindowsService, waitSignalAndStop
	isWindowsService = func() (bool, error) { return false, errors.New("检测失败") }
	waitSignalAndStop = func(chan struct{}) { t.Fatal("检测失败不应等待系统信号") }
	defer func() { isWindowsService, waitSignalAndStop = origSvc, origWait }()

	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	Run(stopCh, &wg, testLogger())
	close(stopCh)
	wg.Wait()
}
