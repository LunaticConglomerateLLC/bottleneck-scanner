package integration_test
// Package integration_test exercises the compiled bottleneck binary end-to-end.
// Run with: go test ./test/
package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binary holds the path to the compiled binary, built once in TestMain.
var binary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "bottleneck-integ-*")
	if err != nil {
		panic("MkdirTemp: " + err.Error())
	}
	defer os.RemoveAll(tmp)

	binary = filepath.Join(tmp, "bottleneck")
	out, err := exec.Command("go", "build", "-o", binary, "..").CombinedOutput()
	if err != nil {
		panic("build failed:\n" + string(out))
	}

	os.Exit(m.Run())
}

// run executes the binary with the given arguments and returns combined output.
func run(args ...string) string {
	out, _ := exec.Command(binary, args...).CombinedOutput()
	return string(out)
}

func TestBinary_NoArgs_PrintsUsage(t *testing.T) {
	out := run()
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected Usage in output, got:\n%s", out)
	}
}

func TestBinary_InvalidReportMode(t *testing.T) {
	out := run("--report", "invalid", "owner/repo")
	if !strings.Contains(out, "--report must be one of") {
		t.Errorf("expected validation error, got:\n%s", out)
	}
}

func TestBinary_InvalidSAMode(t *testing.T) {
	out := run("--service-accounts", "invalid", "owner/repo")
	if !strings.Contains(out, "--service-accounts must be one of") {
		t.Errorf("expected validation error, got:\n%s", out)
	}
}

func TestBinary_TeamAndRepoMutuallyExclusive(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "team-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("name: t\nrepos:\n  - org/repo\n")
	f.Close()

	out := run("--team", f.Name(), "owner/repo")
	if !strings.Contains(out, "mutually exclusive") {
		t.Errorf("expected mutual exclusion error, got:\n%s", out)
	}
}

func TestBinary_InvalidRepoFormat(t *testing.T) {
	out := run("invalid-repo-no-slash")
	if !strings.Contains(out, "owner/repo") {
		t.Errorf("expected format error mentioning owner/repo, got:\n%s", out)
	}
}
