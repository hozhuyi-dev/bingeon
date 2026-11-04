package account

import (
	"context"
	"time"

	"bingeon-lambda/internal/favourites"
	"bingeon-lambda/internal/picks"
)

// Favourites is the favourites slice (favourites.Service).
type Favourites interface {
	State(ctx context.Context, user string) (favourites.State, error)
	List(ctx context.Context, user string) ([]favourites.Favourite, error)
	DeleteAll(ctx context.Context, user string) error
}

// Picks is the picks slice (picks.Service).
type Picks interface {
	State(ctx context.Context, user string, favourites int) (picks.State, error)
	Delete(ctx context.Context, user string) error
}

type Service struct {
	Favourites Favourites
	Picks      Picks
	Now        func() time.Time
}

// Status is what the user may do now: three small reads.
func (s *Service) Status(ctx context.Context, user string) (Status, error) {
	f, err := s.Favourites.State(ctx, user)
	if err != nil {
		return Status{}, err
	}
	p, err := s.Picks.State(ctx, user, f.Stats.Count)
	if err != nil {
		return Status{}, err
	}
	return StatusOf(f, p, s.Now()), nil
}

// Get is the whole account: read once a session, then kept current by each
// mutation's reply.
func (s *Service) Get(ctx context.Context, user string) (View, error) {
	favs, err := s.Favourites.List(ctx, user)
	if err != nil {
		return View{}, err
	}
	f, err := s.Favourites.State(ctx, user)
	if err != nil {
		return View{}, err
	}
	p, err := s.Picks.State(ctx, user, f.Stats.Count)
	if err != nil {
		return View{}, err
	}
	if favs == nil {
		favs = []favourites.Favourite{}
	}
	return View{Favourites: favs, Picks: picks.ViewOf(p.Picks), Status: StatusOf(f, p, s.Now())}, nil
}

// Delete forgets everything bingeon keeps of the user. The web then deletes
// their sign-in (Cognito's user, with its email).
func (s *Service) Delete(ctx context.Context, user string) error {
	if err := s.Picks.Delete(ctx, user); err != nil {
		return err
	}
	return s.Favourites.DeleteAll(ctx, user)
}
