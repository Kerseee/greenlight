package main

import (
	"context"
	"net/http"

	"greenlight.kerseeehuang.com/internal/data"
)

// contextKey is a customized type of context key
type contextKey string

const userContextKey = contextKey("user")

// contextSetUser returns a child context with adding user to r by calling r.WithContext.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser retrieves a user instance from r.
// It assumes that a user exists in the request r.Context.
// Calling with a request that has no user will cause panic.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
