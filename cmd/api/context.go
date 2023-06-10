package main

import (
	"context"
	"net/http"

	"github.com/ricci2511/greenlight-api/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// This helper to retrieve the user struct from the request context should only be used when we expect
// that it is present in the current request context. Otherwise, it is considered an unexpected error.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	// Should not happen, therefore panic.
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
