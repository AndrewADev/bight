// Package git locates a repository on disk by reading .git, gitdir and
// commondir files directly. HEAD of a reftable repository is resolved by
// running git.
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNotRepo is returned by Discover when no .git entry exists in dir or any
// of its parents.
var ErrNotRepo = errors.New(".git not found — are you in a git repo?")

// Repo describes the repository enclosing a directory.
type Repo struct {
	// Root is the absolute path of the worktree root (the directory containing .git).
	Root string
	// GitDir is the per-worktree git directory, where HEAD lives. In the main
	// worktree this is Root/.git; in a linked worktree it is the directory
	// named by the gitdir line of the .git file.
	GitDir string
	// CommonDir is the git directory shared by all worktrees, where hooks live.
	CommonDir string
	// Branch is the checked-out branch name. It is empty on a detached HEAD.
	Branch string
	// Detached is true when HEAD points at a commit rather than a branch, as
	// during a rebase, a bisect or after `git checkout --detach`.
	Detached bool
}

// MainRoot returns the working directory of the main worktree.
func (r *Repo) MainRoot() string {
	return filepath.Dir(r.CommonDir)
}

// HooksDir returns the directory git reads hooks from.
func (r *Repo) HooksDir() string {
	return filepath.Join(r.CommonDir, "hooks")
}

// Discover walks up from dir to the nearest .git entry and resolves the
// repository it belongs to.
func Discover(dir string) (*Repo, error) {
	root, err := findRoot(dir)
	if err != nil {
		return nil, err
	}
	gitDir, err := resolveGitDir(root)
	if err != nil {
		return nil, err
	}
	commonDir, err := resolveCommonDir(gitDir)
	if err != nil {
		return nil, err
	}
	branch, detached, err := readBranch(gitDir)
	if err != nil {
		return nil, err
	}
	return &Repo{
		Root:      root,
		GitDir:    gitDir,
		CommonDir: commonDir,
		Branch:    branch,
		Detached:  detached,
	}, nil
}

func findRoot(dir string) (string, error) {
	current, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		dotGit := filepath.Join(current, ".git")
		_, err := os.Stat(dotGit)
		if err == nil {
			return current, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", ErrNotRepo
		}
		current = parent
	}
}

// resolveGitDir returns root/.git when it is a directory, or the target of
// its "gitdir: <path>" line when it is a file. A relative gitdir is resolved
// against root.
func resolveGitDir(root string) (string, error) {
	dotGit := filepath.Join(root, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return dotGit, nil
	}
	data, err := os.ReadFile(dotGit)
	if err != nil {
		return "", fmt.Errorf("reading .git: %w", err)
	}
	line := strings.TrimSpace(string(data))
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf(".git file has unexpected format: %q", line)
	}
	gitDir := line[len(prefix):]
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(root, gitDir)
	}
	return filepath.Clean(gitDir), nil
}

// resolveCommonDir follows gitDir/commondir when present. Without a commondir
// file (main worktree, submodule) gitDir is its own common dir.
func resolveCommonDir(gitDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(gitDir, "commondir"))
	if errors.Is(err, os.ErrNotExist) {
		return gitDir, nil
	}
	if err != nil {
		return "", fmt.Errorf("reading commondir: %w", err)
	}
	commonDir := strings.TrimSpace(string(data))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(gitDir, commonDir)
	}
	return filepath.Clean(commonDir), nil
}

// readBranch reads gitDir/HEAD and returns the branch it names. It reports
// detached when HEAD holds a commit hash or a ref outside refs/heads/.
func readBranch(gitDir string) (branch string, detached bool, err error) {
	data, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", false, fmt.Errorf("reading HEAD: %w", err)
	}
	head := strings.TrimSpace(string(data))
	if head == reftablePlaceholder {
		head, err = readReftableHead(gitDir)
		if err != nil {
			return "", false, err
		}
	}
	const prefix = "ref: refs/heads/"
	if strings.HasPrefix(head, prefix) {
		return head[len(prefix):], false, nil
	}
	return "", true, nil
}

// reftablePlaceholder is the content git writes to HEAD in repositories using
// the reftable ref backend. The real HEAD lives in the reftable stack.
const reftablePlaceholder = "ref: refs/heads/.invalid"

// gitLocationEnv lists environment variables that would redirect git away
// from the git dir passed on the command line.
var gitLocationEnv = []string{"GIT_DIR=", "GIT_COMMON_DIR=", "GIT_WORK_TREE="}

// readReftableHead resolves HEAD of the worktree owning gitDir with
// `git symbolic-ref` and returns it in the format of a files-backend HEAD:
// "ref: <target>" for a symbolic ref, or an empty string on a detached HEAD.
func readReftableHead(gitDir string) (string, error) {
	cmd := exec.Command("git", "--git-dir="+gitDir, "symbolic-ref", "-q", "HEAD")
	cmd.Env = filterEnv(os.Environ(), gitLocationEnv)
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return "", nil
	}
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return "", fmt.Errorf("reading HEAD from reftable repository with git: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	if err != nil {
		return "", fmt.Errorf("reading HEAD from reftable repository with git: %w", err)
	}
	return "ref: " + strings.TrimSpace(string(out)), nil
}

func filterEnv(env, prefixes []string) []string {
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		drop := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(entry, prefix) {
				drop = true
				break
			}
		}
		if !drop {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
