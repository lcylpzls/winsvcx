package service

import (
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
