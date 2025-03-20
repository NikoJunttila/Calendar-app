package calendar

import (
	"fmt"
	"gothstack/plugins/auth"
	"net/http"
	"strconv"
	"time"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
)

// Validation schema for calendar creation
var calendarSchema = v.Schema{
	"name": v.Rules(v.Min(1), v.Max(100)),
}

// CalendarPageData holds data for the calendar page
type CalendarPageData struct {
	FormValues CalendarFormValues
	FormErrors v.Errors
	Calendars  []Calendar
}

// CalendarFormValues holds form data for calendar creation
type CalendarFormValues struct {
	Name           string  `form:"name"`
	Work           bool    `form:"work"`
	Hours          float64 `form:"hours"`
	SuccessMessage string
}

// HandleCalendarList handles the calendar list page
func HandleCalendarList(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	calendars, err := ListCalendars(userID)
	if err != nil {
		return err
	}

	return kit.Render(CalendarList(CalendarPageData{Calendars: calendars}))
}

// HandleCalendarCreate handles the creation form page
func HandleCalendarCreate(kit *kit.Kit) error {
	return kit.Render(CalendarCreate(CalendarPageData{}))
}

// HandleCalendarCreatePost handles the form submission for creating a calendar
func HandleCalendarCreatePost(kit *kit.Kit) error {
	var values CalendarFormValues
	errors, ok := v.Request(kit.Request, &values, calendarSchema)
	if !ok {
		return kit.Render(CalendarForm(values, errors))
	}
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	calendar, err := CreateCalendar(values.Name, values.Work, values.Hours, userID)
	if err != nil {
		return kit.Render(CalendarForm(values, errors))
	}

	values.SuccessMessage = fmt.Sprintf("New calendar '%s' created with ID %d", calendar.Name, calendar.ID)
	return kit.Render(CalendarForm(CalendarFormValues{SuccessMessage: values.SuccessMessage}, errors))
}

// HandleCalendarView handles viewing a specific calendar
func HandleCalendarView(kit *kit.Kit) error {

	yearStr := kit.Request.URL.Query().Get("year")
	monthStr := kit.Request.URL.Query().Get("month")
	if yearStr != "" || monthStr != "" {
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/calendars/%s/month?year=%s?month=%s", chi.URLParam(kit.Request, "id"), yearStr, monthStr))
	}

	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid meal option ID: %w", err)
	}
	resources, err := ListWorkResourcesByCalendar(uint(id))
	if err != nil {
		return err
	}
	total := 0
	for _, resource := range resources {
		total += resource.ResourcesPercentage
	}
	userID := kit.Auth().(auth.Auth).UserID
	calendar, err := GetCalendarWithEntries(uint(id), userID)
	if err != nil {
		return err
	}
	now := time.Now()
	year, month, _ := now.Date()
	//get current year and month
	return kit.Render(CalendarView(calendar, resources, total, year, int(month)))
}

// WorkMonthStats holds statistics about working hours for a month
type WorkMonthStats struct {
	WorkingDays    int                         // Number of working days in the month (Mon-Fri, excluding holidays)
	Holidays       []FinnishHoliday            // Holidays that fall on weekdays in this month
	TotalWorkHours float64                     // Total work hours for the month
	LoggedHours    float64                     // Total hours already logged
	Progress       float64                     // Percentage of completion
	ResourceStats  map[uint]ResourceMonthStats // Stats per resource
}

// ResourceMonthStats holds statistics for a single resource in a month
type ResourceMonthStats struct {
	ResourceID   uint
	ResourceName string
	Percentage   int
	TargetHours  float64 // Target hours for this resource
	LoggedHours  float64 // Hours logged for this resource
	Progress     float64 // Percentage of completion
}

// HandleCalendarViewByMonth renders the calendar view with entries filtered by month and year
func HandleCalendarViewByMonth(kit *kit.Kit) error {

	// Get the calendar ID from the URL parameter
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}

	// Get the year and month from URL parameters
	yearStr := kit.Request.URL.Query().Get("year")
	monthStr := kit.Request.URL.Query().Get("month")

	// Default to current year and month if not provided
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	// Parse year if provided
	if yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err == nil {
			currentYear = year
		}
	}

	// Parse month if provided
	if monthStr != "" {
		month, err := strconv.Atoi(monthStr)
		if err == nil && month >= 1 && month <= 12 {
			currentMonth = month
		}
	}

	userID := kit.Auth().(auth.Auth).UserID
	// Get the calendar with entries filtered by month and year
	calendar, err := GetCalendarWithEntriesByMonth(uint(calendarID), userID, currentYear, currentMonth)
	if err != nil {
		return err
	}
	// Get resources for this calendar
	resources, err := ListWorkResourcesByCalendar(uint(calendarID))
	if err != nil {
		return err
	}

	// Calculate total resource allocation
	totalResource := 0
	for _, resource := range resources {
		totalResource += resource.ResourcesPercentage
	}

	// Calculate work statistics for the month
	workStats := calculateWorkStats(calendar, resources, currentYear, currentMonth)

	// Render the view
	return kit.Render(CalendarViewMonthly(calendar, resources, totalResource, currentYear, currentMonth, workStats))
}

// calculateWorkStats calculates work statistics for a given month
func calculateWorkStats(calendar Calendar, resources []WorkResource, year, month int) WorkMonthStats {
	// Initialize work stats
	stats := WorkMonthStats{
		ResourceStats: make(map[uint]ResourceMonthStats),
		Holidays:      []FinnishHoliday{},
	}

	// Calculate working days (Monday-Friday, excluding holidays) in the month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Count working days
	workingDays := 0
	for d := firstDay; d.Before(lastDay.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
		weekday := d.Weekday()

		// Check if it's a weekday (Monday-Friday)
		if weekday != time.Saturday && weekday != time.Sunday {
			// Check if it's a holiday
			isHoliday, holiday := IsFinnishHoliday(d)

			if isHoliday {
				// Add to our list of holidays
				stats.Holidays = append(stats.Holidays, holiday)
			} else {
				// If not a holiday, count as working day
				workingDays++
			}
		}
	}

	stats.WorkingDays = workingDays
	stats.TotalWorkHours = float64(workingDays) * calendar.DailyWorkHours

	// Initialize resource stats
	for _, resource := range resources {
		resourceTarget := stats.TotalWorkHours * float64(resource.ResourcesPercentage) / 100
		stats.ResourceStats[resource.ID] = ResourceMonthStats{
			ResourceID:   resource.ID,
			ResourceName: resource.Name,
			Percentage:   resource.ResourcesPercentage,
			TargetHours:  resourceTarget,
			LoggedHours:  0,
			Progress:     0,
		}
	}

	// Calculate logged hours
	totalLogged := 0.0
	for _, entry := range calendar.Entries {
		totalLogged += entry.Hours

		// Add to resource stats if this entry has a resource
		if entry.WorkResourceID > 0 {
			if resourceStats, exists := stats.ResourceStats[entry.WorkResourceID]; exists {
				resourceStats.LoggedHours += entry.Hours
				resourceStats.Progress = (resourceStats.LoggedHours / resourceStats.TargetHours) * 100
				stats.ResourceStats[entry.WorkResourceID] = resourceStats
			}
		}
	}

	stats.LoggedHours = totalLogged
	if stats.TotalWorkHours > 0 {
		stats.Progress = (stats.LoggedHours / stats.TotalWorkHours) * 100
	}

	return stats
}

func CreateDateParam(year, month, day int) string {
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	return date.Format("2006-01-02")
}
