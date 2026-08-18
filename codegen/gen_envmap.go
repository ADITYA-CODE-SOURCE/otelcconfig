// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"
	"go/format"
	"strings"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func genEnvMap(m *manifest.Manifest) (string, []byte, error) {
	var b strings.Builder
	b.WriteString(genHeader)
	b.WriteString("package envmap\n\n")
	b.WriteString("// EnvMapping describes how a single option maps to an environment variable.\n")
	b.WriteString("type EnvMapping struct {\n")
	b.WriteString("\t// OptionPath is the dot-separated declarative path of the option.\n")
	b.WriteString("\tOptionPath string\n")
	b.WriteString("\t// EnvVar is the environment variable name.\n")
	b.WriteString("\tEnvVar string\n")
	b.WriteString("\t// Semantics is the mapping semantics: direct or csv_contains.\n")
	b.WriteString("\tSemantics string\n")
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "// %sEnvMappings lists the environment-variable compatibility mappings\n// for the %s instrumentation. Only existing otelc compatibility variables\n// are used; otelcconfig-specific variables are intentionally not introduced.\n", m.InstrumentationGoName(), m.Instrumentation.Name)
	fmt.Fprintf(&b, "var %sEnvMappings = []EnvMapping{\n", m.InstrumentationGoName())
	for _, o := range m.Options {
		if o.EnvVar == "" {
			continue
		}
		fmt.Fprintf(&b, "\t{OptionPath: %q, EnvVar: %q, Semantics: %q},\n", o.DeclarativePath, o.EnvVar, o.EnvVarSemantics)
	}
	b.WriteString("}\n")

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", nil, fmt.Errorf("format generated env map for %s: %w", m.Instrumentation.Name, err)
	}
	return "envmap/" + m.InstrumentationFileName() + "_gen.go", src, nil
}
