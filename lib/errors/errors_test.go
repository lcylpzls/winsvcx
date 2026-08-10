package errors

import (
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

// TestCodesRegistered 覆盖全部错误码注册与匹配。
func TestCodesRegistered(t *testing.T) {
	codes := []errx.Code{
		CodeInvalidConfig,
		CodeManagerConnect,
		CodeServiceNotFound,
		CodeServiceAlreadyExists,
		CodeServiceControlFailed,
		CodeServiceAlreadyRunning,
		CodeServiceAlreadyStopped,
		CodeServiceStopTimeout,
		CodeEventLogFailed,
		CodeExecutablePath,
		CodeServiceRunFailed,
		CodeAccessDenied,
	}
	for _, code := range codes {
		e := errx.NewCode(code, "测试错误")
		testx.RequireErrCode(t, e, code)
		testx.RequireNotEqual(t, e.Kind(), errx.KindUnknown)
	}
}
