package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/AndrewADev/bight/internal/config"
)

var validCfg = &config.Config{
	Project: "myapp",
	EnvFiles: []config.EnvFile{
		{
			Path: ".env",
			Vars: []config.Var{
				{Name: "DB_NAME", Strategy: "template", On: "checkout"},
				{Name: "JWT_SECRET", Strategy: "random", On: "checkout"},
			},
		},
	},
}

var happyDeps = checkDeps{
	gitOK:            true,
	hookErr:          nil,
	existingEnvFiles: map[string]bool{".env": true},
}

func allOK(results []result) bool {
	for _, r := range results {
		if r.status == "fail" {
			return false
		}
	}
	return true
}

func statusOf(results []result, msg string) string {
	for _, r := range results {
		if r.msg == msg {
			return r.status
		}
	}
	return ""
}

func findByPrefix(results []result, prefix string) (result, bool) {
	for _, r := range results {
		if strings.HasPrefix(r.msg, prefix) {
			return r, true
		}
	}
	return result{}, false
}

func TestRunChecks_AllPass(t *testing.T) {
	results := runChecks(validCfg, nil, happyDeps)
	if !allOK(results) {
		t.Errorf("expected all checks to pass, got: %v", results)
	}
}

func TestRunChecks_GitRepoMissing(t *testing.T) {
	deps := happyDeps
	deps.gitOK = false
	results := runChecks(validCfg, nil, deps)

	r, found := findByPrefix(results, "git repo:")
	if !found {
		t.Fatal("expected git repo check result")
	}
	if r.status != "fail" {
		t.Errorf("expected fail, got %q", r.status)
	}
}

func TestRunChecks_ConfigLoadError(t *testing.T) {
	results := runChecks(nil, errors.New("file not found"), happyDeps)

	if s := statusOf(results, "config: failed to load .bight.yml — file not found"); s != "fail" {
		t.Errorf("expected config check to fail, got %q", s)
	}
	// Checks 5–7 skipped — only a warning about skipping, no env file or var results
	for _, r := range results {
		if r.msg == "vars: all strategies valid" || r.msg == "vars: all triggers valid" {
			t.Errorf("unexpected var check result when config failed: %v", r)
		}
	}
}

func TestRunChecks_ConfigEmptyProject(t *testing.T) {
	cfg := &config.Config{Project: "", EnvFiles: validCfg.EnvFiles}
	results := runChecks(cfg, nil, happyDeps)
	if s := statusOf(results, "config: project field is empty"); s != "fail" {
		t.Errorf("expected fail for empty project, got %q", s)
	}
}

func TestRunChecks_ConfigNoEnvFiles(t *testing.T) {
	cfg := &config.Config{Project: "myapp"}
	results := runChecks(cfg, nil, happyDeps)
	r, found := findByPrefix(results, "config: project = ")
	if !found || r.status != "fail" {
		t.Errorf("expected fail for no env files, got: %v", r)
	}
}

func TestRunChecks_HookNotInstalled(t *testing.T) {
	deps := happyDeps
	deps.hookErr = errors.New("hook not found")
	results := runChecks(validCfg, nil, deps)
	r, found := findByPrefix(results, "hook: not installed")
	if !found || r.status != "info" {
		t.Errorf("expected info for missing hook, got: %v", r)
	}
}

func TestRunChecks_EnvFileMissing(t *testing.T) {
	deps := happyDeps
	deps.existingEnvFiles = map[string]bool{}
	results := runChecks(validCfg, nil, deps)
	r, found := findByPrefix(results, "env file: .env —")
	if !found || r.status != "warn" {
		t.Errorf("expected warn for missing env file, got: %v", r)
	}
}

func TestRunChecks_UnknownStrategy(t *testing.T) {
	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{Path: ".env", Vars: []config.Var{
				{Name: "SECRET", Strategy: "nonexistent", On: "checkout"},
			}},
		},
	}
	results := runChecks(cfg, nil, happyDeps)
	r, found := findByPrefix(results, "vars: unknown strategy in:")
	if !found || r.status != "fail" {
		t.Errorf("expected fail for unknown strategy, got: %v", r)
	}
}

func TestRunChecks_UnknownTrigger(t *testing.T) {
	cfg := &config.Config{
		Project: "myapp",
		EnvFiles: []config.EnvFile{
			{Path: ".env", Vars: []config.Var{
				{Name: "DB_NAME", Strategy: "template", On: "db_create"},
			}},
		},
	}
	results := runChecks(cfg, nil, happyDeps)
	r, found := findByPrefix(results, "vars: unknown trigger in:")
	if !found || r.status != "fail" {
		t.Errorf("expected fail for unknown trigger, got: %v", r)
	}
}
