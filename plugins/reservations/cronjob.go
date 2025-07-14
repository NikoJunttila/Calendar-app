package reservations

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

func SetupCron() {
	fmt.Println("set up cron")
	c := cron.New()
	c.AddFunc("* * * * *", func() {
		timeslots, err := ListTimeSlotsBetweenNowAndTomorrowEnd()
		if err != nil {
			slog.Error("error in cron", "cron", err)
		}
		for _, slot := range timeslots{
			booking, err := GetBookingByTimeslot(slot.ID) 
			if err != nil {
			slog.Error("error fetching booking in cron", "err", err)
				continue
			}
			err = sendReminderEmail(booking,slot,booking.Service)
			if err != nil {
				slog.Error("error sending reminder cron", "err", err)
				continue
			}

			err = UpdateReminderSent(slot.ID, time.Now())
			if err != nil {
				slog.Error("error updating reminder", "err", err)
				continue
			}
			fmt.Println("sent reminder for: ",slot.ID)
		}
	})
	c.Start()
}
