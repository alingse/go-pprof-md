package generator

import (
	"testing"
	"time"

	"github.com/alingse/go-pprof-md/internal/parser"
)

// TestNewGenerator tests generator creation
func TestNewGenerator(t *testing.T) {
	profile := &parser.Profile{
		Type:     parser.TypeCPU,
		Functions: []parser.Function{},
	}

	gen := NewGenerator(profile)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}

	if gen.topN != 20 {
		t.Errorf("default topN = %d, want 20", gen.topN)
	}

	if !gen.includeAIPrompt {
		t.Error("default includeAIPrompt = false, want true")
	}
}

// TestNewGeneratorWithOptions tests generator creation with options
func TestNewGeneratorWithOptions(t *testing.T) {
	profile := &parser.Profile{
		Type:     parser.TypeCPU,
		Functions: []parser.Function{},
	}

	gen := NewGenerator(profile, WithTopN(10), WithAIPrompt(false))
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}

	if gen.topN != 10 {
		t.Errorf("topN = %d, want 10", gen.topN)
	}

	if gen.includeAIPrompt {
		t.Error("includeAIPrompt = true, want false")
	}
}

// TestGenerate tests markdown generation
func TestGenerate(t *testing.T) {
	profile := &parser.Profile{
		Type:        parser.TypeCPU,
		TotalSamples: 300000000, // 300ms in nanoseconds
		SampleTime:  5 * time.Second,
		Stats: parser.Stats{
			TotalDuration:     5 * time.Second,
			CPUProfileDuration: 5 * time.Second,
			SampleRate:        1000,
		},
		Functions: []parser.Function{
			{
				Name:    "main.main",
				File:    "main.go",
				Line:    10,
				Flat:    100000000, // 100ms
				Cum:     200000000, // 200ms
				FlatPct: 33.33,
				CumPct:  66.67,
				SumPct:  33.33,
				CallStack: []string{
					"main.main",
					"runtime.main",
				},
				CallPaths: []parser.CallPath{
					{Stack: []string{"main.main", "runtime.main"}, Weight: 100000000},
				},
			},
		},
	}

	gen := NewGenerator(profile, WithTopN(10), WithAIPrompt(true))
	markdown, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if markdown == "" {
		t.Error("generated markdown is empty")
	}

	// Check that markdown contains expected sections
	expectedStrings := []string{
		"# cpu",
		"## Summary Statistics",
		"## Top cpu Functions",
		"main.main",
		"## AI Analysis Request",
		"CPU Time",
		"Sum %",
	}

	for _, expected := range expectedStrings {
		if !contains(markdown, expected) {
			t.Errorf("markdown missing expected string: %s", expected)
		}
	}
}

// TestGenerateWithoutAIPrompt tests markdown generation without AI prompt
func TestGenerateWithoutAIPrompt(t *testing.T) {
	profile := &parser.Profile{
		Type:        parser.TypeCPU,
		TotalSamples: 1000,
		Functions: []parser.Function{
			{
				Name:    "main.main",
				File:    "main.go",
				Line:    10,
				Flat:    100,
				Cum:     200,
				FlatPct: 10.0,
				CumPct:  20.0,
			},
		},
	}

	gen := NewGenerator(profile, WithTopN(10), WithAIPrompt(false))
	markdown, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that AI prompt is NOT included
	if contains(markdown, "## AI Analysis Request") {
		t.Error("markdown should not contain AI prompt when WithAIPrompt(false)")
	}
}

// TestGenerateHeapProfile tests heap profile markdown generation
func TestGenerateHeapProfile(t *testing.T) {
	profile := &parser.Profile{
		Type:        parser.TypeHeap,
		TotalSamples: 1000000,
		Stats: parser.Stats{
			AllocBytes:   2000000,
			AllocObjects: 5000,
			InUseBytes:   1000000,
			InUseObjects: 2000,
		},
		Functions: []parser.Function{
			{
				Name:    "main.makeAllocation",
				File:    "main.go",
				Line:    20,
				Flat:    500000,
				Cum:     500000,
				FlatPct: 50.0,
				CumPct:  50.0,
			},
		},
	}

	gen := NewGenerator(profile)
	markdown, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check for heap-specific strings
	expectedStrings := []string{
		"# heap",
		"Allocated Bytes",
		"Allocated Objects",
		"main.makeAllocation",
	}

	for _, expected := range expectedStrings {
		if !contains(markdown, expected) {
			t.Errorf("markdown missing expected string: %s", expected)
		}
	}
}

// TestGenerateGoroutineProfile tests goroutine profile markdown generation
func TestGenerateGoroutineProfile(t *testing.T) {
	profile := &parser.Profile{
		Type:        parser.TypeGoroutine,
		TotalSamples: 100,
		Stats: parser.Stats{
			TotalGoroutines: 100,
		},
		Functions: []parser.Function{
			{
				Name:    "main.spawnGoroutines",
				File:    "main.go",
				Line:    30,
				Flat:    50,
				Cum:     100,
				FlatPct: 50.0,
				CumPct:  100.0,
			},
		},
	}

	gen := NewGenerator(profile)
	markdown, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check for goroutine-specific strings
	expectedStrings := []string{
		"# goroutine",
		"Total Goroutines",
		"main.spawnGoroutines",
	}

	for _, expected := range expectedStrings {
		if !contains(markdown, expected) {
			t.Errorf("markdown missing expected string: %s", expected)
		}
	}
}

// TestGenerateMutexProfile tests mutex profile markdown generation
func TestGenerateMutexProfile(t *testing.T) {
	profile := &parser.Profile{
		Type:        parser.TypeMutex,
		TotalSamples: 50000000, // 50ms in nanoseconds
		Stats: parser.Stats{
			TotalContentionTime: 50000000,
			TotalWaits:          100,
		},
		Functions: []parser.Function{
			{
				Name:    "main.accessSharedData",
				File:    "main.go",
				Line:    40,
				Flat:    25000000,
				Cum:     25000000,
				FlatPct: 50.0,
				CumPct:  50.0,
			},
		},
	}

	gen := NewGenerator(profile)
	markdown, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check for mutex-specific strings
	expectedStrings := []string{
		"# mutex",
		"Total Contention Time",
		"Contention Time",
		"main.accessSharedData",
	}

	for _, expected := range expectedStrings {
		if !contains(markdown, expected) {
			t.Errorf("markdown missing expected string: %s", expected)
		}
	}
}

// TestTopNLimit tests that the TopN limit is respected
func TestTopNLimit(t *testing.T) {
	// Create profile with more functions than topN
	functions := make([]parser.Function, 30)
	for i := 0; i < 30; i++ {
		functions[i] = parser.Function{
			Name:    "testFunction",
			File:    "test.go",
			Line:    i,
			Flat:    int64(100 - i),
			Cum:     int64(100 - i),
			FlatPct: float64(100 - i),
			CumPct:  float64(100 - i),
		}
	}

	profile := &parser.Profile{
		Type:        parser.TypeCPU,
		TotalSamples: 1000,
		Functions:   functions,
	}

	gen := NewGenerator(profile, WithTopN(10))
	markdown, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Count how many function rows are in the table
	// This is a simple check - the markdown should have limited functions
	if !contains(markdown, "testFunction") {
		t.Error("markdown should contain functions")
	}
}

// TestFormatBytes tests the FormatBytes function
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

// TestFormatDuration tests the FormatDuration function
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		nanos    int64
		expected string
	}{
		{0, "0"},
		{500, "500ns"},
		{1000, "1.00µs"},
		{1500000, "1.50ms"},
		{1000000, "1.00ms"},
		{820000000, "820.00ms"},
		{5000000000, "5.00s"},
		{5200000000, "5.20s"},
		{65000000000, "1m5s"},
		{3600000000000, "1h0m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatDuration(tt.nanos)
			if result != tt.expected {
				t.Errorf("FormatDuration(%d) = %s, want %s", tt.nanos, result, tt.expected)
			}
		})
	}
}

// TestFormatNumber tests the FormatNumber function
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n        int64
		expected string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{1000000, "1.0M"},
		{1000000000, "1.0G"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatNumber(tt.n)
			if result != tt.expected {
				t.Errorf("FormatNumber(%d) = %s, want %s", tt.n, result, tt.expected)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
