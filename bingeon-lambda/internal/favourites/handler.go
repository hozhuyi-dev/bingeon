package favourites

import (
	"context"

	"bingeon-lambda/internal/platform/router"
	"bingeon-lambda/internal/platform/user"
)

// Status is the account's status (account.Status) after a change, which
// each reply carries so the web reads nothing again.
type Status func(ctx context.Context, user string) (any, error)

type movieRequest struct {
	ID int64 `json:"id"`
}

type (
	addReply struct {
		Favourite Favourite `json:"favourite"`
		Status    any       `json:"status"`
	}
	removeReply struct {
		ID     int64 `json:"id"`
		Status any   `json:"status"`
	}
)

func Routes(r *router.Router, s *Service, status Status) {
	r.Add("favourites.add", router.Typed(func(ctx context.Context, in movieRequest) (any, error) {
		u := user.ID(ctx)
		f, err := s.Add(ctx, u, in.ID)
		if err != nil {
			return nil, err
		}
		st, err := status(ctx, u)
		return addReply{Favourite: f, Status: st}, err
	}))
	r.Add("favourites.remove", router.Typed(func(ctx context.Context, in movieRequest) (any, error) {
		u := user.ID(ctx)
		if err := s.Remove(ctx, u, in.ID); err != nil {
			return nil, err
		}
		st, err := status(ctx, u)
		return removeReply{ID: in.ID, Status: st}, err
	}))
}
