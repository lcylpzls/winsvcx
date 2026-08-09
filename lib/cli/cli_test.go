package cli

import (
	"testing"

	"github.com/lcylpzls/errx"
	wxerr "github.com/lcylpzls/winsvcx/lib/errors"
)

// TestParse 覆盖命令与安静标志解析。
func TestParse(t *testing.T) {
	opts, err := Parse(nil)
	if err != nil || opts.Command != "" || opts.Quiet {
		t.Fatalf("空参数应返回空选项：%+v err=%v", opts, err)
	}

	opts, err = Parse([]string{"app.exe"})
	if err != nil || opts.Command != "" || opts.Quiet {
		t.Fatalf("仅程序名应返回空选项：%+v err=%v", opts, err)
	}

	opts, err = Parse([]string{"app.exe", "install"})
	if err != nil || opts.Command != CommandInstall || opts.Quiet {
		t.Fatalf("命令解析不符：%+v err=%v", opts, err)
	}

	// 大小写不敏感。
	opts, err = Parse([]string{"app.exe", "INSTALL"})
	if err != nil || opts.Command != CommandInstall {
		t.Fatalf("命令应归一为小写：%+v err=%v", opts, err)
	}

	// 安静标志任意位置。
	for _, flag := range []string{"-quiet", "--quiet", "/quiet", "-q", "-QUIET"} {
		opts, err = Parse([]string{"app.exe", flag, "start"})
		if err != nil || !opts.Quiet || opts.Command != CommandStart {
			t.Fatalf("%s 解析不符：%+v err=%v", flag, opts, err)
		}
		opts, err = Parse([]string{"app.exe", "stop", flag})
		if err != nil || !opts.Quiet || opts.Command != CommandStop {
			t.Fatalf("%s 尾部解析不符：%+v err=%v", flag, opts, err)
		}
	}

	// 多参数取首个命令。
	opts, err = Parse([]string{"app.exe", "start", "extra"})
	if err != nil || opts.Command != CommandStart {
		t.Fatalf("首个命令应生效：%+v err=%v", opts, err)
	}

	// 版本参数。
	for _, flag := range []string{"-V", "--version", "/V", "-v"} {
		opts, err = Parse([]string{"app.exe", flag})
		if err != nil || !opts.ShowVersion || opts.Command != "" {
			t.Fatalf("%s 应触发版本输出：%+v err=%v", flag, opts, err)
		}
	}
}

// TestParseErrors 覆盖未知命令与未知开关。
func TestParseErrors(t *testing.T) {
	if _, err := Parse([]string{"app.exe", "unknown"}); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("未知命令应报配置错误，实际：%v", err)
	}
	if _, err := Parse([]string{"app.exe", "-verbose"}); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("未知开关应报配置错误，实际：%v", err)
	}
	if _, err := Parse([]string{"app.exe", "/verbose"}); !errx.Is(err, wxerr.CodeInvalidConfig) {
		t.Fatalf("未知斜杠开关应报配置错误，实际：%v", err)
	}
}

// TestIsSupportedCommand 覆盖命令白名单。
func TestIsSupportedCommand(t *testing.T) {
	for _, c := range SupportedCommands {
		if !IsSupportedCommand(c) {
			t.Fatalf("命令 %s 应在白名单内", c)
		}
	}
	if IsSupportedCommand("run") {
		t.Fatal("非控制命令不应在白名单内")
	}
}
