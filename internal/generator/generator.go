package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/alingse/go-pprof-md/internal/parser"
)

// Generator generates markdown from pprof profiles
type Generator struct {
	profile        *parser.Profile
	topN           int
	includeAIPrompt bool
}

// NewGenerator creates a new markdown generator
func NewGenerator(profile *parser.Profile, opts ...Option) *Generator {
	g := &Generator{
		profile:        profile,
		topN:           20,
		includeAIPrompt: true,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// Option configures a Generator
type Option func(*Generator)

// WithTopN sets the number of top functions to display
func WithTopN(n int) Option {
	return func(g *Generator) {
		g.topN = n
	}
}

// WithAIPrompt enables or disables AI prompt inclusion
func WithAIPrompt(enable bool) Option {
	return func(g *Generator) {
		g.includeAIPrompt = enable
	}
}

// Generate generates markdown from the profile
func (g *Generator) Generate() (string, error) {
	tmpl, err := g.getTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to get template: %w", err)
	}

	data := g.prepareTemplateData()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// prepareTemplateData prepares data for template rendering
func (g *Generator) prepareTemplateData() map[string]interface{} {
	// Limit functions to top N
	functions := g.profile.Functions
	if len(functions) > g.topN {
		functions = functions[:g.topN]
	}

	return map[string]interface{}{
		"Type":             string(g.profile.Type),
		"Stats":            g.profile.Stats,
		"Functions":        functions,
		"TotalSamples":     g.profile.TotalSamples,
		"IncludeAIPrompt":  g.includeAIPrompt,
		"AIAnalysisPrompt": g.getAIAnalysisPrompt(),
	}
}

// getTemplate returns the appropriate template based on profile type
func (g *Generator) getTemplate() (*template.Template, error) {
	// Build full template with main template + stats template
	mainTmpl := g.getMainTemplateContent()
	statsTmpl := getStatsTemplate(g.profile.Type)
	fullTmpl := statsTmpl + "\n" + mainTmpl

	// Create template with functions
	tmpl := template.New("pprof").Funcs(getTemplateFuncs())
	return tmpl.Parse(fullTmpl)
}

// getMainTemplateContent returns the main template content
func (g *Generator) getMainTemplateContent() string {
	return `# {{ .Type }} Profile Analysis

## Summary Statistics

{{- template "stats" . }}

## Top {{ .Type }} Functions

| Rank | Function | File | {{ template "metric-header" .Type }} | % of Total | Sum % | Cumulative | Cumulative % |
|------|----------|------|-------|------------|-------|------------|--------------|
{{- range $i, $fn := .Functions }}
| {{ add $i 1 }} | ` + "`" + `{{ $fn.Name }}` + "`" + ` | {{ $fn.File }}:{{ $fn.Line }} | {{ template "metric-value" $fn }} | {{ printf "%.2f" $fn.FlatPct }}% | {{ printf "%.2f" $fn.SumPct }}% | {{ template "metric-cum" $fn }} | {{ printf "%.2f" $fn.CumPct }}% |
{{- end }}

{{- range $fn := .Functions }}
{{- if ne (len $fn.CallPaths) 0 }}

### {{ $fn.Name }}

{{- range $pi, $path := $fn.CallPaths }}

**Call Path #{{ add $pi 1 }}** (weight: {{ $path.Weight }})
{{- range $i, $call := $path.Stack }}
{{- if eq $i 0 }}
  → {{ $call }}
{{- else }}
    {{ $call }}
{{- end }}
{{- end }}

{{- end }}

{{- end }}
{{- end }}

{{- if .IncludeAIPrompt }}

{{ .AIAnalysisPrompt }}
{{- end }}
`
}

// getTemplateFuncs returns template functions
func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatBytes": FormatBytes,
		"formatDuration": FormatDuration,
		"formatNumber": FormatNumber,
	}
}

// getAIAnalysisPrompt returns AI-friendly analysis prompt
func (g *Generator) getAIAnalysisPrompt() string {
	return getAIPrompt(g.profile.Type)
}

// getStatsTemplate returns template for stats section based on profile type
func getStatsTemplate(profileType parser.ProfileType) string {
	switch profileType {
	case parser.TypeCPU:
		return `
{{- define "stats" }}
- **Profile Duration:** {{ formatDuration .Stats.TotalDuration.Nanoseconds }}
- **Total CPU Time:** {{ formatDuration .TotalSamples }}
- **Sample Rate:** {{ .Stats.SampleRate }} Hz
{{- end }}

{{- define "metric-header" }}CPU Time{{ end }}
{{- define "metric-value" }}{{ formatDuration .Flat }}{{ end }}
{{- define "metric-cum" }}{{ formatDuration .Cum }}{{ end }}
`

	case parser.TypeHeap:
		return `
{{- define "stats" }}
- **Allocated Objects:** {{ formatNumber .Stats.AllocObjects }}
- **Allocated Bytes:** {{ formatBytes .Stats.AllocBytes }}
- **In-Use Objects:** {{ formatNumber .Stats.InUseObjects }}
- **In-Use Bytes:** {{ formatBytes .Stats.InUseBytes }}
{{- end }}

{{- define "metric-header" }}Allocated Bytes{{ end }}
{{- define "metric-value" }}{{ formatBytes .Flat }}{{ end }}
{{- define "metric-cum" }}{{ formatBytes .Cum }}{{ end }}
`

	case parser.TypeGoroutine:
		return `
{{- define "stats" }}
- **Total Goroutines:** {{ formatNumber .Stats.TotalGoroutines }}
- **Current Goroutines:** {{ .TotalSamples }}
{{- end }}

{{- define "metric-header" }}Goroutines{{ end }}
{{- define "metric-value" }}{{ .Flat }}{{ end }}
{{- define "metric-cum" }}{{ .Cum }}{{ end }}
`

	case parser.TypeMutex:
		return `
{{- define "stats" }}
- **Total Contention Time:** {{ formatDuration .Stats.TotalContentionTime }}
- **Total Waits:** {{ formatNumber .Stats.TotalWaits }}
{{- end }}

{{- define "metric-header" }}Contention Time{{ end }}
{{- define "metric-value" }}{{ formatDuration .Flat }}{{ end }}
{{- define "metric-cum" }}{{ formatDuration .Cum }}{{ end }}
`

	default:
		return `
{{- define "stats" }}
- **Total Samples:** {{ .TotalSamples }}
{{- end }}

{{- define "metric-header" }}Value{{ end }}
{{- define "metric-value" }}{{ .Flat }}{{ end }}
{{- define "metric-cum" }}{{ .Cum }}{{ end }}
`
	}
}
