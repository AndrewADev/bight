package strategy

import (
	"bytes"
	"text/template"
)

const defaultTemplate = "{{.Project}}_{{.Branch}}"

func applyTemplate(ctx Context, tmpl string) (string, error) {
	if tmpl == "" {
		tmpl = defaultTemplate
	}
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}
