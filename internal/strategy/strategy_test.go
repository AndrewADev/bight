package strategy

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/AndrewADev/bight/internal/config"
)

var testCfg = &config.Config{
	Project: "myapp",
	Defaults: config.Defaults{
		BranchTemplate: "",
	},
}

func TestApplyTemplate_Default(t *testing.T) {
	ctx := Context{Branch: "feat-login", Project: "myapp"}
	val, err := Apply("template", ctx, testCfg)
	if err != nil {
		t.Fatal(err)
	}
	if val != "myapp_feat-login" {
		t.Errorf("got %q, want %q", val, "myapp_feat-login")
	}
}

func TestApplyTemplate_Custom(t *testing.T) {
	cfg := &config.Config{
		Project:  "myapp",
		Defaults: config.Defaults{BranchTemplate: "{{.Branch}}_db"},
	}
	ctx := Context{Branch: "main", Project: "myapp"}
	val, err := Apply("template", ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if val != "main_db" {
		t.Errorf("got %q, want %q", val, "main_db")
	}
}

func TestApplyRandom(t *testing.T) {
	ctx := Context{Branch: "main", Project: "myapp"}
	val, err := Apply("random", ctx, testCfg)
	if err != nil {
		t.Fatal(err)
	}
	// Should be 64 hex chars (32 bytes)
	if len(val) != 64 {
		t.Errorf("len = %d, want 64", len(val))
	}
	if _, err := hex.DecodeString(val); err != nil {
		t.Errorf("not valid hex: %v", err)
	}
	// Two calls should differ
	val2, _ := Apply("random", ctx, testCfg)
	if val == val2 {
		t.Error("expected different values for two random calls")
	}
}

func TestApplyUnknownStrategy(t *testing.T) {
	ctx := Context{Branch: "main", Project: "myapp"}
	_, err := Apply("nonexistent", ctx, testCfg)
	if err == nil || !strings.Contains(err.Error(), "unknown strategy") {
		t.Errorf("expected unknown strategy error, got %v", err)
	}
}
