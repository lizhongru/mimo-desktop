package cmd

import (
	"fmt"
	"runtime"

	"github.com/mimo-cli/mimo-cli/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("MiMo CLI v%s\n", version.Version)
		fmt.Printf("  Git Commit: %s\n", version.GitCommit)
		fmt.Printf("  Build Date: %s\n", version.BuildDate)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
