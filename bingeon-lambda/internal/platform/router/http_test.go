package router

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"bingeon-lambda/internal/platform/apperr"
)

func get(path string, query map[string]string) events.LambdaFunctionURLRequest {
	req := events.LambdaFunctionURLRequest{RawPath: path, QueryStringParameters: query}
	req.RequestContext.HTTP.Method = "GET"
	return req
}

func TestFunctionURL(t *testing.T) {
	r := New()
	type in struct {
		Page int    `json:"page,string"`
		Q    string `json:"q"`
	}
	r.Cached("movies.search", 24*time.Hour, Typed(func(_ context.Context, in in) (any, error) {
		if in.Q == "" {
			return nil, apperr.Invalid("Search for something.")
		}
		return in, nil
	}))
	r.Cached("movies.get", time.Hour, Typed(func(context.Context, struct{}) (any, error) {
		return nil, apperr.Unavailable("TMDB is down.")
	}))
	serve := r.FunctionURL("/catalog")

	cases := []struct {
		name        string
		req         events.LambdaFunctionURLRequest
		status      int
		cache, body string
	}{
		{"query becomes the payload, cached for the action's ttl",
			get("/catalog/movies.search", map[string]string{"q": "heat", "page": "2"}),
			200, "public, max-age=300, s-maxage=86400", `{"data":{"page":"2","q":"heat"}}`},
		{"a refusal is kept a minute",
			get("/catalog/movies.search", nil),
			400, "public, max-age=60, s-maxage=60", `{"error":{"code":"invalid","message":"Search for something."}}`},
		{"a number that isn't one is invalid",
			get("/catalog/movies.search", map[string]string{"q": "heat", "page": "two"}),
			400, "public, max-age=60, s-maxage=60", ""},
		{"an unknown action is not found",
			get("/catalog/movies.nope", nil),
			404, "public, max-age=60, s-maxage=60", ""},
		{"a failure isn't cached",
			get("/catalog/movies.get", nil),
			502, "no-store", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := serve(context.Background(), c.req)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != c.status || res.Headers["Cache-Control"] != c.cache {
				t.Errorf("got %d %q, want %d %q (%s)", res.StatusCode, res.Headers["Cache-Control"], c.status, c.cache, res.Body)
			}
			if c.body != "" && res.Body != c.body {
				t.Errorf("body = %s, want %s", res.Body, c.body)
			}
			if !json.Valid([]byte(res.Body)) {
				t.Errorf("body isn't JSON: %s", res.Body)
			}
		})
	}
}
