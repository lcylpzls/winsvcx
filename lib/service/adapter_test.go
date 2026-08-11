package service

import (
	"errors"
	"testing"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// fakeMgrOps 是 mgrOps 的测试桩（不触碰真实 SCM）。
type fakeMgrOps struct {
	openErr   error
	createErr error
}

func (m *fakeMgrOps) Disconnect() error { return nil }

func (m *fakeMgrOps) OpenService(string) (*mgr.Service, error) {
	if m.openErr != nil {
		return nil, m.openErr
	}
	return &mgr.Service{}, nil
}

func (m *fakeMgrOps) CreateService(string, string, mgr.Config, ...string) (*mgr.Service, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &mgr.Service{}, nil
}

// fakeServiceOps 是 serviceOps 的测试桩（不触碰真实服务句柄）。
type fakeServiceOps struct{}

func (fakeServiceOps) Close() error { return nil }

func (fakeServiceOps) Query() (svc.Status, error) {
	return svc.Status{State: svc.Running}, nil
}

func (fakeServiceOps) Start(...string) error { return nil }

func (fakeServiceOps) Control(svc.Cmd) (svc.Status, error) {
	return svc.Status{State: svc.Stopped}, nil
}

func (fakeServiceOps) Delete() error { return nil }

func (fakeServiceOps) SetRecoveryActions([]mgr.RecoveryAction, uint32) error { return nil }

// TestManagerAdapterMethods 覆盖管理器适配器的成功与失败分支。
func TestManagerAdapterMethods(t *testing.T) {
	ops := &fakeMgrOps{}
	m := managerAdapter{m: ops}

	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect 失败：%v", err)
	}

	h, err := m.OpenService("svc")
	if err != nil || h == nil {
		t.Fatalf("OpenService 成功分支失败：%v", err)
	}
	ops.openErr = errors.New("打开失败")
	if _, err := m.OpenService("svc"); err == nil {
		t.Fatal("OpenService 失败分支未生效")
	}

	h, err = m.CreateService("svc", `C:\x.exe`, mgr.Config{})
	if err != nil || h == nil {
		t.Fatalf("CreateService 成功分支失败：%v", err)
	}
	ops.createErr = errors.New("创建失败")
	if _, err := m.CreateService("svc", `C:\x.exe`, mgr.Config{}); err == nil {
		t.Fatal("CreateService 失败分支未生效")
	}
}

// TestServiceAdapterMethods 覆盖服务句柄适配器全部方法。
func TestServiceAdapterMethods(t *testing.T) {
	s := serviceAdapter{s: fakeServiceOps{}}

	if err := s.Close(); err != nil {
		t.Fatalf("Close 失败：%v", err)
	}
	st, err := s.Query()
	if err != nil || st.State != svc.Running {
		t.Fatalf("Query 失败：%v %v", st.State, err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start 失败：%v", err)
	}
	st, err = s.Control(svc.Stop)
	if err != nil || st.State != svc.Stopped {
		t.Fatalf("Control 失败：%v %v", st.State, err)
	}
	if err := s.Delete(); err != nil {
		t.Fatalf("Delete 失败：%v", err)
	}
	if err := s.SetRecoveryActions(nil, 0); err != nil {
		t.Fatalf("SetRecoveryActions 失败：%v", err)
	}
}
