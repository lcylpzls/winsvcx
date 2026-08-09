package errors

import (
	"testing"

	"github.com/lcylpzls/errx"
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
	}
	for _, code := range codes {
		e := errx.NewCode(code, "测试错误")
		if !errx.Is(e, code) {
			t.Fatalf("错误码 %s 无法匹配", code)
		}
		if e.Kind() == errx.KindUnknown {
			t.Fatalf("错误码 %s 未注册分类", code)
		}
	}
}
