// Package time provides date, timestamp, and duration functionality for NilLang.
package time

import (
	"time"
)

// NowMs returns current Unix timestamp in milliseconds.
func NowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// NowSec returns current Unix timestamp in seconds.
func NowSec() int64 {
	return time.Now().Unix()
}

// Format returns current time formatted with the specified layout (default ISO 8601).
func Format(layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}
	return time.Now().Format(layout)
}

// Sleep pauses execution for the given number of milliseconds.
func Sleep(ms int64) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// SinceMs returns elapsed time in milliseconds since the given start timestamp (ms).
func SinceMs(startMs int64) int64 {
	return NowMs() - startMs
}
