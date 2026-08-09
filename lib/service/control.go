package service

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// GetServiceStatus 获取服务当前状态
func GetServiceStatus(name string) (svc.State, error) {
	m, err := mgr.Connect()
	if err != nil {
		return 0, fmt.Errorf("无法连接到服务管理器: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return 0, fmt.Errorf("服务 %s 不存在: %v", name, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return 0, fmt.Errorf("无法查询服务状态: %v", err)
	}
	return status.State, nil
}

// IsServiceExist 检查服务是否存在
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

// Install 安装服务
func Install(name, displayName, description string) error {
	// 检查服务是否已经存在
	if IsServiceExist(name) {
		return fmt.Errorf("服务 %s 已存在", name)
	}

	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取可执行文件路径: %v", err)
	}

	// 连接服务管理器
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("无法连接到服务管理器: %v", err)
	}
	defer m.Disconnect()

	// 创建服务
	s, err := m.CreateService(name, exePath, mgr.Config{
		DisplayName: displayName,
		Description: description,
		StartType:   mgr.StartAutomatic,
	})
	if err != nil {
		return fmt.Errorf("无法创建服务: %v", err)
	}
	defer s.Close()

	// 设置服务恢复操作
	err = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 10 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 15 * time.Second},
	}, uint32(60)) // 这里需要使用 uint32 类型的秒数
	if err != nil {
		return fmt.Errorf("无法设置服务恢复操作: %v", err)
	}

	// 创建事件日志
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		// 如果事件日志创建失败，尝试删除服务
		_ = s.Delete()
		return fmt.Errorf("无法设置事件日志: %v", err)
	}

	return nil
}

// Uninstall 卸载服务
func Uninstall(name string) error {
	// 检查服务是否存在
	if !IsServiceExist(name) {
		return fmt.Errorf("服务 %s 不存在", name)
	}

	// 连接服务管理器
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("无法连接到服务管理器: %v", err)
	}
	defer m.Disconnect()

	// 打开服务
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("无法打开服务: %v", err)
	}
	defer s.Close()

	// 删除服务
	err = s.Delete()
	if err != nil {
		return fmt.Errorf("无法删除服务: %v", err)
	}

	// 删除事件日志
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("无法删除事件日志: %v", err)
	}

	return nil
}

// Start 启动服务
func Start(name string) error {
	// 检查服务是否存在
	if !IsServiceExist(name) {
		return fmt.Errorf("服务 %s 不存在", name)
	}

	// 获取服务状态
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}

	// 如果服务已经在运行，直接返回
	if status == svc.Running {
		return fmt.Errorf("服务 %s 已经在运行中", name)
	}

	// 连接服务管理器
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("无法连接到服务管理器: %v", err)
	}
	defer m.Disconnect()

	// 打开服务
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("无法打开服务: %v", err)
	}
	defer s.Close()

	// 启动服务
	err = s.Start()
	if err != nil {
		return fmt.Errorf("无法启动服务: %v", err)
	}

	return nil
}

// Stop 停止服务
func Stop(name string) error {
	// 检查服务是否存在
	if !IsServiceExist(name) {
		return fmt.Errorf("服务 %s 不存在", name)
	}

	// 获取服务状态
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}

	// 如果服务已经停止，直接返回
	if status == svc.Stopped {
		return fmt.Errorf("服务 %s 已经停止", name)
	}

	// 连接服务管理器
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("无法连接到服务管理器: %v", err)
	}
	defer m.Disconnect()

	// 打开服务
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("无法打开服务: %v", err)
	}
	defer s.Close()

	// 构建服务状态请求
	timeout := 30 * time.Second
	statusInfo, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("无法发送停止命令: %v", err)
	}

	// 等待服务停止
	startTime := time.Now()
	for statusInfo.State != svc.Stopped {
		if time.Since(startTime) > timeout {
			return fmt.Errorf("等待服务停止超时")
		}
		time.Sleep(300 * time.Millisecond)
		statusInfo, err = s.Query()
		if err != nil {
			return fmt.Errorf("无法查询服务状态: %v", err)
		}
	}

	return nil
}

// Restart 重启服务
func Restart(name string) error {
	// 检查服务是否存在
	if !IsServiceExist(name) {
		return fmt.Errorf("服务 %s 不存在", name)
	}

	// 获取服务状态
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}

	// 如果服务正在运行，先停止
	if status == svc.Running {
		err = Stop(name)
		if err != nil {
			return fmt.Errorf("无法停止服务: %v", err)
		}
	}

	// 启动服务
	err = Start(name)
	if err != nil {
		return fmt.Errorf("无法启动服务: %v", err)
	}

	return nil
}
