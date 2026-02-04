package parser

import (
	"fmt"
	"os"
	"sort"
	"time"
)

// CPUParser parses CPU pprof profiles
type CPUParser struct{}

// Parse parses a CPU profile file
func (p *CPUParser) Parse(filename string) (*Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	protoProf, err := parseProtoProfile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse protobuf: %w", err)
	}

	profile := &Profile{
		Type:         TypeCPU,
		SampleTime:   time.Duration(protoProf.DurationNanos),
		TotalSamples: 0,
		Functions:    []Function{},
		Stats: Stats{
			TotalDuration:     time.Duration(protoProf.DurationNanos),
			CPUProfileDuration: time.Duration(protoProf.DurationNanos),
			SampleRate:        protoProf.Period,
		},
	}

	// Parse samples
	funcMap := make(map[string]*Function)
	totalSamples := int64(0)

	for _, sample := range protoProf.Sample {
		if len(sample.Value) == 0 {
			continue
		}

		// CPU samples have one value: sample count
		value := sample.Value[0]
		totalSamples += value

		// Process location stack
		for _, locIdx := range sample.LocationIdx {
			loc := findLocation(protoProf, locIdx)
			if loc == nil {
				continue
			}

			for _, line := range loc.Line {
				fn := findFunction(protoProf, line.FunctionId)
				if fn == nil {
					continue
				}

				funcName := getString(protoProf, fn.Name)
				fileName := getString(protoProf, fn.Filename)

				if _, exists := funcMap[funcName]; !exists {
					funcMap[funcName] = &Function{
						Name:      funcName,
						File:      fileName,
						Line:      int(line.Line),
						Flat:      0,
						Cum:       0,
						FlatPct:   0,
						CumPct:    0,
						CallStack: []string{},
					}
				}

				// Add to flat value (leaf function)
				if locIdx == sample.LocationIdx[len(sample.LocationIdx)-1] {
					funcMap[funcName].Flat += value
				}
				funcMap[funcName].Cum += value

				// Build call stack
				callStack := buildCallStack(protoProf, sample.LocationIdx)
				if len(callStack) > 0 {
					funcMap[funcName].CallStack = callStack
				}
			}
		}
	}

	profile.TotalSamples = totalSamples
	profile.Stats.TotalSamples = totalSamples

	// Convert map to slice and calculate percentages
	for _, fn := range funcMap {
		if totalSamples > 0 {
			fn.FlatPct = float64(fn.Flat) / float64(totalSamples) * 100
			fn.CumPct = float64(fn.Cum) / float64(totalSamples) * 100
		}
		profile.Functions = append(profile.Functions, *fn)
	}

	// Sort by cumulative value descending
	sort.Slice(profile.Functions, func(i, j int) bool {
		return profile.Functions[i].Cum > profile.Functions[j].Cum
	})

	return profile, nil
}

// DetectType always returns TypeCPU for CPUParser
func (p *CPUParser) DetectType() (ProfileType, error) {
	return TypeCPU, nil
}
