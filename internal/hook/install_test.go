package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func fakeBight(t *testing.T, path, logFile string) {
	t.Helper()
	script := "#!/bin/sh\necho \"$0 $*\" > " + logFile + "\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
}

func runHook(t *testing.T, exe string) (string, error) {
	t.Helper()
	hookPath := filepath.Join(t.TempDir(), "post-checkout")
	if err := os.WriteFile(hookPath, []byte(fmt.Sprintf(hookScript, exe)), 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", hookPath, "aaa", "bbb", "1")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestHookScript_UsesInstalledPath(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "log")
	exe := filepath.Join(dir, "bight")
	fakeBight(t, exe, logFile)

	if out, err := runHook(t, exe); err != nil {
		t.Fatalf("hook failed: %v\n%s", err, out)
	}
	got, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("installed binary was not invoked: %v", err)
	}
	want := exe + " post-checkout aaa bbb 1\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHookScript_MissingBinaryExitsZero(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone", "bight")

	out, err := runHook(t, missing)
	if err != nil {
		t.Fatalf("expected exit 0, got %v\n%s", err, out)
	}
	want := "bight: " + missing + " not found; run 'bight install' again\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestHookScript_NonExecutableBinaryExitsZero(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "bight")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := runHook(t, exe)
	if err != nil {
		t.Fatalf("expected exit 0, got %v\n%s", err, out)
	}
	want := "bight: " + exe + " is not executable\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}
