package service

import (
	"errors"
	"testing"
	"time"

	"github.com/lcylpzls/errx"
	wxerr "github.com/lcylpzls/winsvcx/lib/errors"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// fakeService 可编程的服务句柄测试桩。
type fakeService struct {
	status       svc.State
	queryErr     error
	startErr     error
	controlErr   error
	controlState svc.State
	deleteErr    error
	recoveryErr  error
	recovery     []mgr.RecoveryAction
	resetPeriod  uint32
	closed       bool
	deleted      bool
}

func (s *fakeService) Close() error {
	s.closed = true
	return nil
}

func (s *fakeService) Query() (svc.Status, error) {
	return svc.Status{State: s.status}, s.queryErr
}

func (s *fakeService) Start() error { return s.startErr }

func (s *fakeService) Control(cmd svc.Cmd) (svc.Status, error) {
	if cmd == svc.Stop && s.controlErr == nil {
		s.status = s.controlState
	}
	return svc.Status{State: s.controlState}, s.controlErr
}

func (s *fakeService) Delete() error {
	s.deleted = true
	return s.deleteErr
}

func (s *fakeService) SetRecoveryActions(actions []mgr.RecoveryAction, resetPeriod uint32) error {
	s.recovery = actions
	s.resetPeriod = resetPeriod
	return s.recoveryErr
}

// fakeManager 可编程的服务管理器测试桩。
type fakeManager struct {
	openErr          error
	createErr        error
	svc              *fakeService
	openCalls        int
	openFailAt       int
	createdStartType uint32
}

func (m *fakeManager) Disconnect() error { return nil }

func (m *fakeManager) OpenService(string) (serviceHandle, error) {
	m.openCalls++
	if m.openFailAt > 0 && m.openCalls == m.openFailAt {
		return nil, errors.New("打开失败")
	}
	if m.openErr != nil {
		return nil, m.openErr
	}
	return m.svc, nil
}

func (m *fakeManager) CreateService(_ string, _ string, c mgr.Config, _ ...string) (serviceHandle, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdStartType = c.StartType
	return m.svc, nil
}

// applyControl 应用控制依赖，返回立即恢复函数。
func applyControl(m manager, installErr, removeErr error, exeErr error, timeout time.Duration) func() {
	origConnect, origInstall, origRemove, origExe, origTimeout :=
		connectManager, installEventLog, removeEventLog, executablePath, stopTimeout
	connectManager = func() (manager, error) { return m, nil }
	installEventLog = func(string, uint32) error { return installErr }
	removeEventLog = func(string) error { return removeErr }
	executablePath = func() (string, error) {
		if exeErr != nil {
			return "", exeErr
		}
		return `C:\test\app.exe`, nil
	}
	stopTimeout = timeout
	return func() {
		connectManager, installEventLog, removeEventLog, executablePath, stopTimeout =
			origConnect, origInstall, origRemove, origExe, origTimeout
	}
}

// applyConnectError 应用连接失败桩，返回恢复函数。
func applyConnectError() func() {
	orig := connectManager
	connectManager = func() (manager, error) { return nil, errors.New("连接失败") }
	return func() { connectManager = orig }
}

func TestGetServiceStatus(t *testing.T) {
	svcStub := &fakeService{status: svc.Running}
	restore := applyControl(&fakeManager{svc: svcStub}, nil, nil, nil, time.Second)
	st, err := GetServiceStatus("svc")
	restore()
	if err != nil || st != svc.Running {
		t.Fatalf("查询失败：%v", err)
	}
	if !svcStub.closed {
		t.Fatal("服务句柄未关闭")
	}

	restore = applyControl(&fakeManager{openErr: errors.New("不存在")}, nil, nil, nil, time.Second)
	_, err = GetServiceStatus("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("应报服务不存在，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{queryErr: errors.New("查询失败")}}, nil, nil, nil, time.Second)
	_, err = GetServiceStatus("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("应报查询失败，实际：%v", err)
	}
}

func TestIsServiceExist(t *testing.T) {
	restore := applyControl(&fakeManager{svc: &fakeService{}}, nil, nil, nil, time.Second)
	got, err := IsServiceExist("svc")
	restore()
	if err != nil || !got {
		t.Fatal("存在服务应返回 true")
	}

	restore = applyControl(&fakeManager{openErr: errors.New("不存在")}, nil, nil, nil, time.Second)
	got, err = IsServiceExist("svc")
	restore()
	if err != nil || got {
		t.Fatal("不存在服务应返回 false")
	}

	restore = applyConnectError()
	_, err = IsServiceExist("svc")
	restore()
	if err == nil {
		t.Fatal("连接失败应返回错误")
	}
}

func TestControlConnectError(t *testing.T) {
	restore := applyConnectError()
	_, err := GetServiceStatus("svc")
	if !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("应报连接失败，实际：%v", err)
	}
	if err := Install("svc", "", ""); !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("应报连接失败，实际：%v", err)
	}
	if err := Uninstall("svc"); !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("应报连接失败，实际：%v", err)
	}
	if err := Start("svc"); !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("应报连接失败，实际：%v", err)
	}
	if err := Stop("svc"); !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("应报连接失败，实际：%v", err)
	}
	if err := Restart("svc"); !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("应报连接失败，实际：%v", err)
	}
	restore()
}

// applySequenceControl 让 connectManager 按序返回 m，第 failAt 次返回连接错误。
func applySequenceControl(m manager, failAt int) func() {
	orig := connectManager
	calls := 0
	connectManager = func() (manager, error) {
		calls++
		if calls == failAt {
			return nil, errors.New("连接失败")
		}
		return m, nil
	}
	return func() { connectManager = orig }
}

// TestSecondConnectError 覆盖 IsServiceExist 之后的连接失败分支。
func TestSecondConnectError(t *testing.T) {
	// Install：存在检查成功（不存在）后连接失败。
	restore := applySequenceControl(&fakeManager{openErr: errors.New("不存在"), svc: &fakeService{}}, 2)
	err := Install("svc", "", "")
	restore()
	if !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("Install 应报连接失败，实际：%v", err)
	}

	// Uninstall：存在检查成功后连接失败。
	restore = applySequenceControl(&fakeManager{svc: &fakeService{}}, 2)
	err = Uninstall("svc")
	restore()
	if !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("Uninstall 应报连接失败，实际：%v", err)
	}

	// Start：存在与状态查询成功后连接失败。
	restore = applySequenceControl(&fakeManager{svc: &fakeService{status: svc.Stopped}}, 3)
	err = Start("svc")
	restore()
	if !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("Start 应报连接失败，实际：%v", err)
	}

	// Stop：存在与状态查询成功后连接失败。
	restore = applySequenceControl(&fakeManager{svc: &fakeService{status: svc.Running}}, 3)
	err = Stop("svc")
	restore()
	if !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("Stop 应报连接失败，实际：%v", err)
	}

	// Restart：存在检查后状态查询连接失败。
	restore = applySequenceControl(&fakeManager{svc: &fakeService{}}, 2)
	err = Restart("svc")
	restore()
	if !errx.Is(err, wxerr.CodeManagerConnect) {
		t.Fatalf("Restart 应报连接失败，实际：%v", err)
	}
}

func TestInstall(t *testing.T) {
	// 服务不存在 → 安装成功。
	ok := &fakeService{}
	restore := applyControl(&fakeManager{openErr: errors.New("不存在"), svc: ok}, nil, nil, nil, time.Second)
	err := Install("svc", "显示名", "描述")
	restore()
	if err != nil {
		t.Fatalf("安装失败：%v", err)
	}
	if !ok.closed {
		t.Fatal("安装后服务句柄未关闭")
	}

	// 服务已存在。
	restore = applyControl(&fakeManager{svc: &fakeService{}}, nil, nil, nil, time.Second)
	err = Install("svc", "", "")
	restore()
	if !errx.Is(err, wxerr.CodeServiceAlreadyExists) {
		t.Fatalf("已存在应报错，实际：%v", err)
	}

	// 可执行文件路径失败。
	restore = applyControl(&fakeManager{openErr: errors.New("不存在"), svc: &fakeService{}},
		nil, nil, errors.New("路径失败"), time.Second)
	err = Install("svc", "", "")
	restore()
	if !errx.Is(err, wxerr.CodeExecutablePath) {
		t.Fatalf("路径失败应报错，实际：%v", err)
	}

	// 创建服务失败。
	restore = applyControl(&fakeManager{openErr: errors.New("不存在"),
		createErr: errors.New("创建失败"), svc: &fakeService{}}, nil, nil, nil, time.Second)
	err = Install("svc", "", "")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("创建失败应报错，实际：%v", err)
	}

	// 恢复配置失败。
	restore = applyControl(&fakeManager{openErr: errors.New("不存在"),
		svc: &fakeService{recoveryErr: errors.New("恢复失败")}}, nil, nil, nil, time.Second)
	err = Install("svc", "", "")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("恢复配置失败应报错，实际：%v", err)
	}

	// 事件日志失败应回滚删除服务。
	rollback := &fakeService{}
	restore = applyControl(&fakeManager{openErr: errors.New("不存在"), svc: rollback},
		errors.New("事件日志失败"), nil, nil, time.Second)
	err = Install("svc", "", "")
	restore()
	if !errx.Is(err, wxerr.CodeEventLogFailed) {
		t.Fatalf("事件日志失败应报错，实际：%v", err)
	}
	if !rollback.deleted {
		t.Fatal("事件日志失败应回滚删除服务")
	}
}

func TestUninstall(t *testing.T) {
	restore := applyControl(&fakeManager{svc: &fakeService{}}, nil, nil, nil, time.Second)
	err := Uninstall("svc")
	restore()
	if err != nil {
		t.Fatalf("卸载失败：%v", err)
	}

	restore = applyControl(&fakeManager{openErr: errors.New("不存在")}, nil, nil, nil, time.Second)
	err = Uninstall("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("不存在应报错，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{deleteErr: errors.New("删除失败")}}, nil, nil, nil, time.Second)
	err = Uninstall("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("删除失败应报错，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{}}, nil, errors.New("日志移除失败"), nil, time.Second)
	err = Uninstall("svc")
	restore()
	if !errx.Is(err, wxerr.CodeEventLogFailed) {
		t.Fatalf("日志移除失败应报错，实际：%v", err)
	}

	// 存在检查通过后打开失败。
	restore = applyControl(&fakeManager{svc: &fakeService{}, openFailAt: 2}, nil, nil, nil, time.Second)
	err = Uninstall("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("打开失败应报服务不存在，实际：%v", err)
	}
}

func TestStart(t *testing.T) {
	ok := &fakeService{}
	restore := applyControl(&fakeManager{svc: ok}, nil, nil, nil, time.Second)
	err := Start("svc")
	restore()
	if err != nil {
		t.Fatalf("启动失败：%v", err)
	}

	restore = applyControl(&fakeManager{openErr: errors.New("不存在")}, nil, nil, nil, time.Second)
	err = Start("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("不存在应报错，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Running}}, nil, nil, nil, time.Second)
	err = Start("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceAlreadyRunning) {
		t.Fatalf("运行中应报错，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{queryErr: errors.New("查询失败")}}, nil, nil, nil, time.Second)
	err = Start("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("状态查询失败应报错，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{startErr: errors.New("启动失败")}}, nil, nil, nil, time.Second)
	err = Start("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("启动失败应报错，实际：%v", err)
	}

	// 存在与状态查询通过后打开失败。
	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Stopped}, openFailAt: 3}, nil, nil, nil, time.Second)
	err = Start("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("打开失败应报服务不存在，实际：%v", err)
	}
}

func TestStop(t *testing.T) {
	ok := &fakeService{status: svc.Running, controlState: svc.Stopped}
	restore := applyControl(&fakeManager{svc: ok}, nil, nil, nil, time.Second)
	err := Stop("svc")
	restore()
	if err != nil {
		t.Fatalf("停止失败：%v", err)
	}

	restore = applyControl(&fakeManager{openErr: errors.New("不存在")}, nil, nil, nil, time.Second)
	err = Stop("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("不存在应报错，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Stopped}}, nil, nil, nil, time.Second)
	err = Stop("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceAlreadyStopped) {
		t.Fatalf("已停止应报错，实际：%v", err)
	}

	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Running, controlErr: errors.New("控制失败")}},
		nil, nil, nil, time.Second)
	err = Stop("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("控制失败应报错，实际：%v", err)
	}

	// 轮询查询失败。
	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Running, controlState: svc.Running,
		queryErr: errors.New("查询失败")}}, nil, nil, nil, time.Second)
	err = Stop("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("轮询查询失败应报错，实际：%v", err)
	}

	// 等待超时。
	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Running, controlState: svc.Running}},
		nil, nil, nil, 10*time.Millisecond)
	err = Stop("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceStopTimeout) {
		t.Fatalf("超时应报错，实际：%v", err)
	}

	// 存在与状态查询通过后打开失败。
	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Running}, openFailAt: 3}, nil, nil, nil, time.Second)
	err = Stop("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("打开失败应报服务不存在，实际：%v", err)
	}
}

func TestRestart(t *testing.T) {
	// 运行中：先停后启。
	restore := applyControl(&fakeManager{svc: &fakeService{status: svc.Running, controlState: svc.Stopped}},
		nil, nil, nil, time.Second)
	err := Restart("svc")
	restore()
	if err != nil {
		t.Fatalf("重启失败：%v", err)
	}

	// 已停止：直接启动。
	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Stopped}}, nil, nil, nil, time.Second)
	err = Restart("svc")
	restore()
	if err != nil {
		t.Fatalf("重启失败：%v", err)
	}

	restore = applyControl(&fakeManager{openErr: errors.New("不存在")}, nil, nil, nil, time.Second)
	err = Restart("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceNotFound) {
		t.Fatalf("不存在应报错，实际：%v", err)
	}

	// 运行中但停止失败。
	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Running, controlErr: errors.New("控制失败")}},
		nil, nil, nil, time.Second)
	err = Restart("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("停止失败应报错，实际：%v", err)
	}

	// 停止成功但启动失败。
	restore = applyControl(&fakeManager{svc: &fakeService{status: svc.Running, controlState: svc.Stopped,
		startErr: errors.New("启动失败")}}, nil, nil, nil, time.Second)
	err = Restart("svc")
	restore()
	if !errx.Is(err, wxerr.CodeServiceControlFailed) {
		t.Fatalf("启动失败应报错，实际：%v", err)
	}
}

// TestRealManagerAdapter 覆盖真实服务管理器适配层（无权限时跳过）。
func TestRealManagerAdapter(t *testing.T) {
	m, err := connectManager()
	if err != nil {
		t.Skipf("无服务管理器访问权限：%v", err)
	}
	if m == nil {
		t.Fatal("服务管理器句柄为空")
	}
	if _, err := m.OpenService("__winsvcx_not_exist__"); err == nil {
		t.Fatal("不存在的服务不应打开成功")
	}
	if err := m.Disconnect(); err != nil {
		t.Fatalf("断开服务管理器失败：%v", err)
	}
}

// TestSetStopTimeout 覆盖停止超时配置校验。
func TestSetStopTimeout(t *testing.T) {
	orig := stopTimeout
	defer func() { stopTimeout = orig }()
	if err := SetStopTimeout(0); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("非法超时应报错，实际：%v", err)
	}
	if err := SetStopTimeout(-time.Second); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("非法超时应报错，实际：%v", err)
	}
	if err := SetStopTimeout(5 * time.Second); err != nil || stopTimeout != 5*time.Second {
		t.Fatalf("合法超时应生效：%v", err)
	}
}

// TestValidateInstallOptions 覆盖安装选项默认值与校验。
func TestValidateInstallOptions(t *testing.T) {
	def, err := validateInstallOptions(InstallOptions{})
	if err != nil {
		t.Fatalf("空选项应补齐默认值：%v", err)
	}
	if def.StartType != mgr.StartAutomatic || len(def.RecoveryActions) != 3 ||
		def.RecoveryResetPeriod != 60 || def.EventLogTypes == 0 {
		t.Fatalf("默认值不符：%+v", def)
	}

	if _, err := validateInstallOptions(InstallOptions{StartType: mgr.StartDisabled + 1}); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("非法启动类型应报错，实际：%v", err)
	}
	if _, err := validateInstallOptions(InstallOptions{
		RecoveryActions: []mgr.RecoveryAction{{Type: 0, Delay: time.Second}},
	}); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("非法恢复动作应报错，实际：%v", err)
	}
	if _, err := validateInstallOptions(InstallOptions{
		RecoveryActions: []mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: 0}},
	}); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("零延迟恢复动作应报错，实际：%v", err)
	}

	custom := InstallOptions{
		StartType:           mgr.StartManual,
		RecoveryActions:     []mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: time.Second}},
		RecoveryResetPeriod: 30,
		EventLogTypes:       windows.EVENTLOG_ERROR_TYPE,
	}
	got, err := validateInstallOptions(custom)
	if err != nil || got.StartType != mgr.StartManual || len(got.RecoveryActions) != 1 ||
		got.RecoveryResetPeriod != 30 || got.EventLogTypes != windows.EVENTLOG_ERROR_TYPE {
		t.Fatalf("自定义选项未被保留：%+v err=%v", got, err)
	}
}

// TestInstallWithOptionsCustom 覆盖自定义选项透传。
func TestInstallWithOptionsCustom(t *testing.T) {
	svcStub := &fakeService{}
	m := &fakeManager{openErr: errors.New("不存在"), svc: svcStub}
	restore := applyControl(m, nil, nil, nil, time.Second)
	opts := InstallOptions{
		StartType:           mgr.StartManual,
		RecoveryActions:     []mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: 7 * time.Second}},
		RecoveryResetPeriod: 45,
		EventLogTypes:       windows.EVENTLOG_ERROR_TYPE,
	}
	err := InstallWithOptions("svc", "", "", opts)
	restore()
	if err != nil {
		t.Fatalf("安装失败：%v", err)
	}
	if m.createdStartType != mgr.StartManual {
		t.Fatalf("启动类型未透传：%v", m.createdStartType)
	}
	if len(svcStub.recovery) != 1 || svcStub.recovery[0].Delay != 7*time.Second {
		t.Fatalf("恢复动作未透传：%+v", svcStub.recovery)
	}
	if svcStub.resetPeriod != 45 {
		t.Fatalf("恢复重置周期未透传：%d", svcStub.resetPeriod)
	}
}

// TestAccessDeniedClassification 覆盖访问被拒绝错误细化。
func TestAccessDeniedClassification(t *testing.T) {
	orig := connectManager
	connectManager = func() (manager, error) { return nil, windows.ERROR_ACCESS_DENIED }
	_, err := GetServiceStatus("svc")
	connectManager = orig
	if !errx.Is(err, wxerr.CodeAccessDenied) {
		t.Fatalf("访问被拒绝应细化错误码，实际：%v", err)
	}

	restore := applyControl(&fakeManager{openErr: errors.New("不存在"),
		createErr: windows.ERROR_ACCESS_DENIED, svc: &fakeService{}}, nil, nil, nil, time.Second)
	err = Install("svc", "", "")
	restore()
	if !errx.Is(err, wxerr.CodeAccessDenied) {
		t.Fatalf("创建服务被拒绝应细化错误码，实际：%v", err)
	}
}
