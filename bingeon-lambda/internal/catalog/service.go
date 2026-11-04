package catalog

import (
	"context"
	"sync"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/tmdb"
)

// Movies is TMDB, as the catalogue reads it (tmdb.Client).
type Movies interface {
	List(ctx context.Context, list string, page int) (tmdb.Page, error)
	Genre(ctx context.Context, genre int64, page int) (tmdb.Page, error)
	Search(ctx context.Context, query string, year, page int) (tmdb.Page, error)
	Genres(ctx context.Context) ([]tmdb.Genre, error)
	Movie(ctx context.Context, id int64, full bool) (tmdb.Movie, error)
}

type Service struct {
	Movies Movies
}

// Home is the first page of each list, and the genres, read at once.
func (s *Service) Home(ctx context.Context) (Home, error) {
	var (
		home  Home
		wg    sync.WaitGroup
		mu    sync.Mutex
		first error
	)
	fail := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		if first == nil {
			first = err
		}
	}
	shelves := map[string]*[]Card{
		"popular": &home.Popular, "now_playing": &home.NowPlaying, "upcoming": &home.Upcoming, "top_rated": &home.TopRated,
	}
	for list, shelf := range shelves {
		wg.Go(func() {
			page, err := s.Movies.List(ctx, list, 1)
			if err != nil {
				fail(err)
				return
			}
			*shelf = cardsOf(page.Results, 0)
		})
	}
	wg.Go(func() {
		genres, err := s.Movies.Genres(ctx)
		if err != nil {
			fail(err)
			return
		}
		home.Genres = genresOf(genres)
	})
	wg.Wait()
	return home, first
}

// List is a page of a list, or of a genre's movies.
func (s *Service) List(ctx context.Context, list string, genre int64, page int) (Page, error) {
	if err := CheckList(list, genre); err != nil {
		return Page{}, err
	}
	page, err := CheckPage(page)
	if err != nil {
		return Page{}, err
	}
	var res tmdb.Page
	if genre != 0 {
		res, err = s.Movies.Genre(ctx, genre, page)
	} else {
		res, err = s.Movies.List(ctx, list, page)
	}
	return pageOf(res), err
}

// Search is a page of the movies matching a query.
func (s *Service) Search(ctx context.Context, query string, page int) (Page, error) {
	query, err := CheckQuery(query)
	if err != nil {
		return Page{}, err
	}
	page, err = CheckPage(page)
	if err != nil {
		return Page{}, err
	}
	res, err := s.Movies.Search(ctx, query, 0, page)
	return pageOf(res), err
}

// Movie is a movie's page, from one call to TMDB.
func (s *Service) Movie(ctx context.Context, id int64) (Movie, error) {
	if id <= 0 {
		return Movie{}, apperr.Invalid("No such movie.")
	}
	m, err := s.Movies.Movie(ctx, id, true)
	if err != nil {
		return Movie{}, err
	}
	out := Movie{
		Card: cardOf(m.Summary), Released: m.ReleaseDate, Runtime: m.Runtime, Tagline: m.Tagline,
		Overview: m.Overview, Backdrop: m.BackdropPath, Genres: genresOf(m.Genres),
		Cast: []CastMember{}, Directors: []string{},
		Similar: cardsOf(m.Similar.Results, maxRelated), Recommended: cardsOf(m.Recommendations.Results, maxRelated),
	}
	for _, c := range m.Credits.Cast[:min(maxCast, len(m.Credits.Cast))] {
		out.Cast = append(out.Cast, CastMember{Name: c.Name, Character: c.Character, Photo: c.ProfilePath})
	}
	for _, c := range m.Credits.Crew {
		if c.Job == "Director" {
			out.Directors = append(out.Directors, c.Name)
		}
	}
	out.Trailer = trailerOf(m)
	return out, nil
}

// trailerOf is the movie's YouTube trailer: an official one if it has one.
func trailerOf(m tmdb.Movie) string {
	found := ""
	for _, v := range m.Videos.Results {
		if v.Site != "YouTube" || v.Type != "Trailer" {
			continue
		}
		if v.Official {
			return v.Key
		}
		if found == "" {
			found = v.Key
		}
	}
	return found
}

func cardOf(m tmdb.Summary) Card {
	return Card{ID: m.ID, Title: m.Title, Year: tmdb.Year(m.ReleaseDate), Poster: m.PosterPath, Rating: m.VoteAverage}
}

// cardsOf is the movies as cards, at most max of them (0: all).
func cardsOf(movies []tmdb.Summary, max int) []Card {
	if max > 0 && len(movies) > max {
		movies = movies[:max]
	}
	out := make([]Card, 0, len(movies))
	for _, m := range movies {
		out = append(out, cardOf(m))
	}
	return out
}

func pageOf(p tmdb.Page) Page {
	return Page{Results: cardsOf(p.Results, 0), Page: p.Page, TotalPages: min(p.TotalPages, MaxPage)}
}

func genresOf(genres []tmdb.Genre) []Genre {
	out := make([]Genre, 0, len(genres))
	for _, g := range genres {
		out = append(out, Genre{ID: g.ID, Name: g.Name})
	}
	return out
}
