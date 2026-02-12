package parser

import (
	"fmt"
	"os"

	"github.com/google/pprof/profile"
)

// GoroutineParser parses goroutine pprof profiles
type GoroutineParser struct{}

// Parse parses a goroutine profile file
func (p *GoroutineParser) Parse(filename string) (*Profile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	prof, err := profile.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return convertProfile(prof, TypeGoroutine)
}

// DetectType always returns TypeGoroutine for GoroutineParser
func (p *GoroutineParser) DetectType() (ProfileType, error) {
	return TypeGoroutine, nil
}
