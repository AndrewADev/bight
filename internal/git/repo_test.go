package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// makeMainRepo creates <dir>/.git with hooks/ and a HEAD on branch.
func makeMainRepo(t *testing.T, dir, branch string) {
	t.Helper()
	mkdir(t, filepath.Join(dir, ".git", "hooks"))
	writeFile(t, filepath.Join(dir, ".git", "HEAD"), "ref: refs/heads/"+branch+"\n")
}

// makeLinkedWorktree registers a worktree named name under main's .git and
// returns the per-worktree git dir. HEAD is set to branch.
func makeLinkedWorktree(t *testing.T, main, name, branch string) string {
	t.Helper()
	worktreeGitDir := filepath.Join(main, ".git", "worktrees", name)
	mkdir(t, worktreeGitDir)
	writeFile(t, filepath.Join(worktreeGitDir, "commondir"), "../..")
	writeFile(t, filepath.Join(worktreeGitDir, "HEAD"), "ref: refs/heads/"+branch+"\n")
	return worktreeGitDir
}

func TestDiscover_MainRepo(t *testing.T) {
	dir := t.TempDir()
	makeMainRepo(t, dir, "my-feature")

	repo, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Repo{
		Root:      dir,
		GitDir:    filepath.Join(dir, ".git"),
		CommonDir: filepath.Join(dir, ".git"),
		Branch:    "my-feature",
	}
	if *repo != want {
		t.Errorf("got %+v, want %+v", *repo, want)
	}
	if got := repo.MainRoot(); got != dir {
		t.Errorf("MainRoot() = %q, want %q", got, dir)
	}
	if got, want := repo.HooksDir(), filepath.Join(dir, ".git", "hooks"); got != want {
		t.Errorf("HooksDir() = %q, want %q", got, want)
	}
}

func TestDiscover_FromSubdir(t *testing.T) {
	dir := t.TempDir()
	makeMainRepo(t, dir, "main")
	sub := filepath.Join(dir, "a", "b")
	mkdir(t, sub)

	repo, err := Discover(sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Root != dir {
		t.Errorf("Root = %q, want %q", repo.Root, dir)
	}
}

func TestDiscover_RelativeDir(t *testing.T) {
	dir := t.TempDir()
	makeMainRepo(t, dir, "main")
	mkdir(t, filepath.Join(dir, "sub"))
	t.Chdir(filepath.Join(dir, "sub"))

	repo, err := Discover(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(repo.Root) {
		t.Errorf("Root = %q, want absolute path", repo.Root)
	}
}

func TestDiscover_DetachedHead(t *testing.T) {
	dir := t.TempDir()
	hash := "abc1234def5678900000000000000000000000000"
	mkdir(t, filepath.Join(dir, ".git"))
	writeFile(t, filepath.Join(dir, ".git", "HEAD"), hash+"\n")

	repo, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Branch != hash {
		t.Errorf("Branch = %q, want %q", repo.Branch, hash)
	}
}

func TestDiscover_LinkedWorktree(t *testing.T) {
	main := t.TempDir()
	makeMainRepo(t, main, "main")
	worktreeGitDir := makeLinkedWorktree(t, main, "my-branch", "worktree-branch")

	wt := t.TempDir()
	sub := filepath.Join(wt, "sub")
	mkdir(t, sub)
	writeFile(t, filepath.Join(wt, ".git"), "gitdir: "+worktreeGitDir+"\n")

	repo, err := Discover(sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Repo{
		Root:      wt,
		GitDir:    worktreeGitDir,
		CommonDir: filepath.Join(main, ".git"),
		Branch:    "worktree-branch",
	}
	if *repo != want {
		t.Errorf("got %+v, want %+v", *repo, want)
	}
	if got := repo.MainRoot(); got != main {
		t.Errorf("MainRoot() = %q, want %q", got, main)
	}
	if got, want := repo.HooksDir(), filepath.Join(main, ".git", "hooks"); got != want {
		t.Errorf("HooksDir() = %q, want %q", got, want)
	}
}

func TestDiscover_LinkedWorktreeRelativeGitdir(t *testing.T) {
	parent := t.TempDir()
	main := filepath.Join(parent, "main")
	makeMainRepo(t, main, "main")
	worktreeGitDir := makeLinkedWorktree(t, main, "wt", "wt-branch")

	wt := filepath.Join(parent, "wt")
	mkdir(t, wt)
	// Relative to the .git file's directory, as written by worktree.useRelativePaths.
	writeFile(t, filepath.Join(wt, ".git"), "gitdir: ../main/.git/worktrees/wt\n")

	repo, err := Discover(wt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.GitDir != worktreeGitDir {
		t.Errorf("GitDir = %q, want %q", repo.GitDir, worktreeGitDir)
	}
	if want := filepath.Join(main, ".git"); repo.CommonDir != want {
		t.Errorf("CommonDir = %q, want %q", repo.CommonDir, want)
	}
	if repo.Branch != "wt-branch" {
		t.Errorf("Branch = %q, want %q", repo.Branch, "wt-branch")
	}
}

func TestDiscover_GitFileWithoutCommondir(t *testing.T) {
	// Submodule layout: .git file points at a git dir with no commondir.
	gitDir := t.TempDir()
	writeFile(t, filepath.Join(gitDir, "HEAD"), "ref: refs/heads/main\n")
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".git"), "gitdir: "+gitDir)

	repo, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.CommonDir != gitDir {
		t.Errorf("CommonDir = %q, want %q", repo.CommonDir, gitDir)
	}
}

func TestDiscover_NoDotGit(t *testing.T) {
	_, err := Discover(t.TempDir())
	if !errors.Is(err, ErrNotRepo) {
		t.Errorf("expected ErrNotRepo, got %v", err)
	}
}

func TestDiscover_BadGitFileFormat(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".git"), "not-a-gitdir-line")

	_, err := Discover(dir)
	if err == nil || !strings.Contains(err.Error(), "unexpected format") {
		t.Errorf("expected format error, got %v", err)
	}
}

func TestDiscover_MissingHead(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, filepath.Join(dir, ".git"))

	_, err := Discover(dir)
	if err == nil || !strings.Contains(err.Error(), "HEAD") {
		t.Errorf("expected HEAD error, got %v", err)
	}
}
