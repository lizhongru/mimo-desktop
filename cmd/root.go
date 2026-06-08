package cmd

import (
	"fmt"
	"os"

	"github.com/mimo-cli/mimo-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	rootCmd = &cobra.Command{
		Use:   "mimo",
		Short: "MiMo CLI — Your AI pair programmer in the terminal",
		Long: `MiMo CLI is an AI-powered command-line development tool.
Use natural language to describe your goals, and MiMo will autonomously
complete code writing, file operations, project building, testing, and debugging.`,
		SilenceUsage: true,
	}
)

func init() {
	// Load config on startup
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Persistent flags
	rootCmd.PersistentFlags().StringP("model", "m", "", "Model to use (overrides config)")
	rootCmd.PersistentFlags().StringP("safety", "s", "", "Safety level: lockdown|confirm|auto")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text|json|markdown")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
