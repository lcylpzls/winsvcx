# winsvcx API 定版

> 版本：v0.3.0 · 以下签名与代码一致。

## 1. lib/logger

```go
type Options struct {
	LogDir        string    // 日志目录（空则仅控制台）
	Filename      string    // 日志文件名
	MaxSize       int       // 单文件最大 MB（0=20）
	MaxBackups    int       // 保留份数（0=10）
	MaxAge        int       // 保留天数（0=60）
	CompressAfter int       // 超过 N 天压缩（0=不压缩）
	Level         logx.Level // 最低级别（0=InfoLevel）
	Console       bool      // 同时输出控制台
}

func Init(opts Options) logx.Logger // 初始化全局日志器
func Get() logx.Logger              // 获取全局日志器（未初始化时 stderr 降级）
func Reset()                        // 重置全局日志器（测试用）
```

## 2. lib/service

```go
type Service struct { StopFlag bool } // 实现 svc.Handler

func Run(name string)                          // 以服务模式运行
func GetServiceStatus(name string) (svc.State, error)
func IsServiceExist(name string) (bool, error)
func Install(name, displayName, description string) error
func InstallWithOptions(name, displayName, description string, opts InstallOptions) error
func DefaultInstallOptions() InstallOptions
func SetStopTimeout(d time.Duration) error
func Uninstall(name string) error
func Start(name string) error
func Stop(name string) error
func Restart(name string) error
```

`InstallOptions`：`StartType`（0=自动）、`RecoveryActions`（nil=默认
三次重启）、`RecoveryResetPeriod`（0=60）、`EventLogTypes`
（0=错误+警告+信息）。

## 3. lib/app

```go
// Run 启动主循环；服务模式由框架关闭 stopCh，应用模式等待系统信号。
func Run(stopCh chan struct{}, wg *sync.WaitGroup, logger logx.Logger)
```

## 4. lib/win32

```go
// MessageBox 显示置顶消息框，返回按钮 ID。
func MessageBox(caption, text string, style uint32) int
```

常量：`MB_OK` 等按钮样式、`MB_ICONERROR` 等图标、
`IDOK` 等返回值。

## 5. lib/errors

错误码常量（`winsvcx_` 前缀）已通过 `errx.RegisterCode` /
`errx.RegisterCodeKind` 注册，使用 `errx.NewCode` / `errx.WrapCode`
构造，`errx.Is(err, code)` 匹配。

## 6. 入口命令

```text
MySampleService.exe install|uninstall|start|stop|restart
```

无参数：Windows 服务模式（由服务管理器启动）或应用模式（等待系统信号）。

## 7. 安静模式

任意参数位置支持 `-quiet` / `--quiet` / `/quiet` / `-q`：
关闭消息框与控制台输出，仅保留文件日志；退出码
`0` 成功、`1` 操作失败、`2` 无效命令。

```text
MySampleService.exe -quiet install
MySampleService.exe install -quiet
```
