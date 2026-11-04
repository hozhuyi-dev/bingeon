package picks

import (
	"context"
	"time"

	"bingeon-lambda/internal/platform/router"
	"bingeon-lambda/internal/platform/user"
)

// Status is the account's status (account.Status) after a change, which
// each reply carries so the web reads nothing again.
type Status func(ctx context.Context, user string) (any, error)

// View is the user's last picks, as the web shows them.
type View struct {
	Movies []Pick     `json:"movies"`
	MadeAt *time.Time `json:"madeAt,omitempty"`
}

// ViewOf is the picks' view.
func ViewOf(p Picks) View {
	v := View{Movies: p.Last}
	if v.Movies == nil {
		v.Movies = []Pick{}
	}
	if !p.MadeAt.IsZero() {
		v.MadeAt = &p.MadeAt
	}
	return v
}

type makeReply struct {
	Picks  View `json:"picks"`
	Status any  `json:"status"`
}

func Routes(r *router.Router, s *Service, status Status) {
	r.Add("picks.make", router.Typed(func(ctx context.Context, _ struct{}) (any, error) {
		u := user.ID(ctx)
		p, err := s.Make(ctx, u)
		if err != nil {
			return nil, err
		}
		st, err := status(ctx, u)
		return makeReply{Picks: ViewOf(p), Status: st}, err
	}))
}
