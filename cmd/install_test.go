package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestPromptInitConfig_NonTTYBailsOut(t *testing.T) {
	// Isolate from any real ~/.bight.yml and from the worktree's .bight.yml
	// so config.Load() returns os.ErrNotExist.
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	withStubPrompt(t, false, "")

	if err := promptInitConfig(); err != nil {
		t.Fatalf("promptInitConfig: %v", err)
	}

	if _, err := os.Stat(".bight.yml"); !os.IsNotExist(err) {
		t.Errorf("expected no .bight.yml to be written under non-TTY, stat err=%v", err)
	}
}

func TestPromptInitConfig_TTYCreatesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	// Accept "Create .bight.yml?" (Y), accept default project, default env
	// file, decline seed, decline add vars.
	withStubPrompt(t, true, strings.Join([]string{
		"y", // Create .bight.yml?
		"",  // Project name (default)
		"",  // Env file path (default)
		"n", // Seed from another path?
		"n", // Add env vars?
		"",
	}, "\n"))

	if err := promptInitConfig(); err != nil {
		t.Fatalf("promptInitConfig: %v", err)
	}

	if _, err := os.Stat(".bight.yml"); err != nil {
		t.Errorf("expected .bight.yml to be written under TTY: %v", err)
	}
}
