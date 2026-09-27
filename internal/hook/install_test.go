package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHooksDir_RegularRepo(t *testing.T) {
	dir := makeHooksDir(t)
	got, err := hooksDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(dir, ".git", "hooks")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHooksDir_Worktree(t *testing.T) {
	// Main repo: <main>/.git/hooks and <main>/.git/worktrees/<name>/commondir
	main := makeHooksDir(t)
	worktreeGitDir := filepath.Join(main, ".git", "worktrees", "my-branch")
	if err := os.MkdirAll(worktreeGitDir, 0755); err != nil {
		t.Fatal(err)
	}
	// commondir is relative to worktreeGitDir, pointing at <main>/.git
	if err := os.WriteFile(filepath.Join(worktreeGitDir, "commondir"), []byte("../.."), 0644); err != nil {
		t.Fatal(err)
	}

	// Worktree dir: .git is a file pointing to worktreeGitDir
	wt := t.TempDir()
	gitFile := "gitdir: " + worktreeGitDir
	if err := os.WriteFile(filepath.Join(wt, ".git"), []byte(gitFile), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := hooksDir(wt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(main, ".git", "hooks")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHooksDir_NoDotGit(t *testing.T) {
	dir := t.TempDir()
	_, err := hooksDir(dir)
	if err == nil || !strings.Contains(err.Error(), "are you in a git repo") {
		t.Errorf("expected git repo error, got %v", err)
	}
}

func TestHooksDir_BadGitFileFormat(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("not-a-gitdir-line"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := hooksDir(dir)
	if err == nil || !strings.Contains(err.Error(), "unexpected format") {
		t.Errorf("expected format error, got %v", err)
	}
}

func TestHooksDir_MissingCommondir(t *testing.T) {
	// worktreeGitDir exists but has no commondir file
	worktreeGitDir := t.TempDir()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: "+worktreeGitDir), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := hooksDir(dir)
	if err == nil || !strings.Contains(err.Error(), "commondir") {
		t.Errorf("expected commondir error, got %v", err)
	}
}

func TestMainWorktreeRoot_RegularRepo(t *testing.T) {
	dir := makeHooksDir(t)
	got, err := mainWorktreeRoot(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// In a regular repo, the main worktree root is the dir containing .git.
	if got != dir {
		t.Errorf("got %q, want %q", got, dir)
	}
}

func TestMainWorktreeRoot_Worktree(t *testing.T) {
	// Layout: <main>/.git is a directory; <wt>/.git is a file pointing at
	// <main>/.git/worktrees/<name>; that dir's commondir → <main>/.git.
	main := makeHooksDir(t)
	worktreeGitDir := filepath.Join(main, ".git", "worktrees", "my-branch")
	if err := os.MkdirAll(worktreeGitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktreeGitDir, "commondir"), []byte("../.."), 0644); err != nil {
		t.Fatal(err)
	}

	wt := t.TempDir()
	gitFile := "gitdir: " + worktreeGitDir
	if err := os.WriteFile(filepath.Join(wt, ".git"), []byte(gitFile), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := mainWorktreeRoot(wt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != main {
		t.Errorf("got %q, want %q", got, main)
	}
}

func readHook(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading hook: %v", err)
	}
	return string(data)
}

func TestInstallFresh(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)

	if err := install(path, "/usr/local/bin/bight"); err != nil {
		t.Fatalf("install() error: %v", err)
	}

	want := fmt.Sprintf(hookScript, "/usr/local/bin/bight")
	if got := readHook(t, path); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0755 {
		t.Errorf("permissions = %o, want %o", got, 0755)
	}
}

func TestInstallReinstallReplacesBlock(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)

	if err := install(path, "/old/bight"); err != nil {
		t.Fatalf("first install() error: %v", err)
	}
	if err := install(path, "/new/bight"); err != nil {
		t.Fatalf("second install() error: %v", err)
	}

	content := readHook(t, path)
	if n := strings.Count(content, blockBegin); n != 1 {
		t.Errorf("found %d bight blocks, want 1:\n%s", n, content)
	}
	if strings.Contains(content, "/old/bight") {
		t.Errorf("old binary path still present:\n%s", content)
	}
	if !strings.Contains(content, "/new/bight") {
		t.Errorf("new binary path missing:\n%s", content)
	}
	if want := fmt.Sprintf(hookScript, "/new/bight"); content != want {
		t.Errorf("got:\n%s\nwant:\n%s", content, want)
	}
}

func TestInstallReplacesLegacyLine(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)
	writeHook(t, path, "#!/bin/sh\n/some/other/tool run \"$@\"\n/old/bight post-checkout \"$@\"\n")

	if err := install(path, "/new/bight"); err != nil {
		t.Fatalf("install() error: %v", err)
	}

	content := readHook(t, path)
	if strings.Contains(content, "/old/bight") {
		t.Errorf("legacy line still present:\n%s", content)
	}
	if n := strings.Count(content, blockBegin); n != 1 {
		t.Errorf("found %d bight blocks, want 1:\n%s", n, content)
	}
	if !strings.Contains(content, "/some/other/tool") {
		t.Error("other hook content was removed")
	}
}

func TestInstallPreservesForeignContent(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)
	writeHook(t, path, "#!/bin/sh\n/some/other/tool run \"$@\"\n")

	if err := install(path, "/usr/local/bin/bight"); err != nil {
		t.Fatalf("install() error: %v", err)
	}

	want := "#!/bin/sh\n/some/other/tool run \"$@\"\n" + fmt.Sprintf(hookBlock, "/usr/local/bin/bight")
	if got := readHook(t, path); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestInstallKeepsExistingShebang(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)
	writeHook(t, path, "#!/bin/bash\n/some/other/tool run \"$@\"\n")

	if err := install(path, "/usr/local/bin/bight"); err != nil {
		t.Fatalf("install() error: %v", err)
	}

	content := readHook(t, path)
	if !strings.HasPrefix(content, "#!/bin/bash\n") {
		t.Errorf("existing shebang not kept:\n%s", content)
	}
	if strings.Contains(content, shebang) {
		t.Errorf("default shebang added alongside existing one:\n%s", content)
	}
}

func TestInstallAddsMissingShebang(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)
	writeHook(t, path, "/some/other/tool run \"$@\"\n")

	if err := install(path, "/usr/local/bin/bight"); err != nil {
		t.Fatalf("install() error: %v", err)
	}

	want := "#!/bin/sh\n/some/other/tool run \"$@\"\n" + fmt.Sprintf(hookBlock, "/usr/local/bin/bight")
	if got := readHook(t, path); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestInstallPreservesPermissions(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)
	writeHook(t, path, "#!/bin/sh\n/some/other/tool run \"$@\"\n")
	if err := os.Chmod(path, 0750); err != nil {
		t.Fatal(err)
	}

	if err := install(path, "/usr/local/bin/bight"); err != nil {
		t.Fatalf("install() error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0750 {
		t.Errorf("permissions = %o, want %o", got, 0750)
	}
}

func TestInstallThenUninstallRoundTrip(t *testing.T) {
	dir := makeHooksDir(t)
	path := hookFile(dir)
	original := "#!/bin/sh\n/some/other/tool run \"$@\"\n"
	writeHook(t, path, original)

	if err := install(path, "/usr/local/bin/bight"); err != nil {
		t.Fatalf("install() error: %v", err)
	}
	if err := uninstall(path); err != nil {
		t.Fatalf("uninstall() error: %v", err)
	}

	if got := readHook(t, path); got != original {
		t.Errorf("got:\n%s\nwant:\n%s", got, original)
	}
}

// fakeBight writes an executable at path that records its arguments to logFile.
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
