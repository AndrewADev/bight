package config

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	out := Generate("myapp", ".env.local", []Var{
		{Name: "DB_NAME", Strategy: "template"},
		{Name: "JWT_SECRET", Strategy: "random"},
	})

	checks := []string{
		"project: myapp",
		"path: .env.local",
		"name: DB_NAME",
		"strategy: template",
		"on: checkout",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("Generate() missing %q in output:\n%s", want, out)
		}
	}
}
