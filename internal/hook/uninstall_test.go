package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hookFile(dir string) string {
	return filepath.Join(dir, ".git", "hooks", "post-checkout")
}

func makeHooksDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeHook(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallNoHookFile(t *testing.T) {
	dir := makeHooksDir(t)

	err := uninstall(hookFile(dir))
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("expected 'not installed' error, got %v", err)
	}
}

func TestUninstallBightLineNotFound(t *testing.T) {
	dir := makeHooksDir(t)
	writeHook(t, hookFile(dir), "#!/bin/sh\n/some/other/tool run \"$@\"\n")

	err := uninstall(hookFile(dir))
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("expected 'not installed' error, got %v", err)
	}
}

func TestUninstallBightOnly(t *testing.T) {
	dir := makeHooksDir(t)
	writeHook(t, hookFile(dir), "#!/bin/sh\n/usr/local/bin/bight post-checkout \"$@\"\n")

	if err := uninstall(hookFile(dir)); err != nil {
		t.Fatalf("uninstall() error: %v", err)
	}
	if _, err := os.Stat(hookFile(dir)); !os.IsNotExist(err) {
		t.Error("expected hook file to be removed")
	}
}

func TestUninstallSharedHook(t *testing.T) {
	dir := makeHooksDir(t)
	writeHook(t, hookFile(dir), "#!/bin/sh\n/some/other/tool run \"$@\"\n/usr/local/bin/bight post-checkout \"$@\"\n")

	if err := uninstall(hookFile(dir)); err != nil {
		t.Fatalf("uninstall() error: %v", err)
	}

	data, err := os.ReadFile(hookFile(dir))
	if err != nil {
		t.Fatalf("reading hook after uninstall: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "post-checkout") {
		t.Error("bight line still present after uninstall")
	}
	if !strings.Contains(content, "/some/other/tool") {
		t.Error("other hook content was removed")
	}
}

func TestUninstallPreservesPermissions(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)
	writeHook(t, path, "#!/bin/sh\n/some/other/tool run \"$@\"\n/usr/local/bin/bight post-checkout \"$@\"\n")
	if err := os.Chmod(path, 0750); err != nil {
		t.Fatal(err)
	}

	if err := uninstall(path); err != nil {
		t.Fatalf("uninstall() error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0750 {
		t.Errorf("permissions = %o, want %o", got, 0750)
	}
}
