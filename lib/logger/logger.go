package logger

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *logrus.Logger

// InitLogger 初始化全局日志对象
func InitLogger(logPath string, maxSize, maxBackups, maxAge int, compress bool, level logrus.Level) {
	log = logrus.New()
	log.SetOutput(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize,    // 单个日志文件最大(MB)
		MaxBackups: maxBackups, // 保留旧文件最大数量
		MaxAge:     maxAge,     // 保留旧文件最大天数
		Compress:   compress,   // 是否压缩
	})
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(level)
	log.SetReportCaller(true)
}

// GetLogger 获取全局日志对象
func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger("app.log", 10, 7, 30, true, logrus.InfoLevel)
	}
	return log
}