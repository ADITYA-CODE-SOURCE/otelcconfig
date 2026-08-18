// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func genDocs(m *manifest.Manifest) (string, []byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — Configuration Catalog\n\n", m.Instrumentation.Name)
	b.WriteString("> Generated from the behavior manifest by otelcconfig. Do not edit.\n\n")
	fmt.Fprintf(&b, "%s\n\n", m.Instrumentation.Description)
	fmt.Fprintf(&b, "Library: %s\n\n", m.Instrumentation.LibraryLink)
	b.WriteString("| Option | Declarative path | Type | Default | Stability | Env var | Description |\n")
	b.WriteString("|--------|------------------|------|---------|-----------|---------|-------------|\n")
	for _, o := range m.Options {
		def := displayDefault(&o)
		env := "—"
		if o.EnvVar != "" {
			env = "`" + o.EnvVar + "`"
		}
		desc := strings.ReplaceAll(o.Description, "|", "\\|")
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | `%s` | `%s` | %s | %s |\n",
			o.Name, o.DeclarativePath, o.Type, def, o.Stability, env, desc)
	}
	return "catalog/" + m.InstrumentationFileName() + ".md", []byte(b.String()), nil
}

func displayDefault(o *manifest.Option) string {
	v, err := o.DefaultValue()
	if err != nil {
		return ""
	}
	switch manifest.OptionType(o.Type) {
	case manifest.TypeBoolean:
		b, _ := v.(bool)
		return strconv.FormatBool(b)
	case manifest.TypeListString:
		items, _ := v.([]string)
		if len(items) == 0 {
			return "[]"
		}
		quoted := make([]string, 0, len(items))
		for _, it := range items {
			quoted = append(quoted, strconv.Quote(it))
		}
		return "[" + strings.Join(quoted, ", ") + "]"
	default:
		return ""
	}
}
