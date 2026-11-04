package testutil

import (
	"context"
	"strings"
	"sync"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/gemini"
	"bingeon-lambda/internal/platform/tmdb"
)

// TMDB is a fake TMDB holding a few movies; every list is all of them.
type TMDB struct {
	Movies []tmdb.Movie
	// Err, when set, is every call's answer: TMDB failing.
	Err error
}

func (f *TMDB) summaries() []tmdb.Summary {
	out := []tmdb.Summary{}
	for _, m := range f.Movies {
		out = append(out, m.Summary)
	}
	return out
}

func (f *TMDB) page() tmdb.Page {
	return tmdb.Page{Page: 1, TotalPages: 1, TotalResults: len(f.Movies), Results: f.summaries()}
}

func (f *TMDB) List(context.Context, string, int) (tmdb.Page, error) { return f.page(), f.Err }

func (f *TMDB) Genre(context.Context, int64, int) (tmdb.Page, error) { return f.page(), f.Err }

func (f *TMDB) Genres(context.Context) ([]tmdb.Genre, error) {
	return []tmdb.Genre{{ID: 18, Name: "Drama"}, {ID: 80, Name: "Crime"}}, f.Err
}

// Search matches titles containing the query, ignoring case, in the year
// if one is given.
func (f *TMDB) Search(_ context.Context, query string, year, page int) (tmdb.Page, error) {
	if f.Err != nil {
		return tmdb.Page{}, f.Err
	}
	out := tmdb.Page{Page: page, TotalPages: 1, Results: []tmdb.Summary{}}
	for _, m := range f.Movies {
		if strings.Contains(strings.ToLower(m.Title), strings.ToLower(query)) && (year == 0 || tmdb.Year(m.ReleaseDate) == year) {
			out.Results = append(out.Results, m.Summary)
		}
	}
	out.TotalResults = len(out.Results)
	return out, nil
}

func (f *TMDB) Movie(_ context.Context, id int64, _ bool) (tmdb.Movie, error) {
	if f.Err != nil {
		return tmdb.Movie{}, f.Err
	}
	for _, m := range f.Movies {
		if m.ID == id {
			return m, nil
		}
	}
	return tmdb.Movie{}, apperr.NotFound("No such movie.")
}

// Movie is a movie for the fake to hold.
func Movie(id int64, title, released string) tmdb.Movie {
	return tmdb.Movie{Summary: tmdb.Summary{ID: id, Title: title, ReleaseDate: released, PosterPath: "/" + title + ".jpg"}}
}

// Gemini is a fake Gemini suggesting the same movies every time.
type Gemini struct {
	mu          sync.Mutex
	Suggestions []gemini.Suggestion
	// Err, when set, is its answer: Gemini failing.
	Err error
	// Liked and Avoid are what the last call was told.
	Liked, Avoid []gemini.Title
	Calls        int
}

func (f *Gemini) Suggest(_ context.Context, liked, avoid []gemini.Title, _ int) ([]gemini.Suggestion, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls++
	f.Liked, f.Avoid = liked, avoid
	return f.Suggestions, f.Err
}

// Suggest is a suggestion for the fake to make.
func Suggest(title string, year int) gemini.Suggestion {
	return gemini.Suggestion{Title: gemini.Title{Title: title, Year: year}, Reason: "Because you liked it."}
}
