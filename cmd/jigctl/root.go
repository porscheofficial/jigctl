package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var recordedExitCode int

var rootCmd = &cobra.Command{
	Use:           "jigctl",
	Short:         "A constraint harness for polyglot monorepos",
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() int {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "jigctl: %v\n", err)
		return 2
	}
	return recordedExitCode
}
