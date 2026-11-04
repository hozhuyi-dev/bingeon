// Package tmdb calls The Movie Database's API v3 with the read access
// token, which never leaves the Lambdas. Its types are TMDB's, trimmed to
// what bingeon reads; the slices turn them into their own views.
package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/meter"
)

const base = "https://api.themoviedb.org/3"

// Region is where now playing and upcoming are for.
const Region = "MY"

type Client struct {
	token string
	http  *http.Client
	// Base is TMDB's API; tests point it at a fake.
	Base string
}

// New is a client with the read access token. Each call gives up after
// 4 seconds, inside the catalogue's 5.
func New(token string) *Client {
	return &Client{token: token, http: &http.Client{Timeout: 4 * time.Second}, Base: base}
}

type Summary struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	ReleaseDate string  `json:"release_date"`
	PosterPath  string  `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
}

// Year is a TMDB date's year ("2026-11-04"); 0 when it has none.
func Year(date string) int {
	if len(date) < 4 {
		return 0
	}
	y, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0
	}
	return y
}

type Page struct {
	Page         int       `json:"page"`
	TotalPages   int       `json:"total_pages"`
	TotalResults int       `json:"total_results"`
	Results      []Summary `json:"results"`
}

type Genre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Movie struct {
	Summary
	Overview     string  `json:"overview"`
	Tagline      string  `json:"tagline"`
	Runtime      int     `json:"runtime"`
	BackdropPath string  `json:"backdrop_path"`
	Genres       []Genre `json:"genres"`
	// Appended (see Movie with full): absent otherwise.
	Credits struct {
		Cast []struct {
			Name        string `json:"name"`
			Character   string `json:"character"`
			ProfilePath string `json:"profile_path"`
		} `json:"cast"`
		Crew []struct {
			Name string `json:"name"`
			Job  string `json:"job"`
		} `json:"crew"`
	} `json:"credits"`
	Videos struct {
		Results []struct {
			Key      string `json:"key"`
			Site     string `json:"site"`
			Type     string `json:"type"`
			Official bool   `json:"official"`
		} `json:"results"`
	} `json:"videos"`
	Similar         Page `json:"similar"`
	Recommendations Page `json:"recommendations"`
}

// List is a page of a list under /movie/<list> (popular, now_playing,
// upcoming, top_rated), in Region.
func (c *Client) List(ctx context.Context, list string, page int) (Page, error) {
	var out Page
	err := c.get(ctx, "/movie/"+list, url.Values{"page": {strconv.Itoa(page)}, "region": {Region}}, &out)
	return out, err
}

// Genre is a page of a genre's movies, most popular first.
func (c *Client) Genre(ctx context.Context, genre int64, page int) (Page, error) {
	var out Page
	err := c.get(ctx, "/discover/movie", url.Values{
		"with_genres": {strconv.FormatInt(genre, 10)}, "sort_by": {"popularity.desc"},
		"page": {strconv.Itoa(page)}, "include_adult": {"false"},
	}, &out)
	return out, err
}

// Search is a page of the movies matching a query, of a year if not 0.
func (c *Client) Search(ctx context.Context, query string, year, page int) (Page, error) {
	q := url.Values{"query": {query}, "page": {strconv.Itoa(page)}, "include_adult": {"false"}}
	if year != 0 {
		q.Set("primary_release_year", strconv.Itoa(year))
	}
	var out Page
	err := c.get(ctx, "/search/movie", q, &out)
	return out, err
}

// Genres are TMDB's movie genres.
func (c *Client) Genres(ctx context.Context) ([]Genre, error) {
	var out struct {
		Genres []Genre `json:"genres"`
	}
	err := c.get(ctx, "/genre/movie/list", nil, &out)
	return out.Genres, err
}

// Movie is one movie; full also brings its credits, videos, similar and
// recommended movies, in the same call.
func (c *Client) Movie(ctx context.Context, id int64, full bool) (Movie, error) {
	q := url.Values{}
	if full {
		q.Set("append_to_response", "credits,videos,similar,recommendations")
	}
	var out Movie
	err := c.get(ctx, "/movie/"+strconv.FormatInt(id, 10), q, &out)
	return out, err
}

func (c *Client) get(ctx context.Context, path string, q url.Values, out any) error {
	if q == nil {
		q = url.Values{}
	}
	q.Set("language", "en-US")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Base+path+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	meter.Add(ctx, "tmdb", 1)
	res, err := c.http.Do(req)
	if err != nil {
		return unavailable(ctx, fmt.Errorf("tmdb %s: %w", path, err))
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusNotFound:
		return apperr.NotFound("No such movie.")
	case res.StatusCode != http.StatusOK:
		return unavailable(ctx, fmt.Errorf("tmdb %s: %s", path, res.Status))
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return unavailable(ctx, fmt.Errorf("tmdb %s: %w", path, err))
	}
	return nil
}

// unavailable is TMDB failing, as the web is shown it; the cause is
// logged here, as the router logs only errors it can't show.
func unavailable(ctx context.Context, cause error) error {
	slog.WarnContext(ctx, "tmdb failed", "error", cause)
	return apperr.Unavailable("The movie database isn't answering. Try again in a minute.")
}
