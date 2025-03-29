package reservations

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

// InitRoutes sets up the routes for the reservations plugin.
func InitRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {
	router.Group(func(auth chi.Router) {
		// Apply authentication middleware - true means authentication is required
		auth.Use(kit.WithAuthentication(authConfig, true))

		// Route for the main reservations overview page
		auth.Get("/reservations", kit.Handler(HandleReservationsIndex))

		// --- Placeholder Routes for Future Actions ---

		// Example route for viewing/handling the booking confirmation page for a specific slot
		// auth.Get("/reservations/book/{slotID}", kit.Handler(HandleBookingCreateForm)) // Needs handler
		// auth.Post("/reservations/book/{slotID}", kit.Handler(HandleBookingCreatePost)) // Needs handler

		// Example route for handling the cancellation action (could be POST or DELETE)
		// auth.Post("/reservations/bookings/{bookingID}/cancel", kit.Handler(HandleBookingCancel)) // Needs handler
		// auth.Delete("/reservations/bookings/{bookingID}", kit.Handler(HandleBookingCancel)) // Alternative method

		// Example routes for managing time slots
		// auth.Get("/reservations/slots/create", kit.Handler(HandleTimeSlotCreateForm)) // Needs handler
		// auth.Post("/reservations/slots", kit.Handler(HandleTimeSlotCreatePost)) // Needs handler
		// auth.Delete("/reservations/slots/{slotID}", kit.Handler(HandleTimeSlotDelete)) // Needs handler

		// Example routes for settings
		// auth.Get("/reservations/settings", kit.Handler(HandleSettingsView)) // Needs handler
		// auth.Post("/reservations/settings", kit.Handler(HandleSettingsUpdate)) // Needs handler

	})
}
