package catalog

import (
	"context"

	"bingeon-lambda/internal/platform/router"
)

// The catalogue's payloads come from a URL's query, so numbers arrive as
// strings (see router.FunctionURL). Its names are CloudFront's cache key:
// keep them in step with CATALOG_QUERY in bingeon-infra's site.ts.
type (
	listRequest struct {
		List  string `json:"list"`
		Genre int64  `json:"genre,string"`
		Page  int    `json:"page,string"`
	}
	searchRequest struct {
		Q    string `json:"q"`
		Page int    `json:"page,string"`
	}
	movieRequest struct {
		ID int64 `json:"id,string"`
	}
)

func Routes(r *router.Router, s *Service) {
	r.Cached("catalog.home", ListTTL, router.Typed(func(ctx context.Context, _ struct{}) (any, error) {
		return s.Home(ctx)
	}))
	r.Cached("movies.list", ListTTL, router.Typed(func(ctx context.Context, in listRequest) (any, error) {
		return s.List(ctx, in.List, in.Genre, in.Page)
	}))
	r.Cached("movies.search", SearchTTL, router.Typed(func(ctx context.Context, in searchRequest) (any, error) {
		return s.Search(ctx, in.Q, in.Page)
	}))
	r.Cached("movies.get", MovieTTL, router.Typed(func(ctx context.Context, in movieRequest) (any, error) {
		return s.Movie(ctx, in.ID)
	}))
}
