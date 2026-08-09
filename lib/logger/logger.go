// Package logger 提供日志初始化与全局日志器（底层为 logx 家族库）。
package logger

import (
	"os"

	"github.com/lcylpzls/logx"
)

// Options 日志配置。
type Options struct {
	// LogDir 日志目录；为空时仅输出到控制台。
	LogDir string
	// Filename 日志文件名。
	Filename string
	// MaxSize 单文件最大容量（MB），0 使用默认值。
	MaxSize int
	// MaxBackups 保留的历史文件数量，0 使用默认值。
	MaxBackups int
	// MaxAge 保留的历史文件天数，0 使用默认值。
	MaxAge int
	// CompressAfter 超过 N 天的历史日志自动压缩为 gz，0 不压缩。
	CompressAfter int
	// Level 启用的最低日志级别，0 使用 InfoLevel。
	Level logx.Level
	// Console 是否同时输出到控制台。
	Console bool
}

var global logx.Logger

// Init 初始化全局日志器并返回实例。
func Init(opts Options) logx.Logger {
	b := logx.NewBuilder()
	level := opts.Level
	if level == 0 {
		level = logx.InfoLevel
	}
	if opts.LogDir != "" && opts.Filename != "" {
		b.EnableFileLog(
			logx.WithLogDir(opts.LogDir),
			logx.WithFilename(opts.Filename),
			logx.WithMaxSize(defaultInt(opts.MaxSize, 20)),
			logx.WithMaxBackups(defaultInt(opts.MaxBackups, 10)),
			logx.WithMaxAge(defaultInt(opts.MaxAge, 60)),
			logx.WithCompressAfter(opts.CompressAfter),
			logx.WithLevels(level),
		)
	}
	if opts.Console {
		b.EnableConsole(level)
	}
	l, _ := b.Build()
	global = l
	return l
}

// Get 返回全局日志器；未初始化时降级为 stderr 输出。
func Get() logx.Logger {
	if global == nil {
		// Build 在 logx 中始终返回非 nil 实例。
		global, _ = logx.NewBuilder().EnableWriter(os.Stderr, logx.InfoLevel).Build()
	}
	return global
}

// Reset 重置全局日志器（测试用）。
func Reset() {
	global = nil
}

// defaultInt 返回非零值，否则使用默认值。
func defaultInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
