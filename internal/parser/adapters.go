package parser

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/pprof/profile"
)

// convertProfile converts a pprof profile.Profile to internal parser.Profile
func convertProfile(prof *profile.Profile, profileType ProfileType) (*Profile, error) {
	if prof == nil {
		return nil, fmt.Errorf("nil profile")
	}

	result := &Profile{
		Type:        profileType,
		SampleTime:  time.Duration(prof.DurationNanos),
		TotalSamples: 0,
		Functions:   []Function{},
		Stats:       Stats{},
	}

	// Build function map by ID for quick lookup
	funcMap := make(map[uint64]*profile.Function)
	for _, fn := range prof.Function {
		funcMap[fn.ID] = fn
	}

	// Track function data by function ID for accurate aggregation
	functionData := make(map[uint64]*FunctionData)

	// Process samples
	for _, sample := range prof.Sample {
		if len(sample.Value) == 0 {
			continue
		}

		// Get value based on profile type
		var value int64
		var value2 int64 // For profiles with 2 values

		switch profileType {
		case TypeCPU:
			value = sample.Value[0] * prof.Period // scale to nanoseconds
			result.TotalSamples += value
		case TypeHeap:
			// Heap profiles: 4 values = [alloc_objects, alloc_space, inuse_objects, inuse_space]
			//                 2 values = [alloc_objects, alloc_space] (inuse stays 0)
			if len(sample.Value) >= 2 {
				result.Stats.AllocObjects += sample.Value[0]
				result.Stats.AllocBytes += sample.Value[1]
				value = sample.Value[0]  // objects
				value2 = sample.Value[1] // bytes
			}
			if len(sample.Value) >= 4 {
				result.Stats.InUseObjects += sample.Value[2]
				result.Stats.InUseBytes += sample.Value[3]
			}
		case TypeGoroutine:
			value = sample.Value[0] // count
			result.Stats.TotalGoroutines += value
			result.TotalSamples += value
		case TypeMutex:
			// Mutex samples: [contentions (count), lock_duration (nanoseconds)]
			if len(sample.Value) >= 2 {
				value = sample.Value[1]  // lock duration (nanoseconds) - primary metric
				value2 = sample.Value[0] // contentions count
				result.Stats.TotalContentionTime += value
				result.Stats.TotalWaits += value2
			}
		}

		// Build call stack
		callStack := buildCallStackFromSample(sample, funcMap)

		// Process each location in the stack
		for i, loc := range sample.Location {
			for _, line := range loc.Line {
				fn := line.Function
				if fn == nil {
					continue
				}

				// Get or create function data
				data, exists := functionData[fn.ID]
				if !exists {
					funcName := fn.Name
					if funcName == "" {
						funcName = fn.SystemName
					}
					data = &FunctionData{
						ID:        fn.ID,
						Name:      funcName,
						File:      fn.Filename,
						Line:      int(line.Line),
						CallStack: callStack,
					}
					functionData[fn.ID] = data
				}

				// Flat value only for the leaf (first) location
				// In pprof, Location[0] is the leaf (innermost frame)
				isLeaf := (i == 0)

				// Use bytes for heap profiles as primary metric
				metricValue := value
				if profileType == TypeHeap {
					metricValue = value2
				}

				if isLeaf {
					data.Flat += metricValue
					// Track this call path for the leaf function
					data.CallPaths = append(data.CallPaths, CallPath{
						Stack:  callStack,
						Weight: metricValue,
					})
				}
				data.Cum += metricValue
			}
		}
	}

	// Set total samples based on profile type
	switch profileType {
	case TypeHeap:
		result.TotalSamples = result.Stats.AllocBytes
	case TypeMutex:
		result.TotalSamples = result.Stats.TotalContentionTime
	}

	// Convert to Function slice and calculate percentages
	for _, data := range functionData {
		var total int64
		switch profileType {
		case TypeHeap:
			total = result.Stats.AllocBytes
		case TypeMutex:
			total = result.Stats.TotalContentionTime
		default:
			total = result.TotalSamples
		}

		// Merge duplicate call paths and sort by weight descending
		callPaths := mergeCallPaths(data.CallPaths)
		sort.Slice(callPaths, func(i, j int) bool {
			return callPaths[i].Weight > callPaths[j].Weight
		})

		// Set CallStack to the heaviest path for backward compatibility
		callStack := data.CallStack
		if len(callPaths) > 0 {
			callStack = callPaths[0].Stack
		}

		fn := Function{
			Name:      data.Name,
			File:      data.File,
			Line:      data.Line,
			Flat:      data.Flat,
			Cum:       data.Cum,
			CallStack: callStack,
			CallPaths: callPaths,
		}

		if total > 0 {
			fn.FlatPct = float64(fn.Flat) / float64(total) * 100
			fn.CumPct = float64(fn.Cum) / float64(total) * 100
		}

		result.Functions = append(result.Functions, fn)
	}

	// Sort by flat value descending (matches `go tool pprof -text` default)
	sort.Slice(result.Functions, func(i, j int) bool {
		return result.Functions[i].Flat > result.Functions[j].Flat
	})

	// Compute cumulative sum percentage (SumPct)
	sumPct := 0.0
	for i := range result.Functions {
		sumPct += result.Functions[i].FlatPct
		result.Functions[i].SumPct = sumPct
	}

	// Set common stats
	result.Stats.TotalSamples = result.TotalSamples
	result.Stats.TotalDuration = time.Duration(prof.DurationNanos)
	result.Stats.CPUProfileDuration = time.Duration(prof.DurationNanos)

	// Set sample rate for CPU profiles (convert period in ns to Hz)
	if profileType == TypeCPU && prof.Period > 0 {
		result.Stats.SampleRate = 1e9 / prof.Period
	}

	return result, nil
}

// FunctionData holds intermediate data during conversion
type FunctionData struct {
	ID        uint64
	Name      string
	File      string
	Line      int
	Flat      int64
	Cum       int64
	CallStack []string
	CallPaths []CallPath
}

// buildCallStackFromSample builds a call stack from a sample
func buildCallStackFromSample(sample *profile.Sample, funcMap map[uint64]*profile.Function) []string {
	stack := []string{}

	// Process locations in reverse order (leaf first)
	for i := len(sample.Location) - 1; i >= 0; i-- {
		loc := sample.Location[i]
		for _, line := range loc.Line {
			fn := line.Function
			if fn != nil {
				name := fn.Name
				if name == "" {
					name = fn.SystemName
				}
				if name != "" {
					stack = append(stack, name)
				}
			}
		}
	}

	return stack
}

// mergeCallPaths merges duplicate call paths by summing their weights
func mergeCallPaths(paths []CallPath) []CallPath {
	if len(paths) == 0 {
		return nil
	}
	merged := make(map[string]*CallPath)
	for _, p := range paths {
		key := strings.Join(p.Stack, "\x00")
		if existing, ok := merged[key]; ok {
			existing.Weight += p.Weight
		} else {
			cp := CallPath{Stack: p.Stack, Weight: p.Weight}
			merged[key] = &cp
		}
	}
	result := make([]CallPath, 0, len(merged))
	for _, cp := range merged {
		result = append(result, *cp)
	}
	return result
}

// detectProfileTypeFromSampleType detects profile type from sample type
func detectProfileTypeFromSampleType(prof *profile.Profile) (ProfileType, error) {
	if len(prof.SampleType) == 0 {
		return "", fmt.Errorf("no sample type in profile")
	}

	// Check sample types for known patterns
	for _, st := range prof.SampleType {
		typeStr := st.Type
		unitStr := st.Unit

		switch typeStr {
		case "cpu", "samples":
			if unitStr == "nanoseconds" || unitStr == "seconds" || unitStr == "count" {
				return TypeCPU, nil
			}
		case "alloc_objects", "inuse_objects", "alloc_space", "inuse_space":
			return TypeHeap, nil
		case "goroutines":
			return TypeGoroutine, nil
		case "contentions", "lock_duration":
			return TypeMutex, nil
		}

		// Fallback: check unit
		switch unitStr {
		case "nanoseconds", "seconds", "milliseconds":
			// Could be CPU or mutex, check type name
			if typeStr == "cpu" || typeStr == "samples" {
				return TypeCPU, nil
			}
		case "bytes":
			return TypeHeap, nil
		case "objects", "count":
			if typeStr == "goroutines" {
				return TypeGoroutine, nil
			}
		case "lock_ns", "contentions":
			return TypeMutex, nil
		}
	}

	// Try period type
	if prof.PeriodType != nil {
		if prof.PeriodType.Type == "cpu" || prof.PeriodType.Type == "samples" {
			return TypeCPU, nil
		}
	}

	// Default: try to infer from sample value count
	if len(prof.Sample) > 0 {
		valueCount := len(prof.Sample[0].Value)
		if valueCount == 2 {
			// Could be heap or mutex
			// Check sample type names
			for _, st := range prof.SampleType {
				if st.Type == "alloc_objects" || st.Type == "inuse_objects" {
					return TypeHeap, nil
				}
				if st.Type == "contentions" {
					return TypeMutex, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unknown profile type: sample types: %v", prof.SampleType)
}
