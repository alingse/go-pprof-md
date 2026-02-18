package parser

import (
	"fmt"
	"os"
	"time"

	"github.com/google/pprof/profile"
)

// ProfileType represents the type of pprof profile
type ProfileType string

const (
	TypeCPU       ProfileType = "cpu"
	TypeHeap      ProfileType = "heap"
	TypeGoroutine ProfileType = "goroutine"
	TypeMutex     ProfileType = "mutex"
)

// Profile represents the parsed pprof data
type Profile struct {
	Type        ProfileType
	SampleTime  time.Duration
	TotalSamples int64
	Functions   []Function
	Stats       Stats
}

// CallPath represents a single call path with its weight
type CallPath struct {
	Stack  []string
	Weight int64
}

// Function represents a function in the profile
type Function struct {
	Name      string
	File      string
	Line      int
	Flat      int64    // Direct resource consumption
	Cum       int64    // Cumulative resource consumption
	FlatPct   float64  // Percentage of total
	CumPct    float64  // Percentage of total
	SumPct    float64  // Cumulative sum of FlatPct (running total)
	CallStack []string // Call stack (heaviest path, for backward compat)
	CallPaths []CallPath // All call paths with weights
}

// Stats contains summary statistics
type Stats struct {
	// Common stats
	TotalSamples int64
	TotalDuration time.Duration

	// CPU specific
	CPUProfileDuration time.Duration
	SampleRate         int64

	// Heap specific
	AllocBytes    int64
	AllocObjects  int64
	InUseBytes    int64
	InUseObjects  int64

	// Goroutine specific
	TotalGoroutines int64

	// Mutex specific
	TotalContentionTime int64
	TotalWaits         int64
}

// Parser interface for parsing different pprof profile types
type Parser interface {
	Parse(filename string) (*Profile, error)
	DetectType() (ProfileType, error)
}

// NewParser creates a parser for the given profile type
func NewParser(profileType ProfileType) Parser {
	switch profileType {
	case TypeCPU:
		return &CPUParser{}
	case TypeHeap:
		return &HeapParser{}
	case TypeGoroutine:
		return &GoroutineParser{}
	case TypeMutex:
		return &MutexParser{}
	default:
		return nil
	}
}

// DetectProfileType auto-detects the profile type from file content
func DetectProfileType(filename string) (ProfileType, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Use official pprof library to parse
	prof, err := profile.Parse(f)
	if err != nil {
		return "", fmt.Errorf("failed to parse profile: %w", err)
	}

	return detectProfileTypeFromSampleType(prof)
}

// Parse parses a pprof file with auto-detected type
func Parse(filename string) (*Profile, error) {
	profileType, err := DetectProfileType(filename)
	if err != nil {
		return nil, err
	}

	parser := NewParser(profileType)
	if parser == nil {
		return nil, fmt.Errorf("unsupported profile type: %s", profileType)
	}

	return parser.Parse(filename)
}
