package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/patricebouillet/jigctl/internal/hcr"
	"github.com/patricebouillet/jigctl/internal/runner"
	"github.com/spf13/cobra"
)

var (
	runAllowExec bool
	runStrict    bool
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

		renderOpts := runner.RenderOptions{
			Out:               os.Stdout,
			Plan:              &plan,
			Verdicts:          verdicts,
			NormalizeDuration: false,
		}
		if renderErr := runner.Render(renderOpts); renderErr != nil {
			return fmt.Errorf("render output: %w", renderErr)
		}

		lookupExceptions := make(map[runner.BindingIdentity][]string)
		var knownServicePaths []string

		for i := range plan.Targets {
			t := &plan.Targets[i]
			if t.Kind == "service" {
				knownServicePaths = append(knownServicePaths, t.Path)
			}
			for j := range t.Bindings {
				b := &t.Bindings[j]
				id := runner.BindingIdentity{RecordPath: b.RecordPath, BindingIndex: b.BindingIndex}

				var exceptions []string
				for _, exc := range b.Exceptions {
					exceptions = append(exceptions, exc.Scope)
				}
				lookupExceptions[id] = exceptions
			}
		}

		var summaries []runner.ExitSummary
		for _, v := range verdicts {
			if v == nil {
				continue
			}

			rep := v.Report()
			exceptions := lookupExceptions[rep.Identity]

			var mutatedFindings []runner.Finding
			for _, f := range rep.Findings {
				mut, applyErr := runner.ApplyExceptions(
					f, rep.Kind, exceptions, rep.Identity.RecordPath, knownServicePaths)
				if applyErr == nil {
					mutatedFindings = append(mutatedFindings, mut)
				} else {
					mutatedFindings = append(mutatedFindings, f)
				}
			}

			var proj runner.Projection
			if v.Completion() == runner.CompletionCompleted {
				proj = runner.ProjectionPass
				for _, f := range mutatedFindings {
					if len(f.WaivedBy) == 0 {
						proj = runner.ProjectionViolation
						break
					}
				}
			} else {
				proj, _ = v.Projection()
			}

			summaries = append(summaries, runner.ExitSummary{
				Projection: proj,
				IsBlocking: rep.Severity != "advisory",
			})
		}

		recordedExitCode = runner.AggregateExitCode(summaries, runStrict)
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&runAllowExec, "allow-exec", false, "Authorize execution of command bindings")
	runCmd.Flags().BoolVar(&runStrict, "strict", false, "Promote expected-unchecked bindings to failure")
	rootCmd.AddCommand(runCmd)
}
