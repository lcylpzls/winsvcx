package service

import (
	"errors"
	testx "github.com/lcylpzls/testx"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/lcylpzls/logx"
	"github.com/lcylpzls/winsvcx/lib/config"
	"golang.org/x/sys/windows/svc"
)

// testLogger 构造写入丢弃目标的日志器。
func testLogger() logx.Logger {
	l, err := logx.NewBuilder().EnableWriter(io.Discard, logx.InfoLevel).Build()
	if err != nil {
		panic(err)
	}
	return l
}

// TestExecuteLifecycle 覆盖服务启动、未知命令与停止生命周期。
func TestExecuteLifecycle(t *testing.T) {
	config.Log = testLogger()
	orig := runApp
	runApp = func(stopCh chan struct{}, wg *sync.WaitGroup, _ logx.Logger) {
		wg.Add(1)
		go func() {
			<-stopCh
			wg.Done()
		}()
	}
	defer func() { runApp = orig }()

	s := &Service{}
	req := make(chan svc.ChangeRequest)
	changes := make(chan svc.Status, 8)
	done := make(chan struct{})
	var ssec bool
	var errno uint32
	go func() {
		ssec, errno = s.Execute(nil, req, changes)
		close(done)
	}()

	if st := <-changes; st.State != svc.StartPending {
		t.Fatalf("初始状态应为 StartPending，实际：%v", st.State)
	}
	if st := <-changes; st.State != svc.Running {
		t.Fatalf("启动后状态应为 Running，实际：%v", st.State)
	}
	req <- svc.ChangeRequest{Cmd: svc.Pause}
	req <- svc.ChangeRequest{Cmd: svc.Stop}
	if st := <-changes; st.State != svc.StopPending {
		t.Fatalf("停止后状态应为 StopPending，实际：%v", st.State)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Execute 未在停止后退出")
	}
	if ssec || errno != 0 {
		t.Fatalf("异常退出：ssec=%v errno=%d", ssec, errno)
	}
	if !s.StopFlag {
		t.Fatal("StopFlag 应被置位")
	}
}

// TestRun 覆盖服务运行成功与失败分支。
func TestRun(t *testing.T) {
	config.Log = testLogger()
	orig := runSvc
	origEvent := writeEventLog
	var events []string
	writeEventLog = func(name, msg string) error {
		events = append(events, name+":"+msg)
		return nil
	}
	var ran bool
	runSvc = func(string, svc.Handler) error { ran = true; return nil }
	Run("svc")
	testx.RequireTrue(t, ran)

	runSvc = func(string, svc.Handler) error { return errors.New("运行失败") }
	Run("svc")
	runSvc = orig
	if len(events) != 1 {
		t.Fatalf("事件日志应写入 1 条，实际：%d", len(events))
	}
	writeEventLog = origEvent
}

// TestRunEventLogFailure 覆盖事件日志写入失败分支。
func TestRunEventLogFailure(t *testing.T) {
	config.Log = testLogger()
	orig := runSvc
	origEvent := writeEventLog
	runSvc = func(string, svc.Handler) error { return errors.New("运行失败") }
	writeEventLog = func(string, string) error { return errors.New("事件日志失败") }
	Run("svc")
	runSvc = orig
	writeEventLog = origEvent
}

// TestExecuteChannelClose 覆盖请求通道关闭（未收到停止命令）分支。
func TestExecuteChannelClose(t *testing.T) {
	config.Log = testLogger()
	orig := runApp
	runApp = func(stopCh chan struct{}, wg *sync.WaitGroup, _ logx.Logger) {
		wg.Add(1)
		go func() {
			<-stopCh
			wg.Done()
		}()
	}
	defer func() { runApp = orig }()

	s := &Service{}
	req := make(chan svc.ChangeRequest)
	changes := make(chan svc.Status, 8)
	done := make(chan struct{})
	var ssec bool
	var errno uint32
	go func() {
		ssec, errno = s.Execute(nil, req, changes)
		close(done)
	}()

	if st := <-changes; st.State != svc.StartPending {
		t.Fatalf("初始状态应为 StartPending，实际：%v", st.State)
	}
	if st := <-changes; st.State != svc.Running {
		t.Fatalf("启动后状态应为 Running，实际：%v", st.State)
	}
	close(req)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("通道关闭后 Execute 未退出")
	}
	if ssec || errno != 0 {
		t.Fatalf("异常退出：ssec=%v errno=%d", ssec, errno)
	}
	if s.StopFlag {
		t.Fatal("未收到停止命令时 StopFlag 不应置位")
	}
}

// TestExecuteNilLogFallback 覆盖日志器未注入时的降级（不 panic）。
func TestExecuteNilLogFallback(t *testing.T) {
	orig := config.Log
	config.Log = nil
	defer func() { config.Log = orig }()
	origRun := runApp
	runApp = func(stopCh chan struct{}, wg *sync.WaitGroup, _ logx.Logger) {
		wg.Add(1)
		go func() {
			<-stopCh
			wg.Done()
		}()
	}
	defer func() { runApp = origRun }()

	s := &Service{}
	req := make(chan svc.ChangeRequest)
	changes := make(chan svc.Status, 8)
	done := make(chan struct{})
	go func() {
		_, _ = s.Execute(nil, req, changes)
		close(done)
	}()
	<-changes
	<-changes
	req <- svc.ChangeRequest{Cmd: svc.Stop}
	<-changes
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Execute 未退出")
	}
}
