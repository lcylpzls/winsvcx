package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lcylpzls/logx"
	"github.com/lcylpzls/testx"
	"github.com/lcylpzls/winsvcx/lib/config"
	"golang.org/x/sys/windows/svc"
)

// traceCall 记录一次追踪调用。
type traceCall struct {
	name  string
	attrs map[string]string
	err   error
	ended bool
}

// fakeTraceHook 内存追踪钩子。
type fakeTraceHook struct {
	mu    sync.Mutex
	calls []traceCall
}

func (h *fakeTraceHook) Start(ctx context.Context, name string, attrs ...TraceAttr) (context.Context, func(error)) {
	h.mu.Lock()
	h.calls = append(h.calls, traceCall{name: name, attrs: map[string]string{}})
	for _, a := range attrs {
		h.calls[len(h.calls)-1].attrs[a.Key] = a.Value
	}
	h.mu.Unlock()
	return ctx, func(err error) {
		h.mu.Lock()
		h.calls[len(h.calls)-1].err = err
		h.calls[len(h.calls)-1].ended = true
		h.mu.Unlock()
	}
}

func (h *fakeTraceHook) snapshot() []traceCall {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]traceCall, len(h.calls))
	copy(out, h.calls)
	return out
}

// TestServiceTraceHook 覆盖服务生命周期埋点。
func TestServiceTraceHook(t *testing.T) {
	config.Log = testLogger()
	hook := &fakeTraceHook{}
	origRun := runApp
	runApp = func(stopCh chan struct{}, wg *sync.WaitGroup, _ logx.Logger) {
		wg.Add(1)
		go func() {
			<-stopCh
			wg.Done()
		}()
	}
	defer func() { runApp = origRun }()

	s := &Service{Name: "MyService", TraceHook: hook}
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

	calls := hook.snapshot()
	testx.RequireLen(t, calls, 1)
	c := calls[0]
	testx.RequireEqual(t, c.name, "winsvcx.service.execute")
	testx.RequireEqual(t, c.attrs["winsvcx.service_name"], "MyService")
	testx.RequireTrue(t, c.ended)
	testx.RequireNil(t, c.err)
}

// TestRunWithHook 覆盖带钩子运行入口。
func TestRunWithHook(t *testing.T) {
	config.Log = testLogger()
	orig := runSvc
	var got *Service
	runSvc = func(_ string, h svc.Handler) error {
		got = h.(*Service)
		return nil
	}
	hook := &fakeTraceHook{}
	RunWithHook("svc", hook)
	runSvc = orig
	testx.RequireNotNil(t, got)
	testx.RequireEqual(t, got.Name, "svc")
	testx.RequireEqual(t, got.TraceHook, hook)
}

// TestErrnoToError 覆盖退出码映射。
func TestErrnoToError(t *testing.T) {
	testx.RequireNil(t, errnoToError(0))
	testx.RequireNotNil(t, errnoToError(1))
}
