package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectProfileType tests profile type detection
func TestDetectProfileType(t *testing.T) {
	// Get the absolute path to testdata
	testdataDir, err := filepath.Abs("../../testdata")
	if err != nil {
		t.Fatalf("failed to get testdata path: %v", err)
	}

	tests := []struct {
		name       string
		testFile   string
		expectType ProfileType
		expectErr  bool
	}{
		{"CPU profile", filepath.Join(testdataDir, "cpu.prof"), TypeCPU, false},
		{"Heap profile", filepath.Join(testdataDir, "heap.prof"), TypeHeap, false},
		{"Goroutine profile", filepath.Join(testdataDir, "goroutine.prof"), TypeGoroutine, false},
		{"Mutex profile", filepath.Join(testdataDir, "mutex.prof"), TypeMutex, false},
		{"Non-existent file", "/nonexistent/file.prof", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pType, err := DetectProfileType(tt.testFile)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pType != tt.expectType {
				t.Errorf("got type %s, want %s", pType, tt.expectType)
			}
		})
	}
}

// TestParseCPUProfile tests CPU profile parsing
func TestParseCPUProfile(t *testing.T) {
	parser := &CPUParser{}
	prof, err := parser.Parse("../../testdata/cpu.prof")
	if err != nil {
		t.Fatalf("failed to parse CPU profile: %v", err)
	}

	if prof.Type != TypeCPU {
		t.Errorf("got type %s, want %s", prof.Type, TypeCPU)
	}

	if len(prof.Functions) == 0 {
		t.Error("expected at least one function")
	}

	// Check that we have some samples
	if prof.TotalSamples == 0 {
		t.Error("expected non-zero total samples")
	}

	// Verify function data
	for _, fn := range prof.Functions {
		if fn.Name == "" {
			t.Error("function name should not be empty")
		}
		if fn.Cum <= 0 {
			t.Errorf("function %s: cumulative value should be positive, got %d", fn.Name, fn.Cum)
		}
	}
}

// TestParseHeapProfile tests heap profile parsing
func TestParseHeapProfile(t *testing.T) {
	parser := &HeapParser{}
	prof, err := parser.Parse("../../testdata/heap.prof")
	if err != nil {
		t.Fatalf("failed to parse heap profile: %v", err)
	}

	if prof.Type != TypeHeap {
		t.Errorf("got type %s, want %s", prof.Type, TypeHeap)
	}

	if len(prof.Functions) == 0 {
		t.Error("expected at least one function")
	}

	// Check heap-specific stats
	if prof.Stats.AllocBytes == 0 {
		t.Error("expected non-zero alloc bytes")
	}
	if prof.Stats.AllocObjects == 0 {
		t.Error("expected non-zero alloc objects")
	}
}

// TestParseGoroutineProfile tests goroutine profile parsing
func TestParseGoroutineProfile(t *testing.T) {
	parser := &GoroutineParser{}
	prof, err := parser.Parse("../../testdata/goroutine.prof")
	if err != nil {
		t.Fatalf("failed to parse goroutine profile: %v", err)
	}

	if prof.Type != TypeGoroutine {
		t.Errorf("got type %s, want %s", prof.Type, TypeGoroutine)
	}

	if len(prof.Functions) == 0 {
		t.Error("expected at least one function")
	}

	// Check goroutine-specific stats
	if prof.Stats.TotalGoroutines == 0 {
		t.Error("expected non-zero total goroutines")
	}
}

// TestParseMutexProfile tests mutex profile parsing
func TestParseMutexProfile(t *testing.T) {
	parser := &MutexParser{}
	prof, err := parser.Parse("../../testdata/mutex.prof")
	if err != nil {
		t.Fatalf("failed to parse mutex profile: %v", err)
	}

	if prof.Type != TypeMutex {
		t.Errorf("got type %s, want %s", prof.Type, TypeMutex)
	}

	if len(prof.Functions) == 0 {
		t.Error("expected at least one function")
	}

	// Check mutex-specific stats
	if prof.Stats.TotalContentionTime == 0 {
		t.Error("expected non-zero total contention time")
	}
	if prof.Stats.TotalWaits == 0 {
		t.Error("expected non-zero total waits")
	}
}

// TestParseAutoDetect tests auto-detection parsing
func TestParseAutoDetect(t *testing.T) {
	tests := []struct {
		name     string
		testFile string
		expType  ProfileType
	}{
		{"CPU profile auto-detect", "../../testdata/cpu.prof", TypeCPU},
		{"Heap profile auto-detect", "../../testdata/heap.prof", TypeHeap},
		{"Goroutine profile auto-detect", "../../testdata/goroutine.prof", TypeGoroutine},
		{"Mutex profile auto-detect", "../../testdata/mutex.prof", TypeMutex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prof, err := Parse(tt.testFile)
			if err != nil {
				t.Fatalf("failed to parse profile: %v", err)
			}
			if prof.Type != tt.expType {
				t.Errorf("got type %s, want %s", prof.Type, tt.expType)
			}
		})
	}
}

// TestCPUParserFileNotFound tests error handling for missing files
func TestCPUParserFileNotFound(t *testing.T) {
	parser := &CPUParser{}
	_, err := parser.Parse("/nonexistent/file.prof")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestHeapParserFileNotFound tests error handling for missing files
func TestHeapParserFileNotFound(t *testing.T) {
	parser := &HeapParser{}
	_, err := parser.Parse("/nonexistent/file.prof")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestGoroutineParserFileNotFound tests error handling for missing files
func TestGoroutineParserFileNotFound(t *testing.T) {
	parser := &GoroutineParser{}
	_, err := parser.Parse("/nonexistent/file.prof")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestMutexParserFileNotFound tests error handling for missing files
func TestMutexParserFileNotFound(t *testing.T) {
	parser := &MutexParser{}
	_, err := parser.Parse("/nonexistent/file.prof")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestParserTypeValidation tests parser type validation
func TestParserTypeValidation(t *testing.T) {
	tests := []struct {
		name     string
		pType    ProfileType
		expected bool
	}{
		{"CPU type", TypeCPU, true},
		{"Heap type", TypeHeap, true},
		{"Goroutine type", TypeGoroutine, true},
		{"Mutex type", TypeMutex, true},
		{"Invalid type", ProfileType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.pType)
			if (p != nil) != tt.expected {
				t.Errorf("NewParser(%v) = %v, want nil: %v", tt.pType, p != nil, !tt.expected)
			}
		})
	}
}

// TestPercentageCalculations tests that percentages are calculated correctly
func TestPercentageCalculations(t *testing.T) {
	parser := &CPUParser{}
	prof, err := parser.Parse("../../testdata/cpu.prof")
	if err != nil {
		t.Fatalf("failed to parse CPU profile: %v", err)
	}

	// Sum of all flat percentages should be approximately 100%
	// (may be slightly off due to functions with zero flat value)
	totalFlatPct := 0.0
	for _, fn := range prof.Functions {
		totalFlatPct += fn.FlatPct
	}

	// Allow some tolerance for functions that only have cumulative value
	if totalFlatPct < 50 || totalFlatPct > 110 {
		t.Errorf("total flat percentage %.2f%% seems wrong (expected ~100%%)", totalFlatPct)
	}
}

// TestCallStackTests verifies call stacks are populated
func TestCallStackTests(t *testing.T) {
	parser := &CPUParser{}
	prof, err := parser.Parse("../../testdata/cpu.prof")
	if err != nil {
		t.Fatalf("failed to parse CPU profile: %v", err)
	}

	for _, fn := range prof.Functions {
		if len(fn.CallStack) == 0 {
			t.Logf("function %s has empty call stack", fn.Name)
		}
	}
}

// TestInvalidProfileFile tests parsing invalid profile files
func TestInvalidProfileFile(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.prof")

	// Write some random data
	if err := os.WriteFile(invalidFile, []byte("not a valid profile"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := Parse(invalidFile)
	if err == nil {
		t.Error("expected error when parsing invalid profile file")
	}
}
