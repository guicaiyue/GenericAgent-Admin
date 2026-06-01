package api

import (
	"reflect"
	"testing"
)

func TestTMWebDriverPythonDependencyPackageNames(t *testing.T) {
	wantModules := []string{"requests", "bottle", "simple_websocket_server"}
	if got := tmwebdriverPythonModules(); !reflect.DeepEqual(got, wantModules) {
		t.Fatalf("modules=%v want=%v", got, wantModules)
	}

	wantPkgs := []string{"requests", "bottle", "simple-websocket-server"}
	if got := tmwebdriverRequiredPipPackages(); !reflect.DeepEqual(got, wantPkgs) {
		t.Fatalf("packages=%v want=%v", got, wantPkgs)
	}
	if got := tmwebdriverPipPackages([]string{"simple_websocket_server"}); !reflect.DeepEqual(got, []string{"simple-websocket-server"}) {
		t.Fatalf("mapped package=%v", got)
	}
}

func TestBuildTMWebDriverInstallCommand(t *testing.T) {
	if got, want := buildTMWebDriverInstallCommand(""), "python -m pip install -i https://pypi.tuna.tsinghua.edu.cn/simple requests bottle simple-websocket-server"; got != want {
		t.Fatalf("command=%q want=%q", got, want)
	}
	if got, want := buildTMWebDriverInstallCommand("C:/GA/.venv/Scripts/python.exe"), "C:/GA/.venv/Scripts/python.exe -m pip install -i https://pypi.tuna.tsinghua.edu.cn/simple requests bottle simple-websocket-server"; got != want {
		t.Fatalf("command=%q want=%q", got, want)
	}
}

func TestBuildGitMirrorArgs(t *testing.T) {
	mirror := "https://gh-proxy.com/https://github.com/"
	if got, want := buildGitMirrorArgs(true, mirror), []string{"config", "--global", "url." + mirror + ".insteadOf", "https://github.com/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("enable args=%v want=%v", got, want)
	}
	if got, want := buildGitMirrorArgs(false, mirror), []string{"config", "--global", "--unset-all", "url." + mirror + ".insteadOf"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("disable args=%v want=%v", got, want)
	}
}
