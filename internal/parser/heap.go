package parser

import (
	"fmt"
	"os"

	"github.com/google/pprof/profile"
)

// HeapParser parses heap pprof profiles
type HeapParser struct{}

// Parse parses a heap profile file
func (p *HeapParser) Parse(filename string) (*Profile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	prof, err := profile.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return convertProfile(prof, TypeHeap)
}

// DetectType always returns TypeHeap for HeapParser
func (p *HeapParser) DetectType() (ProfileType, error) {
	return TypeHeap, nil
}
