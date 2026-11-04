// Package user carries the signed-in user through a request: cmd/api puts
// the invocation's Cognito identity in the context, never an id from the
// payload.
package user

import "context"

type key struct{}

// With is ctx for the user id.
func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key{}, id)
}

// ID is the signed-in user's id; "" outside a signed-in request (cmd/api
// refuses those before any handler runs).
func ID(ctx context.Context) string {
	id, _ := ctx.Value(key{}).(string)
	return id
}
