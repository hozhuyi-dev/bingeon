package picks

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/day"
	"bingeon-lambda/internal/platform/gemini"
	"bingeon-lambda/internal/platform/tmdb"
)

// Store keeps a user's picks and the day's Gemini calls (store.go).
type Store interface {
	// Get reads the user's picks, consistently; zero for a new user.
	Get(ctx context.Context, user string) (Picks, error)
	// Reserve stores next and counts one of the day's Gemini calls, as one:
	// ErrConflict if the picks aren't at read's version, ErrQuotaFull if
	// the day's calls reached limit.
	Reserve(ctx context.Context, user string, read, next Picks, day string, limit int) error
	// Save stores next, on condition the picks are still at read's
	// version; ErrConflict if not.
	Save(ctx context.Context, user string, read, next Picks) error
	// QuotaUsed reads how many of the day's Gemini calls are used.
	QuotaUsed(ctx context.Context, day string) (int, error)
	// Delete deletes the user's picks.
	Delete(ctx context.Context, user string) error
}

// Favourites are the user's favourites, newest first (the favourites
// slice, joined in internal/app).
type Favourites interface {
	Seeds(ctx context.Context, user string) ([]Seed, error)
}

// Picker is Gemini (gemini.Client).
type Picker interface {
	Suggest(ctx context.Context, liked, avoid []gemini.Title, n int) ([]gemini.Suggestion, error)
}

// Movies is TMDB, where each suggestion is found.
type Movies interface {
	Search(ctx context.Context, query string, year, page int) (tmdb.Page, error)
}

type Service struct {
	Store      Store
	Favourites Favourites
	Picker     Picker
	Movies     Movies
	// DailyLimit is the day's Gemini calls across all users
	// (GEMINI_DAILY_LIMIT).
	DailyLimit int
	Now        func() time.Time
}

// attempts is how many times a change tries before giving up on others
// going first.
const attempts = 3

// State is the user's picks and whether they may make more now.
type State struct {
	Picks   Picks
	CanPick apperr.Verdict
	Left    int
}

// State is the user's, who has favourites favourites.
func (s *Service) State(ctx context.Context, user string, favourites int) (State, error) {
	p, err := s.Store.Get(ctx, user)
	if err != nil {
		return State{}, err
	}
	now := s.Now()
	used, err := s.Store.QuotaUsed(ctx, day.Of(now))
	if err != nil {
		return State{}, err
	}
	return State{Picks: p, CanPick: CanPick(p, now, favourites, used, s.DailyLimit), Left: p.Left(now)}, nil
}

// Make asks Gemini for movies like the user's favourites and keeps those
// found in TMDB as the user's last picks. A failure gives the pick back.
func (s *Service) Make(ctx context.Context, user string) (Picks, error) {
	seeds, err := s.Favourites.Seeds(ctx, user)
	if err != nil {
		return Picks{}, err
	}
	reserved, err := s.reserve(ctx, user, len(seeds))
	if err != nil {
		return Picks{}, err
	}
	picks, err := s.pick(ctx, seeds, reserved.Last)
	if err != nil {
		return Picks{}, errors.Join(err, s.refund(ctx, user))
	}
	return s.change(ctx, user, func(p Picks) Picks { return p.With(picks, s.Now()) })
}

// reserve counts the pick, the user's and the day's, before Gemini is
// called: two at once can't both take the last one.
func (s *Service) reserve(ctx context.Context, user string, favourites int) (Picks, error) {
	for range attempts {
		p, err := s.Store.Get(ctx, user)
		if err != nil {
			return Picks{}, err
		}
		now := s.Now()
		used, err := s.Store.QuotaUsed(ctx, day.Of(now))
		if err != nil {
			return Picks{}, err
		}
		if err := CanPick(p, now, favourites, used, s.DailyLimit).Err(); err != nil {
			return Picks{}, err
		}
		next := p.Reserved(now)
		err = s.Store.Reserve(ctx, user, p, next, day.Of(now), s.DailyLimit)
		switch {
		case errors.Is(err, ErrQuotaFull):
			return Picks{}, CanPick(p, now, favourites, s.DailyLimit, s.DailyLimit).Err()
		case !errors.Is(err, ErrConflict):
			return next, err
		}
	}
	return Picks{}, apperr.Conflict("Your picks changed at the same moment. Try again.")
}

func (s *Service) refund(ctx context.Context, user string) error {
	_, err := s.change(ctx, user, func(p Picks) Picks { return p.Refunded(s.Now()) })
	return err
}

// change applies fn to the user's picks as they are now, and stores it.
func (s *Service) change(ctx context.Context, user string, fn func(Picks) Picks) (Picks, error) {
	for range attempts {
		p, err := s.Store.Get(ctx, user)
		if err != nil {
			return Picks{}, err
		}
		next := fn(p)
		if err := s.Store.Save(ctx, user, p, next); !errors.Is(err, ErrConflict) {
			return next, err
		}
	}
	return Picks{}, apperr.Conflict("Your picks changed at the same moment. Try again.")
}

// pick asks Gemini for movies like the newest seeds, avoiding them and the
// last picks, and finds each in TMDB at once. It keeps the first Size
// found that aren't favourites, in Gemini's order.
func (s *Service) pick(ctx context.Context, seeds []Seed, last []Pick) ([]Pick, error) {
	liked := make([]gemini.Title, 0, min(Seeds, len(seeds)))
	for _, f := range seeds[:min(Seeds, len(seeds))] {
		liked = append(liked, gemini.Title{Title: f.Title, Year: f.Year})
	}
	avoid := make([]gemini.Title, 0, len(last))
	for _, p := range last {
		avoid = append(avoid, gemini.Title{Title: p.Title, Year: p.Year})
	}
	suggestions, err := s.Picker.Suggest(ctx, liked, avoid, Ask)
	if err != nil {
		return nil, err
	}

	found := make([]*Pick, len(suggestions))
	errs := make([]error, len(suggestions))
	var wg sync.WaitGroup
	for i, sg := range suggestions {
		wg.Go(func() { found[i], errs[i] = s.find(ctx, sg) })
	}
	wg.Wait()

	skip := map[int64]bool{}
	for _, f := range seeds {
		skip[f.ID] = true
	}
	var out []Pick
	for _, p := range found {
		if p == nil || skip[p.ID] || len(out) == Size {
			continue
		}
		skip[p.ID] = true
		out = append(out, *p)
	}
	if len(out) == 0 {
		if err := errors.Join(errs...); err != nil {
			return nil, err
		}
		return nil, apperr.Unavailable("No picks this time: none of the suggestions were found. Try again.")
	}
	return out, nil
}

// find is a suggestion as TMDB has it: the top match of its title in its
// year; nil when there is none.
func (s *Service) find(ctx context.Context, sg gemini.Suggestion) (*Pick, error) {
	if strings.TrimSpace(sg.Title.Title) == "" {
		return nil, nil
	}
	page, err := s.Movies.Search(ctx, sg.Title.Title, sg.Year, 1)
	if err != nil || len(page.Results) == 0 {
		return nil, err
	}
	m := page.Results[0]
	return &Pick{ID: m.ID, Title: m.Title, Year: tmdb.Year(m.ReleaseDate), Poster: m.PosterPath, Reason: sg.Reason}, nil
}

// Delete forgets the user's picks.
func (s *Service) Delete(ctx context.Context, user string) error {
	return s.Store.Delete(ctx, user)
}
