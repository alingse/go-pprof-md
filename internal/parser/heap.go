package parser

import (
	"fmt"
	"os"
	"sort"
)

// HeapParser parses heap pprof profiles
type HeapParser struct{}

// Parse parses a heap profile file
func (p *HeapParser) Parse(filename string) (*Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	protoProf, err := parseProtoProfile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse protobuf: %w", err)
	}

	profile := &Profile{
		Type:         TypeHeap,
		TotalSamples: 0,
		Functions:    []Function{},
		Stats: Stats{
			AllocBytes:   0,
			AllocObjects: 0,
			InUseBytes:   0,
			InUseObjects: 0,
		},
	}

	// Parse samples
	funcMap := make(map[string]*Function)

	for _, sample := range protoProf.Sample {
		if len(sample.Value) < 2 {
			continue
		}

		// Heap samples have 2 values: alloc_objects, alloc_bytes (or inuse_*)
		// Value[0] is typically objects/count
		// Value[1] is typically bytes
		objects := sample.Value[0]
		bytes := sample.Value[1]

		profile.Stats.AllocObjects += objects
		profile.Stats.AllocBytes += bytes
		profile.Stats.InUseObjects += objects
		profile.Stats.InUseBytes += bytes

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
					funcMap[funcName].Flat += bytes
				}
				funcMap[funcName].Cum += bytes

				// Build call stack
				callStack := buildCallStack(protoProf, sample.LocationIdx)
				if len(callStack) > 0 {
					funcMap[funcName].CallStack = callStack
				}
			}
		}
	}

	profile.TotalSamples = profile.Stats.AllocBytes

	// Convert map to slice and calculate percentages
	totalBytes := profile.Stats.AllocBytes
	for _, fn := range funcMap {
		if totalBytes > 0 {
			fn.FlatPct = float64(fn.Flat) / float64(totalBytes) * 100
			fn.CumPct = float64(fn.Cum) / float64(totalBytes) * 100
		}
		profile.Functions = append(profile.Functions, *fn)
	}

	// Sort by cumulative bytes descending
	sort.Slice(profile.Functions, func(i, j int) bool {
		return profile.Functions[i].Cum > profile.Functions[j].Cum
	})

	return profile, nil
}

// DetectType always returns TypeHeap for HeapParser
func (p *HeapParser) DetectType() (ProfileType, error) {
	return TypeHeap, nil
}
