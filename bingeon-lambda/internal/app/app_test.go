package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"bingeon-lambda/internal/app"
	"bingeon-lambda/internal/platform/gemini"
	"bingeon-lambda/internal/platform/router"
	"bingeon-lambda/internal/platform/tmdb"
	"bingeon-lambda/internal/platform/user"
	"bingeon-lambda/internal/testutil"
)

// api is the signed-in API on memory stores and fakes; call sends an
// action as the web does, as one user, and decodes the reply's data.
func api(t *testing.T) func(action string, payload any, out any) *router.ErrorBody {
	movies := &testutil.TMDB{Movies: []tmdb.Movie{
		testutil.Movie(949, "Heat", "1995-12-15"), testutil.Movie(1, "Thief", "1981-03-27"),
	}}
	r := app.API(app.Deps{
		Favourites: testutil.NewFavourites(), Picks: testutil.NewPicks(), Movies: movies,
		Picker:           &testutil.Gemini{Suggestions: []gemini.Suggestion{testutil.Suggest("Thief", 1981)}},
		GeminiDailyLimit: 50,
		Now:              func() time.Time { return time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC) },
	})
	ctx := user.With(context.Background(), "u1")
	return func(action string, payload any, out any) *router.ErrorBody {
		t.Helper()
		body, _ := json.Marshal(payload)
		res, _ := r.Handle(ctx, router.Request{Action: action, Payload: body})
		if res.Error != nil {
			if res.Error.Code == "internal" {
				t.Fatalf("%s failed; see the log", action)
			}
			return res.Error
		}
		data, _ := json.Marshal(res.Data)
		if out != nil {
			if err := json.Unmarshal(data, out); err != nil {
				t.Fatal(err)
			}
		}
		return nil
	}
}

type verdict struct {
	OK  bool   `json:"ok"`
	Why string `json:"why"`
}

type status struct {
	Can struct {
		Add, Remove, Pick verdict
	} `json:"can"`
	Today struct {
		PicksLeft   int `json:"picksLeft"`
		ChangesLeft int `json:"changesLeft"`
	} `json:"today"`
}

type account struct {
	Favourites []struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	} `json:"favourites"`
	Picks struct {
		Movies []struct {
			Title  string `json:"title"`
			Reason string `json:"reason"`
		} `json:"movies"`
	} `json:"picks"`
	status
}

func TestAUsersDay(t *testing.T) {
	call := api(t)

	var fresh account
	if err := call("account.get", nil, &fresh); err != nil {
		t.Fatal(err.Message)
	}
	if len(fresh.Favourites) != 0 || !fresh.Can.Add.OK || fresh.Can.Pick.OK || fresh.Today.PicksLeft != 20 {
		t.Fatalf("a new account = %+v; want nothing, may add, may not pick yet", fresh)
	}
	if err := call("picks.make", nil, nil); err == nil || err.Code != "refused" || err.Message != fresh.Can.Pick.Why {
		t.Fatalf("picks.make without favourites: %+v; want refused, as can.pick said", err)
	}

	var added struct {
		Favourite struct {
			Title string `json:"title"`
		} `json:"favourite"`
		Status status `json:"status"`
	}
	if err := call("favourites.add", map[string]any{"id": 949}, &added); err != nil {
		t.Fatal(err.Message)
	}
	if added.Favourite.Title != "Heat" || !added.Status.Can.Pick.OK || added.Status.Today.ChangesLeft != 999 {
		t.Fatalf("favourites.add = %+v; want Heat, and picks allowed now", added)
	}

	var made struct {
		Picks struct {
			Movies []struct {
				Title string `json:"title"`
			} `json:"movies"`
		} `json:"picks"`
		Status status `json:"status"`
	}
	if err := call("picks.make", nil, &made); err != nil {
		t.Fatal(err.Message)
	}
	if len(made.Picks.Movies) != 1 || made.Picks.Movies[0].Title != "Thief" || made.Status.Today.PicksLeft != 19 {
		t.Fatalf("picks.make = %+v", made)
	}

	var now account
	if err := call("account.get", nil, &now); err != nil {
		t.Fatal(err.Message)
	}
	if len(now.Favourites) != 1 || len(now.Picks.Movies) != 1 || now.Picks.Movies[0].Reason == "" {
		t.Fatalf("account = %+v", now)
	}

	if err := call("account.delete", nil, nil); err != nil {
		t.Fatal(err.Message)
	}
	var gone account
	if err := call("account.get", nil, &gone); err != nil {
		t.Fatal(err.Message)
	}
	if len(gone.Favourites) != 0 || len(gone.Picks.Movies) != 0 || gone.Today.PicksLeft != 20 {
		t.Fatalf("after account.delete = %+v; want a new account", gone)
	}
}

func TestEveryActionIsRegistered(t *testing.T) {
	r := app.API(app.Deps{})
	want := []string{"account.delete", "account.get", "favourites.add", "favourites.remove", "picks.make"}
	if got := r.Actions(); len(got) != len(want) {
		t.Errorf("api actions = %v, want %v", got, want)
	}
	if got := app.Catalog(nil).Actions(); len(got) != 4 {
		t.Errorf("catalog actions = %v", got)
	}
}
