package backoff

import (
	"log/slog"
	"os"
	"time"

	"github.com/brody192/logger"
	"github.com/sethvargo/go-retry"
)

// BackoffParams holds the parameters for the exponential backoff
type BackoffParams struct {
	// InitialInterval is the starting interval for the backoff
	InitialInterval time.Duration
	// MaxInterval is the upper limit of the backoff duration
	MaxInterval time.Duration
	// After MaxElapsedTime the ExponentialBackOff stops. It never stops if MaxElapsedTime == 0.
	MaxElapsedTime time.Duration
}

// GetBackoffParams reads backoff parameters from environment variables
// or returns default values if not set
func GetBackoffParams() retry.Backoff {
	params := BackoffParams{
		InitialInterval: time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  2 * time.Minute,
	}

	if initialIntervalStr := os.Getenv("BACKOFF_INITIAL_INTERVAL"); initialIntervalStr != "" {
		if val, err := time.ParseDuration(initialIntervalStr); err == nil && val > 0 {
			params.InitialInterval = val
		}
	}

	if maxIntervalStr := os.Getenv("BACKOFF_MAX_INTERVAL"); maxIntervalStr != "" {
		if val, err := time.ParseDuration(maxIntervalStr); err == nil && val > 0 {
			params.MaxInterval = val
		}
	}

	if maxElapsedTimeStr := os.Getenv("BACKOFF_MAX_ELAPSED_TIME"); maxElapsedTimeStr != "" {
		if val, err := time.ParseDuration(maxElapsedTimeStr); err == nil && val > 0 {
			params.MaxElapsedTime = val
		}
	}

	logger.Stdout.Info("backoff params:",
		slog.Duration("initial_interval", params.InitialInterval),
		slog.Duration("max_interval", params.MaxInterval),
		slog.Duration("max_elapsed_time", params.MaxElapsedTime),
	)

	b := retry.NewExponential(params.InitialInterval)
	b = retry.WithCappedDuration(params.MaxInterval, b)
	b = retry.WithMaxDuration(params.MaxElapsedTime, b)

	return b
}
