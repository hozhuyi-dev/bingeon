// Package meter counts what a request costs (table units, TMDB and Gemini
// calls), for the one line the router logs per request.
package meter

import (
	"context"
	"maps"
	"sync"
)

type key struct{}

type counts struct {
	mu sync.Mutex
	m  map[string]float64
}

// Start gives a request its own counts.
func Start(ctx context.Context) context.Context {
	return context.WithValue(ctx, key{}, &counts{m: map[string]float64{}})
}

// Add counts n of something; nothing outside a request.
func Add(ctx context.Context, name string, n float64) {
	c, ok := ctx.Value(key{}).(*counts)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[name] += n
}

// Counts is what the request counted.
func Counts(ctx context.Context) map[string]float64 {
	c, ok := ctx.Value(key{}).(*counts)
	if !ok {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return maps.Clone(c.m)
}
