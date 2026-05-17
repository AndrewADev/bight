package config

import (
	"fmt"
	"strings"
)

// Generate returns the contents of a fresh .bight.yml.
//
// `copySource`, if non-empty, emits a `copy:` block on the env_file so a
// freshly-created worktree seeds the dest from that path (resolved against
// the main worktree root). When empty, a commented `copy:` hint is emitted
// instead so the option is discoverable later.
func Generate(project, envFilePath, copySource string, vars []Var) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "project: %s\nenv_files:\n  - path: %s\n    # backup: true\n", project, envFilePath)

	if copySource != "" {
		fmt.Fprintf(&sb, "    copy:\n      source: %s\n      # overwrite: false   # set true to clobber an existing dest\n", copySource)
	} else {
		sb.WriteString("    # copy: .env   # seed this file from the main worktree on first checkout (path is resolved against the main worktree root)\n")
	}

	sb.WriteString("    vars:\n")
	for _, v := range vars {
		fmt.Fprintf(&sb, "      - name: %s\n        strategy: %s\n        on: checkout\n        # sensitive: true\n", v.Name, v.Strategy)
	}

	return sb.String()
}
