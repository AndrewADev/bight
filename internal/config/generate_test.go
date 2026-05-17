package config

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	out := Generate("myapp", ".env.local", "", []Var{
		{Name: "DB_NAME", Strategy: "template"},
		{Name: "JWT_SECRET", Strategy: "random"},
	})

	checks := []string{
		"project: myapp",
		"path: .env.local",
		"# backup: true",
		"name: DB_NAME",
		"strategy: template",
		"on: checkout",
		"# sensitive: true",
		// When no copy source is given, the commented hint should appear.
		"# copy: ",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("Generate() missing %q in output:\n%s", want, out)
		}
	}

	// Without an explicit source, the active copy block must NOT be present.
	if strings.Contains(out, "copy:\n      source:") {
		t.Errorf("Generate() emitted an active copy block when copySource was empty:\n%s", out)
	}
}

func TestGenerate_WithCopySource(t *testing.T) {
	out := Generate("myapp", ".env", "../main/.env", []Var{
		{Name: "DB_NAME", Strategy: "template"},
	})

	checks := []string{
		"copy:",
		"source: ../main/.env",
		"# overwrite: false",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("Generate() missing %q in output:\n%s", want, out)
		}
	}

	// The commented hint version should NOT also be emitted when an active
	// copy block is generated.
	if strings.Contains(out, "# copy: .env") {
		t.Errorf("Generate() emitted both an active copy block and the commented hint:\n%s", out)
	}
}
