package service

import (
	"os"
	"time"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/winsvcx/lib/errors"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// GetServiceStatus 获取服务当前状态。
func GetServiceStatus(name string) (svc.State, error) {
	m, err := mgr.Connect()
	if err != nil {
		return 0, errx.WrapCode(err, errors.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return 0, errx.WrapCode(err, errors.CodeServiceNotFound, "服务不存在："+name)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return 0, errx.WrapCode(err, errors.CodeServiceControlFailed, "无法查询服务状态")
	}
	return status.State, nil
}

// IsServiceExist 检查服务是否存在。
func IsServiceExist(name string) bool {
	m, err := mgr.Connect()
	if err != nil {
		return false
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return false
	}
	defer s.Close()
	return true
}

// Install 安装服务并配置自动启动与崩溃恢复。
func Install(name, displayName, description string) error {
	if IsServiceExist(name) {
		return errx.NewCode(errors.CodeServiceAlreadyExists, "服务已存在："+name)
	}

	exePath, err := os.Executable()
	if err != nil {
		return errx.WrapCode(err, errors.CodeExecutablePath, "无法获取可执行文件路径")
	}

	m, err := mgr.Connect()
	if err != nil {
		return errx.WrapCode(err, errors.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.CreateService(name, exePath, mgr.Config{
		DisplayName: displayName,
		Description: description,
		StartType:   mgr.StartAutomatic,
	})
	if err != nil {
		return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法创建服务")
	}
	defer s.Close()

	err = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 10 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 15 * time.Second},
	}, uint32(60))
	if err != nil {
		return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法设置服务恢复操作")
	}

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		// 事件日志创建失败时回滚已创建的服务。
		_ = s.Delete()
		return errx.WrapCode(err, errors.CodeEventLogFailed, "无法设置事件日志")
	}
	return nil
}

// Uninstall 卸载服务。
func Uninstall(name string) error {
	if !IsServiceExist(name) {
		return errx.NewCode(errors.CodeServiceNotFound, "服务不存在："+name)
	}

	m, err := mgr.Connect()
	if err != nil {
		return errx.WrapCode(err, errors.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return errx.WrapCode(err, errors.CodeServiceNotFound, "无法打开服务："+name)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法删除服务")
	}
	if err := eventlog.Remove(name); err != nil {
		return errx.WrapCode(err, errors.CodeEventLogFailed, "无法删除事件日志")
	}
	return nil
}

// Start 启动服务。
func Start(name string) error {
	if !IsServiceExist(name) {
		return errx.NewCode(errors.CodeServiceNotFound, "服务不存在："+name)
	}
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}
	if status == svc.Running {
		return errx.NewCode(errors.CodeServiceAlreadyRunning, "服务已在运行："+name)
	}

	m, err := mgr.Connect()
	if err != nil {
		return errx.WrapCode(err, errors.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return errx.WrapCode(err, errors.CodeServiceNotFound, "无法打开服务："+name)
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法启动服务")
	}
	return nil
}

// Stop 停止服务并等待其完全停止。
func Stop(name string) error {
	if !IsServiceExist(name) {
		return errx.NewCode(errors.CodeServiceNotFound, "服务不存在："+name)
	}
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}
	if status == svc.Stopped {
		return errx.NewCode(errors.CodeServiceAlreadyStopped, "服务已停止："+name)
	}

	m, err := mgr.Connect()
	if err != nil {
		return errx.WrapCode(err, errors.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return errx.WrapCode(err, errors.CodeServiceNotFound, "无法打开服务："+name)
	}
	defer s.Close()

	timeout := 30 * time.Second
	statusInfo, err := s.Control(svc.Stop)
	if err != nil {
		return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法发送停止命令")
	}

	startTime := time.Now()
	for statusInfo.State != svc.Stopped {
		if time.Since(startTime) > timeout {
			return errx.NewCode(errors.CodeServiceStopTimeout, "等待服务停止超时")
		}
		time.Sleep(300 * time.Millisecond)
		statusInfo, err = s.Query()
		if err != nil {
			return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法查询服务状态")
		}
	}
	return nil
}

// Restart 重启服务（未运行时直接启动）。
func Restart(name string) error {
	if !IsServiceExist(name) {
		return errx.NewCode(errors.CodeServiceNotFound, "服务不存在："+name)
	}
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}
	if status == svc.Running {
		if err := Stop(name); err != nil {
			return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法停止服务")
		}
	}
	if err := Start(name); err != nil {
		return errx.WrapCode(err, errors.CodeServiceControlFailed, "无法启动服务")
	}
	return nil
}
