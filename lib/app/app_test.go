package app

import (
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
