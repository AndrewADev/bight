package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveGitDir_Directory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := resolveGitDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Join(dir, ".git") {
		t.Errorf("got %q, want %q", got, filepath.Join(dir, ".git"))
	}
}

func TestResolveGitDir_Worktree(t *testing.T) {
	dir := t.TempDir()
	realGitDir := "/some/repo/.git/worktrees/my-worktree"
	content := "gitdir: " + realGitDir + "\n"
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveGitDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != realGitDir {
		t.Errorf("got %q, want %q", got, realGitDir)
	}
}

func TestResolveGitDir_UnexpectedFileContent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("not a gitdir pointer\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := resolveGitDir(dir)
	if err == nil {
		t.Fatal("expected error for unexpected .git file content, got nil")
	}
}

func TestResolveBranch_Normal(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/my-feature\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-feature" {
		t.Errorf("got %q, want %q", got, "my-feature")
	}
}

func TestResolveBranch_DetachedHead(t *testing.T) {
	dir := t.TempDir()
	hash := "abc1234def5678900000000000000000000000000"
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte(hash+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != hash {
		t.Errorf("got %q, want %q", got, hash)
	}
}

func TestResolveBranch_Worktree(t *testing.T) {
	// simulate the worktree structure:
	// <worktree>/.git  (file)  →  <realGitDir>/HEAD
	realGitDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(realGitDir, "HEAD"), []byte("ref: refs/heads/worktree-branch\n"), 0644); err != nil {
		t.Fatal(err)
	}

	worktreeDir := t.TempDir()
	gitFile := "gitdir: " + realGitDir + "\n"
	if err := os.WriteFile(filepath.Join(worktreeDir, ".git"), []byte(gitFile), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveBranch(worktreeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "worktree-branch" {
		t.Errorf("got %q, want %q", got, "worktree-branch")
	}
}

func TestResolveBranch_MissingHead(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	// no HEAD file written

	_, err := resolveBranch(dir)
	if err == nil {
		t.Fatal("expected error for missing HEAD, got nil")
	}
}
