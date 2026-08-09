// Package errors 定义 winsvcx 统一错误码（errx 家族规范）。
package errors

import "github.com/lcylpzls/errx"

// 错误码统一以 winsvcx_ 为前缀。
const (
	CodeInvalidConfig         errx.Code = "winsvcx_invalid_config"
	CodeManagerConnect        errx.Code = "winsvcx_manager_connect"
	CodeServiceNotFound       errx.Code = "winsvcx_service_not_found"
	CodeServiceAlreadyExists  errx.Code = "winsvcx_service_already_exists"
	CodeServiceControlFailed  errx.Code = "winsvcx_service_control_failed"
	CodeServiceAlreadyRunning errx.Code = "winsvcx_service_already_running"
	CodeServiceAlreadyStopped errx.Code = "winsvcx_service_already_stopped"
	CodeServiceStopTimeout    errx.Code = "winsvcx_service_stop_timeout"
	CodeEventLogFailed        errx.Code = "winsvcx_event_log_failed"
	CodeExecutablePath        errx.Code = "winsvcx_executable_path"
	CodeServiceRunFailed      errx.Code = "winsvcx_service_run_failed"
	CodeAccessDenied          errx.Code = "winsvcx_access_denied"
)

func init() {
	errx.RegisterCode(CodeInvalidConfig, "配置非法")
	errx.RegisterCodeKind(CodeInvalidConfig, errx.KindInvalid)
	errx.RegisterCode(CodeManagerConnect, "连接服务管理器失败")
	errx.RegisterCodeKind(CodeManagerConnect, errx.KindUnavailable)
	errx.RegisterCode(CodeServiceNotFound, "服务不存在")
	errx.RegisterCodeKind(CodeServiceNotFound, errx.KindNotFound)
	errx.RegisterCode(CodeServiceAlreadyExists, "服务已存在")
	errx.RegisterCodeKind(CodeServiceAlreadyExists, errx.KindAlreadyExists)
	errx.RegisterCode(CodeServiceControlFailed, "服务控制操作失败")
	errx.RegisterCodeKind(CodeServiceControlFailed, errx.KindUnavailable)
	errx.RegisterCode(CodeServiceAlreadyRunning, "服务已在运行")
	errx.RegisterCodeKind(CodeServiceAlreadyRunning, errx.KindConflict)
	errx.RegisterCode(CodeServiceAlreadyStopped, "服务已停止")
	errx.RegisterCodeKind(CodeServiceAlreadyStopped, errx.KindConflict)
	errx.RegisterCode(CodeServiceStopTimeout, "等待服务停止超时")
	errx.RegisterCodeKind(CodeServiceStopTimeout, errx.KindTimeout)
	errx.RegisterCode(CodeEventLogFailed, "事件日志操作失败")
	errx.RegisterCodeKind(CodeEventLogFailed, errx.KindUnavailable)
	errx.RegisterCode(CodeExecutablePath, "无法获取可执行文件路径")
	errx.RegisterCodeKind(CodeExecutablePath, errx.KindUnavailable)
	errx.RegisterCode(CodeServiceRunFailed, "服务运行失败")
	errx.RegisterCodeKind(CodeServiceRunFailed, errx.KindUnavailable)
	errx.RegisterCode(CodeAccessDenied, "访问被拒绝（需要管理员权限）")
	errx.RegisterCodeKind(CodeAccessDenied, errx.KindForbidden)
}
