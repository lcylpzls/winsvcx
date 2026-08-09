package service

import (
	"testing"
	"time"

	"golang.org/x/sys/windows/svc/mgr"
)

// FuzzValidateInstallOptions 保证安装选项校验对任意输入不 panic。
func FuzzValidateInstallOptions(f *testing.F) {
	f.Add(int32(0), int32(5000), uint32(0), uint32(0))
	f.Add(int32(99), int32(-1), uint32(0), uint32(0))
	f.Fuzz(func(t *testing.T, startType, delayMs int32, reset, eventTypes uint32) {
		opts := InstallOptions{
			StartType: uint32(startType),
			RecoveryActions: []mgr.RecoveryAction{
				{Type: int(startType), Delay: time.Duration(delayMs) * time.Millisecond},
			},
			RecoveryResetPeriod: reset,
			EventLogTypes:       eventTypes,
		}
		_, _ = validateInstallOptions(opts)
	})
}
