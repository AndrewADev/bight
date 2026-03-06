package config

import (
	"os"
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

	cfg, err := Load(f.Name())
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
	_, err := Load("/nonexistent/.bight.yml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
