// Package catalog is the movie catalogue anyone can browse: home, lists,
// search and a movie's page, as views over TMDB that CloudFront caches.
package catalog

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"bingeon-lambda/internal/platform/apperr"
)

// How long CloudFront keeps each reply: lists move daily, a search's
// results and a movie's page hardly at all.
const (
	ListTTL   = 6 * time.Hour
	SearchTTL = 24 * time.Hour
	MovieTTL  = 24 * time.Hour
)

// A search is 2–100 characters: one letter matches everything, and the
// bound keeps cache keys short.
const (
	MinQuery = 2
	MaxQuery = 100
)

// MaxPage is the last page TMDB serves of any list.
const MaxPage = 500

// Lists are the lists a page of movies can be of, besides a genre.
var Lists = []string{"popular", "now_playing", "upcoming", "top_rated"}

// Shown on a movie's page, at most.
const (
	maxCast    = 12
	maxRelated = 12
)

type Card struct {
	ID     int64   `json:"id"`
	Title  string  `json:"title"`
	Year   int     `json:"year,omitempty"`
	Poster string  `json:"poster,omitempty"`
	Rating float64 `json:"rating,omitempty"`
}

type Page struct {
	Results    []Card `json:"results"`
	Page       int    `json:"page"`
	TotalPages int    `json:"totalPages"`
}

type Genre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Home struct {
	Popular    []Card  `json:"popular"`
	NowPlaying []Card  `json:"nowPlaying"`
	Upcoming   []Card  `json:"upcoming"`
	TopRated   []Card  `json:"topRated"`
	Genres     []Genre `json:"genres"`
}

type CastMember struct {
	Name      string `json:"name"`
	Character string `json:"character,omitempty"`
	Photo     string `json:"photo,omitempty"`
}

type Movie struct {
	Card
	Released    string       `json:"released,omitempty"`
	Runtime     int          `json:"runtime,omitempty"`
	Tagline     string       `json:"tagline,omitempty"`
	Overview    string       `json:"overview,omitempty"`
	Backdrop    string       `json:"backdrop,omitempty"`
	Genres      []Genre      `json:"genres"`
	Cast        []CastMember `json:"cast"`
	Directors   []string     `json:"directors"`
	Trailer     string       `json:"trailer,omitempty"` // a YouTube video id
	Similar     []Card       `json:"similar"`
	Recommended []Card       `json:"recommended"`
}

// CheckPage is a requested page: the first when none is asked for.
func CheckPage(page int) (int, error) {
	switch {
	case page == 0:
		return 1, nil
	case page < 1 || page > MaxPage:
		return 0, apperr.Invalid("Pages run from 1 to %d.", MaxPage)
	}
	return page, nil
}

// CheckList is a requested list: one of Lists, or a genre's.
func CheckList(list string, genre int64) error {
	switch {
	case list != "" && genre != 0:
		return apperr.Invalid("Ask for a list or a genre, not both.")
	case genre < 0:
		return apperr.Invalid("No such genre.")
	case genre == 0 && !slices.Contains(Lists, list):
		return apperr.Invalid("No such list: %q.", list)
	}
	return nil
}

// CheckQuery is a search as TMDB is asked it: words single-spaced, and
// 2–100 characters. The web asks in that form too, so the same search is
// one cache entry.
func CheckQuery(q string) (string, error) {
	q = strings.Join(strings.Fields(q), " ")
	if n := utf8.RuneCountInString(q); n < MinQuery || n > MaxQuery {
		return "", apperr.Invalid("Search for %d to %d characters.", MinQuery, MaxQuery)
	}
	return q, nil
}
