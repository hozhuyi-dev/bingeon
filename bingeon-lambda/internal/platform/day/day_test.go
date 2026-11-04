package day

import (
	"testing"
	"time"
)

func TestDayIsPacific(t *testing.T) {
	// 07:59 UTC on 4 Nov is 23:59 on 3 Nov in Los Angeles (PST, UTC-8).
	before := time.Date(2026, 11, 4, 7, 59, 0, 0, time.UTC)
	after := before.Add(time.Minute)
	if got := Of(before); got != "2026-11-03" {
		t.Errorf("Of(%v) = %s, want 2026-11-03", before, got)
	}
	if got := Of(after); got != "2026-11-04" {
		t.Errorf("Of(%v) = %s, want 2026-11-04", after, got)
	}
	if got := End(before); !got.Equal(after) {
		t.Errorf("End(%v) = %v, want %v", before, got, after)
	}
}

func TestEndFollowsDaylightSaving(t *testing.T) {
	// PDT (UTC-7) in July: midnight is 07:00 UTC.
	noon := time.Date(2026, 7, 1, 19, 0, 0, 0, time.UTC)
	if got, want := End(noon), time.Date(2026, 7, 2, 7, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("End(%v) = %v, want %v", noon, got, want)
	}
}
