package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

var (
	// at this time Railway only supports JSON log messages if they have a "message" key
	replacer = func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == "msg" {
			a.Key = "message"
		}
		return a
	}

	stdoutHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: replacer,
	})
	//enable source
	stdoutHandlerWithSource = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replacer,
	})

	stderrHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		ReplaceAttr: replacer,
	})
	// enable source
	stderrHandlerWithSource = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replacer,
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
