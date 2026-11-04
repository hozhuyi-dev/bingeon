package picks

import (
	"testing"
	"time"

	"bingeon-lambda/internal/platform/day"
)

var now = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

func TestCanPick(t *testing.T) {
	today := Picks{Day: day.Of(now)}
	spent := Picks{Day: day.Of(now), Made: PerDay}
	yesterday := Picks{Day: "2026-11-01", Made: PerDay}
	cases := []struct {
		name       string
		p          Picks
		favourites int
		used       int
		ok         bool
		frees      bool
	}{
		{"with favourites", today, 3, 0, true, false},
		{"without favourites", today, 0, 0, false, false},
		{"after today's 20", spent, 3, 0, false, true},
		{"after yesterday's 20", yesterday, 3, 0, true, false},
		{"with everyone's calls used", today, 3, 50, false, true},
	}
	for _, c := range cases {
		v := CanPick(c.p, now, c.favourites, c.used, 50)
		if v.OK != c.ok || (v.Until != nil) != c.frees {
			t.Errorf("%s: %+v; want ok %v, freeing up %v", c.name, v, c.ok, c.frees)
		}
		if c.frees && !v.Until.Equal(day.End(now)) {
			t.Errorf("%s frees up at %v, want the day's end", c.name, v.Until)
		}
	}
}

func TestARefundIsOnlyForToday(t *testing.T) {
	p := Picks{}.Reserved(now)
	if p.Made != 1 || p.Refunded(now).Made != 0 {
		t.Errorf("reserved %d, refunded %d; want 1, then 0", p.Made, p.Refunded(now).Made)
	}
	if got := p.Refunded(now.Add(24 * time.Hour)).Made; got != 1 {
		t.Errorf("refunded tomorrow: %d, want yesterday's count left alone", got)
	}
}
