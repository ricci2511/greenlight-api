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
	r.Use(middleware.CleanPath)
	r.Use(app.authenticate)
	r.Use(app.rateLimit)
	r.Use(app.enableCors)
	r.Use(app.recoverPanic)

	r.Get("/v1/healthcheck", app.healthcheckHandler)

	r.Route("/v1/users", func(r chi.Router) {
		r.Post("/", app.createUserHandler)
		r.Put("/activate", app.activateUserHandler)
	})

	r.Route("/v1/tokens", func(r chi.Router) {
		r.Post("/authentication", app.createAuthenticationTokenHandler)
	})

	r.Route("/v1/movies", func(r chi.Router) {
		r.Post("/", app.requirePermission("movies:write", app.createMovieHandler))
		r.Get("/", app.requirePermission("movies:read", app.listMoviesHandler))

		r.Get("/{id}", app.requirePermission("movies:read", app.showMovieHandler))
		r.Patch("/{id}", app.requirePermission("movies:write", app.updateMovieHandler))
		r.Delete("/{id}", app.requirePermission("movies:write", app.deleteMovieHandler))
	})

	return r
}
