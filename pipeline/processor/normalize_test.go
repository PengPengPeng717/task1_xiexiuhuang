package processor_test

import (
	"testing"
	"time"

	"mining-pipeline/pipeline/processor"
)

func TestWithinLast30Days(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, processor.ShanghaiLoc)
	old := now.AddDate(0, 0, -40)
	recent := now.AddDate(0, 0, -5)
	if processor.WithinLast30Days(old, now) {
		t.Fatal("old should be excluded")
	}
	if !processor.WithinLast30Days(recent, now) {
		t.Fatal("recent should be included")
	}
}
