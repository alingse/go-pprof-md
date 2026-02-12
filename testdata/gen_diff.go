// +build ignore

package main

import (
	"os"

	"github.com/google/pprof/profile"
)

func main() {
	// Generate base heap profile
	baseProf := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
		Sample: []*profile.Sample{
			{
				Location: []*profile.Location{
					{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go", StartLine: 20}, Line: 25}}},
				},
				Value: []int64{1000, 80000},
			},
			{
				Location: []*profile.Location{
					{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "main.processData", Filename: "main.go", StartLine: 30}, Line: 35}}},
				},
				Value: []int64{500, 40000},
			},
		},
		Location: []*profile.Location{
			{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go", StartLine: 20}, Line: 25}}},
			{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "main.processData", Filename: "main.go", StartLine: 30}, Line: 35}}},
		},
		Function: []*profile.Function{
			{ID: 1, Name: "main.makeSlice", SystemName: "main.makeSlice", Filename: "main.go", StartLine: 20},
			{ID: 2, Name: "main.processData", SystemName: "main.processData", Filename: "main.go", StartLine: 30},
		},
		DurationNanos: 10_000_000_000,
	}

	// Generate new heap profile with changes
	newProf := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
		Sample: []*profile.Sample{
			{
				// makeSlice increased - regression
				Location: []*profile.Location{
					{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go", StartLine: 20}, Line: 25}}},
				},
				Value: []int64{2000, 160000}, // Doubled!
			},
			{
				// processData improved
				Location: []*profile.Location{
					{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "main.processData", Filename: "main.go", StartLine: 30}, Line: 35}}},
				},
				Value: []int64{100, 8000}, // Reduced!
			},
			{
				// New function!
				Location: []*profile.Location{
					{ID: 3, Line: []profile.Line{{Function: &profile.Function{ID: 3, Name: "main.newFeature", Filename: "main.go", StartLine: 40}, Line: 45}}},
				},
				Value: []int64{300, 24000},
			},
		},
		Location: []*profile.Location{
			{ID: 1, Line: []profile.Line{{Function: &profile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go", StartLine: 20}, Line: 25}}},
			{ID: 2, Line: []profile.Line{{Function: &profile.Function{ID: 2, Name: "main.processData", Filename: "main.go", StartLine: 30}, Line: 35}}},
			{ID: 3, Line: []profile.Line{{Function: &profile.Function{ID: 3, Name: "main.newFeature", Filename: "main.go", StartLine: 40}, Line: 45}}},
		},
		Function: []*profile.Function{
			{ID: 1, Name: "main.makeSlice", SystemName: "main.makeSlice", Filename: "main.go", StartLine: 20},
			{ID: 2, Name: "main.processData", SystemName: "main.processData", Filename: "main.go", StartLine: 30},
			{ID: 3, Name: "main.newFeature", SystemName: "main.newFeature", Filename: "main.go", StartLine: 40},
		},
		DurationNanos: 10_000_000_000,
	}

	// Write base profile
	f1, _ := os.Create("testdata/base.prof")
	baseProf.Write(f1)
	f1.Close()
	println("Generated testdata/base.prof")

	// Write new profile
	f2, _ := os.Create("testdata/new.prof")
	newProf.Write(f2)
	f2.Close()
	println("Generated testdata/new.prof")
}
