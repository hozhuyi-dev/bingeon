// Package router dispatches actions to handlers.
//
// The signed-in API is invoked with {"action": "favourites.add", "payload":
// ...} and always answers {"data": ...} or {"error": ...}. Errors the web may
// show (apperr) go in the body, not as a Lambda function error, so the web
// has one shape to handle; anything else is logged and reported as internal.
// The catalogue serves the same router over HTTP (see FunctionURL).
package router

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"maps"
	"slices"
	"time"

	"bingeon-lambda/internal/platform/apperr"
	"bingeon-lambda/internal/platform/meter"
)

type Request struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type Response struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    apperr.Code `json:"code"`
	Message string      `json:"message"`
	// Until is when a refusal frees up.
	Until *time.Time `json:"until,omitempty"`
}

type Handler func(ctx context.Context, payload json.RawMessage) (any, error)

type route struct {
	handler Handler
	// ttl is how long a shared cache may keep the reply (see FunctionURL);
	// 0 for an action that isn't cached.
	ttl time.Duration
}

type Router struct {
	routes map[string]route
	// LogRequests logs a line per request (see logRequest): the Lambda's,
	// not the tests'.
	LogRequests bool
}

func New() *Router {
	return &Router{routes: map[string]route{}}
}

// Add registers an action.
func (r *Router) Add(action string, h Handler) { r.register(action, route{handler: h}) }

// Cached registers an action whose reply a shared cache may keep for ttl.
func (r *Router) Cached(action string, ttl time.Duration, h Handler) {
	r.register(action, route{handler: h, ttl: ttl})
}

func (r *Router) register(action string, rt route) {
	if _, exists := r.routes[action]; exists {
		panic("router: duplicate action " + action)
	}
	r.routes[action] = rt
}

// Actions lists the registered actions, sorted.
func (r *Router) Actions() []string {
	return slices.Sorted(maps.Keys(r.routes))
}

// Handle runs one request. Its error is always nil: failures go in the body.
func (r *Router) Handle(ctx context.Context, req Request) (Response, error) {
	res, _ := r.handle(ctx, req)
	return res, nil
}

// handle runs one request and also says how long it may be cached.
func (r *Router) handle(ctx context.Context, req Request) (Response, time.Duration) {
	rt, ok := r.routes[req.Action]
	if !ok {
		return Failed(apperr.NotFound("No such action: %q.", req.Action)), 0
	}
	start := time.Now()
	ctx = meter.Start(ctx)
	res := respond(ctx, req.Action, rt.handler, req.Payload)
	if r.LogRequests {
		logRequest(ctx, req.Action, time.Since(start), res)
	}
	return res, rt.ttl
}

// logRequest is one line per request: what it was, how long it took, how
// it ended, and what it cost (table units, calls to TMDB and Gemini), so a
// slow or costly one says why.
func logRequest(ctx context.Context, action string, took time.Duration, res Response) {
	attrs := []any{"action", action, "ms", took.Milliseconds()}
	if res.Error != nil {
		attrs = append(attrs, "error", res.Error.Code)
	}
	counts := meter.Counts(ctx)
	for _, name := range slices.Sorted(maps.Keys(counts)) {
		attrs = append(attrs, name, counts[name])
	}
	slog.InfoContext(ctx, "request", attrs...)
}

func respond(ctx context.Context, action string, h Handler, payload json.RawMessage) Response {
	data, err := h(ctx, payload)
	var shown *apperr.Error
	switch {
	case errors.As(err, &shown):
		return Failed(shown)
	case err != nil:
		slog.ErrorContext(ctx, "action failed", "action", action, "error", err)
		return Response{Error: &ErrorBody{Code: "internal", Message: "Something went wrong."}}
	}
	return Response{Data: data}
}

// Failed is the response for an error the web may show.
func Failed(err *apperr.Error) Response {
	return Response{Error: &ErrorBody{Code: err.Code, Message: err.Message, Until: err.Until}}
}

// Typed adapts a handler that takes a decoded payload.
func Typed[In any](fn func(ctx context.Context, in In) (any, error)) Handler {
	return func(ctx context.Context, payload json.RawMessage) (any, error) {
		var in In
		if len(payload) > 0 && string(payload) != "null" {
			if err := json.Unmarshal(payload, &in); err != nil {
				return nil, apperr.Invalid("Malformed request: %v", err)
			}
		}
		return fn(ctx, in)
	}
}
