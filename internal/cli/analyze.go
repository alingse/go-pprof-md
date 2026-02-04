package cli

import (
	"fmt"
	"os"

	"github.com/alingse/go-pprof-md/internal/generator"
	"github.com/alingse/go-pprof-md/internal/parser"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	topN       int
	noAIPrompt bool
	profileType string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <pprof-file>",
	Short: "Analyze a pprof file and generate markdown report",
	Long: `Analyze a pprof file (CPU, heap, goroutine, or mutex) and
generate a markdown report optimized for AI analysis.

The profile type is auto-detected from the file content.`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	analyzeCmd.Flags().IntVarP(&topN, "top", "n", 20, "Number of top functions to display")
	analyzeCmd.Flags().BoolVar(&noAIPrompt, "no-ai-prompt", false, "Disable AI analysis prompt")
	analyzeCmd.Flags().StringVarP(&profileType, "type", "t", "", "Profile type (cpu, heap, goroutine, mutex). Auto-detected if not specified")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Parse profile
	var profile *parser.Profile
	var err error

	if profileType != "" {
		p := parser.NewParser(parser.ProfileType(profileType))
		if p == nil {
			return fmt.Errorf("invalid profile type: %s", profileType)
		}
		profile, err = p.Parse(filename)
	} else {
		profile, err = parser.Parse(filename)
	}

	if err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	// Generate markdown
	gen := generator.NewGenerator(
		profile,
		generator.WithTopN(topN),
		generator.WithAIPrompt(!noAIPrompt),
	)

	markdown, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}

	// Write output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(markdown), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Markdown report written to: %s\n", outputFile)
	} else {
		fmt.Print(markdown)
	}

	return nil
}
