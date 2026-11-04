package picks_test

import (
	"context"
	"testing"
	"time"

	"bingeon-lambda/internal/picks"
	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/day"
	"bingeon-lambda/internal/platform/gemini"
	"bingeon-lambda/internal/platform/tmdb"
	"bingeon-lambda/internal/testutil"
)

const user = "u1"

var now = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

type seeds []picks.Seed

func (s seeds) Seeds(context.Context, string) ([]picks.Seed, error) { return s, nil }

type fixture struct {
	service *picks.Service
	store   *testutil.Picks
	gemini  *testutil.Gemini
}

// setup is a user whose favourite is Heat, and a TMDB holding it and
// eight more movies.
func setup(favourites ...picks.Seed) fixture {
	if favourites == nil {
		favourites = []picks.Seed{{ID: 949, Title: "Heat", Year: 1995}}
	}
	movies := []tmdb.Movie{testutil.Movie(949, "Heat", "1995-12-15")}
	for i, title := range []string{"Thief", "Collateral", "Ronin", "Sicario", "Drive", "The Insider", "Manhunter", "Miami Vice"} {
		movies = append(movies, testutil.Movie(int64(i+1), title, "2000-01-01"))
	}
	f := fixture{store: testutil.NewPicks(), gemini: &testutil.Gemini{}}
	f.service = &picks.Service{
		Store: f.store, Favourites: seeds(favourites), Picker: f.gemini, Movies: &testutil.TMDB{Movies: movies},
		DailyLimit: 50, Now: func() time.Time { return now },
	}
	return f
}

func titles(ps []picks.Pick) []string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Title)
	}
	return out
}

func TestMakeKeepsTheFirstSixFoundThatArentFavourites(t *testing.T) {
	f := setup()
	f.gemini.Suggestions = []gemini.Suggestion{
		testutil.Suggest("Heat", 1995), // a favourite
		testutil.Suggest("Thief", 2000),
		testutil.Suggest("Not A Real Film", 2001),
		testutil.Suggest("Collateral", 2000),
		testutil.Suggest("Thief", 2000), // twice
		testutil.Suggest("Ronin", 2000),
		testutil.Suggest("Sicario", 2000),
		testutil.Suggest("Drive", 2000),
		testutil.Suggest("The Insider", 2000),
		testutil.Suggest("Manhunter", 2000),
	}
	p, err := f.service.Make(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Thief", "Collateral", "Ronin", "Sicario", "Drive", "The Insider"}
	if got := titles(p.Last); len(got) != len(want) || got[0] != want[0] || got[5] != want[5] {
		t.Errorf("picks = %v, want %v", got, want)
	}
	if p.Last[0].Reason == "" || p.Last[0].Poster != "/Thief.jpg" || p.Made != 1 || !p.MadeAt.Equal(now) {
		t.Errorf("first pick %+v, made %d at %v", p.Last[0], p.Made, p.MadeAt)
	}
	if len(f.gemini.Liked) != 1 || f.gemini.Liked[0].Title != "Heat" {
		t.Errorf("Gemini was told %v liked", f.gemini.Liked)
	}

	// The next asks Gemini to avoid these.
	if _, err := f.service.Make(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	if len(f.gemini.Avoid) != 6 || f.gemini.Avoid[0].Title != "Thief" {
		t.Errorf("Gemini was told to avoid %v", f.gemini.Avoid)
	}
}

func TestOnlyTheNewestFavouritesAreSent(t *testing.T) {
	var many []picks.Seed
	for i := range 150 {
		many = append(many, picks.Seed{ID: int64(1000 + i), Title: "Film"})
	}
	f := setup(many...)
	f.gemini.Suggestions = []gemini.Suggestion{testutil.Suggest("Thief", 2000)}
	if _, err := f.service.Make(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	if len(f.gemini.Liked) != picks.Seeds {
		t.Errorf("Gemini was sent %d favourites, want %d", len(f.gemini.Liked), picks.Seeds)
	}
}

func TestRefusedBeforeGeminiIsCalled(t *testing.T) {
	ctx := context.Background()
	none := setup()
	none.service.Favourites = seeds{}
	spent := setup()
	if err := spent.store.Save(ctx, user, picks.Picks{}, picks.Picks{Day: day.Of(now), Made: picks.PerDay, Version: 1}); err != nil {
		t.Fatal(err)
	}
	resting := setup()
	resting.store.SetQuota(day.Of(now), 50)

	for name, f := range map[string]fixture{
		"without favourites": none, "after today's 20": spent, "with everyone's calls used": resting,
	} {
		before, _ := f.store.Get(ctx, user)
		if _, err := f.service.Make(ctx, user); !apperr.Is(err, apperr.CodeRefused) {
			t.Errorf("%s: %v, want refused", name, err)
		}
		if after, _ := f.store.Get(ctx, user); f.gemini.Calls != 0 || after.Version != before.Version {
			t.Errorf("%s: Gemini called %d times, picks %+v → %+v; want nothing done", name, f.gemini.Calls, before, after)
		}
	}
}

func TestAFailureGivesThePickBack(t *testing.T) {
	cases := map[string]func(f fixture){
		"Gemini busy":   func(f fixture) { f.gemini.Err = apperr.Busy("busy") },
		"nothing found": func(f fixture) { f.gemini.Suggestions = []gemini.Suggestion{testutil.Suggest("Not A Real Film", 2001)} },
	}
	for name, arrange := range cases {
		t.Run(name, func(t *testing.T) {
			f := setup()
			arrange(f)
			_, err := f.service.Make(context.Background(), user)
			if err == nil {
				t.Fatal("made picks")
			}
			p, _ := f.store.Get(context.Background(), user)
			used, _ := f.store.QuotaUsed(context.Background(), day.Of(now))
			if p.Made != 0 || used != 1 {
				t.Errorf("made %d, Gemini calls counted %d; want the pick given back, the call still counted", p.Made, used)
			}
		})
	}
}
