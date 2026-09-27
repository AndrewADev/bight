package strategy

import (
	"bytes"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const defaultTemplate = "{{.Project}}_{{.Branch}}"

var templateFuncs = template.FuncMap{
	"slug": slug,
}

func applyTemplate(ctx Context, tmpl string) (string, error) {
	if tmpl == "" {
		tmpl = defaultTemplate
	}
	t, err := template.New("").Funcs(templateFuncs).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// foldAccents decomposes s (NFD) and removes combining marks, turning "é"
// into "e" and "ü" into "u".
var foldAccents = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)))

// latinLetters maps lowercase Latin letters that have no Unicode
// decomposition to ASCII spellings.
var latinLetters = map[rune]string{
	'ß': "ss",
	'æ': "ae",
	'œ': "oe",
	'ø': "o",
	'ł': "l",
	'đ': "d",
	'ð': "d",
	'þ': "th",
	'ı': "i",
}

// slug lowercases s, transliterates accented Latin letters to ASCII, and
// replaces each run of characters outside [a-z0-9] with a single
// underscore, trimming leading and trailing underscores.
// "feat/Café-Straße" becomes "feat_cafe_strasse".
func slug(s string) string {
	lower := strings.ToLower(s)
	folded, _, err := transform.String(foldAccents, lower)
	if err != nil {
		folded = lower
	}
	var b strings.Builder
	pendingSep := false
	for _, r := range folded {
		var part string
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			part = string(r)
		case latinLetters[r] != "":
			part = latinLetters[r]
		default:
			pendingSep = true
			continue
		}
		if pendingSep && b.Len() > 0 {
			b.WriteByte('_')
		}
		pendingSep = false
		b.WriteString(part)
	}
	return b.String()
}
