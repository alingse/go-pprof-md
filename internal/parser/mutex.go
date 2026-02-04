package parser

import (
	"fmt"
	"os"
	"sort"
)

// MutexParser parses mutex/lock pprof profiles
type MutexParser struct{}

// Parse parses a mutex profile file
func (p *MutexParser) Parse(filename string) (*Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	protoProf, err := parseProtoProfile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse protobuf: %w", err)
	}

	profile := &Profile{
		Type:         TypeMutex,
		TotalSamples: 0,
		Functions:    []Function{},
		Stats: Stats{
			TotalContentionTime: 0,
			TotalWaits:          0,
		},
	}

	// Parse samples
	funcMap := make(map[string]*Function)
	totalContention := int64(0)
	totalWaits := int64(0)

	for _, sample := range protoProf.Sample {
		if len(sample.Value) < 2 {
			continue
		}

		// Mutex samples have 2 values: contention time (nanoseconds), wait count
		contentionTime := sample.Value[0]
		waits := sample.Value[1]

		totalContention += contentionTime
		totalWaits += waits
		profile.Stats.TotalContentionTime += contentionTime
		profile.Stats.TotalWaits += waits

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

				// Add to flat value (leaf function) - use contention time as primary metric
				if locIdx == sample.LocationIdx[len(sample.LocationIdx)-1] {
					funcMap[funcName].Flat += contentionTime
				}
				funcMap[funcName].Cum += contentionTime

				// Build call stack
				callStack := buildCallStack(protoProf, sample.LocationIdx)
				if len(callStack) > 0 {
					funcMap[funcName].CallStack = callStack
				}
			}
		}
	}

	profile.TotalSamples = totalContention

	// Convert map to slice and calculate percentages
	for _, fn := range funcMap {
		if totalContention > 0 {
			fn.FlatPct = float64(fn.Flat) / float64(totalContention) * 100
			fn.CumPct = float64(fn.Cum) / float64(totalContention) * 100
		}
		profile.Functions = append(profile.Functions, *fn)
	}

	// Sort by cumulative contention time descending
	sort.Slice(profile.Functions, func(i, j int) bool {
		return profile.Functions[i].Cum > profile.Functions[j].Cum
	})

	return profile, nil
}

// DetectType always returns TypeMutex for MutexParser
func (p *MutexParser) DetectType() (ProfileType, error) {
	return TypeMutex, nil
}
