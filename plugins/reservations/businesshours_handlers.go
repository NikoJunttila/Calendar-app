package reservations

import (
	"fmt"
	"log/slog"

	"gothstack/plugins/auth"

	"github.com/anthdm/superkit/kit"
)

// HandleBusinessHoursView displays the business hours settings form
func HandleBusinessHoursView(kit *kit.Kit) error {
	// Get the authenticated user's ID
	userID := kit.Auth().(auth.Auth).UserID

	// Fetch existing business hours
	businessHours, err := GetBusinessHours(userID)
	if err != nil {
		slog.Error("Failed to fetch business hours", "userID", userID, "err", err)
		data := BusinessHoursPageData{
			UserID:   userID,
			UserName: kit.Auth().(auth.Auth).Email,
			Error:    "Failed to load business hours settings.",
		}
		return kit.Render(BusinessHoursSettings(data))
	}

	data := BusinessHoursPageData{
		UserID:        userID,
		UserName:      kit.Auth().(auth.Auth).Email,
		BusinessHours: businessHours,
	}

	return kit.Render(BusinessHoursSettings(data))
}

func HandleBusinessHoursUpdate(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	// Process each day's settings
	for day := 0; day < 7; day++ {
		// Get form values
		isWorking := kit.Request.FormValue(fmt.Sprintf("is_working_%d", day)) == "on"
		startTime := kit.Request.FormValue(fmt.Sprintf("start_time_%d", day))
		endTime := kit.Request.FormValue(fmt.Sprintf("end_time_%d", day))

		// Update or create business hour record
		_, err := CreateBusinessHour(userID, day, startTime, endTime, isWorking)
		if err != nil {
			slog.Error("Failed to update business hours",
				"userID", userID,
				"day", day,
				"err", err)
			// Fetch current business hours for the form
			businessHours, _ := GetBusinessHours(userID)
			data := BusinessHoursPageData{
				UserID:        userID,
				UserName:      kit.Auth().(auth.Auth).Email,
				BusinessHours: businessHours,
				Error:         "Failed to update business hours settings.",
			}
			return kit.Render(BusinessHoursSettings(data))
		}
	}

	// Success - redirect back to the form with success message
	businessHours, _ := GetBusinessHours(userID)
	data := BusinessHoursPageData{
		UserID:        userID,
		UserName:      kit.Auth().(auth.Auth).Email,
		BusinessHours: businessHours,
		Success:       "Business hours updated successfully.",
	}
	return kit.Render(BusinessHoursSettings(data))
}
