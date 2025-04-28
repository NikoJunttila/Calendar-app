package reservations

import (
	"fmt" // Added for string formatting
	"gothstack/plugins/auth"
	"log/slog" // Added for parsing form and redirect
	"strconv"  // Added for string conversion
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/anthdm/superkit/kit"
)

// ReservationsPageData holds data for the main reservations overview page.
type ReservationsPageData struct {
	AvailableSlots []TimeSlot // Renamed from TimeSlots for clarity
	Bookings       []Booking
	Settings       Setting // User settings might be relevant
	ErrorMessage string
}

// SettingsFormData holds data for the settings form
type SettingsFormData struct {
	Setting        Setting // Current settings to display
	SuccessMessage string
	ErrorMessage   string
	// Can add validation errors map here: FormErrors v.Errors
}

// HandleReservationsIndex displays the main reservations overview page.
func HandleReservationsIndex(kit *kit.Kit) error {
	// Get authenticated user's ID
	userID := kit.Auth().(auth.Auth).UserID
	var data  ReservationsPageData
	// Fetch existing bookings for the user
	bookings, err := ListBookings(userID)
	if err != nil {
		// Use Error level as this is a core data fetching failure
		slog.Error(
			"Failed to list bookings",
			"userID", userID,
			"err", err,
		)
		// Decide how to handle
		data.ErrorMessage = fmt.Sprintf("Failed to get bookings %v", err)
		return kit.Render(ReservationsIndex(data))
	}
	data.Bookings = bookings
	fmt.Println(data)
	// Render the Templ component
	return kit.Render(ReservationsIndex(data))
}

// HandleSettingsView displays the user's reservation settings form.
func HandleSettingsView(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID

	settings, err := GetSettings(userID)
	if err != nil {
		// Log and potentially show an error page or redirect
		slog.Error(
			"Failed to get settings for view",
			"userID", userID,
			"err", err,
		)
		// Render the form anyway, maybe with an error message or default values
		// Or return a proper error response:
		// kit.Render(ErrorPage("Could not load settings.")) // Example
		return fmt.Errorf("failed to get settings for user %d: %w", userID, err) // Or render form with error
	}

	data := SettingsFormData{
		Setting: settings,
	}

	return kit.Render(SettingsForm(data))
}

// HandleSettingsUpdate processes the submission of the settings form.
func HandleSettingsUpdate(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID

	// Parse the form data
	if err := kit.Request.ParseForm(); err != nil {
		slog.Error("Failed to parse settings form", "userID", userID, "err", err)
		// Re-render form with a generic error message
		settings, _ := GetSettings(userID) // Attempt to get current settings for re-render
		data := SettingsFormData{
			Setting:      settings,
			ErrorMessage: "Failed to process form submission.",
		}
		return kit.Render(SettingsForm(data))
	}

	// Extract values and create Setting struct
	// Note: Production code should have more robust parsing & validation
	minNotice, _ := strconv.Atoi(kit.Request.FormValue("min_scheduling_notice"))
	maxAdvance, _ := strconv.Atoi(kit.Request.FormValue("max_scheduling_advance"))

	updatedSetting := Setting{
		UserID:               userID, // Crucial: Associate with the correct user
		Timezone:             kit.Request.FormValue("timezone"),
		NotificationEmail:    kit.Request.FormValue("notification_email") == "on", // Checkbox value
		NotificationSMS:      kit.Request.FormValue("notification_sms") == "on",   // Checkbox value
		CalendarView:         kit.Request.FormValue("calendar_view"),
		MinSchedulingNotice:  minNotice,
		MaxSchedulingAdvance: maxAdvance,
		// ID, CreatedAt, UpdatedAt will be handled by UpdateSettings/GORM
	}

	// Fetch the existing settings to get the ID (UpdateSettings needs it implicitly via Model)
	// Alternatively, pass the ID in the form as a hidden field.
	existingSettings, err := GetSettings(userID)
	if err != nil {
		slog.Error("Failed to get existing settings before update", "userID", userID, "err", err)
		data := SettingsFormData{
			Setting:      updatedSetting, // Show submitted values
			ErrorMessage: "Could not save settings: failed to retrieve existing record.",
		}
		return kit.Render(SettingsForm(data))
	}
	updatedSetting.ID = existingSettings.ID               // Ensure ID is set for update
	updatedSetting.CreatedAt = existingSettings.CreatedAt // Preserve CreatedAt

	// Call the UpdateSettings function from types.go
	savedSettings, err := UpdateSettings(updatedSetting)
	if err != nil {
		slog.Error(
			"Failed to update settings",
			"userID", userID,
			"err", err,
		)
		// Re-render form with submitted values and an error message
		data := SettingsFormData{
			Setting:      updatedSetting, // Show the values the user tried to save
			ErrorMessage: fmt.Sprintf("Failed to save settings: %v", err),
		}
		return kit.Render(SettingsForm(data))
	}

	// Success! Re-render the form with the updated settings and a success message
	data := SettingsFormData{
		Setting:        savedSettings, // Show the newly saved settings
		SuccessMessage: "Settings updated successfully!",
	}
	return kit.Render(SettingsForm(data))
}

func handleDeleteBooking(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	bookingIDStr := chi.URLParam(kit.Request, "id")
	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid service ID requested for delete", "id", bookingIDStr, "err", err)
		// Respond appropriately for HTMX if used (e.g., no content, error message)
		return kit.Redirect(http.StatusSeeOther, "/reservations/services") // Fallback redirect
	}

	// Optional: Check if service can be deleted (e.g., no future bookings)

	err = CancelBooking(uint(bookingID), userID)
	if err != nil {
		slog.Error("Failed to delete service", "userID", userID, "serviceID", bookingID, "err", err)
		// Respond with error for HTMX or set flash message and redirect
		// For HTMX, could return a toast message component
		//kit.Session.Put(kit.Request.Context(), "flash_error", "Failed to delete service. It might be in use.")
		return kit.Redirect(http.StatusSeeOther, "/reservations")
	}

	// Success
	// For HTMX: Return HTTP 200 OK (often with no content, as target element is removed)
	// Or set flash message and redirect
	//kit.Session.Put(kit.Request.Context(), "flash_success", "Service deleted successfully.")
	// If using HTMX to remove the row, we might not redirect, just return OK.
	// For now, redirecting to the list:
	return kit.Redirect(http.StatusSeeOther, "/reservations")
}
