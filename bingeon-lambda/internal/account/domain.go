// Package account is the signed-in user's account as the web shows it:
// their favourites and picks, what they may do now, and deleting it all.
// It keeps nothing of its own: it composes the favourites and picks slices.
package account

import (
	"time"

	"bingeon-lambda/internal/favourites"
	"bingeon-lambda/internal/picks"
	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/day"
)

// Can is what the user may do now: each verdict comes from the rule the
// action itself checks, so the web offers only what the API would take.
type Can struct {
	Add    apperr.Verdict `json:"add"`
	Remove apperr.Verdict `json:"remove"`
	Pick   apperr.Verdict `json:"pick"`
}

// Today is what is left of the user's day, and when it resets (midnight
// Pacific time, shown in the user's own).
type Today struct {
	PicksLeft   int       `json:"picksLeft"`
	ChangesLeft int       `json:"changesLeft"`
	Resets      time.Time `json:"resets"`
}

// Status is what the user may do now, which every mutation's reply carries.
type Status struct {
	Can   Can   `json:"can"`
	Today Today `json:"today"`
}

type View struct {
	Favourites []favourites.Favourite `json:"favourites"`
	Picks      picks.View             `json:"picks"`
	Status
}

// StatusOf is the status from each slice's state.
func StatusOf(f favourites.State, p picks.State, now time.Time) Status {
	return Status{
		Can:   Can{Add: f.CanAdd, Remove: f.CanRemove, Pick: p.CanPick},
		Today: Today{PicksLeft: p.Left, ChangesLeft: f.ChangesLeft, Resets: day.End(now)},
	}
}
