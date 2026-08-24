package main

import (
	"fmt"
	"io"

	"github.com/porscheofficial/jigctl/internal/runner"
)

func render(out io.Writer, rows []runner.Row, screen tty, style runner.Style, format runner.Format, root string) error {
	if format == runner.FormatJSON {
		exitCode := runner.AggregateExitCode(runner.ExitSummaries(rows), runStrict)
		err := runner.RenderJSON(out, runner.JSONOptions{
			Root:         root,
			Rows:         rows,
			ExitCode:     exitCode,
			OnlyFailures: runOnlyFailures,
		})
		if err != nil {
			return fmt.Errorf("render json: %w", err)
		}
		return nil
	}

	if runPlain || format == runner.FormatPlain {
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
