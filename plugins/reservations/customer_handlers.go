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
	UserID      uint       // Owner of the service
	UserName    string     // Placeholder for User's public/business name
	Service     Service    // Details of the selected service
	Slots       []TimeSlot // Available time slots for the service
	Error       string
	WeekStart   time.Time // Start of the current week being displayed
	WeekEnd     time.Time // End of the current week being displayed
	CurrentWeek uint
	Days        []struct {
		Date     time.Time
		DateStr  string
		DaySlots []TimeSlot
	}
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
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	// 2. Fetch user details (Optional - Placeholder)
	userName := fmt.Sprintf("Book appointment with User %d", userID)

	// 3. Fetch the selected service details
	service, err := GetService(uint(serviceID), uint(userID))
	if err != nil {
		slog.Error("Customer view: Failed to get service", "userID", userID, "serviceID", serviceID, "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}
	if !service.IsActive {
		slog.Warn("Customer view: Attempted to book inactive service", "userID", userID, "serviceID", serviceID)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	// Get user's settings for scheduling advance notice
	settings, err := GetSettings(uint(userID))
	if err != nil {
		slog.Warn("Failed to get user settings, using defaults",
			"userID", userID,
			"err", err)
		settings = Setting{
			MinSchedulingNotice:  24, // 24 hours
			MaxSchedulingAdvance: 60, // 60 days
		}
	}

	// Calculate the week range
	now := time.Now()
	weekOffset := 0
	if offsetStr := kit.Request.URL.Query().Get("week"); offsetStr != "" {
		weekOffset, _ = strconv.Atoi(offsetStr)
	}
	// Calculate the start of the week (Monday) based on the offset
	weekStart := now.AddDate(0, 0, settings.MinSchedulingNotice+weekOffset*7)
	// Move to Monday (1) if current day is Sunday (0), otherwise move back to previous Monday
	if weekStart.Weekday() == time.Sunday {
		weekStart = weekStart.AddDate(0, 0, 1)
	} else {
		weekStart = weekStart.AddDate(0, 0, -int(weekStart.Weekday()-1))
	}
	weekEnd := weekStart.AddDate(0, 0, 6) // Move to Sunday

	// Ensure we don't exceed the maximum scheduling advance
	maxAdvanceDate := now.AddDate(0, 0, settings.MaxSchedulingAdvance)
	if weekEnd.After(maxAdvanceDate) {
		weekEnd = maxAdvanceDate
	}

	slots, err := ListAvailableTimeSlotsForService(uint(userID), uint(serviceID), weekStart, weekEnd)
	if err != nil {
		slog.Error("Customer view: Failed to list available time slots",
			"userID", userID,
			"serviceID", serviceID,
			"err", err)
		data := CustomerTimeSlotPageData{
			UserID:      uint(userID),
			UserName:    userName,
			Service:     service,
			CurrentWeek: uint(weekOffset),
			Error:       "Could not load available times for this service.",
		}
		return kit.Render(CustomerTimeSlotSelection(data))
	}

	// Prepare data for the template
	data := CustomerTimeSlotPageData{
		UserID:      uint(userID),
		UserName:    userName,
		Service:     service,
		Slots:       slots,
		WeekStart:   weekStart,
		WeekEnd:     weekEnd,
		CurrentWeek: uint(weekOffset),
	}

	// Prepare days data
	for i := 0; i < 7; i++ {
		currentDate := weekStart.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		daySlots := filterSlotsForDate(slots, dateStr)
		data.Days = append(data.Days, struct {
			Date     time.Time
			DateStr  string
			DaySlots []TimeSlot
		}{
			Date:     currentDate,
			DateStr:  dateStr,
			DaySlots: daySlots,
		})
	}

	return kit.Render(CustomerTimeSlotSelection(data))
}
