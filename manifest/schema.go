// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// OptionType is the kind of value an option accepts.
type OptionType string

const (
	// TypeBoolean is a true/false option.
	TypeBoolean OptionType = "boolean"
	// TypeListString is a list of strings option.
	TypeListString OptionType = "list_string"
)

// Stability marks the lifecycle stage of an option.
type Stability string

const (
	// StabilityStable marks a stable, non-experimental option.
	StabilityStable Stability = "stable"
	// StabilityExperimental marks an experimental option. Experimental options
	// correspond to OpenTelemetry declarative configuration keys that are not
	// yet stable, or to prototype-owned go: keys that may change.
	StabilityExperimental Stability = "experimental"
)

// EnvVarSemantics describes how an environment variable value maps onto an option.
type EnvVarSemantics string

const (
	// SemanticsDirect means the environment variable holds the option value
	// directly.
	SemanticsDirect EnvVarSemantics = "direct"
	// SemanticsCSVContains means the environment variable holds a comma-separated
	// list and the option is enabled when it contains the instrumentation name.
	// This matches otelc's OTEL_GO_ENABLED_INSTRUMENTATIONS semantics.
	SemanticsCSVContains EnvVarSemantics = "csv_contains"
)

// Manifest is the single source of truth for an instrumentation's configurable
// behavior options. It follows the Java agent metadata.yaml pattern referenced
// by otelc Issue #705: one manifest per instrumentation, from which the go:
// validation schema, the catalog, and defaults are derived.
type Manifest struct {
	// ManifestVersion is the version of the manifest format itself.
	ManifestVersion string `yaml:"manifest_version"`
	// Instrumentation describes the instrumentation the manifest applies to.
	Instrumentation Instrumentation `yaml:"instrumentation"`
	// Options are the configurable behavior options of the instrumentation.
	Options []Option `yaml:"options"`
}

// Instrumentation identifies the instrumentation a manifest describes.
type Instrumentation struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	LibraryLink string `yaml:"library_link"`
}

// Option describes one configurable behavior option.
type Option struct {
	// Name is a short, stable identifier for the option (for example
	// request_captured_headers).
	Name string `yaml:"name"`
	// DeclarativePath is the dot-separated path of the option inside the
	// instrumentation/development node. Paths starting with general. must use
	// official OpenTelemetry declarative configuration keys; paths starting with
	// go. are owned by this prototype.
	DeclarativePath string `yaml:"declarative_path"`
	// Type is the option value type.
	Type string `yaml:"type"`
	// Default is the option's default value, interpreted according to Type.
	Default yaml.Node `yaml:"default"`
	// EnvVar is an optional environment variable that can override the option.
	// In this release only existing otelc compatibility variables are used;
	// otelcconfig-specific variables are intentionally not introduced.
	EnvVar string `yaml:"env_var,omitempty"`
	// EnvVarSemantics describes how EnvVar maps to the option.
	EnvVarSemantics string `yaml:"env_var_semantics,omitempty"`
	// Description documents the option for generated catalogs and schemas.
	Description string `yaml:"description"`
	// Stability marks the option lifecycle stage.
	Stability string `yaml:"stability"`
	// UpstreamRef records the official OpenTelemetry declarative configuration
	// key when DeclarativePath is a general.* cross-language key.
	UpstreamRef string `yaml:"upstream_ref,omitempty"`
}

// Load parses and validates a manifest from raw YAML bytes.
func Load(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// LoadFile loads and validates a manifest from a file path.
func LoadFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}
	m, err := Load(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return m, nil
}

// Discover loads every manifest at <dir>/<instrumentation>/metadata.yaml.
// Results are sorted by path for deterministic generation.
func Discover(dir string) ([]*Manifest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read manifest dir %s: %w", dir, err)
	}
	var paths []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(dir, e.Name(), "metadata.yaml")
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	manifests := make([]*Manifest, 0, len(paths))
	for _, p := range paths {
		m, err := LoadFile(p)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
}

// Validate checks that the manifest is structurally sound.
func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.ManifestVersion) == "" {
		return fmt.Errorf("manifest_version is required")
	}
	if strings.TrimSpace(m.Instrumentation.Name) == "" {
		return fmt.Errorf("instrumentation.name is required")
	}
	if strings.TrimSpace(m.Instrumentation.Description) == "" {
		return fmt.Errorf("instrumentation.description is required")
	}
	if len(m.Options) == 0 {
		return fmt.Errorf("instrumentation %s declares no options", m.Instrumentation.Name)
	}

	seenNames := make(map[string]bool, len(m.Options))
	seenPaths := make(map[string]bool, len(m.Options))
	for _, o := range m.Options {
		if err := o.Validate(); err != nil {
			return fmt.Errorf("instrumentation %s option %q: %w", m.Instrumentation.Name, o.Name, err)
		}
		if seenNames[o.Name] {
			return fmt.Errorf("instrumentation %s declares option %q more than once", m.Instrumentation.Name, o.Name)
		}
		if seenPaths[o.DeclarativePath] {
			return fmt.Errorf("instrumentation %s declares declarative_path %q more than once", m.Instrumentation.Name, o.DeclarativePath)
		}
		seenNames[o.Name] = true
		seenPaths[o.DeclarativePath] = true
	}
	return nil
}

// Validate checks a single option.
func (o *Option) Validate() error {
	if strings.TrimSpace(o.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(o.DeclarativePath) == "" {
		return fmt.Errorf("declarative_path is required")
	}
	if strings.TrimSpace(o.Description) == "" {
		return fmt.Errorf("description is required")
	}
	switch OptionType(o.Type) {
	case TypeBoolean, TypeListString:
	default:
		return fmt.Errorf("unsupported type %q (allowed: boolean, list_string)", o.Type)
	}
	switch Stability(o.Stability) {
	case StabilityStable, StabilityExperimental:
	default:
		return fmt.Errorf("unsupported stability %q (allowed: stable, experimental)", o.Stability)
	}
	if o.EnvVar != "" && o.EnvVarSemantics == "" {
		return fmt.Errorf("env_var_semantics is required when env_var is set")
	}
	if o.EnvVar != "" {
		switch EnvVarSemantics(o.EnvVarSemantics) {
		case SemanticsDirect, SemanticsCSVContains:
		default:
			return fmt.Errorf("unsupported env_var_semantics %q (allowed: direct, csv_contains)", o.EnvVarSemantics)
		}
	}
	if _, err := o.DefaultValue(); err != nil {
		return err
	}
	return nil
}

// DefaultValue decodes the option default into a typed Go value.
func (o *Option) DefaultValue() (any, error) {
	switch OptionType(o.Type) {
	case TypeBoolean:
		var v bool
		if err := o.Default.Decode(&v); err != nil {
			return nil, fmt.Errorf("default for boolean option %q: %w", o.Name, err)
		}
		return v, nil
	case TypeListString:
		var v []string
		if err := o.Default.Decode(&v); err != nil {
			return nil, fmt.Errorf("default for list_string option %q: %w", o.Name, err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type %q for default", o.Type)
	}
}

// GoName returns the exported Go identifier for the option.
func (o *Option) GoName() string {
	return goIdent(o.Name)
}

// InstrumentationGoName returns the exported Go identifier for the
// instrumentation, suitable as the prefix of a generated config type.
func (m *Manifest) InstrumentationGoName() string {
	parts := strings.Split(m.Instrumentation.Name, ".")
	for i, p := range parts {
		parts[i] = goIdent(p)
	}
	return strings.Join(parts, "")
}

// InstrumentationFileName returns a snake_case identifier for the
// instrumentation, suitable for generated file names.
func (m *Manifest) InstrumentationFileName() string {
	return fileIdent(m.Instrumentation.Name)
}

// StructName returns the generated config struct name for the instrumentation.
func (m *Manifest) StructName() string {
	return m.InstrumentationGoName() + "Config"
}

// DefaultsFuncName returns the generated defaults constructor name.
func (m *Manifest) DefaultsFuncName() string {
	return "Default" + m.InstrumentationGoName() + "Config"
}

func goIdent(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == '/'
	})
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if acr, ok := acronyms[strings.ToLower(p)]; ok {
			b.WriteString(acr)
			continue
		}
		p = strings.ToLower(p)
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

func fileIdent(s string) string {
	return strings.NewReplacer(".", "_", "-", "_", "/", "_").Replace(s)
}

var acronyms = map[string]string{
	"http":     "HTTP",
	"https":    "HTTPS",
	"nethttp":  "NetHTTP",
	"sql":      "SQL",
	"url":      "URL",
	"grpc":     "GRPC",
	"id":       "ID",
	"api":      "API",
	"db":       "DB",
	"dns":      "DNS",
	"tls":      "TLS",
	"json":     "JSON",
	"ip":       "IP",
	"dsn":      "DSN",
	"kafka":    "Kafka",
	"redis":    "Redis",
	"mongodb":  "MongoDB",
	"genai":    "GenAI",
	"otelc":    "OTelC",
	"semconv":  "SemConv",
	"config":   "Config",
	"metadata": "Metadata",
}
