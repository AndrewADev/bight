package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AndrewADev/bight/internal/config"
)

func TestDryRunEnvFiles_TemplateVar(t *testing.T) {
	cfg := &config.Config{
		Project: "myapp",
		Defaults: config.Defaults{
			BranchTemplate: "{{.Project}}_{{.Branch}}",
		},
		EnvFiles: []config.EnvFile{
			{
				Path: ".env",
				Vars: []config.Var{
					{Name: "DB_NAME", Strategy: "template", On: "checkout"},
				},
			},
		},
	}

	results := dryRunEnvFiles(cfg, "feature-x")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
	if r.path != ".env" {
		t.Errorf("path: got %q, want %q", r.path, ".env")
	}
	if r.varName != "DB_NAME" {
		t.Errorf("varName: got %q, want %q", r.varName, "DB_NAME")
	}
	if r.value != "myapp_feature-x" {
		t.Errorf("value: got %q, want %q", r.value, "myapp_feature-x")
	}
}

func TestDryRunEnvFiles_SkipsNonCheckoutTrigger(t *testing.T) {
	cfg := &config.Config{
		EnvFiles: []config.EnvFile{
			{
				Path: ".env",
				Vars: []config.Var{
					{Name: "DB_NAME", Strategy: "template", On: "db_create"},
				},
			},
		},
	}

	results := dryRunEnvFiles(cfg, "main")

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDryRunEnvFiles_PropagatesStrategyError(t *testing.T) {
	cfg := &config.Config{
		EnvFiles: []config.EnvFile{
			{
				Path: ".env",
				Vars: []config.Var{
					{Name: "SOME_VAR", Strategy: "nonexistent", On: "checkout"},
				},
			},
		},
	}

	results := dryRunEnvFiles(cfg, "main")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].err == nil {
		t.Error("expected error for unknown strategy, got nil")
	}
}

func TestDryRunEnvFiles_SensitiveFlagThreaded(t *testing.T) {
	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{
				Path: ".env",
				Vars: []config.Var{
					{Name: "JWT_SECRET", Strategy: "random", On: "checkout", Sensitive: true},
					{Name: "DB_NAME", Strategy: "deterministic", On: "checkout"},
				},
			},
		},
	}

	results := dryRunEnvFiles(cfg, "main")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].sensitive {
		t.Errorf("results[0] (JWT_SECRET) sensitive = false, want true")
	}
	if results[0].value == "" {
		t.Error("results[0] value should be non-empty even when sensitive")
	}
	if results[1].sensitive {
		t.Errorf("results[1] (DB_NAME) sensitive = true, want false")
	}
}

func TestDryRunEnvFiles_MultipleFilesAndVars(t *testing.T) {
	cfg := &config.Config{
		Project: "proj",
		Defaults: config.Defaults{
			BranchTemplate: "{{.Project}}_{{.Branch}}",
		},
		EnvFiles: []config.EnvFile{
			{
				Path: ".env",
				Vars: []config.Var{
					{Name: "DB_NAME", Strategy: "template", On: "checkout"},
					{Name: "SKIP_ME", Strategy: "template", On: "db_create"},
				},
			},
			{
				Path: ".env.test",
				Vars: []config.Var{
					{Name: "TEST_DB", Strategy: "template", On: "checkout"},
				},
			},
		},
	}

	results := dryRunEnvFiles(cfg, "mybranch")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].path != ".env" || results[0].varName != "DB_NAME" {
		t.Errorf("result[0]: got %q/%q", results[0].path, results[0].varName)
	}
	if results[1].path != ".env.test" || results[1].varName != "TEST_DB" {
		t.Errorf("result[1]: got %q/%q", results[1].path, results[1].varName)
	}
}

func TestPatchEnvFiles_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("DB_NAME=old\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project: "myapp",
		Defaults: config.Defaults{
			BranchTemplate: "{{.Project}}_{{.Branch}}",
		},
		EnvFiles: []config.EnvFile{
			{
				Path:   envPath,
				Backup: true,
				Vars:   []config.Var{{Name: "DB_NAME", Strategy: "template", On: "checkout"}},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feature-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	data, err := os.ReadFile(envPath + ".bak")
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(data) != "DB_NAME=old\n" {
		t.Errorf("backup content = %q, want %q", string(data), "DB_NAME=old\n")
	}
}

func TestDryRunEnvFiles_EventFieldOnInit(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env") // missing → init fires

	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{
				Path: envPath,
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
					{Name: "ON_SWITCH", Strategy: "random", On: "checkout"},
				},
			},
		},
	}

	results := dryRunEnvFiles(cfg, "feat-x")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %+v", len(results), results)
	}
	// Order: worktree-init pass first, then checkout pass.
	if results[0].varName != "ON_INIT" || results[0].event != "worktree-init" {
		t.Errorf("results[0] = %+v, want ON_INIT/worktree-init", results[0])
	}
	if results[1].varName != "ON_SWITCH" || results[1].event != "checkout" {
		t.Errorf("results[1] = %+v, want ON_SWITCH/checkout", results[1])
	}
}

func TestDryRunEnvFiles_EventFieldOnExistingFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("X=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{
				Path: envPath,
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
					{Name: "ON_SWITCH", Strategy: "random", On: "checkout"},
				},
			},
		},
	}

	results := dryRunEnvFiles(cfg, "feat-x")
	if len(results) != 1 {
		t.Fatalf("expected 1 result (init suppressed), got %d: %+v", len(results), results)
	}
	if results[0].varName != "ON_SWITCH" || results[0].event != "checkout" {
		t.Errorf("results[0] = %+v, want ON_SWITCH/checkout", results[0])
	}
}

func TestPatchEnvFiles_SkipsWhenNoVars(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	original := []byte("DB_NAME=keep\n# a comment\nOTHER=val\n")
	if err := os.WriteFile(envPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{Path: envPath, Vars: nil},
		},
	}

	if err := patchEnvFiles(cfg, "feature-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	got, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Errorf("file content changed: got %q, want %q", got, original)
	}
	after, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Errorf("mtime changed: before %v, after %v", before.ModTime(), after.ModTime())
	}
}

func TestPatchEnvFiles_SkipsWhenAllVarsNonCheckout(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	original := []byte("DB_NAME=keep\n")
	if err := os.WriteFile(envPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{
				Path: envPath,
				Vars: []config.Var{
					{Name: "DB_NAME", Strategy: "template", On: "db_create"},
				},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feature-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	got, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Errorf("file content changed: got %q, want %q", got, original)
	}
	after, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Errorf("mtime changed: before %v, after %v", before.ModTime(), after.ModTime())
	}
}

// Regression test for the backup over-trigger: when the file exists, has
// only worktree-init vars, and backup is enabled, no events that fire have
// any matching vars — so no .bak should be written.
func TestPatchEnvFiles_NoBackupWhenNoVarsFire(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("X=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		EnvFiles: []config.EnvFile{
			{
				Path:   envPath,
				Backup: true,
				// File exists + no Copy + only init vars → events=[checkout]
				// and no var matches checkout → nothing should change, no .bak.
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
				},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feat-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	if _, err := os.Stat(envPath + ".bak"); !os.IsNotExist(err) {
		t.Errorf("expected no backup file (nothing fired), but it exists")
	}
}

func TestPatchEnvFiles_NoBackupWhenSkipped(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("DB_NAME=keep\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{Path: envPath, Backup: true, Vars: nil},
		},
	}

	if err := patchEnvFiles(cfg, "feature-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	if _, err := os.Stat(envPath + ".bak"); !os.IsNotExist(err) {
		t.Errorf("expected no backup file, but it exists (or stat err: %v)", err)
	}
}

func TestPatchEnvFiles_CopyOnInit(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(srcPath, []byte("SEEDED=from-source\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".env") // does not exist yet

	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{
				Path: envPath,
				Copy: &config.Copy{Source: srcPath},
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
					{Name: "ON_SWITCH", Strategy: "random", On: "checkout"},
				},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feat-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("reading dest: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "SEEDED=") {
		t.Errorf("expected dest to contain SEEDED= (from copy), got: %q", s)
	}
	if !strings.Contains(s, "ON_INIT=") {
		t.Errorf("expected dest to contain ON_INIT= (worktree-init var), got: %q", s)
	}
	if !strings.Contains(s, "ON_SWITCH=") {
		t.Errorf("expected dest to contain ON_SWITCH= (checkout var), got: %q", s)
	}
}

func TestPatchEnvFiles_NoCopyWhenDestExistsAndOverwriteFalse(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(srcPath, []byte("SEEDED=from-source\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("PREEXISTING=hi\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		EnvFiles: []config.EnvFile{
			{
				Path: envPath,
				Copy: &config.Copy{Source: srcPath, Overwrite: false},
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
					{Name: "ON_SWITCH", Strategy: "random", On: "checkout"},
				},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feat-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	data, _ := os.ReadFile(envPath)
	s := string(data)
	if strings.Contains(s, "SEEDED=") {
		t.Errorf("dest unexpectedly contains SEEDED= (copy should have been skipped): %q", s)
	}
	if strings.Contains(s, "ON_INIT=") {
		t.Errorf("dest unexpectedly contains ON_INIT= (init should not have fired): %q", s)
	}
	if !strings.Contains(s, "PREEXISTING=") {
		t.Errorf("dest lost PREEXISTING=: %q", s)
	}
	if !strings.Contains(s, "ON_SWITCH=") {
		t.Errorf("dest missing ON_SWITCH= (checkout vars should still fire): %q", s)
	}
}

func TestPatchEnvFiles_OverwriteClobbersAndFiresInit(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, ".env")
	if err := os.WriteFile(srcPath, []byte("SEEDED=fresh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("OLD=ghost\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		EnvFiles: []config.EnvFile{
			{
				Path:   envPath,
				Backup: true,
				Copy:   &config.Copy{Source: srcPath, Overwrite: true},
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
				},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feat-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	data, _ := os.ReadFile(envPath)
	s := string(data)
	if !strings.Contains(s, "SEEDED=") {
		t.Errorf("dest should contain SEEDED= after overwrite: %q", s)
	}
	if strings.Contains(s, "OLD=") {
		t.Errorf("dest should not contain OLD= after overwrite: %q", s)
	}
	if !strings.Contains(s, "ON_INIT=") {
		t.Errorf("dest should contain ON_INIT= (init fires after overwrite): %q", s)
	}

	bak, err := os.ReadFile(envPath + ".bak")
	if err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if string(bak) != "OLD=ghost\n" {
		t.Errorf("backup = %q, want %q", bak, "OLD=ghost\n")
	}
}

func TestPatchEnvFiles_InitFiresWhenFileMissingNoCopy(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env") // missing

	cfg := &config.Config{
		EnvFiles: []config.EnvFile{
			{
				Path: envPath,
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
					{Name: "ON_SWITCH", Strategy: "random", On: "checkout"},
				},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feat-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("dest missing: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "ON_INIT=") {
		t.Errorf("init var missing (file was created from missing): %q", s)
	}
	if !strings.Contains(s, "ON_SWITCH=") {
		t.Errorf("checkout var missing: %q", s)
	}
}

func TestPatchEnvFiles_NoInitWhenFileExistsNoCopy(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("X=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		EnvFiles: []config.EnvFile{
			{
				Path: envPath,
				Vars: []config.Var{
					{Name: "ON_INIT", Strategy: "random", On: "worktree-init"},
					{Name: "ON_SWITCH", Strategy: "random", On: "checkout"},
				},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feat-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	data, _ := os.ReadFile(envPath)
	s := string(data)
	if strings.Contains(s, "ON_INIT=") {
		t.Errorf("init var should not fire when file exists: %q", s)
	}
	if !strings.Contains(s, "ON_SWITCH=") {
		t.Errorf("checkout var missing: %q", s)
	}
}

func TestPatchEnvFiles_NoBackupByDefault(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("DB_NAME=old\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project: "myapp",
		Defaults: config.Defaults{
			BranchTemplate: "{{.Project}}_{{.Branch}}",
		},
		EnvFiles: []config.EnvFile{
			{
				Path:   envPath,
				Backup: false,
				Vars:   []config.Var{{Name: "DB_NAME", Strategy: "template", On: "checkout"}},
			},
		},
	}

	if err := patchEnvFiles(cfg, "feature-x"); err != nil {
		t.Fatalf("patchEnvFiles: %v", err)
	}

	if _, err := os.Stat(envPath + ".bak"); !os.IsNotExist(err) {
		t.Errorf("expected no backup file, but it exists")
	}
}
