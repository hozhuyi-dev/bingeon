package account

import (
	"context"

	"bingeon-lambda/internal/platform/router"
	"bingeon-lambda/internal/platform/user"
)

func Routes(r *router.Router, s *Service) {
	r.Add("account.get", router.Typed(func(ctx context.Context, _ struct{}) (any, error) {
		return s.Get(ctx, user.ID(ctx))
	}))
	r.Add("account.delete", router.Typed(func(ctx context.Context, _ struct{}) (any, error) {
		return nil, s.Delete(ctx, user.ID(ctx))
	}))
}
