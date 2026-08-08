package filter

import (
	"testing"

	"github.com/b92c/gowatch/internal/docker"
	"github.com/b92c/gowatch/pkg/metrics"
)

func TestFilterMinLogLevel(t *testing.T) {
	fs := NewFilterState()
	if fs.Active {
		t.Errorf("expected initial FilterState to be inactive")
	}

	fs.SetMinLogLevel(metrics.LogLevelWarn)
	if !fs.Active {
		t.Errorf("expected FilterState to be active after SetMinLogLevel")
	}
	if fs.MinLogLevel != metrics.LogLevelWarn {
		t.Errorf("expected MinLogLevel = LogLevelWarn, got %v", fs.MinLogLevel)
	}

	fs.CycleMinLogLevel() // Warn -> Error
	if fs.MinLogLevel != metrics.LogLevelError {
		t.Errorf("expected MinLogLevel = LogLevelError after cycle, got %v", fs.MinLogLevel)
	}

	fs.CycleMinLogLevel() // Error -> Unknown
	if fs.MinLogLevel != metrics.LogLevelUnknown {
		t.Errorf("expected MinLogLevel = LogLevelUnknown after cycle, got %v", fs.MinLogLevel)
	}

	fs.CycleMinLogLevel() // Unknown -> Info
	if fs.MinLogLevel != metrics.LogLevelInfo {
		t.Errorf("expected MinLogLevel = LogLevelInfo after cycle, got %v", fs.MinLogLevel)
	}
}

func TestFilterContainersLogLevel(t *testing.T) {
	cnts := docker.Containers{
		C: []docker.Container{
			{
				ID:      "123456789012",
				Service: "web",
				Log: []string{
					"[INFO] App started",
					"[WARN] High latency",
					"[ERROR] Database error",
					"[DEBUG] Processing request",
				},
			},
		},
	}

	fs := NewFilterState()
	fs.SetMinLogLevel(metrics.LogLevelWarn)

	filtered := FilterContainers(cnts, fs)

	if len(filtered.C) != 1 {
		t.Fatalf("expected 1 container, got %d", len(filtered.C))
	}

	if len(filtered.FlatLogs) != 2 {
		t.Fatalf("expected 2 flat logs (WARN and ERROR), got %d", len(filtered.FlatLogs))
	}

	if filtered.FlatLogs[0].Level != metrics.LogLevelWarn {
		t.Errorf("expected first log to be WARN, got %v", filtered.FlatLogs[0].Level)
	}
	if filtered.FlatLogs[1].Level != metrics.LogLevelError {
		t.Errorf("expected second log to be ERROR, got %v", filtered.FlatLogs[1].Level)
	}
}
