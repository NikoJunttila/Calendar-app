package calendar

import (
	"fmt"
	"gothstack/plugins/auth"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
)

// Validation schema for calendar entry creation and editing
var calendarEntrySchema = v.Schema{
	"date": v.Rules(v.Min(1)), // ensure a non-empty date string
	"text": v.Rules(v.Min(1)), // ensure a non-empty text string
}

// CalendarEntryPageData holds data for the calendar entry pages
type CalendarEntryPageData struct {
	Calendar      Calendar
	WorkResources []WorkResource
	FormValues    CalendarEntryFormValues
	FormErrors    v.Errors
	EntryID       uint
}

// CalendarEntryFormValues holds form data for calendar entries
type CalendarEntryFormValues struct {
	Date           string  `form:"date"` // expected in "2006-01-02" format
	Hours          float64 `form:"hours"`
	Text           string  `form:"text"`
	WorkResourceID uint    `form:"resource"`
	SuccessMessage string
}

// HandleCalendarEntryCreate renders the entry creation form (GET request)
func HandleCalendarEntryCreate(kit *kit.Kit) error {
	// Get the calendar ID from the URL parameter
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}

	// Retrieve the calendar details
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	calendar, err := GetCalendar(uint(calendarID), userID)
	if err != nil {
		return err
	}

	resources, err := ListWorkResourcesByCalendar(uint(calendarID))
	if err != nil {
		return err
	}

	dateStr := kit.Request.URL.Query().Get("date")
	var selectedDate time.Time
	fmt.Println(dateStr)
	if dateStr != "" {
		// Parse the date string (format: "2006-01-02")
		selectedDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			// If invalid date, use current date
			selectedDate = time.Now()
		}
	} else {
		// No date provided, use current date
		selectedDate = time.Now()
	}
	formValues := CalendarEntryFormValues{
		Date:  selectedDate.Format("2006-01-02"), // Current day
		Hours: calendar.DailyWorkHours,           // Default hours from calendar
	}

	// Render the entry creation form
	data := CalendarEntryPageData{
		Calendar:      calendar,
		WorkResources: resources,
		FormValues:    formValues,
	}
	return kit.Render(CalendarEntryCreate(data))
}

// HandleCalendarEntryCreatePost processes the form submission (POST request) for creating a calendar entry
func HandleCalendarEntryCreatePost(kit *kit.Kit) error {
	// Get the calendar ID from the URL parameter
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}

	// Parse and validate the form values
	var values CalendarEntryFormValues
	errors, ok := v.Request(kit.Request, &values, calendarEntrySchema)

	// Retrieve the calendar details for re-rendering the form if needed
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	calendar, err := GetCalendar(uint(calendarID), userID)
	if err != nil {
		slog.Error("Failed to get calendar", "error", err)
	}
	resources, err := ListWorkResourcesByCalendar(uint(calendarID))
	if err != nil {
		slog.Error("Failed to list work resources", "error", err)
	}

	if !ok {
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources, 0))
	}

	// Convert the date string to a time.Time value
	entryDate, err := time.Parse("2006-01-02", values.Date)
	if err != nil {
		errors.Add("date", "Invalid date format. Please use YYYY-MM-DD.")
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources, 0))
	}

	// Create the new calendar entry
	entry, err := CreateCalendarEntry(uint(calendarID), entryDate, values.Text, values.Hours, values.WorkResourceID)
	if err != nil {
		errors.Add("general", "Failed to create calendar entry.")
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources, 0))
	}

	// Set a success message and re-render the form
	values.SuccessMessage = fmt.Sprintf("New entry created on %s with ID %d", entryDate.Format("2006-01-02"), entry.ID)
	return kit.Render(CalendarEntryForm(CalendarEntryFormValues{SuccessMessage: values.SuccessMessage}, errors, calendar, resources, 0))
}

// HandleCalendarEntryEdit renders the entry edit form (GET request)
func HandleCalendarEntryEdit(kit *kit.Kit) error {
	// Get the entry ID from the URL parameter
	entryIDStr := chi.URLParam(kit.Request, "entry_id")
	entryID, err := strconv.ParseUint(entryIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid entry ID: %w", err)
	}

	// Retrieve the calendar entry details
	entry, err := GetCalendarEntry(uint(entryID))
	if err != nil {
		return err
	}

	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID

	// Retrieve the calendar details
	calendar, err := GetCalendar(entry.CalendarID, userID)
	if err != nil {
		return err
	}

	// Get work resources for the calendar
	resources, err := ListWorkResourcesByCalendar(calendar.ID)
	if err != nil {
		return err
	}

	// Populate form values from the existing entry
	values := CalendarEntryFormValues{
		Date:           entry.Date.Format("2006-01-02"),
		Text:           entry.Text,
		Hours:          entry.Hours,
		WorkResourceID: entry.WorkResourceID,
	}

	// Render the calendar entry edit form
	data := CalendarEntryPageData{
		Calendar:      calendar,
		WorkResources: resources,
		FormValues:    values,
		EntryID:       uint(entryID),
	}
	return kit.Render(CalendarEntryEdit(data))
}

// HandleCalendarEntryEditPost processes the form submission (POST request) for updating a calendar entry
func HandleCalendarEntryEditPost(kit *kit.Kit) error {
	// Get the entry ID from the URL parameter
	entryIDStr := chi.URLParam(kit.Request, "entry_id")
	entryID, err := strconv.ParseUint(entryIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid entry ID: %w", err)
	}

	// Retrieve the calendar entry details
	entry, err := GetCalendarEntry(uint(entryID))
	if err != nil {
		return err
	}

	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID

	// Retrieve the calendar details
	calendar, err := GetCalendar(entry.CalendarID, userID)
	if err != nil {
		return err
	}

	// Parse and validate the form values
	var values CalendarEntryFormValues
	errors, ok := v.Request(kit.Request, &values, calendarEntrySchema)

	// Get work resources for the calendar
	resources, err := ListWorkResourcesByCalendar(calendar.ID)
	if err != nil {
		slog.Error("Failed to list work resources", "error", err)
	}

	if !ok {
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources, uint(entryID)))
	}

	// Parse the date
	entryDate, err := time.Parse("2006-01-02", values.Date)
	if err != nil {
		errors.Add("date", "Invalid date format. Please use YYYY-MM-DD.")
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources, uint(entryID)))
	}

	// Update the calendar entry
	// year, month, week := getDateComponents(entryDate)
	updatedEntry, err := UpdateCalendarEntry(uint(entryID), entryDate, values.Text, values.Hours, values.WorkResourceID)
	if err != nil {
		errors.Add("general", "Failed to update calendar entry.")
		return kit.Render(CalendarEntryForm(values, errors, calendar, resources, uint(entryID)))
	}

	// Set a success message
	values.SuccessMessage = fmt.Sprintf("Entry updated successfully on %s", updatedEntry.Date.Format("2006-01-02"))
	return kit.Render(CalendarEntryForm(values, errors, calendar, resources, uint(entryID)))
}

// HandleCalendarEntryDelete processes the request to delete a calendar entry
func HandleCalendarEntryDelete(kit *kit.Kit) error {
	// Get the entry ID from the URL parameter
	entryIDStr := chi.URLParam(kit.Request, "entry_id")
	fmt.Println("entry id: ", entryIDStr)
	entryID, err := strconv.ParseUint(entryIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid entry ID: %w", err)
	}

	// Retrieve the calendar entry details to get the calendar ID
	entry, err := GetCalendarEntry(uint(entryID))
	if err != nil {
		return err
	}
	calendarID := entry.CalendarID

	// Delete the calendar entry
	err = DeleteCalendarEntry(uint(entryID))
	if err != nil {
		return err
	}

	// Redirect to the calendar entries list page
	return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/calendars/%d", calendarID))
}
