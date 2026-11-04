// Package app wires the slices into the two routers: the catalogue, which
// anyone reaches through CloudFront, and the signed-in API. cmd/catalog
// and cmd/api build them on TMDB, Gemini and DynamoDB; the tests on fakes.
package app

import (
	"context"
	"time"

	"bingeon-lambda/internal/account"
	"bingeon-lambda/internal/catalog"
	"bingeon-lambda/internal/favourites"
	"bingeon-lambda/internal/picks"
	"bingeon-lambda/internal/platform/router"
)

// Catalog is the catalogue's router.
func Catalog(movies catalog.Movies) *router.Router {
	r := router.New()
	catalog.Routes(r, &catalog.Service{Movies: movies})
	return r
}

// Movies is TMDB, as the signed-in API reads it.
type Movies interface {
	favourites.Movies
	picks.Movies
}

type Deps struct {
	Favourites favourites.Store
	Picks      picks.Store
	Movies     Movies
	Picker     picks.Picker
	// GeminiDailyLimit is the day's Gemini calls across all users.
	GeminiDailyLimit int
	Now              func() time.Time
}

// API is the signed-in API's router.
func API(d Deps) *router.Router {
	favs := &favourites.Service{Store: d.Favourites, Movies: d.Movies, Now: d.Now}
	pks := &picks.Service{
		Store: d.Picks, Favourites: seeds{favs}, Picker: d.Picker, Movies: d.Movies,
		DailyLimit: d.GeminiDailyLimit, Now: d.Now,
	}
	acct := &account.Service{Favourites: favs, Picks: pks, Now: d.Now}
	status := func(ctx context.Context, user string) (any, error) { return acct.Status(ctx, user) }

	r := router.New()
	account.Routes(r, acct)
	favourites.Routes(r, favs, status)
	picks.Routes(r, pks, status)
	return r
}

// seeds joins picks to favourites: the user's favourites, newest first, as
// picks are made from them.
type seeds struct{ favourites *favourites.Service }

func (s seeds) Seeds(ctx context.Context, user string) ([]picks.Seed, error) {
	favs, err := s.favourites.List(ctx, user)
	if err != nil {
		return nil, err
	}
	out := make([]picks.Seed, 0, len(favs))
	for _, f := range favs {
		out = append(out, picks.Seed{ID: f.ID, Title: f.Title, Year: f.Year})
	}
	return out, nil
}
