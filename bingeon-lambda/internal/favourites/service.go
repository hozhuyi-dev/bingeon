package favourites

import (
	"context"
	"errors"
	"time"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/tmdb"
)

// Store keeps a user's favourites (store.go).
type Store interface {
	// Stats reads the user's stats, consistently; zero for a new user.
	Stats(ctx context.Context, user string) (Stats, error)
	// Get reads one favourite, consistently.
	Get(ctx context.Context, user string, id int64) (Favourite, bool, error)
	// List reads every favourite, in no order.
	List(ctx context.Context, user string) ([]Favourite, error)
	// Add stores a favourite and next, on condition the stats are still at
	// read's version; ErrConflict if not.
	Add(ctx context.Context, user string, read, next Stats, f Favourite) error
	// Remove deletes a favourite and stores next, on condition the stats are
	// still at read's version; ErrConflict if not.
	Remove(ctx context.Context, user string, read, next Stats, id int64) error
	// DeleteAll deletes the user's favourites and stats.
	DeleteAll(ctx context.Context, user string) error
}

// Movies is TMDB, where a favourite's title, year and poster come from.
type Movies interface {
	Movie(ctx context.Context, id int64, full bool) (tmdb.Movie, error)
}

type Service struct {
	Store  Store
	Movies Movies
	Now    func() time.Time
}

// attempts is how many times a change tries before giving up on others
// going first.
const attempts = 3

// State is the user's stats and what they may do now.
type State struct {
	Stats       Stats
	CanAdd      apperr.Verdict
	CanRemove   apperr.Verdict
	ChangesLeft int
}

func (s *Service) State(ctx context.Context, user string) (State, error) {
	st, err := s.Store.Stats(ctx, user)
	if err != nil {
		return State{}, err
	}
	now := s.Now()
	return State{Stats: st, CanAdd: CanAdd(st, now), CanRemove: CanRemove(st, now), ChangesLeft: st.ChangesLeft(now)}, nil
}

// List is the user's favourites, newest first.
func (s *Service) List(ctx context.Context, user string) ([]Favourite, error) {
	favs, err := s.Store.List(ctx, user)
	if err != nil {
		return nil, err
	}
	Newest(favs)
	return favs, nil
}

// Add makes a movie a favourite; one already is stays as it was.
func (s *Service) Add(ctx context.Context, user string, id int64) (Favourite, error) {
	if id <= 0 {
		return Favourite{}, apperr.Invalid("No such movie.")
	}
	var movie *tmdb.Movie
	for range attempts {
		st, err := s.Store.Stats(ctx, user)
		if err != nil {
			return Favourite{}, err
		}
		f, has, err := s.Store.Get(ctx, user, id)
		if err != nil || has {
			return f, err
		}
		now := s.Now()
		if err := CanAdd(st, now).Err(); err != nil {
			return Favourite{}, err
		}
		if movie == nil {
			m, err := s.Movies.Movie(ctx, id, false)
			if err != nil {
				return Favourite{}, err
			}
			movie = &m
		}
		f = Favourite{ID: id, Title: movie.Title, Year: tmdb.Year(movie.ReleaseDate), Poster: movie.PosterPath, Added: now}
		err = s.Store.Add(ctx, user, st, st.Added(now), f)
		if !errors.Is(err, ErrConflict) {
			return f, err
		}
	}
	return Favourite{}, apperr.Conflict("Your favourites changed at the same moment. Try again.")
}

// Remove makes a movie no longer a favourite; one that isn't is fine.
func (s *Service) Remove(ctx context.Context, user string, id int64) error {
	for range attempts {
		st, err := s.Store.Stats(ctx, user)
		if err != nil {
			return err
		}
		_, has, err := s.Store.Get(ctx, user, id)
		if err != nil || !has {
			return err
		}
		now := s.Now()
		if err := CanRemove(st, now).Err(); err != nil {
			return err
		}
		err = s.Store.Remove(ctx, user, st, st.Removed(now), id)
		if !errors.Is(err, ErrConflict) {
			return err
		}
	}
	return apperr.Conflict("Your favourites changed at the same moment. Try again.")
}

// DeleteAll forgets the user's favourites.
func (s *Service) DeleteAll(ctx context.Context, user string) error {
	return s.Store.DeleteAll(ctx, user)
}
