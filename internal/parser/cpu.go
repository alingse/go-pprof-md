package parser

import (
	"fmt"
	"os"

	"github.com/google/pprof/profile"
)

// CPUParser parses CPU pprof profiles
type CPUParser struct{}

// Parse parses a CPU profile file
func (p *CPUParser) Parse(filename string) (*Profile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	prof, err := profile.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return convertProfile(prof, TypeCPU)
}

// DetectType always returns TypeCPU for CPUParser
func (p *CPUParser) DetectType() (ProfileType, error) {
	return TypeCPU, nil
}
