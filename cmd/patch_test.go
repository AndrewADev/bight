package cmd

import (
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
