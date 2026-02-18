package jobs

import (
	"testing"
	"time"
)

func TestNextBackoff(t *testing.T) {
	interval, next := nextBackoff(0)
	if interval != initialTaggerBackoff {
		t.Fatalf("expected interval %s, got %s", initialTaggerBackoff, interval)
	}
	if next != 2*time.Second {
		t.Fatalf("expected next 2s, got %s", next)
	}

	interval, next = nextBackoff(next)
	if interval != 2*time.Second {
		t.Fatalf("expected interval 2s, got %s", interval)
	}
	if next != 4*time.Second {
		t.Fatalf("expected next 4s, got %s", next)
	}

	interval, next = nextBackoff(maxTaggerBackoff)
	if interval != maxTaggerBackoff {
		t.Fatalf("expected interval max %s, got %s", maxTaggerBackoff, interval)
	}
	if next != maxTaggerBackoff {
		t.Fatalf("expected next capped at max %s, got %s", maxTaggerBackoff, next)
	}
}

func TestDefaultInterval(t *testing.T) {
	job := &TaggerJob{df: 0}
	if got := job.defaultInterval(); got != initialTaggerBackoff {
		t.Fatalf("expected fallback interval %s, got %s", initialTaggerBackoff, got)
	}

	job.df = 150 * time.Millisecond
	if got := job.defaultInterval(); got != 150*time.Millisecond {
		t.Fatalf("expected configured interval 150ms, got %s", got)
	}
}
