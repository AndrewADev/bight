package hook

import (
	"os"
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
