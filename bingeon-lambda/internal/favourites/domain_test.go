package favourites

import (
	"testing"
	"time"

	"bingeon-lambda/internal/platform/day"
)

var now = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

func TestANewDayStartsWithNoChanges(t *testing.T) {
	s := Stats{Count: 3, Day: "2026-11-01", Changes: MaxChangesPerDay}
	if got := s.ChangesLeft(now); got != MaxChangesPerDay {
		t.Errorf("changes left = %d, want %d", got, MaxChangesPerDay)
	}
	if next := s.Added(now); next.Day != day.Of(now) || next.Changes != 1 || next.Count != 4 || next.Version != 1 {
		t.Errorf("Added = %+v", next)
	}
}

func TestLimits(t *testing.T) {
	full := Stats{Count: MaxFavourites}
	if v := CanAdd(full, now); v.OK || v.Until != nil {
		t.Errorf("CanAdd at 500 = %+v; want refused, not freeing by itself", v)
	}
	if v := CanRemove(full, now); !v.OK {
		t.Errorf("CanRemove at 500 = %+v; want allowed", v)
	}
	busy := Stats{Count: 1, Day: day.Of(now), Changes: MaxChangesPerDay}
	if CanAdd(busy, now).OK || CanRemove(busy, now).OK {
		t.Error("a change allowed after 1,000 today")
	}
	if v := CanAdd(busy, now); v.Until == nil || !v.Until.Equal(day.End(now)) {
		t.Errorf("CanAdd frees up at %v, want the day's end", v.Until)
	}
}

func TestRemovingNeverCountsBelowNone(t *testing.T) {
	if got := (Stats{}).Removed(now); got.Count != 0 {
		t.Errorf("count = %d", got.Count)
	}
}
