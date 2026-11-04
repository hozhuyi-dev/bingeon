package catalog_test

import (
	"context"
	"strings"
	"testing"

	"bingeon-lambda/internal/catalog"
	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/tmdb"
	"bingeon-lambda/internal/testutil"
)

var ctx = context.Background()

func service(movies ...tmdb.Movie) *catalog.Service {
	if len(movies) == 0 {
		movies = []tmdb.Movie{testutil.Movie(949, "Heat", "1995-12-15"), testutil.Movie(680, "Pulp Fiction", "1994-09-10")}
	}
	return &catalog.Service{Movies: &testutil.TMDB{Movies: movies}}
}

func TestHomeFillsEveryShelf(t *testing.T) {
	home, err := service().Home(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for name, shelf := range map[string][]catalog.Card{
		"popular": home.Popular, "now playing": home.NowPlaying, "upcoming": home.Upcoming, "top rated": home.TopRated,
	} {
		if len(shelf) != 2 || shelf[0] != (catalog.Card{ID: 949, Title: "Heat", Year: 1995, Poster: "/Heat.jpg"}) {
			t.Errorf("%s = %+v", name, shelf)
		}
	}
	if len(home.Genres) != 2 {
		t.Errorf("genres = %+v", home.Genres)
	}
}

func TestRequestsAreChecked(t *testing.T) {
	s := service()
	refused := map[string]error{}
	_, refused["an unknown list"] = s.List(ctx, "trending", 0, 1)
	_, refused["a list and a genre"] = s.List(ctx, "popular", 18, 1)
	_, refused["page 501"] = s.List(ctx, "popular", 0, 501)
	_, refused["a one-letter search"] = s.Search(ctx, "  h ", 1)
	_, refused["a 101-letter search"] = s.Search(ctx, strings.Repeat("a", 101), 1)
	_, refused["movie 0"] = s.Movie(ctx, 0)
	for name, err := range refused {
		if !apperr.Is(err, apperr.CodeInvalid) {
			t.Errorf("%s: %v, want invalid", name, err)
		}
	}

	allowed := map[string]error{}
	_, allowed["a genre"] = s.List(ctx, "", 18, 0)
	_, allowed["the last page"] = s.List(ctx, "top_rated", 0, 500)
	_, allowed["a spaced search"] = s.Search(ctx, "  pulp   fiction ", 0)
	_, allowed["a two-letter search"] = s.Search(ctx, "up", 0)
	for name, err := range allowed {
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

type (
	castMember = struct {
		Name        string `json:"name"`
		Character   string `json:"character"`
		ProfilePath string `json:"profile_path"`
	}
	crewMember = struct {
		Name string `json:"name"`
		Job  string `json:"job"`
	}
	video = struct {
		Key      string `json:"key"`
		Site     string `json:"site"`
		Type     string `json:"type"`
		Official bool   `json:"official"`
	}
)

func TestMoviePage(t *testing.T) {
	m := testutil.Movie(949, "Heat", "1995-12-15")
	for i := range 15 {
		m.Credits.Cast = append(m.Credits.Cast, castMember{Name: "Actor " + string(rune('A'+i))})
	}
	m.Credits.Crew = []crewMember{{"Michael Mann", "Director"}, {"Dante Spinotti", "Director of Photography"}}
	m.Videos.Results = []video{
		{Key: "teaser", Site: "YouTube", Type: "Teaser", Official: true},
		{Key: "fan", Site: "YouTube", Type: "Trailer"},
		{Key: "vimeo", Site: "Vimeo", Type: "Trailer", Official: true},
		{Key: "official", Site: "YouTube", Type: "Trailer", Official: true},
	}
	for i := range 20 {
		m.Similar.Results = append(m.Similar.Results, tmdb.Summary{ID: int64(i + 1)})
	}

	got, err := service(m).Movie(ctx, 949)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cast) != 12 || got.Cast[0].Name != "Actor A" {
		t.Errorf("cast = %d, first %+v; want the first 12", len(got.Cast), got.Cast[0])
	}
	if len(got.Directors) != 1 || got.Directors[0] != "Michael Mann" {
		t.Errorf("directors = %v", got.Directors)
	}
	if got.Trailer != "official" {
		t.Errorf("trailer = %q, want the official YouTube trailer", got.Trailer)
	}
	if len(got.Similar) != 12 || got.Recommended == nil {
		t.Errorf("similar = %d, recommended = %v; want 12 and an empty list", len(got.Similar), got.Recommended)
	}
}

func TestTMDBFailingIsUnavailable(t *testing.T) {
	s := &catalog.Service{Movies: &testutil.TMDB{Err: apperr.Unavailable("down")}}
	if _, err := s.Home(ctx); !apperr.Is(err, apperr.CodeUnavailable) {
		t.Errorf("Home: %v, want unavailable", err)
	}
}
