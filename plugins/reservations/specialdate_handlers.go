package reservations

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"gothstack/plugins/auth"

	"github.com/anthdm/superkit/kit"
)

// HandleSpecialDateView displays the special dates settings form
func HandleSpecialDateView(kit *kit.Kit) error {
	// Get the authenticated user's ID
	userID := kit.Auth().(auth.Auth).UserID

	// Fetch existing special dates
	specialDates, err := GetSpecialDates(userID)
	if err != nil {
		slog.Error("Failed to fetch special dates", "userID", userID, "err", err)
		data := SpecialDatePageData{
			UserID:   userID,
			UserName: kit.Auth().(auth.Auth).Email,
			Error:    "Failed to load special dates settings.",
		}
		return kit.Render(SpecialDateSettings(data))
	}

	data := SpecialDatePageData{
		UserID:       userID,
		UserName:     kit.Auth().(auth.Auth).Email,
		SpecialDates: specialDates,
	}

	return kit.Render(SpecialDateSettings(data))
}

// HandleSpecialDateCreate creates a new special date
func HandleSpecialDateCreate(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID

	// Parse form data
	dateStr := kit.Request.FormValue("date")
	description := kit.Request.FormValue("description")
	isWorkingDay := kit.Request.FormValue("is_working_day") == "on"

	// Parse the date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		slog.Error("Failed to parse date", "date", dateStr, "err", err)
		data := SpecialDatePageData{
			UserID:   userID,
			UserName: kit.Auth().(auth.Auth).Email,
			Error:    "Invalid date format.",
		}
		return kit.Render(SpecialDateSettings(data))
	}

	// Create the special date
	_, err = CreateSpecialDate(userID, date.Format("2006-01-02"), description, isWorkingDay)
	if err != nil {
		slog.Error("Failed to create special date",
			"userID", userID,
			"date", date,
			"err", err)
		// Fetch current special dates for the form
		specialDates, _ := GetSpecialDates(userID)
		data := SpecialDatePageData{
			UserID:       userID,
			UserName:     kit.Auth().(auth.Auth).Email,
			SpecialDates: specialDates,
			Error:        "Failed to create special date.",
		}
		return kit.Render(SpecialDateSettings(data))
	}

	// Success - redirect back to the form with success message
	specialDates, _ := GetSpecialDates(userID)
	data := SpecialDatePageData{
		UserID:       userID,
		UserName:     kit.Auth().(auth.Auth).Email,
		SpecialDates: specialDates,
		Success:      "Special date added successfully.",
	}
	return kit.Render(SpecialDateSettings(data))
}

// HandleSpecialDateDelete deletes a special date
func HandleSpecialDateDelete(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID

	// Get special date ID from URL
	specialDateIDStr := kit.Request.URL.Path[len("/reservations/special-dates/"):]
	specialDateID, err := strconv.ParseUint(specialDateIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse special date ID", "id", specialDateIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, "/reservations/special-dates")
	}

	// Delete the special date
	err = DeleteSpecialDate(uint(specialDateID), fmt.Sprintf("%d", userID))
	if err != nil {
		slog.Error("Failed to delete special date",
			"userID", userID,
			"specialDateID", specialDateID,
			"err", err)
		return kit.Redirect(http.StatusSeeOther, "/reservations/special-dates")
	}

	return kit.Redirect(http.StatusSeeOther, "/reservations/special-dates")
}
