// Package testutil holds what the tests run on instead of AWS, TMDB and
// Gemini: in-memory stores that behave as the slices' DynamoDB stores do
// (each slice's store test runs against both), and fakes of TMDB and
// Gemini.
package testutil

import (
	"context"
	"maps"
	"slices"
	"sync"

	"bingeon-lambda/internal/favourites"
	"bingeon-lambda/internal/picks"
)

// Favourites is favourites.Store in memory.
type Favourites struct {
	mu    sync.Mutex
	stats map[string]favourites.Stats
	favs  map[string]map[int64]favourites.Favourite
	// BeforeWrite, when set, runs before each write is applied: a test
	// changes the stats in it to make the write conflict.
	BeforeWrite func(user string)
}

func NewFavourites() *Favourites {
	return &Favourites{stats: map[string]favourites.Stats{}, favs: map[string]map[int64]favourites.Favourite{}}
}

func (m *Favourites) Stats(_ context.Context, user string) (favourites.Stats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stats[user], nil
}

func (m *Favourites) Get(_ context.Context, user string, id int64) (favourites.Favourite, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.favs[user][id]
	return f, ok, nil
}

func (m *Favourites) List(_ context.Context, user string) ([]favourites.Favourite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Collect(maps.Values(m.favs[user])), nil
}

func (m *Favourites) Add(_ context.Context, user string, read, next favourites.Stats, f favourites.Favourite) error {
	return m.write(user, read, next, func(favs map[int64]favourites.Favourite) { favs[f.ID] = f })
}

func (m *Favourites) Remove(_ context.Context, user string, read, next favourites.Stats, id int64) error {
	return m.write(user, read, next, func(favs map[int64]favourites.Favourite) { delete(favs, id) })
}

func (m *Favourites) write(user string, read, next favourites.Stats, change func(map[int64]favourites.Favourite)) error {
	if m.BeforeWrite != nil {
		m.BeforeWrite(user)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stats[user].Version != read.Version {
		return favourites.ErrConflict
	}
	m.stats[user] = next
	if m.favs[user] == nil {
		m.favs[user] = map[int64]favourites.Favourite{}
	}
	change(m.favs[user])
	return nil
}

func (m *Favourites) DeleteAll(_ context.Context, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.stats, user)
	delete(m.favs, user)
	return nil
}

// Bump moves a user's stats on, as another change would.
func (m *Favourites) Bump(user string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.stats[user]
	st.Version++
	m.stats[user] = st
}

// Picks is picks.Store in memory.
type Picks struct {
	mu    sync.Mutex
	picks map[string]picks.Picks
	quota map[string]int
}

func NewPicks() *Picks {
	return &Picks{picks: map[string]picks.Picks{}, quota: map[string]int{}}
}

func (m *Picks) Get(_ context.Context, user string) (picks.Picks, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.picks[user], nil
}

func (m *Picks) Reserve(_ context.Context, user string, read, next picks.Picks, day string, limit int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch {
	case m.quota[day] >= limit:
		return picks.ErrQuotaFull
	case m.picks[user].Version != read.Version:
		return picks.ErrConflict
	}
	m.picks[user] = next
	m.quota[day]++
	return nil
}

func (m *Picks) Save(_ context.Context, user string, read, next picks.Picks) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.picks[user].Version != read.Version {
		return picks.ErrConflict
	}
	m.picks[user] = next
	return nil
}

func (m *Picks) QuotaUsed(_ context.Context, day string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.quota[day], nil
}

func (m *Picks) Delete(_ context.Context, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.picks, user)
	return nil
}

// SetQuota sets how many of a day's Gemini calls are used.
func (m *Picks) SetQuota(day string, used int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.quota[day] = used
}
