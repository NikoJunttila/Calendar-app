package reservations

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time" // Needed for date range in ListAvailableTimeSlotsForService

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
	// Assuming User model might be needed later for display names, etc.
	// "gothstack/plugins/auth"
)

// CustomerServiceListPageData holds data for the customer service list page
type CustomerServiceListPageData struct {
	UserID   uint   // To display whose services these are
	UserName string // Placeholder for User's public/business name
	Services []Service
	Error    string // For displaying errors on the page
}

// CustomerTimeSlotPageData holds data for the time slot selection page
type CustomerTimeSlotPageData struct {
	UserID   uint       // Owner of the service
	UserName string     // Placeholder for User's public/business name
	Service  Service    // Details of the selected service
	Slots    []TimeSlot // Available time slots for the service
	Error    string
}

// HandleCustomerServiceList displays the list of active services for a specific user (public view).
// Route example: /book/{userID}
func HandleCustomerServiceList(kit *kit.Kit) error {
	// 1. Get userID from URL
	userIDStr := chi.URLParam(kit.Request, "userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid user ID requested for booking", "id", userIDStr, "err", err)
		data := CustomerServiceListPageData{Error: "Invalid user ID."}
		// Render the list view with an error. Assumes CustomerServiceList template handles Error field.
		// You might want a dedicated error page template later.
		return kit.Render(CustomerServiceList(data)) // Defined in customer_view.templ
	}

	// 2. Fetch user details (Optional - Placeholder)
	// TODO: Fetch user's public/business name from your auth/user system if needed.
	userName := fmt.Sprintf("User %d's Services", userID) // Placeholder

	// 3. Fetch *active* services for this user
	services, err := ListServices(uint(userID), true) // true for activeOnly
	if err != nil {
		slog.Error("Failed to list active services for customer view", "userID", userID, "err", err)
		data := CustomerServiceListPageData{
			UserID:   uint(userID),
			UserName: userName,
			Error:    "Could not load available services at this time.",
		}
		return kit.Render(CustomerServiceList(data)) // Render list view with error
	}

	// 4. Prepare data for the template
	data := CustomerServiceListPageData{
		UserID:   uint(userID),
		UserName: userName,
		Services: services,
	}

	// 5. Render the customer service list view template
	return kit.Render(CustomerServiceList(data)) // Defined in customer_view.templ
}

// HandleCustomerTimeSlotSelection displays available time slots for a selected service.
// Route example: /book/{userID}/service/{serviceID}
func HandleCustomerTimeSlotSelection(kit *kit.Kit) error {
	// 1. Get userID and serviceID from URL
	userIDStr := chi.URLParam(kit.Request, "userID")
	serviceIDStr := chi.URLParam(kit.Request, "serviceID")

	userID, errUser := strconv.ParseUint(userIDStr, 10, 32)
	serviceID, errService := strconv.ParseUint(serviceIDStr, 10, 32)

	if errUser != nil || errService != nil {
		slog.Warn("Invalid user or service ID in time slot request", "userID", userIDStr, "serviceID", serviceIDStr, "errUser", errUser, "errService", errService)
		// Redirect to a safe place, maybe home or a generic error page.
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	// 2. Fetch user details (Optional - Placeholder)
	// TODO: Fetch user's public/business name
	userName := fmt.Sprintf("Book appointment with User %d", userID) // Placeholder

	// 3. Fetch the selected service details (ensure it belongs to userID and is active)
	service, err := GetService(uint(serviceID), uint(userID))
	if err != nil {
		slog.Error("Customer view: Failed to get service", "userID", userID, "serviceID", serviceID, "err", err)
		// Redirect back to the service list for the user, as the service might not exist or belong to them.
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}
	if !service.IsActive {
		slog.Warn("Customer view: Attempted to book inactive service", "userID", userID, "serviceID", serviceID)
		// Service exists but isn't bookable. Redirect back to the list.
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	// 4. Fetch available time slots for this specific service
	// Define a date range (e.g., today to X days in the future)
	// Consider using user settings for MaxSchedulingAdvance later.
	now := time.Now()
	// Default lookahead: 60 days. Fetch from Settings if implemented.
	lookaheadDays := 60
	future := now.AddDate(0, 0, lookaheadDays)

	slots, err := ListAvailableTimeSlotsForService(uint(userID), uint(serviceID), now, future)
	if err != nil {
		slog.Error("Customer view: Failed to list available time slots", "userID", userID, "serviceID", serviceID, "err", err)
		data := CustomerTimeSlotPageData{
			UserID:   uint(userID),
			UserName: userName,
			Service:  service,
			Error:    "Could not load available times for this service.",
		}
		// Render the selection page with an error message.
		return kit.Render(CustomerTimeSlotSelection(data)) // Defined in customer_view.templ
	}

	// 5. Prepare data for the template
	data := CustomerTimeSlotPageData{
		UserID:   uint(userID),
		UserName: userName,
		Service:  service,
		Slots:    slots,
	}

	// 6. Render the time slot selection view template
	return kit.Render(CustomerTimeSlotSelection(data)) // Defined in customer_view.templ
}
