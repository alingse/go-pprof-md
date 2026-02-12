package cli

import (
	"fmt"
	"os"

	"github.com/alingse/go-pprof-md/internal/generator"
	"github.com/alingse/go-pprof-md/internal/parser"
	"github.com/spf13/cobra"
)

var (
	outputFile  string
	topN        int
	noAIPrompt  bool
	profileType string
)

var showCmd = &cobra.Command{
	Use:   "show <pprof-file>",
	Short: "Show a pprof file as markdown report",
	Long: `Show a pprof file (CPU, heap, goroutine, or mutex) as
a markdown report optimized for AI analysis.

The profile type is auto-detected from the file content.`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	showCmd.Flags().IntVarP(&topN, "top", "n", 20, "Number of top functions to display")
	showCmd.Flags().BoolVar(&noAIPrompt, "no-ai-prompt", false, "Disable AI analysis prompt")
	showCmd.Flags().StringVarP(&profileType, "type", "t", "", "Profile type (cpu, heap, goroutine, mutex). Auto-detected if not specified")
}

func runShow(cmd *cobra.Command, args []string) error {
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
