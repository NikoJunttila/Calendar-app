package reservations

import (
	"fmt"
	"gothstack/plugins/auth"
	"log/slog"
)

func sendBookingConfirmationEmail(booking Booking, timeSlot TimeSlot, service Service) error {
	subject := fmt.Sprintf("Ajanvaraus vahvistus - %s", service.Name)
	to := booking.ClientEmail

	plainTextContent, htmlContent := generateBookingConfirmationEmail(booking, timeSlot, service)

	err := sendMail(subject, to, plainTextContent, htmlContent)
	if err != nil {
		slog.Error("Failed to send confirmation email", "booking_ref", booking.BookingRef, "err", err)
		return err
	}

	slog.Info("Confirmation email sent successfully", "booking_ref", booking.BookingRef)
	return nil
}

// sendReminderEmail sends a reminder email for an upcoming booking
func sendOwnerBookingNotificationEmail(booking Booking, timeSlot TimeSlot, service Service, owner auth.User) error {
	subject := fmt.Sprintf("Uusi ajanvaraus - %s", service.Name)
	to := owner.Email

	plainTextContent, htmlContent := generateOwnerBookingNotificationEmail(booking, timeSlot, service, owner)

	err := sendMail(subject, to, plainTextContent, htmlContent)
	if err != nil {
		slog.Error("Failed to send owner notification email", "booking_ref", booking.BookingRef, "err", err)
		return err
	}

	slog.Info("Owner notification email sent successfully", "booking_ref", booking.BookingRef, "owner_email", owner.Email)
	return nil
}

// sendReminderEmail sends a reminder email for an upcoming booking
func sendReminderEmail(booking Booking, timeSlot TimeSlot, service Service) error {
	subject := fmt.Sprintf("Muistutus huomisesta ajanvarauksesta - %s", service.Name)
	to := booking.ClientEmail
	plainTextContent, htmlContent := generateReminderEmail(booking, timeSlot, service)
	err := sendMail(subject,to,plainTextContent,htmlContent)

	if err != nil {
		slog.Error("Failed to send reminder email", "booking_ref", booking.BookingRef, "err", err)
		return err
	}
	slog.Info("Reminder email sent successfully",
		"booking_ref", booking.BookingRef)

	return nil
}
