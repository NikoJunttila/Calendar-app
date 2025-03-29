package reservations

import (
	"gothstack/plugins/auth"
	"log/slog"
	"time" // Added for date calculations

	"github.com/anthdm/superkit/kit"
	// v "github.com/anthdm/superkit/validate" // Not used in this basic example yet
	// "github.com/go-chi/chi/v5" // Not used in this basic example yet
	// "net/http" // Not used directly yet
	// "strconv" // Not used yet
	// "fmt" // Not used yet
)

// ReservationsPageData holds data for the main reservations page
type ReservationsPageData struct {
	AvailableSlots []TimeSlot // Renamed from TimeSlots for clarity
	Bookings       []Booking
	Settings       Setting // User settings might be relevant
	// Add FormValues/Errors here later if forms are added
}

// HandleReservationsIndex displays the main reservations overview page.
// It lists available time slots and current bookings.
func HandleReservationsIndex(kit *kit.Kit) error {
	// Get authenticated user's ID

	userID := kit.Auth().(auth.Auth).UserID

	// Fetch user settings (might influence displayed slots, e.g., timezone, notice period)
	settings, err := GetSettings(userID)
	if err != nil {
		// Log the error or handle it gracefully
		// For now, use default settings if fetch fails
		settings = Setting{Timezone: "UTC", MaxSchedulingAdvance: 60} // Example default
		slog.Info("Failed to get settings for user %d: %v", userID, err)
		// Optionally return error: return fmt.Errorf("could not load settings: %w", err)
	}

	// Fetch available time slots (e.g., for the next N days based on settings)
	// Using ListAvailableTimeSlots requires a date range.
	startDate := time.Now()
	// Use MaxSchedulingAdvance from settings (defaulting if needed)
	endDate := startDate.AddDate(0, 0, settings.MaxSchedulingAdvance)
	availableSlots, err := ListAvailableTimeSlots(userID, startDate, endDate)
	if err != nil {
		slog.Info("Failed to list available time slots for user %d: %v", userID, err)
		// Decide how to handle: return error, show message, etc.
		// For now, continue with an empty list
		availableSlots = []TimeSlot{}
		// Optionally return err
	}

	// Fetch existing bookings for the user
	bookings, err := ListBookings(userID)
	if err != nil {
		slog.Info("Failed to list bookings for user %d: %v", userID, err)
		// Decide how to handle
		bookings = []Booking{}
		// Optionally return err
	}

	// Prepare data for the template
	data := ReservationsPageData{
		AvailableSlots: availableSlots,
		Bookings:       bookings,
		Settings:       settings,
	}

	// Render the Templ component
	return kit.Render(ReservationsIndex(data))
}
