package git

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo runs `git init` with the given ref format in a temp dir, creates a
// commit on branch main and returns the worktree root. The test is skipped
// when git is missing or does not support the ref format.
func initRepo(t *testing.T, refFormat string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, "repo")
	if out, err := exec.Command("git", "init", "-q", "-b", "main", "--ref-format="+refFormat, root).CombinedOutput(); err != nil {
		t.Skipf("git init --ref-format=%s unsupported: %v: %s", refFormat, err, out)
	}
	runGit(t, root, "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "--allow-empty", "-m", "init")
	return root
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = filterEnv(cmd.Environ(), gitLocationEnv)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func assertHead(t *testing.T, dir, wantBranch string, wantDetached bool) {
	t.Helper()
	repo, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover(%q): %v", dir, err)
	}
	if repo.Branch != wantBranch || repo.Detached != wantDetached {
		t.Errorf("Discover(%q): got Branch=%q Detached=%v, want Branch=%q Detached=%v",
			dir, repo.Branch, repo.Detached, wantBranch, wantDetached)
	}
}

func testRealRepo(t *testing.T, refFormat string) {
	root := initRepo(t, refFormat)
	wt := filepath.Join(filepath.Dir(root), "wt")
	runGit(t, root, "worktree", "add", "-q", "-b", "feature/x", wt)

	assertHead(t, root, "main", false)
	assertHead(t, wt, "feature/x", false)

	runGit(t, wt, "checkout", "-q", "--detach")
	assertHead(t, wt, "", true)
	assertHead(t, root, "main", false)

	runGit(t, root, "checkout", "-q", "--detach")
	assertHead(t, root, "", true)

	runGit(t, wt, "checkout", "-q", "-b", "other")
	assertHead(t, wt, "other", false)
	assertHead(t, root, "", true)
}

func TestDiscover_RealRepoFiles(t *testing.T) {
	testRealRepo(t, "files")
}

func TestDiscover_RealRepoReftable(t *testing.T) {
	testRealRepo(t, "reftable")
}

func TestDiscover_ReftableIgnoresGitDirEnv(t *testing.T) {
	root := initRepo(t, "reftable")
	other := initRepo(t, "reftable")
	runGit(t, other, "checkout", "-q", "-b", "elsewhere")
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_COMMON_DIR", filepath.Join(other, ".git"))

	assertHead(t, root, "main", false)
}

func TestDiscover_ReftableWithoutGit(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, filepath.Join(dir, ".git"))
	writeFile(t, filepath.Join(dir, ".git", "HEAD"), reftablePlaceholder+"\n")
	t.Setenv("PATH", "")

	_, err := Discover(dir)
	if err == nil || !strings.Contains(err.Error(), "reftable") {
		t.Errorf("expected reftable error, got %v", err)
	}
}
