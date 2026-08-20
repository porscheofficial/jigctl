package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/patricebouillet/jigctl/internal/hcr"
	"github.com/patricebouillet/jigctl/internal/runner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	runAllowExec bool
	runStrict    bool
	runPlain     bool
	runNoColor   bool
)

var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Execute HCR records in the tree",
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

		plan, diagnostics, err := hcr.ExecutionPlan(root, time.Now())
		if err != nil {
			return fmt.Errorf("execution plan for %s: %w", root, err)
		}

		if len(diagnostics) > 0 {
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
			recordedExitCode = 1
			return nil
		}

		authorized := runAllowExec || os.Getenv("JIGCTL_ALLOW_EXEC") == "1"

		verdicts := runner.EvaluatePlan(plan, authorized)
		rows := runner.BuildRows(&plan, verdicts)

		enableColor := shouldEnableColor(colorDecisionInputs{
			IsTerminal:  term.IsTerminal(int(os.Stdout.Fd())),
			FlagNoColor: runNoColor,
			FlagPlain:   runPlain,
			LookupEnv:   os.LookupEnv,
		})

		var renderErr error
		if runPlain {
			renderErr = runner.RenderPlain(os.Stdout, rows)
		} else {
			renderOpts := runner.RenderOptions{
				Out:               os.Stdout,
				Plan:              &plan,
				Verdicts:          verdicts,
				NormalizeDuration: false,
				Style:             runner.Style{Colour: enableColor},
			}
			renderErr = runner.Render(renderOpts)
		}

		if renderErr != nil {
			return fmt.Errorf("render output: %w", renderErr)
		}

		recordedExitCode = runner.AggregateExitCode(runner.ExitSummaries(rows), runStrict)
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&runAllowExec, "allow-exec", false, "Authorize execution of command bindings")
	runCmd.Flags().BoolVar(&runStrict, "strict", false, "Promote expected-unchecked bindings to failure")
	runCmd.Flags().BoolVar(&runPlain, "plain", false, "Render plain one-line-per-record output")
	runCmd.Flags().BoolVar(&runNoColor, "no-color", false, "Disable color output")
	rootCmd.AddCommand(runCmd)
}
