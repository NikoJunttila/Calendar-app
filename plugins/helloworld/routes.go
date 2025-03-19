package helloworld

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {
	// Public endpoint - no authentication required
	router.Get("/hello", kit.Handler(handleHello))
	router.Get("/hello/read", kit.Handler(handleReadHello))
	router.Post("/helloworld", kit.Handler(handlePostHello))
	// Protected endpoint - authentication required
	// Using router.Group to apply authentication middleware
	router.Group(func(auth chi.Router) {
		// Apply authentication middleware with the true parameter to require authentication
		auth.Use(kit.WithAuthentication(authConfig, true))

		// This endpoint requires authentication
		auth.Get("/hello/authenticated", kit.Handler(handleAuthenticatedHello))
	})
}
