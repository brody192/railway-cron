package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

var (
	stdoutHandler = slog.NewTextHandler(os.Stdout, nil)
	//enable source
	stdoutHandlerWithSource = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})

	stderrHandler = slog.NewTextHandler(os.Stderr, nil)
	// enable source
	stderrHandlerWithSource = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	})

	// sends logs to stdout
	Stdout = slog.New(stdoutHandler)
	// sends logs to stdout with source info
	StdoutWithSource = slog.New(stdoutHandlerWithSource)

	// sends logs to stderr
	Stderr = slog.New(stderrHandler)
	// sends logs to stderr with source info
	StderrWithSource = slog.New(stderrHandlerWithSource)
)
