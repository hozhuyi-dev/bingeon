// Package favourites is a signed-in user's favourite movies: a heart on
// any movie, kept with its title, year and poster so the list needs no
// TMDB call.
package favourites

import (
	"cmp"
	"errors"
	"slices"
	"time"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/day"
)

const (
	// MaxFavourites keeps a user's list one query page (about 75 KB).
	MaxFavourites = 500
	// MaxChangesPerDay stops a script; no one hearts a thousand movies a day.
	MaxChangesPerDay = 1000
)

// ErrConflict is another change to the same user's favourites going
// first; trying again from a fresh read gets past it.
var ErrConflict = errors.New("favourites: changed meanwhile")

type Favourite struct {
	ID     int64     `json:"id"`
	Title  string    `json:"title"`
	Year   int       `json:"year,omitempty"`
	Poster string    `json:"poster,omitempty"`
	Added  time.Time `json:"added"`
}

// Stats are a user's favourites counted: how many, and the changes made
// on Day. Version moves on with every change.
type Stats struct {
	Count   int
	Day     string
	Changes int
	Version int64
}

// Today is the stats as of now: a new day's changes start from none.
func (s Stats) Today(now time.Time) Stats {
	if today := day.Of(now); s.Day != today {
		s.Day, s.Changes = today, 0
	}
	return s
}

// ChangesLeft are the changes the user can still make today.
func (s Stats) ChangesLeft(now time.Time) int {
	return MaxChangesPerDay - s.Today(now).Changes
}

// CanAdd says whether one more favourite may be added now.
func CanAdd(s Stats, now time.Time) apperr.Verdict {
	switch {
	case s.Count >= MaxFavourites:
		return apperr.No(time.Time{}, "You have 500 favourites, the most there can be. Remove one to add another.")
	case s.ChangesLeft(now) <= 0:
		return apperr.No(day.End(now), "You've changed your favourites 1,000 times today.")
	}
	return apperr.Allowed
}

// CanRemove says whether a favourite may be removed now.
func CanRemove(s Stats, now time.Time) apperr.Verdict {
	if s.ChangesLeft(now) <= 0 {
		return apperr.No(day.End(now), "You've changed your favourites 1,000 times today.")
	}
	return apperr.Allowed
}

// Added is the stats after adding a favourite now.
func (s Stats) Added(now time.Time) Stats {
	s = s.Today(now)
	s.Count++
	s.Changes++
	s.Version++
	return s
}

// Removed is the stats after removing a favourite now.
func (s Stats) Removed(now time.Time) Stats {
	s = s.Today(now)
	s.Count = max(0, s.Count-1)
	s.Changes++
	s.Version++
	return s
}

// Newest sorts favourites newest first.
func Newest(favs []Favourite) {
	slices.SortFunc(favs, func(a, b Favourite) int {
		return cmp.Or(b.Added.Compare(a.Added), cmp.Compare(b.ID, a.ID))
	})
}
