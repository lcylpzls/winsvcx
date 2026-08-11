// Package cli 提供命令行参数解析与控制命令检测
// （安静模式、安装/卸载/启动/停止/重启）。
package cli

import (
	"strings"

	"github.com/lcylpzls/errx"
	wxerr "github.com/lcylpzls/winsvcx/lib/errors"
)

// 控制命令常量。
const (
	CommandInstall   = "install"
	CommandUninstall = "uninstall"
	CommandStart     = "start"
	CommandStop      = "stop"
	CommandRestart   = "restart"
)

// SupportedCommands 支持的控制命令列表（用于帮助信息）。
var SupportedCommands = []string{
	CommandInstall,
	CommandUninstall,
	CommandStart,
	CommandStop,
	CommandRestart,
}

// Version 构建版本，随发版同步更新，可用 ldflags 覆盖：
// -ldflags "-X github.com/lcylpzls/winsvcx/lib/cli.Version=v1.0.0"
var Version = "1.1.1"

// Options 命令行解析结果。
type Options struct {
	// Command 控制命令；空串表示运行模式（服务/应用）。
	Command string
	// Quiet 安静模式：关闭消息框与控制台输出，仅保留文件日志与退出码。
	Quiet bool
	// ShowVersion 是否输出版本号（-V / --version）。
	ShowVersion bool
}

// Parse 解析命令行参数（args 应包含程序名 args[0]）。
// 支持 -quiet / --quiet / /quiet / -q 与 -V / --version / /V（大小写不敏感，位置任意）；
// 首个非开关参数作为控制命令；未知开关或命令返回配置错误。
func Parse(args []string) (Options, error) {
	var opts Options
	if len(args) <= 1 {
		return opts, nil
	}
	for _, a := range args[1:] {
		if a == "" {
			continue
		}
		switch strings.ToLower(a) {
		case "-quiet", "--quiet", "/quiet", "-q":
			opts.Quiet = true
		case "-v", "-V", "--version", "/v", "/V", "/version":
			opts.ShowVersion = true
		default:
			if strings.HasPrefix(a, "-") || strings.HasPrefix(a, "/") {
				return opts, errx.NewCode(wxerr.CodeInvalidConfig, "不支持的参数："+a)
			}
			if opts.Command == "" {
				opts.Command = strings.ToLower(a)
			}
		}
	}
	if opts.Command != "" && !IsSupportedCommand(opts.Command) {
		return opts, errx.NewCode(wxerr.CodeInvalidConfig, "不支持的命令："+opts.Command)
	}
	return opts, nil
}

// IsSupportedCommand 判断是否为受支持的控制命令。
func IsSupportedCommand(cmd string) bool {
	for _, c := range SupportedCommands {
		if c == cmd {
			return true
		}
	}
	return false
}
