package book

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {
	router.Group(func(auth chi.Router) {
		// Apply authentication middleware with the true parameter to require authentication
		auth.Use(kit.WithAuthentication(authConfig, true))

		auth.Get("/books", kit.Handler(HandleBookList))

		// Book creation
		auth.Get("/books/create", kit.Handler(HandleBookCreate))
		auth.Post("/books/create", kit.Handler(HandleBookCreatePost))

		// Book view/read/delete
		auth.Get("/books/{id}", kit.Handler(HandleBookView))
		auth.Get("/books/{id}/read", kit.Handler(HandleBookRead))
		auth.Delete("/books/{id}", kit.Handler(HandleBookDelete))

	})
}
