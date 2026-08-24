package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/porscheofficial/jigctl/internal/hcr"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate HCR records in the tree",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("canonicalize path %s: %w", path, err)
		}

		var root string
		current := absPath
		for {
			if _, statErr := os.Stat(filepath.Join(current, "jig.toml")); statErr == nil {
				root = current
				break
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}

		if root == "" {
			return fmt.Errorf("no jig.toml found in %s or any parent", path)
		}

		diagnostics, err := hcr.ValidateTree(root)
		if err != nil {
			return fmt.Errorf("validate tree at %s: %w", root, err)
		}

		for _, d := range diagnostics {
			relPath, relErr := filepath.Rel(root, d.File)
			if relErr != nil {
				relPath = d.File
			}
			ptr := d.Pointer
			if ptr == "" {
				ptr = "\"\""
			}
			fmt.Fprintf(os.Stdout, "%s:%s: %s: %s\n", relPath, ptr, d.Code, d.Message)
		}

		if len(diagnostics) == 0 {
			fmt.Fprintln(os.Stdout, "no findings")
		} else {
			wordFindings := "findings"
			if len(diagnostics) == 1 {
				wordFindings = "finding"
			}

			filesMap := make(map[string]struct{})
			for _, d := range diagnostics {
				filesMap[d.File] = struct{}{}
			}
			numFiles := len(filesMap)
			wordFiles := "files"
			if numFiles == 1 {
				wordFiles = "file"
			}

			fmt.Fprintf(os.Stdout, "%d %s in %d %s\n", len(diagnostics), wordFindings, numFiles, wordFiles)
			recordedExitCode = 1
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
