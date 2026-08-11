package service

import (
	stderrors "errors"
	"os"
	"time"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/validx"
	wxerr "github.com/lcylpzls/winsvcx/lib/errors"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// mgrOps 是 *mgr.Mgr 的最小操作接口（供适配器与测试桩实现）。
type mgrOps interface {
	Disconnect() error
	OpenService(name string) (*mgr.Service, error)
	CreateService(name, exePath string, c mgr.Config, args ...string) (*mgr.Service, error)
}

// serviceOps 是 *mgr.Service 的最小操作接口（供适配器与测试桩实现）。
type serviceOps interface {
	Close() error
	Query() (svc.Status, error)
	Start(args ...string) error
	Control(cmd svc.Cmd) (svc.Status, error)
	Delete() error
	SetRecoveryActions(actions []mgr.RecoveryAction, resetPeriod uint32) error
}

// manager 服务管理器窄接口（*mgr.Mgr 经 adapter 适配）。
type manager interface {
	Disconnect() error
	OpenService(name string) (serviceHandle, error)
	CreateService(name, exePath string, c mgr.Config, args ...string) (serviceHandle, error)
}

// serviceHandle 服务句柄窄接口（*mgr.Service 经 adapter 适配）。
type serviceHandle interface {
	Close() error
	Query() (svc.Status, error)
	Start() error
	Control(cmd svc.Cmd) (svc.Status, error)
	Delete() error
	SetRecoveryActions(actions []mgr.RecoveryAction, resetPeriod uint32) error
}

// managerAdapter 适配 mgrOps 到 manager 接口。
type managerAdapter struct{ m mgrOps }

func (a managerAdapter) Disconnect() error { return a.m.Disconnect() }

func (a managerAdapter) OpenService(name string) (serviceHandle, error) {
	s, err := a.m.OpenService(name)
	if err != nil {
		return nil, err
	}
	return serviceAdapter{s: s}, nil
}

func (a managerAdapter) CreateService(name, exePath string, c mgr.Config, args ...string) (serviceHandle, error) {
	s, err := a.m.CreateService(name, exePath, c, args...)
	if err != nil {
		return nil, err
	}
	return serviceAdapter{s: s}, nil
}

// serviceAdapter 适配 serviceOps 到 serviceHandle 接口。
type serviceAdapter struct{ s serviceOps }

func (a serviceAdapter) Close() error { return a.s.Close() }

func (a serviceAdapter) Query() (svc.Status, error) { return a.s.Query() }

func (a serviceAdapter) Start() error { return a.s.Start() }

func (a serviceAdapter) Control(cmd svc.Cmd) (svc.Status, error) { return a.s.Control(cmd) }

func (a serviceAdapter) Delete() error { return a.s.Delete() }

func (a serviceAdapter) SetRecoveryActions(actions []mgr.RecoveryAction, resetPeriod uint32) error {
	return a.s.SetRecoveryActions(actions, resetPeriod)
}

// 可替换系统函数（测试注入用）。
var (
	mgrConnect     = mgr.Connect
	connectManager = func() (manager, error) {
		m, err := mgrConnect()
		if err != nil {
			return nil, err
		}
		return managerAdapter{m: m}, nil
	}
	installEventLog  = eventlog.InstallAsEventCreate
	removeEventLog   = eventlog.Remove
	executablePath   = os.Executable
	stopTimeout      = 30 * time.Second
	stopPollInterval = 300 * time.Millisecond
)

// InstallOptions 安装服务选项。
type InstallOptions struct {
	// StartType 启动类型（mgr.Start* 常量），0 使用自动启动。
	StartType uint32
	// RecoveryActions 崩溃恢复动作，nil 使用默认三次重启。
	RecoveryActions []mgr.RecoveryAction
	// RecoveryResetPeriod 恢复计数重置秒数，0 使用 60。
	RecoveryResetPeriod uint32
	// EventLogTypes 事件日志类别，0 使用错误+警告+信息。
	EventLogTypes uint32
}

// DefaultInstallOptions 返回默认安装选项。
func DefaultInstallOptions() InstallOptions {
	return InstallOptions{
		StartType: mgr.StartAutomatic,
		RecoveryActions: []mgr.RecoveryAction{
			{Type: int(mgr.ServiceRestart), Delay: 5 * time.Second},
			{Type: int(mgr.ServiceRestart), Delay: 10 * time.Second},
			{Type: int(mgr.ServiceRestart), Delay: 15 * time.Second},
		},
		RecoveryResetPeriod: uint32(60),
		EventLogTypes:       eventlog.Error | eventlog.Warning | eventlog.Info,
	}
}

// SetStopTimeout 设置停止等待超时（必须大于 0）。
func SetStopTimeout(d time.Duration) error {
	if d <= 0 {
		return errx.NewCode(wxerr.CodeInvalidConfig, "停止等待超时必须大于 0")
	}
	stopTimeout = d
	return nil
}

// init 注册安装选项校验规则到 validx 全局规则表，错误码保持 winsvcx 语义。
func init() {
	_ = validx.RegisterRule("winsvcx_install_options", func(value any, param, path string) error {
		// 内部调用保证 value 为 InstallOptions。
		opts := value.(InstallOptions)
		if opts.StartType > mgr.StartDisabled {
			return errx.NewCode(wxerr.CodeInvalidConfig, "非法启动类型")
		}
		for _, a := range opts.RecoveryActions {
			if a.Type == 0 || a.Delay <= 0 {
				return errx.NewCode(wxerr.CodeInvalidConfig, "非法恢复动作")
			}
		}
		return nil
	})
}

// validateInstallOptions 校验并补齐安装选项（纯函数，供测试与 fuzz；
// 校验统一走 validx 规则）。
func validateInstallOptions(opts InstallOptions) (InstallOptions, error) {
	def := DefaultInstallOptions()
	if opts.StartType == 0 {
		opts.StartType = def.StartType
	}
	if len(opts.RecoveryActions) == 0 {
		opts.RecoveryActions = def.RecoveryActions
	}
	if opts.RecoveryResetPeriod == 0 {
		opts.RecoveryResetPeriod = def.RecoveryResetPeriod
	}
	if opts.EventLogTypes == 0 {
		opts.EventLogTypes = def.EventLogTypes
	}
	if err := validx.ValidateField(opts, "winsvcx_install_options"); err != nil {
		return opts, err
	}
	return opts, nil
}

// classifyMgrError 按系统错误细化错误码（访问被拒绝 → 权限分类）。
func classifyMgrError(err error, code errx.Code, msg string) error {
	if stderrors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return errx.WrapCode(err, wxerr.CodeAccessDenied, "访问被拒绝："+msg)
	}
	return errx.WrapCode(err, code, msg)
}

// GetServiceStatus 获取服务当前状态。
func GetServiceStatus(name string) (svc.State, error) {
	m, err := connectManager()
	if err != nil {
		return 0, classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return 0, classifyMgrError(err, wxerr.CodeServiceNotFound, "服务不存在："+name)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return 0, classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法查询服务状态")
	}
	return status.State, nil
}

// IsServiceExist 检查服务是否存在；连接失败返回错误而非误报不存在。
func IsServiceExist(name string) (bool, error) {
	m, err := connectManager()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return false, nil
	}
	defer s.Close()
	return true, nil
}

// Install 以默认选项安装服务。
func Install(name, displayName, description string) error {
	return InstallWithOptions(name, displayName, description, DefaultInstallOptions())
}

// InstallWithOptions 按选项安装服务并配置崩溃恢复。
func InstallWithOptions(name, displayName, description string, opts InstallOptions) error {
	exists, err := IsServiceExist(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	if exists {
		return errx.NewCode(wxerr.CodeServiceAlreadyExists, "服务已存在："+name)
	}
	opts, err = validateInstallOptions(opts)
	if err != nil {
		return err
	}

	exePath, err := executablePath()
	if err != nil {
		return classifyMgrError(err, wxerr.CodeExecutablePath, "无法获取可执行文件路径")
	}

	m, err := connectManager()
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.CreateService(name, exePath, mgr.Config{
		DisplayName: displayName,
		Description: description,
		StartType:   opts.StartType,
	})
	if err != nil {
		return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法创建服务")
	}
	defer s.Close()

	err = s.SetRecoveryActions(opts.RecoveryActions, opts.RecoveryResetPeriod)
	if err != nil {
		// 恢复配置失败时回滚已创建的服务。
		_ = s.Delete()
		return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法设置服务恢复操作")
	}

	err = installEventLog(name, opts.EventLogTypes)
	if err != nil {
		// 事件日志创建失败时回滚已创建的服务。
		_ = s.Delete()
		return classifyMgrError(err, wxerr.CodeEventLogFailed, "无法设置事件日志")
	}
	return nil
}

// Uninstall 卸载服务。
func Uninstall(name string) error {
	exists, err := IsServiceExist(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	if !exists {
		return errx.NewCode(wxerr.CodeServiceNotFound, "服务不存在："+name)
	}
	// 运行中的服务先停止，避免删除失败。
	if status, statusErr := GetServiceStatus(name); statusErr == nil && status == svc.Running {
		if err := Stop(name); err != nil {
			return classifyMgrError(err, wxerr.CodeServiceControlFailed, "卸载前停止服务失败")
		}
	}

	m, err := connectManager()
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeServiceNotFound, "无法打开服务："+name)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法删除服务")
	}
	if err := removeEventLog(name); err != nil {
		return classifyMgrError(err, wxerr.CodeEventLogFailed, "无法删除事件日志")
	}
	return nil
}

// Start 启动服务。
func Start(name string) error {
	exists, err := IsServiceExist(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	if !exists {
		return errx.NewCode(wxerr.CodeServiceNotFound, "服务不存在："+name)
	}
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}
	if status == svc.Running {
		return errx.NewCode(wxerr.CodeServiceAlreadyRunning, "服务已在运行："+name)
	}

	m, err := connectManager()
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeServiceNotFound, "无法打开服务："+name)
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法启动服务")
	}
	return nil
}

// Stop 停止服务并等待其完全停止。
func Stop(name string) error {
	exists, err := IsServiceExist(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	if !exists {
		return errx.NewCode(wxerr.CodeServiceNotFound, "服务不存在："+name)
	}
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}
	if status == svc.Stopped {
		return errx.NewCode(wxerr.CodeServiceAlreadyStopped, "服务已停止："+name)
	}

	m, err := connectManager()
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeServiceNotFound, "无法打开服务："+name)
	}
	defer s.Close()

	statusInfo, err := s.Control(svc.Stop)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法发送停止命令")
	}

	startTime := time.Now()
	for statusInfo.State != svc.Stopped {
		if time.Since(startTime) > stopTimeout {
			return errx.NewCode(wxerr.CodeServiceStopTimeout, "等待服务停止超时")
		}
		time.Sleep(stopPollInterval)
		statusInfo, err = s.Query()
		if err != nil {
			return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法查询服务状态")
		}
	}
	return nil
}

// Restart 重启服务（未运行时直接启动）。
func Restart(name string) error {
	exists, err := IsServiceExist(name)
	if err != nil {
		return classifyMgrError(err, wxerr.CodeManagerConnect, "无法连接到服务管理器")
	}
	if !exists {
		return errx.NewCode(wxerr.CodeServiceNotFound, "服务不存在："+name)
	}
	status, err := GetServiceStatus(name)
	if err != nil {
		return err
	}
	if status == svc.Running {
		if err := Stop(name); err != nil {
			return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法停止服务")
		}
	}
	if err := Start(name); err != nil {
		return classifyMgrError(err, wxerr.CodeServiceControlFailed, "无法启动服务")
	}
	return nil
}
