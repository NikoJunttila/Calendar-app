package reservations

import (
	"database/sql"
	"fmt"
	"gothstack/plugins/auth"
	"log/slog"
	"net/http"
	"strconv"
	"time" // Needed for date range in ListAvailableTimeSlotsForService

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
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
	UserID         uint       // Owner of the service
	UserName       string     // Placeholder for User's public/business name
	Service        Service    // Details of the selected service
	Slots          []TimeSlot // Available time slots for the service
	Error          string
	WeekStart      time.Time // Start of the current week being displayed
	WeekEnd        time.Time // End of the current week being displayed
	CurrentWeek    uint
	MaxAdvanceDate time.Time // Maximum date that can be booked
	Days           []struct {
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
	fmt.Println(now)
	weekOffset := 0
	if offsetStr := kit.Request.URL.Query().Get("week"); offsetStr != "" {
		weekOffset, _ = strconv.Atoi(offsetStr)
	}
	// Calculate the start of the week (Monday) based on the offset
	weekStart := now.AddDate(0, 0, weekOffset*7)
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

	// Generate all possible time slots for the week
	var allSlots []TimeSlot
	for day := 0; day < 7; day++ {
		currentDate := weekStart.AddDate(0, 0, day)
		dateStr := currentDate.Format("2006-01-02")

		// Skip past dates
		if currentDate.Before(time.Now().Truncate(24 * time.Hour)) {
			slog.Debug("Skipping past date", "date", dateStr)
			continue
		}

		// Get business hours for the user
		businessHours, err := GetBusinessHours(uint(userID))
		if err != nil {
			slog.Error("Failed to get business hours",
				"userID", userID,
				"err", err)
			// Continue with default hours if we can't get business hours
			businessHours = []BusinessHour{}
		}

		// Create a map of business hours by day of week
		businessHoursByDay := make(map[int]BusinessHour)
		for _, bh := range businessHours {
			businessHoursByDay[bh.DayOfWeek] = bh
		}

		// Get the business hours for this day of week
		dayOfWeek := int(currentDate.Weekday())
		businessHour, exists := businessHoursByDay[dayOfWeek]
		if !exists || !businessHour.IsWorkingDay {
			// Skip this day if it's not a working day
			continue
		}

		// Parse business hours
		startTime, err := time.Parse("15:04", businessHour.StartTime)
		if err != nil {
			slog.Error("Failed to parse start time",
				"startTime", businessHour.StartTime,
				"err", err)
			continue
		}
		endTime, err := time.Parse("15:04", businessHour.EndTime)
		if err != nil {
			slog.Error("Failed to parse end time",
				"endTime", businessHour.EndTime,
				"err", err)
			continue
		}

		// Calculate the minimum allowed time for this date
		minAllowedTime := time.Now().Add(time.Duration(settings.MinSchedulingNotice) * time.Hour)

		// Generate time slots for this day
		for hour := startTime.Hour(); hour < endTime.Hour(); hour++ {
			for minute := 0; minute < 60; minute += 30 {
				// Skip if we're past the end time
				if hour == endTime.Hour() && minute >= endTime.Minute() {
					break
				}

				// Create the slot time
				slotTime := time.Date(
					currentDate.Year(), currentDate.Month(), currentDate.Day(),
					hour, minute, 0, 0, time.Now().Location(),
				)

				// Skip if the slot is within the minimum notice period
				if slotTime.Before(minAllowedTime) {
					slog.Debug("Skipping slot within minimum notice period",
						"date", dateStr,
						"time", fmt.Sprintf("%02d:%02d", hour, minute),
						"minAllowedTime", minAllowedTime.Format("2006-01-02 15:04"))
					continue
				}

				timeStr := fmt.Sprintf("%02d:%02d", hour, minute)
				slot := TimeSlot{
					UserID:    uint(userID),
					ServiceID: sql.NullInt64{Int64: int64(serviceID), Valid: true},
					Date:      dateStr,
					Time:      timeStr,
					Duration:  service.Duration,
				}
				allSlots = append(allSlots, slot)
			}
		}
	}

	// Check each slot against booked time ranges
	for i := range allSlots {
		slot := &allSlots[i]
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
		if ranges, exists := bookedTimeRanges[slot.Date]; exists {
			for _, timeRange := range ranges {
				// Check for overlap:
				// If the service would start during a booked slot OR
				// If the service would end during a booked slot OR
				// If the service would completely encompass a booked slot
				if (slotDateTime.Before(timeRange.end) && serviceEndTime.After(timeRange.start)) ||
					(slotDateTime.Equal(timeRange.start)) ||
					(serviceEndTime.Equal(timeRange.end)) {
					slot.IsBooked = true
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
	}

	slog.Debug("Time slots summary",
		"totalSlots", len(allSlots),
		"bookedSlots", len(bookedSlots),
		"weekStart", weekStart.Format("2006-01-02"),
		"weekEnd", weekEnd.Format("2006-01-02"))

	// Log slots per day for detailed debugging
	slotsByDay := make(map[string]struct {
		total  int
		booked int
	})
	for _, slot := range allSlots {
		dayStats := slotsByDay[slot.Date]
		dayStats.total++
		if slot.IsBooked {
			dayStats.booked++
		}
		slotsByDay[slot.Date] = dayStats
	}

	for date, stats := range slotsByDay {
		slog.Debug("Slots for day",
			"date", date,
			"totalSlots", stats.total,
			"bookedSlots", stats.booked,
			"availableSlots", stats.total-stats.booked)
	}

	// Prepare data for the template
	data := CustomerTimeSlotPageData{
		UserID:         uint(userID),
		UserName:       userName,
		Service:        service,
		Slots:          allSlots, // Use filtered slots
		WeekStart:      weekStart,
		WeekEnd:        weekEnd,
		CurrentWeek:    uint(weekOffset),
		MaxAdvanceDate: maxAdvanceDate,
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

// HandleLandingPage displays the landing page for the reservations package
func HandleLandingPage(kit *kit.Kit) error {
	data := LandingPageData{
		Features: []struct {
			Title       string
			Description string
			Icon        string
		}{
			{
				Title:       "Easy Availability Management",
				Description: "Set your working hours, create service slots, and manage your availability with a few clicks.",
				Icon:        "M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z",
			},
			{
				Title:       "24/7 Booking",
				Description: "Let your clients book appointments anytime, anywhere. Reduce manual scheduling and eliminate double bookings.",
				Icon:        "M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z",
			},
			{
				Title:       "Smart Notifications",
				Description: "Automated reminders and notifications keep your clients informed and reduce no-shows.",
				Icon:        "M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9",
			},
		},
	}
	return kit.Render(LandingPage(kit.Request.Context(), data))
}

// HandleDashboard displays the main dashboard for managing reservation settings
func HandleDashboard(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	userEmail := kit.Auth().(auth.Auth).Email

	// Get user's services
	services, err := ListServices(userID, false) // false to get all services, not just active ones
	if err != nil {
		slog.Error("Failed to list services", "userID", userID, "err", err)
		return kit.Render(DashboardPage(DashboardPageData{
			UserID:   userID,
			UserName: userEmail,
			Error:    "Failed to load services",
		}))
	}

	// Get user's bookings
	bookings, err := ListBookings(userID)
	if err != nil {
		slog.Error("Failed to list bookings", "userID", userID, "err", err)
		return kit.Render(DashboardPage(DashboardPageData{
			UserID:   userID,
			UserName: userEmail,
			Error:    "Failed to load bookings",
		}))
	}

	// Calculate stats
	activeServices := 0
	for _, service := range services {
		if service.IsActive {
			activeServices++
		}
	}

	upcomingBookings := 0
	now := time.Now()
	for _, booking := range bookings {
		// Parse the booking date and time
		bookingDateTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", booking.TimeSlot.Date, booking.TimeSlot.Time))
		if err != nil {
			slog.Error("Failed to parse booking datetime", "bookingID", booking.ID, "err", err)
			continue
		}
		if bookingDateTime.After(now) && booking.Status == "confirmed" {
			upcomingBookings++
		}
	}

	// Prepare data for the template
	data := DashboardPageData{
		UserID:   userID,
		UserName: userEmail,
		Stats: struct {
			TotalServices    int
			ActiveServices   int
			TotalBookings    int
			UpcomingBookings int
		}{
			TotalServices:    len(services),
			ActiveServices:   activeServices,
			TotalBookings:    len(bookings),
			UpcomingBookings: upcomingBookings,
		},
		RecentBookings: bookings[:min(5, len(bookings))], // Show only the 5 most recent bookings
	}

	return kit.Render(DashboardPage(data))
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
