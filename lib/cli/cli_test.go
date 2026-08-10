package cli

import (
	"testing"

	"github.com/lcylpzls/testx"
	wxerr "github.com/lcylpzls/winsvcx/lib/errors"
)

// TestParse 覆盖命令与安静标志解析。
func TestParse(t *testing.T) {
	opts, err := Parse(nil)
	testx.RequireNoError(t, err)
	testx.RequireEqual(t, opts.Command, "")
	testx.RequireFalse(t, opts.Quiet)

	opts, err = Parse([]string{"app.exe"})
	testx.RequireNoError(t, err)
	testx.RequireEqual(t, opts.Command, "")
	testx.RequireFalse(t, opts.Quiet)

	opts, err = Parse([]string{"app.exe", "install"})
	testx.RequireNoError(t, err)
	testx.RequireEqual(t, opts.Command, CommandInstall)
	testx.RequireFalse(t, opts.Quiet)

	// 大小写不敏感。
	opts, err = Parse([]string{"app.exe", "INSTALL"})
	testx.RequireNoError(t, err)
	testx.RequireEqual(t, opts.Command, CommandInstall)

	// 安静标志任意位置。
	for _, flag := range []string{"-quiet", "--quiet", "/quiet", "-q", "-QUIET"} {
		opts, err = Parse([]string{"app.exe", flag, "start"})
		testx.RequireNoError(t, err)
		testx.RequireTrue(t, opts.Quiet)
		testx.RequireEqual(t, opts.Command, CommandStart)
		opts, err = Parse([]string{"app.exe", "stop", flag})
		testx.RequireNoError(t, err)
		testx.RequireTrue(t, opts.Quiet)
		testx.RequireEqual(t, opts.Command, CommandStop)
	}

	// 多参数取首个命令。
	opts, err = Parse([]string{"app.exe", "start", "extra"})
	testx.RequireNoError(t, err)
	testx.RequireEqual(t, opts.Command, CommandStart)

	// 版本参数。
	for _, flag := range []string{"-V", "--version", "/V", "-v"} {
		opts, err = Parse([]string{"app.exe", flag})
		testx.RequireNoError(t, err)
		testx.RequireTrue(t, opts.ShowVersion)
		testx.RequireEqual(t, opts.Command, "")
	}
}

// TestParseErrors 覆盖未知命令与未知开关。
func TestParseErrors(t *testing.T) {
	_, err := Parse([]string{"app.exe", "unknown"})
	testx.RequireErrCode(t, err, wxerr.CodeInvalidConfig)
	_, err = Parse([]string{"app.exe", "-verbose"})
	testx.RequireErrCode(t, err, wxerr.CodeInvalidConfig)
	_, err = Parse([]string{"app.exe", "/verbose"})
	testx.RequireErrCode(t, err, wxerr.CodeInvalidConfig)
}

// TestIsSupportedCommand 覆盖命令白名单。
func TestIsSupportedCommand(t *testing.T) {
	for _, c := range SupportedCommands {
		testx.RequireTrue(t, IsSupportedCommand(c))
	}
	testx.RequireFalse(t, IsSupportedCommand("run"))
}
