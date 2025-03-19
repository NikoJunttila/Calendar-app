package calendar

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
)

// Validation schema for calendar entry creation.
var calendarEntrySchema = v.Schema{
	"date": v.Rules(v.Min(1)), // ensure a non-empty date string
	"text": v.Rules(v.Min(1)), // ensure a non-empty text string
}

// CalendarEntryPageData holds data for the calendar entry creation page.
type CalendarEntryPageData struct {
	Calendar      Calendar
	WorkResources []WorkResource
	FormValues    CalendarEntryFormValues
	FormErrors    v.Errors
}

// CalendarEntryFormValues holds form data for creating a calendar entry.
type CalendarEntryFormValues struct {
	Date           string  `form:"date"` // expected in "2006-01-02" format
	Hours          float64 `form:"hours"`
	Text           string  `form:"text"`
	WorkResourceID uint    `form:"resource"`
	SuccessMessage string
}

// HandleCalendarEntryCreate renders the entry creation form (GET request).
func HandleCalendarEntryCreate(kit *kit.Kit) error {
	// Get the calendar ID from the URL parameter.
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}
	// Retrieve the calendar details.
	calendar, err := GetCalendar(uint(calendarID))
	if err != nil {
		return err
	}
	// calendar.DailyWorkHours
	resources, err := ListWorkResourcesByCalendar(uint(calendarID))
	if err != nil {
		return err
	}
	formValues := CalendarEntryFormValues{
		Date:  time.Now().Format("2006-01-02"), // Current day
		Hours: calendar.DailyWorkHours,         // Default hours from calendar
	}
	// Render the entry creation form.
	data := CalendarEntryPageData{
		Calendar:      calendar,
		WorkResources: resources,
		FormValues:    formValues,
	}
	return kit.Render(CalendarEntryCreate(data))
}

// HandleCalendarEntryCreatePost processes the form submission (POST request) for creating a calendar entry.
func HandleCalendarEntryCreatePost(kit *kit.Kit) error {
	// Get the calendar ID from the URL parameter.
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}

	// Parse and validate the form values.
	var values CalendarEntryFormValues
	errors, ok := v.Request(kit.Request, &values, calendarEntrySchema)
	// Retrieve the calendar details for re-rendering the form if needed.
	calendar, err := GetCalendar(uint(calendarID))
	if err != nil {
		slog.Error("Failed to calendar", "error", err)
	}
	resources, err := ListWorkResourcesByCalendar(uint(calendarID))
	if err != nil {
		slog.Error("Failed to list work resources", "error", err)
	}
	if !ok {
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources))
	}

	// Convert the date string to a time.Time value.
	entryDate, err := time.Parse("2006-01-02", values.Date)
	if err != nil {
		errors.Add("date", "Invalid date format. Please use YYYY-MM-DD.")
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources))
	}
	fmt.Println(values)
	// Create the new calendar entry.
	entry, err := CreateCalendarEntry(uint(calendarID), entryDate, values.Text, values.Hours, values.WorkResourceID)
	if err != nil {
		errors.Add("general", "Failed to create calendar entry.")
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources))
	}

	// Set a success message and re-render the form.
	values.SuccessMessage = fmt.Sprintf("New entry created on %s with ID %d", entryDate.Format("2006-01-02"), entry.ID)
	return kit.Render(CalendarEntryForm(CalendarEntryFormValues{SuccessMessage: values.SuccessMessage}, errors, calendar, resources))
}
