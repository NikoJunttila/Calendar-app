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

	// Get available slots
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

	// Get booked slots for the week
	bookedSlots, err := ListBookedTimeSlotsForWeek(uint(userID), weekStart)
	if err != nil {
		slog.Error("Failed to get booked slots",
			"userID", userID,
			"weekStart", weekStart,
			"err", err)
		// Continue with available slots even if we can't get booked ones
	}

	slog.Debug("Retrieved booked slots for week",
		"userID", userID,
		"weekStart", weekStart.Format("2006-01-02"),
		"weekEnd", weekEnd.Format("2006-01-02"),
		"bookedSlotsCount", len(bookedSlots))

	// Create a map of booked time ranges for efficient lookup
	bookedTimeRanges := make(map[string][]struct {
		start time.Time
		end   time.Time
	})

	for _, slot := range bookedSlots {
		// Parse the date and time of the booked slot
		slotDateTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", slot.Date, slot.Time))
		if err != nil {
			slog.Error("Failed to parse booked slot datetime",
				"date", slot.Date,
				"time", slot.Time,
				"err", err)
			continue
		}

		// Calculate end time based on the slot's duration
		endTime := slotDateTime.Add(time.Duration(slot.Duration) * time.Minute)

		// Store the time range for this date
		bookedTimeRanges[slot.Date] = append(bookedTimeRanges[slot.Date], struct {
			start time.Time
			end   time.Time
		}{
			start: slotDateTime,
			end:   endTime,
		})

		slog.Debug("Added booked time range",
			"date", slot.Date,
			"startTime", slotDateTime.Format("15:04"),
			"endTime", endTime.Format("15:04"),
			"duration", slot.Duration)
	}

	// Instead of using a simple map, check for time range overlaps
	var allSlots []TimeSlot
	for _, slot := range slots {
		slotDateTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", slot.Date, slot.Time))
		if err != nil {
			slog.Error("Failed to parse slot datetime",
				"date", slot.Date,
				"time", slot.Time,
				"err", err)
			continue
		}

		// Calculate the end time of this service if it were to start at this slot
		serviceEndTime := slotDateTime.Add(time.Duration(service.Duration) * time.Minute)

		// Check if this slot overlaps with any booked time ranges
		isBooked := false
		if ranges, exists := bookedTimeRanges[slot.Date]; exists {
			for _, timeRange := range ranges {
				// Check for overlap:
				// If the service would start during a booked slot OR
				// If the service would end during a booked slot OR
				// If the service would completely encompass a booked slot
				if (slotDateTime.Before(timeRange.end) && serviceEndTime.After(timeRange.start)) ||
					(slotDateTime.Equal(timeRange.start)) ||
					(serviceEndTime.Equal(timeRange.end)) {
					isBooked = true
					slog.Debug("Slot overlaps with booked time range",
						"date", slot.Date,
						"slotTime", slot.Time,
						"serviceDuration", service.Duration,
						"bookedStart", timeRange.start.Format("15:04"),
						"bookedEnd", timeRange.end.Format("15:04"))
					break
				}
			}
		}

		slot.IsBooked = isBooked
		allSlots = append(allSlots, slot)
	}

	slog.Debug("Time slots summary",
		"totalSlots", len(slots),
		"bookedSlots", len(bookedSlots),
		"availableSlots", len(allSlots),
		"weekStart", weekStart.Format("2006-01-02"),
		"weekEnd", weekEnd.Format("2006-01-02"))

	// Log available slots per day for detailed debugging
	slotsByDay := make(map[string]int)
	for _, slot := range allSlots {
		slotsByDay[slot.Date]++
	}
	for date, count := range slotsByDay {
		slog.Debug("Available slots for day",
			"date", date,
			"count", count)
	}

	// Prepare data for the template
	data := CustomerTimeSlotPageData{
		UserID:      uint(userID),
		UserName:    userName,
		Service:     service,
		Slots:       allSlots, // Use filtered slots
		WeekStart:   weekStart,
		WeekEnd:     weekEnd,
		CurrentWeek: uint(weekOffset),
	}

	// Prepare days data
	for i := 0; i < 7; i++ {
		currentDate := weekStart.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		daySlots := filterSlotsForDate(allSlots, dateStr) // Use filtered slots
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
