package picks_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"bingeon-lambda/internal/picks"
	"bingeon-lambda/internal/testutil"
)

// TestStore runs against the memory store, which the service tests use,
// and DynamoStore when TABLE_NAME names a table, so the two agree.
func TestStore(t *testing.T) {
	stores := map[string]picks.Store{"memory": testutil.NewPicks()}
	if db := testutil.Table(t); db != nil {
		stores["dynamo"] = &picks.DynamoStore{DB: db}
	}
	for name, store := range stores {
		t.Run(name, func(t *testing.T) { checkStore(t, store) })
	}
}

func checkStore(t *testing.T, store picks.Store) {
	ctx, user := context.Background(), testutil.TestUser()
	// The quota of a day of its own (expiring by itself).
	day := user
	t.Cleanup(func() { _ = store.Delete(ctx, user) })
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

	read, err := store.Get(ctx, user)
	if err != nil || read.Version != 0 || read.Made != 0 {
		t.Fatalf("a new user's picks = %+v, %v; want zero", read, err)
	}
	reserved := read.Reserved(now)
	if err := store.Reserve(ctx, user, read, reserved, day, 1); err != nil {
		t.Fatal(err)
	}
	if err := store.Reserve(ctx, user, reserved, reserved.Reserved(now), day, 1); !errors.Is(err, picks.ErrQuotaFull) {
		t.Fatalf("reserving past the limit: %v, want ErrQuotaFull", err)
	}
	if err := store.Reserve(ctx, user, read, reserved, day, 5); !errors.Is(err, picks.ErrConflict) {
		t.Fatalf("reserving from a stale read: %v, want ErrConflict", err)
	}
	if used := quotaUsed(t, store, day, 1); used != 1 {
		t.Fatalf("quota used = %d, want 1", used)
	}

	last := []picks.Pick{{ID: 949, Title: "Heat", Year: 1995, Poster: "/heat.jpg", Reason: "Crime, done right."}}
	saved := reserved.With(last, now)
	if err := store.Save(ctx, user, reserved, saved); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ctx, user, reserved, saved); !errors.Is(err, picks.ErrConflict) {
		t.Fatalf("saving from a stale read: %v, want ErrConflict", err)
	}
	got, err := store.Get(ctx, user)
	if err != nil || got.Made != 1 || got.Version != saved.Version || !got.MadeAt.Equal(now) || !slices.Equal(got.Last, last) {
		t.Fatalf("Get = %+v, %v; want %+v", got, err, saved)
	}

	if err := store.Delete(ctx, user); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.Get(ctx, user); got.Version != 0 {
		t.Fatalf("after Delete, picks = %+v", got)
	}
}

// quotaUsed reads the quota, which DynamoDB reads eventually: until it
// says want, or a second has passed.
func quotaUsed(t *testing.T, store picks.Store, day string, want int) int {
	var used int
	for range 10 {
		var err error
		if used, err = store.QuotaUsed(context.Background(), day); err != nil {
			t.Fatal(err)
		}
		if used == want {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return used
}
