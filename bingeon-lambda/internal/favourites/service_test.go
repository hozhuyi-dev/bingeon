package favourites_test

import (
	"context"
	"testing"
	"time"

	"bingeon-lambda/internal/favourites"
	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/tmdb"
	"bingeon-lambda/internal/testutil"
)

const user = "u1"

var now = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

func service() (*favourites.Service, *testutil.Favourites) {
	store := testutil.NewFavourites()
	movies := &testutil.TMDB{Movies: []tmdb.Movie{testutil.Movie(949, "Heat", "1995-12-15"), testutil.Movie(680, "Pulp Fiction", "1994-09-10")}}
	return &favourites.Service{Store: store, Movies: movies, Now: func() time.Time { return now }}, store
}

func TestAddKeepsTheTitleFromTMDB(t *testing.T) {
	s, store := service()
	ctx := context.Background()
	f, err := s.Add(ctx, user, 949)
	if err != nil {
		t.Fatal(err)
	}
	if want := (favourites.Favourite{ID: 949, Title: "Heat", Year: 1995, Poster: "/Heat.jpg", Added: now}); f != want {
		t.Errorf("Add = %+v, want %+v", f, want)
	}
	again, err := s.Add(ctx, user, 949)
	if err != nil || again != f {
		t.Errorf("adding again = %+v, %v; want it as it was", again, err)
	}
	if st, _ := store.Stats(ctx, user); st.Count != 1 || st.Changes != 1 {
		t.Errorf("stats = %+v; want one favourite, one change", st)
	}
}

func TestRemove(t *testing.T) {
	s, store := service()
	ctx := context.Background()
	if _, err := s.Add(ctx, user, 949); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := s.Remove(ctx, user, 949); err != nil {
			t.Fatal(err)
		}
	}
	if st, _ := store.Stats(ctx, user); st.Count != 0 || st.Changes != 2 {
		t.Errorf("stats = %+v; want none, and removing what isn't there not counted", st)
	}
}

func TestRefusalsChangeNothing(t *testing.T) {
	s, store := service()
	ctx := context.Background()
	if _, err := s.Add(ctx, user, 1); !apperr.Is(err, apperr.CodeNotFound) {
		t.Errorf("adding a movie TMDB doesn't have: %v, want not_found", err)
	}
	full := favourites.Stats{Count: favourites.MaxFavourites}
	if err := store.Add(ctx, user, favourites.Stats{}, full, favourites.Favourite{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(ctx, user, 949); !apperr.Is(err, apperr.CodeRefused) {
		t.Errorf("adding a 501st: %v, want refused", err)
	}
	if st, _ := store.Stats(ctx, user); st != full {
		t.Errorf("stats = %+v, want unchanged", st)
	}
}

func TestAChangeGoingFirstIsRetried(t *testing.T) {
	s, store := service()
	ctx := context.Background()
	bumps := 1
	store.BeforeWrite = func(u string) {
		if bumps > 0 {
			bumps--
			store.Bump(u)
		}
	}
	if _, err := s.Add(ctx, user, 949); err != nil {
		t.Fatalf("after one conflict: %v", err)
	}
	store.BeforeWrite = store.Bump
	if _, err := s.Add(ctx, user, 680); !apperr.Is(err, apperr.CodeConflict) {
		t.Errorf("conflicting every time: %v, want conflict", err)
	}
}
