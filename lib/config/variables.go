// Package config 提供全局运行配置。
package config

import "github.com/lcylpzls/logx"

// Log 全局日志器（由入口初始化）。
var Log logx.Logger

// 编译期断言：Log 与 logx.Logger 类型一致。
var _ logx.Logger = Log
