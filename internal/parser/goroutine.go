package parser

import (
	"fmt"
	"os"
	"sort"
)

// GoroutineParser parses goroutine pprof profiles
type GoroutineParser struct{}

// Parse parses a goroutine profile file
func (p *GoroutineParser) Parse(filename string) (*Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	protoProf, err := parseProtoProfile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse protobuf: %w", err)
	}

	profile := &Profile{
		Type:         TypeGoroutine,
		TotalSamples: 0,
		Functions:    []Function{},
		Stats: Stats{
			TotalGoroutines: 0,
		},
	}

	// Parse samples
	funcMap := make(map[string]*Function)
	totalGoroutines := int64(0)

	for _, sample := range protoProf.Sample {
		if len(sample.Value) == 0 {
			continue
		}

		// Goroutine samples have one value: count
		count := sample.Value[0]
		totalGoroutines += count
		profile.Stats.TotalGoroutines += count

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
					funcMap[funcName].Flat += count
				}
				funcMap[funcName].Cum += count

				// Build call stack
				callStack := buildCallStack(protoProf, sample.LocationIdx)
				if len(callStack) > 0 {
					funcMap[funcName].CallStack = callStack
				}
			}
		}
	}

	profile.TotalSamples = totalGoroutines

	// Convert map to slice and calculate percentages
	for _, fn := range funcMap {
		if totalGoroutines > 0 {
			fn.FlatPct = float64(fn.Flat) / float64(totalGoroutines) * 100
			fn.CumPct = float64(fn.Cum) / float64(totalGoroutines) * 100
		}
		profile.Functions = append(profile.Functions, *fn)
	}

	// Sort by cumulative count descending
	sort.Slice(profile.Functions, func(i, j int) bool {
		return profile.Functions[i].Cum > profile.Functions[j].Cum
	})

	return profile, nil
}

// DetectType always returns TypeGoroutine for GoroutineParser
func (p *GoroutineParser) DetectType() (ProfileType, error) {
	return TypeGoroutine, nil
}
