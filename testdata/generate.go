// +build ignore

package main

import (
	"os"

	"github.com/google/pprof/profile"
)

func main() {
	// Generate test CPU profile
	if err := generateCPUProfile("testdata/cpu.prof"); err != nil {
		panic(err)
	}
	println("Generated testdata/cpu.prof")

	// Generate test heap profile
	if err := generateHeapProfile("testdata/heap.prof"); err != nil {
		panic(err)
	}
	println("Generated testdata/heap.prof")

	// Generate test goroutine profile
	if err := generateGoroutineProfile("testdata/goroutine.prof"); err != nil {
		panic(err)
	}
	println("Generated testdata/goroutine.prof")

	// Generate test mutex profile
	if err := generateMutexProfile("testdata/mutex.prof"); err != nil {
		panic(err)
	}
	println("Generated testdata/mutex.prof")
}

func generateCPUProfile(filename string) error {
	prof := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "samples", Unit: "count"},
		},
		Sample: []*profile.Sample{
			{
				Location: []*profile.Location{
					{
						ID: 1,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 1, Name: "main.main", Filename: "main.go", StartLine: 10}, Line: 15},
						},
					},
					{
						ID: 2,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 2, Name: "runtime.main", Filename: "runtime/proc.go", StartLine: 100}, Line: 120},
						},
					},
				},
				Value: []int64{100},
			},
			{
				Location: []*profile.Location{
					{
						ID: 3,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 3, Name: "math/big.nat.set", Filename: "math/big/nat.go", StartLine: 50}, Line: 55},
						},
					},
					{
						ID: 1,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 1, Name: "main.main", Filename: "main.go", StartLine: 10}, Line: 15},
						},
					},
				},
				Value: []int64{50},
			},
		},
		Location: []*profile.Location{
			{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.main", Filename: "main.go", StartLine: 10}, Line: 15}}},
			{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "runtime.main", Filename: "runtime/proc.go", StartLine: 100}, Line: 120}}},
			{ID: 3, Line: []profile.Line{{Function: &profile.Function{ID: 3, Name: "math/big.nat.set", Filename: "math/big/nat.go", StartLine: 50}, Line: 55}}},
		},
		Function: []*profile.Function{
			{ID: 1, Name: "main.main", SystemName: "main.main", Filename: "main.go", StartLine: 10},
			{ID: 2, Name: "runtime.main", SystemName: "runtime.main", Filename: "runtime/proc.go", StartLine: 100},
			{ID: 3, Name: "math/big.nat.set", SystemName: "math/big.nat.set", Filename: "math/big/nat.go", StartLine: 50},
		},
		PeriodType: &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:     1000000, // 1ms sampling rate
		DurationNanos: 30_000_000_000, // 30 seconds
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return prof.Write(f)
}

func generateHeapProfile(filename string) error {
	prof := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
		Sample: []*profile.Sample{
			{
				Location: []*profile.Location{
					{
						ID: 1,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go", StartLine: 20}, Line: 25},
						},
					},
				},
				Value: []int64{1000, 80000}, // 1000 objects, 80000 bytes
			},
			{
				Location: []*profile.Location{
					{
						ID: 2,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 2, Name: "main.makeStruct", Filename: "main.go", StartLine: 30}, Line: 35},
						},
					},
				},
				Value: []int64{500, 40000}, // 500 objects, 40000 bytes
			},
		},
		Location: []*profile.Location{
			{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go", StartLine: 20}, Line: 25}}},
			{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "main.makeStruct", Filename: "main.go", StartLine: 30}, Line: 35}}},
		},
		Function: []*profile.Function{
			{ID: 1, Name: "main.makeSlice", SystemName: "main.makeSlice", Filename: "main.go", StartLine: 20},
			{ID: 2, Name: "main.makeStruct", SystemName: "main.makeStruct", Filename: "main.go", StartLine: 30},
		},
		DurationNanos: 10_000_000_000, // 10 seconds
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return prof.Write(f)
}

func generateGoroutineProfile(filename string) error {
	prof := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "goroutines", Unit: "count"},
		},
		Sample: []*profile.Sample{
			{
				Location: []*profile.Location{
					{
						ID: 1,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 1, Name: "main.worker", Filename: "main.go", StartLine: 40}, Line: 45},
						},
					},
				},
				Value: []int64{5}, // 5 goroutines
			},
			{
				Location: []*profile.Location{
					{
						ID: 2,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 2, Name: "main.handler", Filename: "main.go", StartLine: 50}, Line: 55},
						},
					},
				},
				Value: []int64{3}, // 3 goroutines
			},
		},
		Location: []*profile.Location{
			{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.worker", Filename: "main.go", StartLine: 40}, Line: 45}}},
			{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "main.handler", Filename: "main.go", StartLine: 50}, Line: 55}}},
		},
		Function: []*profile.Function{
			{ID: 1, Name: "main.worker", SystemName: "main.worker", Filename: "main.go", StartLine: 40},
			{ID: 2, Name: "main.handler", SystemName: "main.handler", Filename: "main.go", StartLine: 50},
		},
		DurationNanos: 5_000_000_000, // 5 seconds
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return prof.Write(f)
}

func generateMutexProfile(filename string) error {
	prof := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "contentions", Unit: "count"},
			{Type: "lock_duration", Unit: "nanoseconds"},
		},
		Sample: []*profile.Sample{
			{
				Location: []*profile.Location{
					{
						ID: 1,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 1, Name: "main.lockedSection", Filename: "main.go", StartLine: 60}, Line: 65},
						},
					},
				},
				Value: []int64{100, 50000000}, // 100 contentions, 50ms total
			},
			{
				Location: []*profile.Location{
					{
						ID: 2,
						Line: []profile.Line{
							{Function: &profile.Function{ID: 2, Name: "main.anotherLock", Filename: "main.go", StartLine: 70}, Line: 75},
						},
					},
				},
				Value: []int64{50, 25000000}, // 50 contentions, 25ms total
			},
		},
		Location: []*profile.Location{
			{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.lockedSection", Filename: "main.go", StartLine: 60}, Line: 65}}},
			{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "main.anotherLock", Filename: "main.go", StartLine: 70}, Line: 75}}},
		},
		Function: []*profile.Function{
			{ID: 1, Name: "main.lockedSection", SystemName: "main.lockedSection", Filename: "main.go", StartLine: 60},
			{ID: 2, Name: "main.anotherLock", SystemName: "main.anotherLock", Filename: "main.go", StartLine: 70},
		},
		DurationNanos: 15_000_000_000, // 15 seconds
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return prof.Write(f)
}
