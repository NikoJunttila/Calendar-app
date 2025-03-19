package app

import (
	"gothstack/app/handlers"
	"gothstack/app/views/errors"
	"gothstack/plugins/auth"
	"gothstack/plugins/calendar"
	"gothstack/plugins/helloworld"
	"log/slog"

	"github.com/anthdm/superkit/kit"
	"github.com/anthdm/superkit/kit/middleware"
	"github.com/go-chi/chi/v5"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Define your global middleware
func InitializeMiddleware(router *chi.Mux) {
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.WithRequest)
}

// Define your routes in here
func InitializeRoutes(router *chi.Mux) {
	// Authentication plugin
	//
	// By default the auth plugin is active, to disable the auth plugin
	// you will need to pass your own handler in the `AuthFunc`` field
	// of the `kit.AuthenticationConfig`.
	authConfig := kit.AuthenticationConfig{
		AuthFunc:    auth.AuthenticateUser,
		RedirectURL: "/login",
	}
	auth.InitializeRoutes(router, authConfig)
	helloworld.InitRoutes(router, authConfig)
	calendar.InitRoutes(router, authConfig)
	// Routes that "might" have an authenticated user
	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, false)) // strict set to false
		app.Get("/unauthorized", kit.Handler(handlers.HandleUnauthorized))
		// Routes
		app.Get("/", kit.Handler(handlers.HandleLandingIndex))
	})

	// Authenticated routes
	//
	// Routes that "must" have an authenticated user or else they
	// will be redirected to the configured redirectURL, set in the
	// AuthenticationConfig.
	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, true)) // strict set to true

		// Routes
		// app.Get("/path", kit.Handler(myHandler.HandleIndex))
	})
}

// NotFoundHandler that will be called when the requested path could
// not be found.
func NotFoundHandler(kit *kit.Kit) error {
	return kit.Render(errors.Error404())
}

// ErrorHandler that will be called on errors return from application handlers.
func ErrorHandler(kit *kit.Kit, err error) {
	slog.Error("internal server error", "err", err.Error(), "path", kit.Request.URL.Path)
	kit.Render(errors.Error500())
}
