// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"
	"go/format"
	"strconv"
	"strings"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func genDefaults(m *manifest.Manifest) (string, []byte, error) {
	var b strings.Builder
	b.WriteString(genHeader)
	b.WriteString("package defaults\n\n")
	fmt.Fprintf(&b, "import (\n\t\"github.com/ADITYA-CODE-SOURCE/otelcconfig/generated/types\"\n)\n\n")

	fmt.Fprintf(&b, "// %s returns the default behavior configuration for the %s\n// instrumentation.\n", m.DefaultsFuncName(), m.Instrumentation.Name)
	fmt.Fprintf(&b, "func %s() types.%s {\n", m.DefaultsFuncName(), m.StructName())
	fmt.Fprintf(&b, "\treturn types.%s{\n", m.StructName())
	for _, o := range m.Options {
		lit, err := goDefaultLiteral(&o)
		if err != nil {
			return "", nil, err
		}
		fmt.Fprintf(&b, "\t\t%s: %s,\n", o.GoName(), lit)
	}
	b.WriteString("\t}\n}\n")

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", nil, fmt.Errorf("format generated defaults for %s: %w", m.Instrumentation.Name, err)
	}
	return "defaults/" + m.InstrumentationFileName() + "_gen.go", src, nil
}

func goDefaultLiteral(o *manifest.Option) (string, error) {
	v, err := o.DefaultValue()
	if err != nil {
		return "", err
	}
	switch manifest.OptionType(o.Type) {
	case manifest.TypeBoolean:
		b, ok := v.(bool)
		if !ok {
			return "", fmt.Errorf("default for %q is not a boolean", o.Name)
		}
		return strconv.FormatBool(b), nil
	case manifest.TypeListString:
		items, ok := v.([]string)
		if !ok {
			return "", fmt.Errorf("default for %q is not a string list", o.Name)
		}
		if len(items) == 0 {
			return "[]string{}", nil
		}
		quoted := make([]string, 0, len(items))
		for _, it := range items {
			quoted = append(quoted, strconv.Quote(it))
		}
		return "[]string{" + strings.Join(quoted, ", ") + "}", nil
	default:
		return "", fmt.Errorf("unsupported type %q for default", o.Type)
	}
}
