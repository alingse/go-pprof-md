package generator

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"

	"github.com/alingse/go-pprof-md/internal/parser"
)

// DiffGenerator generates markdown comparing two profiles
type DiffGenerator struct {
	baseProfile *parser.Profile
	newProfile  *parser.Profile
	topN        int
}

// NewDiffGenerator creates a new diff generator
func NewDiffGenerator(base, new *parser.Profile, opts ...DiffOption) *DiffGenerator {
	g := &DiffGenerator{
		baseProfile: base,
		newProfile:  new,
		topN:        20,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// DiffOption configures a DiffGenerator
type DiffOption func(*DiffGenerator)

// WithDiffTopN sets the number of top changed functions to display
func WithDiffTopN(n int) DiffOption {
	return func(g *DiffGenerator) {
		g.topN = n
	}
}

// FunctionDiff represents the difference for a single function
type FunctionDiff struct {
	Name          string
	File          string
	Line          int
	BaseFlat      int64
	BaseCum       int64
	NewFlat       int64
	NewCum        int64
	FlatDelta     int64
	FlatDeltaPct  float64
	CumDelta      int64
	CumDeltaPct   float64
	IsNew         bool
	IsRemoved     bool
	IsImproved    bool // flat/cum decreased
	IsRegressed   bool // flat/cum increased
}

// Generate generates diff markdown
func (g *DiffGenerator) Generate() (string, error) {
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
func (g *DiffGenerator) prepareTemplateData() map[string]interface{} {
	diffs := g.computeDiff()

	// Sort by absolute cumulative change, then by name
	sort.Slice(diffs, func(i, j int) bool {
		absI := abs(diffs[i].CumDelta)
		absJ := abs(diffs[j].CumDelta)
		if absI != absJ {
			return absI > absJ
		}
		return diffs[i].Name < diffs[j].Name
	})

	// Limit to top N
	if len(diffs) > g.topN {
		diffs = diffs[:g.topN]
	}

	return map[string]interface{}{
		"Type":         string(g.baseProfile.Type),
		"BaseStats":    g.baseProfile.Stats,
		"NewStats":     g.newProfile.Stats,
		"BaseTotal":    g.baseProfile.TotalSamples,
		"NewTotal":     g.newProfile.TotalSamples,
		"Diffs":        diffs,
		"TotalDelta":   g.newProfile.TotalSamples - g.baseProfile.TotalSamples,
		"FormatBytes":  FormatBytes,
		"FormatNumber": FormatNumber,
		"FormatDelta":  FormatDelta,
	}
}

// computeDiff computes the differences between the two profiles
func (g *DiffGenerator) computeDiff() []FunctionDiff {
	// Build map of base functions
	baseFuncs := make(map[string]*parser.Function)
	for i := range g.baseProfile.Functions {
		fn := &g.baseProfile.Functions[i]
		baseFuncs[fn.Name] = fn
	}

	// Build map of new functions
	newFuncs := make(map[string]*parser.Function)
	for i := range g.newProfile.Functions {
		fn := &g.newProfile.Functions[i]
		newFuncs[fn.Name] = fn
	}

	// Collect all function names
	allNames := make(map[string]bool)
	for name := range baseFuncs {
		allNames[name] = true
	}
	for name := range newFuncs {
		allNames[name] = true
	}

	var diffs []FunctionDiff

	for name := range allNames {
		baseFn, hasBase := baseFuncs[name]
		newFn, hasNew := newFuncs[name]

		diff := FunctionDiff{
			Name: name,
		}

		if !hasBase {
			// New function
			diff.IsNew = true
			diff.File = newFn.File
			diff.Line = newFn.Line
			diff.NewFlat = newFn.Flat
			diff.NewCum = newFn.Cum
			diff.FlatDelta = newFn.Flat
			diff.CumDelta = newFn.Cum
			diff.IsRegressed = newFn.Flat > 0 || newFn.Cum > 0
		} else if !hasNew {
			// Removed function
			diff.IsRemoved = true
			diff.File = baseFn.File
			diff.Line = baseFn.Line
			diff.BaseFlat = baseFn.Flat
			diff.BaseCum = baseFn.Cum
			diff.FlatDelta = -baseFn.Flat
			diff.CumDelta = -baseFn.Cum
			diff.IsImproved = true
		} else {
			// Changed function
			diff.File = baseFn.File
			diff.Line = baseFn.Line
			diff.BaseFlat = baseFn.Flat
			diff.BaseCum = baseFn.Cum
			diff.NewFlat = newFn.Flat
			diff.NewCum = newFn.Cum
			diff.FlatDelta = newFn.Flat - baseFn.Flat
			diff.CumDelta = newFn.Cum - baseFn.Cum
			diff.IsImproved = diff.FlatDelta < 0 || diff.CumDelta < 0
			diff.IsRegressed = diff.FlatDelta > 0 || diff.CumDelta > 0
		}

		// Calculate percentage changes
		if diff.BaseFlat != 0 {
			diff.FlatDeltaPct = float64(diff.FlatDelta) / float64(diff.BaseFlat) * 100
		} else if diff.NewFlat != 0 {
			diff.FlatDeltaPct = 100.0 // new function
		}

		if diff.BaseCum != 0 {
			diff.CumDeltaPct = float64(diff.CumDelta) / float64(diff.BaseCum) * 100
		} else if diff.NewCum != 0 {
			diff.CumDeltaPct = 100.0 // new function
		}

		diffs = append(diffs, diff)
	}

	return diffs
}

// getTemplate returns the diff template
func (g *DiffGenerator) getTemplate() (*template.Template, error) {
	tmpl := `# {{ .Type }} Profile Diff: Base vs New

## Summary

| Metric | Base | New | Delta |
|--------|------|------|-------|
{{- if eq .Type "cpu" }}
| Total CPU Time | {{ formatDuration .BaseTotal }} | {{ formatDuration .NewTotal }} | {{ formatDuration .TotalDelta }} |
| Duration | {{ formatDuration .BaseStats.TotalDuration.Nanoseconds }} | {{ formatDuration .NewStats.TotalDuration.Nanoseconds }} | - |
{{- else if eq .Type "heap" }}
| Allocated Bytes | {{ FormatBytes .BaseStats.AllocBytes }} | {{ FormatBytes .NewStats.AllocBytes }} | {{ FormatDelta (subtract .NewStats.AllocBytes .BaseStats.AllocBytes) }} |
| Allocated Objects | {{ FormatNumber .BaseStats.AllocObjects }} | {{ FormatNumber .NewStats.AllocObjects }} | {{ FormatDelta (subtract .NewStats.AllocObjects .BaseStats.AllocObjects) }} |
| In-Use Bytes | {{ FormatBytes .BaseStats.InUseBytes }} | {{ FormatBytes .NewStats.InUseBytes }} | {{ FormatDelta (subtract .NewStats.InUseBytes .BaseStats.InUseBytes) }} |
| In-Use Objects | {{ FormatNumber .BaseStats.InUseObjects }} | {{ FormatNumber .NewStats.InUseObjects }} | {{ FormatDelta (subtract .NewStats.InUseObjects .BaseStats.InUseObjects) }} |
{{- else if eq .Type "goroutine" }}
| Total Goroutines | {{ FormatNumber .BaseStats.TotalGoroutines }} | {{ FormatNumber .NewStats.TotalGoroutines }} | {{ FormatDelta (subtract .NewStats.TotalGoroutines .BaseStats.TotalGoroutines) }} |
{{- else if eq .Type "mutex" }}
| Contention Time | {{ formatDuration .BaseStats.TotalContentionTime }} | {{ formatDuration .NewStats.TotalContentionTime }} | {{ formatDuration (subtract .NewStats.TotalContentionTime .BaseStats.TotalContentionTime) }} |
| Total Waits | {{ FormatNumber .BaseStats.TotalWaits }} | {{ FormatNumber .NewStats.TotalWaits }} | {{ FormatDelta (subtract .NewStats.TotalWaits .BaseStats.TotalWaits) }} |
{{- end }}

## Top Changed Functions

| Rank | Function | Base | New | Flat Δ | Flat Δ% | Cum Δ | Cum Δ% |
|------|----------|------|------|---------|---------|-------|--------|
{{- range $i, $d := .Diffs }}
| {{ add $i 1 }} | ` + "`" + `{{ $d.Name }}` + "`" + ` | {{ template "base-val" $d }} | {{ template "new-val" $d }} | {{ FormatDelta $d.FlatDelta }} | {{ printf "%+.1f" $d.FlatDeltaPct }}% | {{ FormatDelta $d.CumDelta }} | {{ printf "%+.1f" $d.CumDeltaPct }}% |
{{- end }}

{{- define "base-val" }}
{{- if .IsRemoved }}{{ .BaseCum }}{{ else if .IsNew }}-{{ else }}{{ .BaseCum }}{{ end }}
{{- end }}

{{- define "new-val" }}
{{- if .IsRemoved }}-{{ else if .IsNew }}{{ .NewCum }}{{ else }}{{ .NewCum }}{{ end }}
{{- end }}

---

## AI Analysis Request

Please analyze this pprof diff and provide:

1. **Overall Impact**: How did the performance change between base and new? Is this an improvement or regression?

2. **Key Changes**:
   - Which functions showed the most significant increases (potential regressions)?
   - Which functions showed the most significant decreases (improvements)?
   - Are there any new functions that appeared in the new profile?

3. **Root Cause Analysis**:
   - For the biggest regressions, what might have caused the increase?
   - For the biggest improvements, what optimizations were effective?

4. **Recommendations**:
   - What specific code changes should be investigated further?
   - Are there any unexpected changes that need attention?
   - What should be the priority for fixing regressions?

Focus on actionable insights to understand the performance change.
`

	funcs := template.FuncMap{
		"add":             func(a, b int) int { return a + b },
		"subtract":        func(a, b int64) int64 { return a - b },
		"abs":             abs,
		"FormatBytes":     FormatBytes,
		"FormatNumber":    FormatNumber,
		"FormatDelta":     FormatDelta,
		"formatDuration":  FormatDuration,
	}

	return template.New("diff").Funcs(funcs).Parse(tmpl)
}

// FormatDelta formats a delta value with + or - sign and formatting
func FormatDelta(delta int64) string {
	if delta == 0 {
		return "0"
	}
	sign := ""
	if delta > 0 {
		sign = "+"
	}

	absDelta := abs(delta)
	if absDelta < 1000 {
		return fmt.Sprintf("%s%d", sign, delta)
	}
	if absDelta < 1_000_000 {
		return fmt.Sprintf("%s%.1fK", sign, float64(delta)/1000)
	}
	return fmt.Sprintf("%s%.1fM", sign, float64(delta)/1_000_000)
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
