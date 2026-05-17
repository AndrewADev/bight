package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeCfg writes a minimal valid config file at the given path.
func writeCfg(t *testing.T, path, project string) {
	t.Helper()
	content := "project: " + project + "\nenv_files:\n  - path: .env\n    vars:\n      - { name: FOO, strategy: template, on: checkout }\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// isolateConfigEnv resets the --config flag var and clears BIGHT_CONFIG for
// the duration of the test, restoring both afterwards. Also chdir's into an
// empty temp dir so auto-discovery has nothing to find.
func isolateConfigEnv(t *testing.T) string {
	t.Helper()
	origFlag := configPath
	configPath = ""
	t.Cleanup(func() { configPath = origFlag })

	t.Setenv(envConfigPath, "")
	// Setenv with empty still keeps the var set; explicitly unset.
	os.Unsetenv(envConfigPath)

	// Chdir into a clean empty dir so .bight.yml in the worktree doesn't leak.
	clean := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(clean); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Point HOME at an empty dir too, so the global ~/.bight.yml merge is a no-op.
	t.Setenv("HOME", t.TempDir())

	return clean
}

func TestLoadConfig_EnvVar(t *testing.T) {
	isolateConfigEnv(t)

	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.bight.yml")
	writeCfg(t, custom, "envapp")

	t.Setenv(envConfigPath, custom)

	cfg, path, source, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.Project != "envapp" {
		t.Errorf("Project = %q, want envapp", cfg.Project)
	}
	if path != custom {
		t.Errorf("path = %q, want %q", path, custom)
	}
	if source != sourceEnv {
		t.Errorf("source = %v, want sourceEnv", source)
	}
}

func TestLoadConfig_FlagBeatsEnv(t *testing.T) {
	isolateConfigEnv(t)

	dir := t.TempDir()
	envFile := filepath.Join(dir, "env.bight.yml")
	flagFile := filepath.Join(dir, "flag.bight.yml")
	writeCfg(t, envFile, "envapp")
	writeCfg(t, flagFile, "flagapp")

	t.Setenv(envConfigPath, envFile)
	configPath = flagFile
	t.Cleanup(func() { configPath = "" })

	cfg, path, source, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.Project != "flagapp" {
		t.Errorf("flag should win — Project = %q, want flagapp", cfg.Project)
	}
	if path != flagFile {
		t.Errorf("path = %q, want %q", path, flagFile)
	}
	if source != sourceFlag {
		t.Errorf("source = %v, want sourceFlag", source)
	}
}

func TestLoadConfig_EnvMissingFileFails(t *testing.T) {
	isolateConfigEnv(t)

	t.Setenv(envConfigPath, "/nonexistent/bight.yml")

	_, _, _, err := loadConfig()
	if err == nil {
		t.Fatal("expected error when BIGHT_CONFIG points to a missing file")
	}
	if errors.Is(err, os.ErrNotExist) {
		// good — propagates filesystem error
	}
}

func TestLoadConfig_AutoDiscoverWhenUnset(t *testing.T) {
	clean := isolateConfigEnv(t)

	// Drop a .bight.yml in the (now-current) clean dir.
	writeCfg(t, filepath.Join(clean, ".bight.yml"), "autoapp")

	cfg, path, source, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.Project != "autoapp" {
		t.Errorf("Project = %q, want autoapp", cfg.Project)
	}
	if path != ".bight.yml" {
		t.Errorf("path = %q, want .bight.yml", path)
	}
	if source != sourceAuto {
		t.Errorf("source = %v, want sourceAuto", source)
	}
}

func TestRunChecks_ConfigPathShownForEnv(t *testing.T) {
	deps := happyDeps
	deps.cfgPath = "/tmp/custom.bight.yml"
	deps.cfgSource = sourceEnv

	results := runChecks(validCfg, nil, deps)
	r, found := findByPrefix(results, "config: ")
	if !found {
		t.Fatal("no config result")
	}
	if !strings.Contains(r.msg, "/tmp/custom.bight.yml") {
		t.Errorf("expected path in message, got %q", r.msg)
	}
	if !strings.Contains(r.msg, "(from BIGHT_CONFIG)") {
		t.Errorf("expected env source hint, got %q", r.msg)
	}
}

func TestRunChecks_ConfigPathShownForFlag(t *testing.T) {
	deps := happyDeps
	deps.cfgPath = "/tmp/custom.bight.yml"
	deps.cfgSource = sourceFlag

	results := runChecks(validCfg, nil, deps)
	r, found := findByPrefix(results, "config: ")
	if !found {
		t.Fatal("no config result")
	}
	if !strings.Contains(r.msg, "(from --config)") {
		t.Errorf("expected flag source hint, got %q", r.msg)
	}
}
