package cli

import (
	"fmt"
	"os"

	"github.com/alingse/go-pprof-md/internal/generator"
	"github.com/alingse/go-pprof-md/internal/parser"
	"github.com/spf13/cobra"
)

var (
	diffOutput    string
	diffTopN      int
	diffBaseType  string
	diffNewType   string
)

var diffCmd = &cobra.Command{
	Use:   "diff <base.prof> <new.prof>",
	Short: "Compare two pprof files",
	Long: `Compare two pprof files and show the differences.
This is useful for regression testing and performance analysis.

The profile types are auto-detected, but must match between files.

Example:
  go-pprof-md diff base.prof new.prof
  go-pprof-md diff -o diff.md before.prof after.prof`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&diffOutput, "output", "o", "", "Output file (default: stdout)")
	diffCmd.Flags().IntVarP(&diffTopN, "top", "n", 20, "Number of top changed functions to display")
	diffCmd.Flags().StringVarP(&diffBaseType, "base-type", "b", "", "Base profile type (auto-detected if not specified)")
	diffCmd.Flags().StringVarP(&diffNewType, "new-type", "t", "", "New profile type (auto-detected if not specified)")
}

func runDiff(cmd *cobra.Command, args []string) error {
	baseFile := args[0]
	newFile := args[1]

	// Check if files exist
	for _, f := range []string{baseFile, newFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", f)
		}
	}

	// Parse base profile
	var baseProfile, newProfile *parser.Profile
	var err error

	if diffBaseType != "" {
		p := parser.NewParser(parser.ProfileType(diffBaseType))
		if p == nil {
			return fmt.Errorf("invalid base profile type: %s", diffBaseType)
		}
		baseProfile, err = p.Parse(baseFile)
	} else {
		baseProfile, err = parser.Parse(baseFile)
	}
	if err != nil {
		return fmt.Errorf("failed to parse base profile: %w", err)
	}

	// Parse new profile
	if diffNewType != "" {
		p := parser.NewParser(parser.ProfileType(diffNewType))
		if p == nil {
			return fmt.Errorf("invalid new profile type: %s", diffNewType)
		}
		newProfile, err = p.Parse(newFile)
	} else {
		newProfile, err = parser.Parse(newFile)
	}
	if err != nil {
		return fmt.Errorf("failed to parse new profile: %w", err)
	}

	// Check profile types match
	if baseProfile.Type != newProfile.Type {
		return fmt.Errorf("profile types do not match: base is %s, new is %s", baseProfile.Type, newProfile.Type)
	}

	// Generate diff markdown
	gen := generator.NewDiffGenerator(baseProfile, newProfile,
		generator.WithDiffTopN(diffTopN),
	)

	markdown, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	// Write output
	if diffOutput != "" {
		if err := os.WriteFile(diffOutput, []byte(markdown), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Diff report written to: %s\n", diffOutput)
	} else {
		fmt.Print(markdown)
	}

	return nil
}
