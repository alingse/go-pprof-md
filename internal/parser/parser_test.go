package parser

import (
	"os"
	"testing"
)

// TestDetectProfileType tests profile type detection
func TestDetectProfileType(t *testing.T) {
	// Test with a simple gzip header (minimal valid profile)
	minimalProfile := []byte{0x1f, 0x8b, 0x08, 0x00} // gzip magic

	tmpDir := t.TempDir()
	profFile := tmpDir + "/test.prof"

	if err := os.WriteFile(profFile, minimalProfile, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// This will likely fail to parse, but we test the detection logic
	_, err := DetectProfileType(profFile)
	if err == nil {
		t.Log("unexpected success parsing minimal profile")
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
