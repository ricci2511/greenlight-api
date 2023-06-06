package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	// Handle 404 and 405 errors.
	r.NotFound(app.notFoundReponse)
	r.MethodNotAllowed(app.methodNotAllowedRespone)

	// Standard middleware stack.
	r.Use(app.recoverPanic)
	r.Use(app.rateLimit)
	r.Use(app.authenticate)
	r.Use(middleware.CleanPath)

	r.Get("/v1/healthcheck", app.healthcheckHandler)

	r.Route("/v1/users", func(r chi.Router) {
		r.Post("/", app.createUserHandler)
		r.Put("/activate", app.activateUserHandler)
	})

	r.Route("/v1/tokens", func(r chi.Router) {
		r.Post("/authentication", app.createAuthenticationTokenHandler)
	})

	r.Route("/v1/movies", func(r chi.Router) {
		// Any /v1/movies request requires an activated user.
		r.Use(app.requireActivatedUser)

		r.Post("/", app.createMovieHandler)
		r.Get("/", app.listMoviesHandler)
		r.Get("/{id}", app.showMovieHandler)
		r.Patch("/{id}", app.updateMovieHandler)
		r.Delete("/{id}", app.deleteMovieHandler)
	})

	return r
}
