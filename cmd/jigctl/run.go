package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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
		return runAction(args, describeStdout(), os.Stdout)
	},
}

// tty is everything the renderer is told about its destination. The
// detection happens here and nowhere else: internal/runner never learns what
// a terminal is (ADR-0013).
type tty struct {
	IsTerminal bool
	Width      int
	Height     int
}

func describeStdout() tty {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return tty{}
	}
	width, height, err := term.GetSize(fd)
	if err != nil {
		return tty{IsTerminal: true}
	}
	return tty{IsTerminal: true, Width: width, Height: height}
}

func findRoot(args []string) (string, error) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("canonicalize path %s: %w", path, err)
	}

	current := absPath
	for {
		if _, statErr := os.Stat(filepath.Join(current, "jig.toml")); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("no jig.toml found in %s or any parent", path)
}

func runAction(args []string, screen tty, out io.Writer) error {
	root, err := findRoot(args)
	if err != nil {
		return err
	}

	plan, diagnostics, err := hcr.ExecutionPlan(root, time.Now())
	if err != nil {
		return fmt.Errorf("execution plan for %s: %w", root, err)
	}

	if len(diagnostics) > 0 {
		printDiagnostics(out, root, diagnostics)
		recordedExitCode = 1
		return nil
	}

	authorized := runAllowExec || os.Getenv("JIGCTL_ALLOW_EXEC") == "1"
	style := runner.Style{Colour: shouldEnableColor(colorDecisionInputs{
		IsTerminal:  screen.IsTerminal,
		FlagNoColor: runNoColor,
		FlagPlain:   runPlain,
		LookupEnv:   os.LookupEnv,
	})}

	rows := evaluate(&plan, authorized, liveOptions(out, &plan, screen, style))

	if renderErr := render(out, rows, screen, style); renderErr != nil {
		return renderErr
	}

	recordedExitCode = runner.AggregateExitCode(runner.ExitSummaries(rows), runStrict)
	return nil
}

func printDiagnostics(out io.Writer, root string, diagnostics []hcr.Diagnostic) {
	for _, d := range diagnostics {
		relPath, relErr := filepath.Rel(root, d.File)
		if relErr != nil {
			relPath = d.File
		}
		ptr := d.Pointer
		if ptr == "" {
			ptr = "\"\""
		}
		fmt.Fprintf(out, "%s:%s: %s: %s\n", relPath, ptr, d.Code, d.Message)
	}
}

// liveOptions returns the zero LiveOptions when a live view has no business
// running: a pipe has no cursor to move, --plain is the shape scripts pin to,
// and a dumb terminal cannot honour the escapes the view is painted with.
func liveOptions(out io.Writer, plan *hcr.Plan, screen tty, style runner.Style) runner.LiveOptions {
	if !screen.IsTerminal || runPlain || os.Getenv("TERM") == "dumb" {
		return runner.LiveOptions{}
	}
	return runner.LiveOptions{
		Out:    out,
		Plan:   plan,
		Style:  style,
		Width:  screen.Width,
		Height: screen.Height,
	}
}

func evaluate(plan *hcr.Plan, authorized bool, live runner.LiveOptions) []runner.Row {
	var progress runner.Progress
	var view *runner.LiveView
	if v, ok := runner.NewLiveView(live); ok {
		view, progress = v, v
	}
	defer restoreOnSignal(view)()

	verdicts := runner.EvaluatePlanWithProgress(*plan, authorized, progress)
	rows := runner.BuildRows(plan, verdicts)

	if view != nil {
		view.Close()
	}
	return rows
}

// restoreOnSignal gives the cursor back when a run is killed midway. The live
// view hides it for the duration of the paint, so without this a Ctrl-C during
// a slow binding leaves the shell with no cursor until the user types `reset`.
// It returns the func that retires the handler.
func restoreOnSignal(view *runner.LiveView) func() {
	if view == nil {
		return func() {}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	finished := make(chan struct{})

	go func() {
		select {
		case received := <-signals:
			view.Close()
			reraise(received)
		case <-finished:
		}
	}()

	return func() {
		close(finished)
		signal.Stop(signals)
	}
}

// reraise redelivers a signal with its default disposition restored, so the
// process dies of the signal the user actually sent. Exiting here instead
// would give the CLI a second exit path, which HCR-0406 forbids: main.go is
// the only place permitted to turn a result into a process status.
func reraise(received os.Signal) {
	signal.Reset(os.Interrupt, syscall.SIGTERM)

	signum, ok := received.(syscall.Signal)
	if !ok {
		signum = syscall.SIGINT
	}
	if killErr := syscall.Kill(syscall.Getpid(), signum); killErr != nil {
		fmt.Fprintf(os.Stderr, "jigctl: could not re-raise %s: %v\n", received, killErr)
	}
}

func render(out io.Writer, rows []runner.Row, screen tty, style runner.Style) error {
	if runPlain {
		if err := runner.RenderPlain(out, rows); err != nil {
			return fmt.Errorf("render plain output: %w", err)
		}
		return nil
	}
	err := runner.Render(runner.RenderOptions{
		Out:               out,
		Rows:              rows,
		NormalizeDuration: false,
		Width:             screen.Width,
		Style:             style,
	})
	if err != nil {
		return fmt.Errorf("render output: %w", err)
	}
	return nil
}

func init() {
	runCmd.Flags().BoolVar(&runAllowExec, "allow-exec", false, "Authorize execution of command bindings")
	runCmd.Flags().BoolVar(&runStrict, "strict", false, "Promote expected-unchecked bindings to failure")
	runCmd.Flags().BoolVar(&runPlain, "plain", false, "Render plain one-line-per-record output")
	runCmd.Flags().BoolVar(&runNoColor, "no-color", false, "Disable color output")
	rootCmd.AddCommand(runCmd)
}
