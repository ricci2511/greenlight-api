package main

import (
	"expvar"
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
	// Get client's real IP through the X-Forwarded-For header set by Caddy's reverse proxy.
	// If not used, the rate limiter will limit the ip of the reverse proxy instead of the client.
	r.Use(middleware.RealIP)
	r.Use(app.recoverPanic)
	r.Use(app.metrics)

	r.Get("/v1/healthcheck", app.healthcheckHandler)

	r.Route("/v1/users", func(r chi.Router) {
		r.Post("/", app.createUserHandler)
		r.Put("/activate", app.activateUserHandler)
		r.Put("/password", app.updateUserPasswordHandler)
	})

	r.Route("/v1/tokens", func(r chi.Router) {
		r.Post("/authentication", app.createAuthenticationTokenHandler)
		r.Post("/password-reset", app.createPasswordResetTokenHandler)
	})

	r.Route("/v1/movies", func(r chi.Router) {
		r.Post("/", app.requirePermission("movies:write", app.createMovieHandler))
		r.Get("/", app.requirePermission("movies:read", app.listMoviesHandler))

		r.Get("/{id}", app.requirePermission("movies:read", app.showMovieHandler))
		r.Patch("/{id}", app.requirePermission("movies:write", app.updateMovieHandler))
		r.Delete("/{id}", app.requirePermission("movies:write", app.deleteMovieHandler))
	})

	r.Mount("/debug/vars", expvar.Handler())

	return r
}
