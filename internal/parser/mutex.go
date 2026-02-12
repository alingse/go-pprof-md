package parser

import (
	"fmt"
	"os"

	"github.com/google/pprof/profile"
)

// MutexParser parses mutex/lock pprof profiles
type MutexParser struct{}

// Parse parses a mutex profile file
func (p *MutexParser) Parse(filename string) (*Profile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	prof, err := profile.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return convertProfile(prof, TypeMutex)
}

// DetectType always returns TypeMutex for MutexParser
func (p *MutexParser) DetectType() (ProfileType, error) {
	return TypeMutex, nil
}
