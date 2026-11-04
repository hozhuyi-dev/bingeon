// Package day is bingeon's "today": the Pacific-time day, when Gemini's
// free quota resets. Every daily count uses it, so they all reset together.
package day

import (
	"time"
	_ "time/tzdata"
)

var pacific = func() *time.Location {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
	return loc
}()

// Of is the day t is in, as "2026-11-04".
func Of(t time.Time) string {
	return t.In(pacific).Format(time.DateOnly)
}

// End is when the day t is in ends: the next Pacific midnight, in UTC.
func End(t time.Time) time.Time {
	y, m, d := t.In(pacific).Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, pacific).UTC()
}
