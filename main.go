//go:build !wails

package main

import (
	"errors"
	"os"

	"github.com/mimo-cli/mimo-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// 语义化退出码：检查是否是 runExitError
		var exitCoder interface{ ExitCode() int }
		if errors.As(err, &exitCoder) {
			os.Exit(exitCoder.ExitCode())
		}
		os.Exit(1)
	}
}
