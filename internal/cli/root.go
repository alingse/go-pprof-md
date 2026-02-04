package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-pprof-md",
	Short: "Convert Go pprof profiles to AI-readable markdown",
	Long: `go-pprof-md converts Go pprof profiles (CPU, heap, goroutine, mutex)
into markdown format optimized for AI analysis.

The tool extracts key metrics, function hotspots, and call stacks,
and includes AI-optimized prompts for automated performance analysis.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
