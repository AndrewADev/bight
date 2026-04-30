package config

import (
	"fmt"
	"strings"
)

func Generate(project, envFilePath string, vars []Var) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "project: %s\nenv_files:\n  - path: %s\n    # backup: true\n    vars:\n", project, envFilePath)
	for _, v := range vars {
		fmt.Fprintf(&sb, "      - name: %s\n        strategy: %s\n        on: checkout\n        # sensitive: true\n", v.Name, v.Strategy)
	}
	return sb.String()
}
