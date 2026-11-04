// Package picks is AI picks: Gemini suggests movies like a user's
// favourites, each found in TMDB, kept as the user's last picks.
package picks

import (
	"errors"
	"time"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/day"
)

const (
	// PerDay are the picks a user may make a day: 10 users making all of
	// theirs use half production's Gemini calls (GEMINI_DAILY_LIMIT).
	PerDay = 20
	// Size is how many movies one pick keeps.
	Size = 6
	// Ask is how many Gemini is asked for: a few more than Size, as some
	// aren't found in TMDB or are already favourites.
	Ask = Size + 2
	// Seeds is how many of the newest favourites Gemini is shown: about
	// 4,000 tokens.
	Seeds = 100
)

var (
	// ErrConflict is another change to the user's picks going first.
	ErrConflict = errors.New("picks: changed meanwhile")
	// ErrQuotaFull is the day's Gemini calls, across all users, used up.
	ErrQuotaFull = errors.New("picks: today's Gemini calls are used up")
)

// Pick is a movie Gemini suggested, as found in TMDB.
type Pick struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Year   int    `json:"year,omitempty"`
	Poster string `json:"poster,omitempty"`
	Reason string `json:"reason"`
}

// Seed is a favourite, as picks are made from it.
type Seed struct {
	ID    int64
	Title string
	Year  int
}

// Picks are a user's picks: how many made on Day, and the last ones.
// Version moves on with every change.
type Picks struct {
	Day     string
	Made    int
	Last    []Pick
	MadeAt  time.Time
	Version int64
}

// Today is the picks as of now: a new day's count starts from none.
func (p Picks) Today(now time.Time) Picks {
	if today := day.Of(now); p.Day != today {
		p.Day, p.Made = today, 0
	}
	return p
}

// Left are the picks the user can still make today.
func (p Picks) Left(now time.Time) int {
	return PerDay - p.Today(now).Made
}

// CanPick says whether the user may make picks now, with favourites
// favourites, when quotaUsed of the day's limit of Gemini calls are used.
func CanPick(p Picks, now time.Time, favourites, quotaUsed, limit int) apperr.Verdict {
	switch {
	case favourites == 0:
		return apperr.No(time.Time{}, "Add a favourite first: picks are made from your favourites.")
	case p.Left(now) <= 0:
		return apperr.No(day.End(now), "You've made today's 20 picks.")
	case quotaUsed >= limit:
		return apperr.No(day.End(now), "AI picks are resting: today's are all used up.")
	}
	return apperr.Allowed
}

// Reserved is the picks with one more made now.
func (p Picks) Reserved(now time.Time) Picks {
	p = p.Today(now)
	p.Made++
	p.Version++
	return p
}

// Refunded is the picks with one made now given back, after it failed.
func (p Picks) Refunded(now time.Time) Picks {
	if p.Day == day.Of(now) && p.Made > 0 {
		p.Made--
	}
	p.Version++
	return p
}

// With is the picks with picks as the last ones, made now.
func (p Picks) With(picks []Pick, now time.Time) Picks {
	p.Last, p.MadeAt = picks, now
	p.Version++
	return p
}
