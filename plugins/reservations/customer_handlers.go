package reservations

import (
	"fmt"
	"gothstack/app/translations"
	"gothstack/plugins/auth"
	"gothstack/plugins/calendar"
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
		Date      time.Time
		DateStr   string
		DaySlots  []TimeSlot
		IsHoliday bool
		Holiday   calendar.FinnishHoliday
	}
}

// HandleCustomerServiceList displays the list of active services for a specific user (public view).
// Route example: /book/{userID}
func HandleCustomerServiceList(kit *kit.Kit) error {
	// 1. Get userID from URL
	currentLangCode := translations.M.DefaultLanguage() // Default if not found
	if langVal := kit.Request.Context().Value(translations.ContextKey{}); langVal != nil {
		if code, ok := langVal.(string); ok {
			currentLangCode = code
		}
	}
	availableLangs := translations.M.GetLanguages()

	currentPath := kit.Request.URL.Path

	userIDStr := chi.URLParam(kit.Request, "userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid user ID requested for booking", "id", userIDStr, "err", err)
		data := CustomerServiceListPageData{Error: "Invalid user ID."}
		// Render the list view with an error. Assumes CustomerServiceList template handles Error field.
		// You might want a dedicated error page template later.
		return kit.Render(CustomerServiceList(data, kit.Request.Context(), availableLangs, currentLangCode, currentPath))
	}
	query := kit.Request.URL.Query()
	serviceIDparams := query["id"]

	// 2. Fetch user details
	set, err := GetSettings(uint(userID))
	var userName string
	if err != nil {
		userName = "error"
	} else {
		userName = set.BusinessName
	}

	// 3. Fetch *active* services for this user
	services, err := ListServices(uint(userID), true) // true for activeOnly
	if err != nil {
		slog.Error("Failed to list active services for customer view", "userID", userID, "err", err)
		data := CustomerServiceListPageData{
			UserID:   uint(userID),
			UserName: userName,
			Error:    "Could not load available services at this time.",
		}
		return kit.Render(CustomerServiceList(data, kit.Request.Context(), availableLangs, currentLangCode, currentPath))
	}
	if len(serviceIDparams) > 0 {
		var newServices []Service
		for _, sId := range serviceIDparams {
			id, err := strconv.ParseUint(sId, 10, 64)
			if err != nil {
				fmt.Println("error: ", err)
			}
			for _, serv := range services {
				if serv.ID == uint(id) {
					newServices = append(newServices, serv)
				}
			}
		}
		services = newServices
	}

	// 4. Prepare data for the template
	data := CustomerServiceListPageData{
		UserID:   uint(userID),
		UserName: userName,
		Services: services,
	}

	// 5. Render the customer service list view template
	return kit.Render(CustomerServiceList(data, kit.Request.Context(), availableLangs, currentLangCode, currentPath)) // Defined in customer_view.templ
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

	// Use the existing function to generate time slots (this handles special dates)
	allSlots, err := ListAvailableTimeSlotsForService(uint(userID), uint(serviceID), weekStart, weekEnd)
	if err != nil {
		slog.Error("Failed to get available time slots", "userID", userID, "serviceID", serviceID, "err", err)
		return err
	}

	// Apply minimum scheduling notice filter
	minAllowedTime := now.Add(time.Duration(settings.MinSchedulingNotice) * time.Hour)
	var filteredSlots []TimeSlot
	for _, slot := range allSlots {
		slotDateTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", slot.Date, slot.Time))
		if err != nil {
			slog.Error("Failed to parse slot datetime", "date", slot.Date, "time", slot.Time, "err", err)
			continue
		}
		
		// Skip if the slot is within the minimum notice period
		if slotDateTime.Before(minAllowedTime) {
			continue
		}
		
		filteredSlots = append(filteredSlots, slot)
	}

	slog.Debug("Time slots summary",
		"totalSlots", len(filteredSlots),
		"weekStart", weekStart.Format("2006-01-02"),
		"weekEnd", weekEnd.Format("2006-01-02"))

	// Log slots per day for detailed debugging
	slotsByDay := make(map[string]int)
	for _, slot := range filteredSlots {
		slotsByDay[slot.Date]++
	}

	for date, count := range slotsByDay {
		slog.Debug("Slots for day", "date", date, "availableSlots", count)
	}

	// Prepare data for the template
	data := CustomerTimeSlotPageData{
		UserID:         uint(userID),
		UserName:       userName,
		Service:        service,
		Slots:          filteredSlots,
		WeekStart:      weekStart,
		WeekEnd:        weekEnd,
		CurrentWeek:    uint(weekOffset),
		MaxAdvanceDate: maxAdvanceDate,
	}

	// Prepare days data
	for i := range 7 {
		currentDate := weekStart.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		daySlots := filterSlotsForDate(filteredSlots, dateStr)
		isHoliday, holiday := calendar.IsFinnishHoliday(currentDate)
		data.Days = append(data.Days, struct {
			Date      time.Time
			DateStr   string
			DaySlots  []TimeSlot
			IsHoliday bool
			Holiday   calendar.FinnishHoliday
		}{
			Date:      currentDate,
			DateStr:   dateStr,
			DaySlots:  daySlots,
			IsHoliday: isHoliday,
			Holiday:   holiday,
		})
	}
	return kit.Render(CustomerTimeSlotSelection(data, kit.Request.Context()))
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
	currentLangCode := translations.M.DefaultLanguage() // Default if not found
	if langVal := kit.Request.Context().Value(translations.ContextKey{}); langVal != nil {
		if code, ok := langVal.(string); ok {
			currentLangCode = code
		}
	}
	availableLangs := translations.M.GetLanguages()
	currentPath := kit.Request.URL.Path

	// Get user's services
	services, err := ListServices(userID, false) // false to get all services, not just active ones
	if err != nil {
		slog.Error("Failed to list services", "userID", userID, "err", err)
		return kit.Render(DashboardPage(DashboardPageData{
			UserID:   userID,
			UserName: userEmail,
			Error:    "Failed to load services",
		}, kit.Request.Context(), availableLangs, currentLangCode, currentPath))
	}

	// Get user's bookings
	bookings, err := ListBookings(userID)
	if err != nil {
		slog.Error("Failed to list bookings", "userID", userID, "err", err)
		return kit.Render(DashboardPage(DashboardPageData{
			UserID:   userID,
			UserName: userEmail,
			Error:    "Failed to load bookings",
		}, kit.Request.Context(), availableLangs, currentLangCode, currentPath))
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
		if booking.Status == "canceled" {
			continue
		}
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
		RecentBookings: bookings[:min(10, len(bookings))], // Show only the 10 most recent bookings
	}

	return kit.Render(DashboardPage(data, kit.Request.Context(), availableLangs, currentLangCode, currentPath))
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
