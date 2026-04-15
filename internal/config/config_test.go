package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	yaml := `
project: myapp
defaults:
  branch_template: "{{.Project}}_{{.Branch}}"
env_files:
  - path: .env
    vars:
      - name: DB_NAME
        strategy: template
        on: checkout
      - name: JWT_SECRET
        strategy: random
        on: db_create
`
	f, err := os.CreateTemp("", "bight-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(yaml)
	f.Close()

	cfg, err := load(f.Name())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Project != "myapp" {
		t.Errorf("Project = %q, want %q", cfg.Project, "myapp")
	}
	if cfg.Defaults.BranchTemplate != "{{.Project}}_{{.Branch}}" {
		t.Errorf("BranchTemplate = %q", cfg.Defaults.BranchTemplate)
	}
	if len(cfg.EnvFiles) != 1 {
		t.Fatalf("EnvFiles len = %d, want 1", len(cfg.EnvFiles))
	}
	ef := cfg.EnvFiles[0]
	if ef.Path != ".env" {
		t.Errorf("Path = %q, want %q", ef.Path, ".env")
	}
	if len(ef.Vars) != 2 {
		t.Fatalf("Vars len = %d, want 2", len(ef.Vars))
	}
	if ef.Vars[0].Name != "DB_NAME" || ef.Vars[0].Strategy != "template" || ef.Vars[0].On != "checkout" {
		t.Errorf("Vars[0] = %+v", ef.Vars[0])
	}
	if ef.Vars[1].Name != "JWT_SECRET" || ef.Vars[1].Strategy != "random" || ef.Vars[1].On != "db_create" {
		t.Errorf("Vars[1] = %+v", ef.Vars[1])
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := load("/nonexistent/.bight.yml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// writeYAML writes content to a file at dir/name and returns the path.
func writeYAML(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// withHome overrides userHomeDir and the working directory for the duration of
// the test, then restores both.
func withHome(t *testing.T, homeDir, repoDir string) {
	t.Helper()
	orig := userHomeDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	userHomeDir = func() (string, error) { return homeDir, nil }
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		userHomeDir = orig
		os.Chdir(origDir)
	})
}

func TestLoadRepoOnly(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	writeYAML(t, repo, ".bight.yml", `
project: myapp
defaults:
  branch_template: "repo_{{.Branch}}"
env_files:
  - path: .env
    vars:
      - name: DB_NAME
        strategy: template
        on: checkout
`)
	withHome(t, home, repo)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Project != "myapp" {
		t.Errorf("Project = %q, want myapp", cfg.Project)
	}
	if cfg.Defaults.BranchTemplate != "repo_{{.Branch}}" {
		t.Errorf("BranchTemplate = %q", cfg.Defaults.BranchTemplate)
	}
}

func TestLoadGlobalOnly(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	writeYAML(t, home, ".bight.yml", `
defaults:
  branch_template: "global_{{.Branch}}"
`)
	withHome(t, home, repo)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Defaults.BranchTemplate != "global_{{.Branch}}" {
		t.Errorf("BranchTemplate = %q, want global_{{.Branch}}", cfg.Defaults.BranchTemplate)
	}
}

func TestLoadNeither(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	withHome(t, home, repo)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when no config files exist")
	}
}

func TestLoadGlobalMissingIsSilent(t *testing.T) {
	home := t.TempDir() // no .bight.yml written here
	repo := t.TempDir()
	writeYAML(t, repo, ".bight.yml", `project: myapp`)
	withHome(t, home, repo)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Project != "myapp" {
		t.Errorf("Project = %q, want myapp", cfg.Project)
	}
}

func TestLoadRepoOverridesGlobal(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	writeYAML(t, home, ".bight.yml", `
defaults:
  branch_template: "global_{{.Branch}}"
`)
	writeYAML(t, repo, ".bight.yml", `
defaults:
  branch_template: "repo_{{.Branch}}"
`)
	withHome(t, home, repo)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Defaults.BranchTemplate != "repo_{{.Branch}}" {
		t.Errorf("BranchTemplate = %q, want repo_{{.Branch}}", cfg.Defaults.BranchTemplate)
	}
}

func TestLoadGlobalFillsGap(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	writeYAML(t, home, ".bight.yml", `
defaults:
  branch_template: "global_{{.Branch}}"
`)
	writeYAML(t, repo, ".bight.yml", `project: myapp`)
	withHome(t, home, repo)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Defaults.BranchTemplate != "global_{{.Branch}}" {
		t.Errorf("BranchTemplate = %q, want global_{{.Branch}}", cfg.Defaults.BranchTemplate)
	}
	if cfg.Project != "myapp" {
		t.Errorf("Project = %q, want myapp", cfg.Project)
	}
}

func TestLoadGlobalEnvFilesWarning(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	writeYAML(t, home, ".bight.yml", `
defaults:
  branch_template: "global_{{.Branch}}"
env_files:
  - path: .env
    vars:
      - name: DB_NAME
        strategy: template
        on: checkout
`)
	writeYAML(t, repo, ".bight.yml", `project: myapp`)
	withHome(t, home, repo)

	// Capture stderr.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStderr := os.Stderr
	os.Stderr = w

	cfg, loadErr := Load()

	w.Close()
	os.Stderr = origStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	r.Close()
	stderr := string(buf[:n])

	if loadErr != nil {
		t.Fatalf("Load: %v", loadErr)
	}
	if len(cfg.EnvFiles) != 0 {
		t.Errorf("expected global env_files to be stripped, got %d entries", len(cfg.EnvFiles))
	}
	if !strings.Contains(stderr, "env_files in ~/.bight.yml is not supported") {
		t.Errorf("expected warning about env_files, got: %q", stderr)
	}
}
