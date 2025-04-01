package reservations

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

// BookingConfirmationPageData holds data for the booking confirmation page
type BookingConfirmationPageData struct {
	UserID   uint
	UserName string
	Service  Service
	TimeSlot TimeSlot
	Error    string
	Success  string
}

// HandleBookingConfirmationView displays the booking confirmation form
func HandleBookingConfirmationView(kit *kit.Kit) error {
	// Get user ID and service ID from URL
	userIDStr := chi.URLParam(kit.Request, "userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse user ID", "id", userIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	serviceIDStr := chi.URLParam(kit.Request, "serviceID")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse service ID", "id", serviceIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	// Get date and time from query parameters
	date := kit.Request.URL.Query().Get("date")
	time := kit.Request.URL.Query().Get("time")
	if date == "" || time == "" {
		slog.Error("Missing date or time parameters")
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d/service/%d", userID, serviceID))
	}

	// Fetch service details
	service, err := GetService(uint(serviceID), uint(userID))
	if err != nil {
		slog.Error("Failed to fetch service", "serviceID", serviceID, "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	// Create a TimeSlot struct with the selected date and time
	timeSlot := TimeSlot{
		UserID:    uint(userID),
		ServiceID: sql.NullInt64{Int64: int64(serviceID), Valid: true},
		Date:      date,
		Time:      time,
		Duration:  service.Duration,
		IsBooked:  false,
	}

	data := BookingConfirmationPageData{
		UserID:   uint(userID),
		UserName: "TODO figureout this",
		Service:  service,
		TimeSlot: timeSlot,
	}

	return kit.Render(BookingConfirmation(data))
}

// HandleBookingConfirmationPost processes the booking confirmation form submission
func HandleBookingConfirmationPost(kit *kit.Kit) error {
	// Get user ID and service ID from URL
	userIDStr := chi.URLParam(kit.Request, "userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse user ID", "id", userIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	serviceIDStr := chi.URLParam(kit.Request, "serviceID")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse service ID", "id", serviceIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	// Get form data
	date := kit.Request.FormValue("date")
	timeStr := kit.Request.FormValue("time")
	clientName := kit.Request.FormValue("client_name")
	clientEmail := kit.Request.FormValue("client_email")
	clientPhone := kit.Request.FormValue("client_phone")
	notes := kit.Request.FormValue("notes")

	// Validate required fields
	if date == "" || timeStr == "" || clientName == "" || clientEmail == "" {
		slog.Error("Missing required fields", "date", date, "time", timeStr, "clientName", clientName, "clientEmail", clientEmail)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d/service/%d/confirm?date=%s&time=%s", userID, serviceID, date, timeStr))
	}

	// Fetch service to get duration
	service, err := GetService(uint(serviceID), uint(userID))
	if err != nil {
		slog.Error("Failed to fetch service", "serviceID", serviceID, "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	// Create the time slot first
	timeSlot, err := CreateTimeSlot(
		uint(userID),
		sql.NullInt64{Int64: int64(serviceID), Valid: true},
		date,
		timeStr,
		service.Duration,
	)
	if err != nil {
		slog.Error("Failed to create time slot", "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d/service/%d/confirm?date=%s&time=%s", userID, serviceID, date, timeStr))
	}

	// Create the booking
	bookingRef := fmt.Sprintf("BK-%d-%d", timeSlot.ID, time.Now().Unix())
	_, err = CreateBooking(
		uint(userID),
		timeSlot.ID,
		sql.NullInt64{Int64: int64(serviceID), Valid: true},
		clientName,
		clientEmail,
		clientPhone,
		bookingRef,
		notes,
	)
	if err != nil {
		slog.Error("Failed to create booking",
			"userID", userID,
			"serviceID", serviceID,
			"timeSlotID", timeSlot.ID,
			"err", err)
		// If booking fails, we should probably delete the time slot we just created
		if deleteErr := DeleteTimeSlot(timeSlot.ID); deleteErr != nil {
			slog.Error("Failed to delete time slot after booking failure", "timeSlotID", timeSlot.ID, "err", deleteErr)
		}
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d/service/%d/confirm?date=%s&time=%s", userID, serviceID, date, timeStr))
	}

	// Success - redirect to a success page or show success message
	return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d/service/%d/success", userID, serviceID))
}

// HandleBookingSuccess displays the booking success page
func HandleBookingSuccess(kit *kit.Kit) error {
	// Get user ID and service ID from URL
	userIDStr := chi.URLParam(kit.Request, "userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse user ID", "id", userIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	serviceIDStr := chi.URLParam(kit.Request, "serviceID")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse service ID", "id", serviceIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	// Fetch service details
	service, err := GetService(uint(serviceID), uint(userID))
	if err != nil {
		slog.Error("Failed to fetch service", "serviceID", serviceID, "err", err)
		return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/book/%d", userID))
	}

	data := BookingConfirmationPageData{
		UserID:   uint(userID),
		UserName: "TODO figureout this",
		Service:  service,
		Success:  "Your booking has been confirmed! You will receive a confirmation email shortly.",
	}

	return kit.Render(BookingSuccess(data))
}
