package auth

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func InitializeRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {
	/* 	authConfig := kit.AuthenticationConfig{
		AuthFunc:    AuthenticateUser,
		RedirectURL: "/login",
	} */
	// Routes that don't require any authentication
	// These endpoints are publicly accessible
	router.Get("/email/verify", kit.Handler(HandleEmailVerify))
	router.Post("/resend-email-verification", kit.Handler(HandleResendVerificationCode))

	// First router group: Authentication-related routes (login/signup flows)
	// The false parameter in WithAuthentication means authentication is NOT required
	// These routes are for unauthenticated users who need to authenticate
	router.Group(func(auth chi.Router) {
		auth.Use(kit.WithAuthentication(authConfig, false))
		auth.Get("/login", kit.Handler(HandleLoginIndex))      // Show login page
		auth.Post("/login", kit.Handler(HandleLoginCreate))    // Process login form
		auth.Delete("/logout", kit.Handler(HandleLoginDelete)) // Log user out
		auth.Get("/signup", kit.Handler(HandleSignupIndex))    // Show signup page
		auth.Post("/signup", kit.Handler(HandleSignupCreate))  // Process signup form
	})

	// Second router group: Protected routes (require authentication)
	// The true parameter in WithAuthentication means authentication IS required
	// These routes are for already authenticated users
	router.Group(func(auth chi.Router) {
		auth.Use(kit.WithAuthentication(authConfig, true))
		auth.Get("/profile", kit.Handler(HandleProfileShow))   // View user profile
		auth.Put("/profile", kit.Handler(HandleProfileUpdate)) // Update user profile
	})
}
