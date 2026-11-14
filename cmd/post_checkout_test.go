package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

const detachedTestConfig = `project: p
defaults:
  branch_template: "db_{{slug .Branch}}"
env_files:
  - path: .env
    copy: .env.example
    vars:
      - name: DB_NAME
        strategy: template
        on: checkout
`

// setupPostCheckoutRepo creates a repo in a temp dir whose HEAD file holds
// head, writes a .bight.yml with a copy source and a template var, and
// changes into it.
func setupPostCheckoutRepo(t *testing.T, head string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv(envConfigPath, "")
	os.Unsetenv(envConfigPath)
	dir := t.TempDir()
	files := map[string]string{
		".git/HEAD":    head + "\n",
		".bight.yml":   detachedTestConfig,
		".env.example": "DB_NAME=main_db\n",
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	return dir
}

func runPostCheckout(t *testing.T) {
	t.Helper()
	cmd := postCheckoutCmd()
	cmd.SetArgs([]string{"a", "b", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("post-checkout: %v", err)
	}
}

func TestPostCheckout_DetachedLeavesEnvUnchanged(t *testing.T) {
	dir := setupPostCheckoutRepo(t, "abc1234def5678900000000000000000000000000")
	envPath := filepath.Join(dir, ".env")
	original := []byte("DB_NAME=\"db_feat_x\"\n")
	if err := os.WriteFile(envPath, original, 0o644); err != nil {
		t.Fatal(err)
	}

	runPostCheckout(t)

	got, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Errorf(".env = %q, want %q", got, original)
	}
}

func TestPostCheckout_DetachedSkipsCopy(t *testing.T) {
	dir := setupPostCheckoutRepo(t, "abc1234def5678900000000000000000000000000")

	runPostCheckout(t)

	if _, err := os.Stat(filepath.Join(dir, ".env")); !os.IsNotExist(err) {
		t.Errorf("expected no .env on detached HEAD, stat err = %v", err)
	}
}

func TestPostCheckout_BranchCopiesAndPatches(t *testing.T) {
	dir := setupPostCheckoutRepo(t, "ref: refs/heads/feat/y")

	runPostCheckout(t)

	got, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "DB_NAME=\"db_feat_y\"\n" {
		t.Errorf(".env = %q, want DB_NAME rendered for feat/y", got)
	}
}
