package reservations

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

// InitRoutes sets up the routes for the reservations plugin.
func InitRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {
	// --- Public Customer-Facing Routes ---
	router.Get("/book/{userID}", kit.Handler(HandleCustomerServiceList))
	router.Get("/book/{userID}/service/{serviceID}", kit.Handler(HandleCustomerTimeSlotSelection))
	router.Get("/book/{userID}/service/{serviceID}/confirm", kit.Handler(HandleBookingConfirmationView))
	router.Post("/book/{userID}/service/{serviceID}/confirm", kit.Handler(HandleBookingConfirmationPost))
	router.Get("/book/{userID}/service/{serviceID}/success", kit.Handler(HandleBookingSuccess))
	// TODO: Add route for booking confirmation GET/POST, e.g.:
	// router.Get("/book/{userID}/service/{serviceID}/confirm", kit.Handler(HandleBookingConfirmationView))
	// router.Post("/book/{userID}/service/{serviceID}/confirm", kit.Handler(HandleBookingConfirmationPost))

	// --- Authenticated Admin Routes ---
	router.Group(func(auth chi.Router) {
		// Apply authentication middleware - true means authentication is required
		auth.Use(kit.WithAuthentication(authConfig, true))

		// Route for the main reservations overview page
		auth.Get("/reservations", kit.Handler(HandleReservationsIndex))

		// --- Settings Routes ---
		auth.Get("/reservations/settings", kit.Handler(HandleSettingsView))
		auth.Post("/reservations/settings", kit.Handler(HandleSettingsUpdate))

		// --- Service Routes ---
		auth.Get("/reservations/services", kit.Handler(HandleServiceList))                       // List services
		auth.Get("/reservations/services/create", kit.Handler(HandleServiceCreateView))          // Show create form
		auth.Post("/reservations/services/create", kit.Handler(HandleServiceCreatePost))         // Handle create form submission
		auth.Get("/reservations/services/{serviceID}/edit", kit.Handler(HandleServiceEditView))  // Show edit form
		auth.Post("/reservations/services/{serviceID}/edit", kit.Handler(HandleServiceEditPost)) // Handle edit form submission
		auth.Delete("/reservations/services/{serviceID}", kit.Handler(HandleServiceDelete))      // Handle deletion

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

		// Business hours management
		auth.Get("/reservations/business-hours", kit.Handler(HandleBusinessHoursView))
		auth.Post("/reservations/business-hours", kit.Handler(HandleBusinessHoursUpdate))

		// Special dates routes
		auth.Get("/reservations/special-dates", kit.Handler(HandleSpecialDateView))
		auth.Post("/reservations/special-dates", kit.Handler(HandleSpecialDateCreate))
		auth.Post("/reservations/special-dates/range", kit.Handler(HandleSpecialDateRangeCreate))
		auth.Post("/reservations/special-dates/{id}/delete", kit.Handler(HandleSpecialDateDelete))
	})
}
