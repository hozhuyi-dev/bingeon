package favourites_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"bingeon-lambda/internal/favourites"
	"bingeon-lambda/internal/testutil"
)

// TestStore runs against the memory store, which the service tests use,
// and DynamoStore when TABLE_NAME names a table, so the two agree.
func TestStore(t *testing.T) {
	stores := map[string]favourites.Store{"memory": testutil.NewFavourites()}
	if db := testutil.Table(t); db != nil {
		stores["dynamo"] = &favourites.DynamoStore{DB: db}
	}
	for name, store := range stores {
		t.Run(name, func(t *testing.T) { checkStore(t, store) })
	}
}

func checkStore(t *testing.T, store favourites.Store) {
	ctx, user := context.Background(), testutil.TestUser()
	t.Cleanup(func() { _ = store.DeleteAll(ctx, user) })
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	heat := favourites.Favourite{ID: 949, Title: "Heat", Year: 1995, Poster: "/heat.jpg", Added: now}

	read, err := store.Stats(ctx, user)
	if err != nil || read != (favourites.Stats{}) {
		t.Fatalf("a new user's stats = %+v, %v; want zero", read, err)
	}
	added := read.Added(now)
	if err := store.Add(ctx, user, read, added, heat); err != nil {
		t.Fatal(err)
	}
	if err := store.Add(ctx, user, read, added, heat); !errors.Is(err, favourites.ErrConflict) {
		t.Fatalf("adding from a stale read: %v, want ErrConflict", err)
	}
	if got, err := store.Stats(ctx, user); err != nil || got != added {
		t.Fatalf("stats = %+v, %v; want %+v", got, err, added)
	}
	got, ok, err := store.Get(ctx, user, heat.ID)
	if err != nil || !ok || got.Title != heat.Title || got.Year != heat.Year || got.Poster != heat.Poster || !got.Added.Equal(now) {
		t.Fatalf("Get = %+v, %v, %v; want %+v", got, ok, err, heat)
	}
	if list, err := store.List(ctx, user); err != nil || len(list) != 1 || list[0].ID != heat.ID {
		t.Fatalf("List = %+v, %v; want Heat alone", list, err)
	}

	removed := added.Removed(now)
	if err := store.Remove(ctx, user, added, removed, heat.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.Get(ctx, user, heat.ID); err != nil || ok {
		t.Fatalf("after Remove, Get found it (%v)", err)
	}
	if got, _ := store.Stats(ctx, user); got != removed {
		t.Fatalf("stats = %+v, want %+v", got, removed)
	}

	if err := store.DeleteAll(ctx, user); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.Stats(ctx, user); got != (favourites.Stats{}) {
		t.Fatalf("after DeleteAll, stats = %+v", got)
	}
}
