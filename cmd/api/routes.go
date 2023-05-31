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

	r.Get("/v1/healthcheck", app.healthcheckHandler)

	r.Route("/v1/movies", func(r chi.Router) {
		r.Post("/", app.createMovieHandler)
		r.Get("/", app.listMoviesHandler)
		r.Get("/{id}", app.showMovieHandler)
		r.Patch("/{id}", app.updateMovieHandler)
		r.Delete("/{id}", app.deleteMovieHandler)
	})

	return r
}
