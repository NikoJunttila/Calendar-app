package calendar

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {
	router.Group(func(auth chi.Router) {
		// Apply authentication middleware with the true parameter to require authentication
		auth.Use(kit.WithAuthentication(authConfig, true))
		auth.Get("/calendars", kit.Handler(HandleCalendarList))
		auth.Get("/calendars/create", kit.Handler(HandleCalendarCreate))
		auth.Post("/calendars/create", kit.Handler(HandleCalendarCreatePost))
		auth.Get("/calendars/{id}", kit.Handler(HandleCalendarView))
		auth.Get("/calendars/{id}/entries/create", kit.Handler(HandleCalendarEntryCreate))
		auth.Post("/calendars/{id}/entries/create", kit.Handler(HandleCalendarEntryCreatePost))

		auth.Get("/calendars/{id}/month", kit.Handler(HandleCalendarViewByMonth))
		auth.Get("/calendars/{id}/{year}/{month}", kit.Handler(HandleCalendarViewByMonth))

		// Work resources
		auth.Get("/calendars/{id}/resources", kit.Handler(HandleWorkResourceList))

		// Create a new work resource
		auth.Get("/calendars/{id}/resources/create", kit.Handler(HandleWorkResourceCreate))
		auth.Post("/calendars/{id}/resources/create", kit.Handler(HandleWorkResourceCreatePost))

		// Edit a work resource
		auth.Get("/calendars/{id}/resources/{resource_id}/edit", kit.Handler(HandleWorkResourceEdit))
		auth.Post("/calendars/{id}/resources/{resource_id}/edit", kit.Handler(HandleWorkResourceEditPost))

		// Delete a work resource
		auth.Delete("/calendars/{id}/resources/{resource_id}", kit.Handler(HandleWorkResourceDelete))

	})
}
