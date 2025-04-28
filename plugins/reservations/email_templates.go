package reservations

import (
	"fmt"
	"log/slog"
	"os"
	"gothstack/plugins/auth"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)
// generateBookingConfirmationEmail creates the content for booking confirmation emails
func generateBookingConfirmationEmail(booking Booking, timeSlot TimeSlot, service Service) (string, string) {
	// Generate plain text version
	plainTextContent := fmt.Sprintf(`
Booking Confirmation - %s

Dear %s,

Thank you for your booking! Your appointment has been confirmed.

BOOKING DETAILS:
Service: %s
Date: %s
Time: %s
Duration: %d minutes
Booking Reference: %s

CLIENT INFORMATION:
Name: %s
Email: %s
`, 
	service.Name,
	booking.ClientName,
	service.Name,
	timeSlot.Date,
	timeSlot.Time,
	timeSlot.Duration,
	booking.BookingRef,
	booking.ClientName,
	booking.ClientEmail)

	// Add phone if available
	if booking.ClientPhone != "" {
		plainTextContent += fmt.Sprintf("Phone: %s\n", booking.ClientPhone)
	}

	// Add notes if available
	if booking.Notes != "" {
		plainTextContent += fmt.Sprintf("\nNOTES:\n%s\n", booking.Notes)
	}

	plainTextContent += `
If you need to reschedule or cancel your appointment, please contact us with your booking reference number.

Thank you for choosing our service!

Best regards,
The Team
`

	// Generate HTML version
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
        }
        .header {
            background-color: #4299e1;
            padding: 20px;
            text-align: center;
            color: white;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
        }
        .booking-details, .client-info {
            background-color: white;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .notes {
            background-color: #f0f4f8;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 5px;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 14px;
            color: #666;
        }
        h2 {
            color: #2b6cb0;
            margin-top: 0;
        }
        table {
            width: 100%%;
        }
        td {
            padding: 8px 0;
        }
        .label {
            font-weight: bold;
            width: 40%%;
        }
        .checkmark {
            color: #48bb78;
            font-size: 24px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Booking Confirmation</h1>
        <div class="checkmark">✓</div>
    </div>
    
    <div class="content">
        <p>Dear <strong>%s</strong>,</p>
        
        <p>Thank you for your booking! Your appointment has been confirmed.</p>
        
        <div class="booking-details">
            <h2>Booking Details</h2>
            <table>
                <tr>
                    <td class="label">Service:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Date:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Time:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Duration:</td>
                    <td>%d minutes</td>
                </tr>
                <tr>
                    <td class="label">Booking Reference:</td>
                    <td>%s</td>
                </tr>
            </table>
        </div>
        
        <div class="client-info">
            <h2>Client Information</h2>
            <table>
                <tr>
                    <td class="label">Name:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Email:</td>
                    <td>%s</td>
                </tr>`,
        booking.ClientName,
        service.Name,
        timeSlot.Date,
        timeSlot.Time,
        timeSlot.Duration,
        booking.BookingRef,
        booking.ClientName,
        booking.ClientEmail)

    // Add phone if available
    if booking.ClientPhone != "" {
        htmlContent += fmt.Sprintf(`
                <tr>
                    <td class="label">Phone:</td>
                    <td>%s</td>
                </tr>`, booking.ClientPhone)
    }

    htmlContent += `
            </table>
        </div>`

    // Add notes if available
    if booking.Notes != "" {
        htmlContent += fmt.Sprintf(`
        <div class="notes">
            <h2>Notes</h2>
            <p>%s</p>
        </div>`, booking.Notes)
    }

    htmlContent += `
        <p>If you need to reschedule or cancel your appointment, please contact us with your booking reference number.</p>
        
        <p>Thank you for choosing our service!</p>
        
        <p>Best regards,<br>The Team</p>
    </div>
    
    <div class="footer">
        <p>&copy; 2025 Our Service. All rights reserved.</p>
    </div>
</body>
</html>
`

	return plainTextContent, htmlContent
}

func sendBookingConfirmationEmail(booking Booking, timeSlot TimeSlot, service Service, email string) error {
    from := mail.NewEmail("Hyvinvointikeskusluxus", email)
    subject := fmt.Sprintf("Booking Confirmation - %s", service.Name)
    to := mail.NewEmail(booking.ClientName, booking.ClientEmail)
    
    plainTextContent, htmlContent := generateBookingConfirmationEmail(booking, timeSlot, service)
    message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
    client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
    
    response, err := client.Send(message)
		fmt.Println(response)
    if err != nil {
        slog.Error("Failed to send confirmation email", "booking_ref", booking.BookingRef, "err", err)
        return err
    }
    
    slog.Info("Confirmation email sent successfully", 
        "booking_ref", booking.BookingRef, 
        "status_code", response.StatusCode)
    
    return nil
}
func generateOwnerBookingNotificationEmail(booking Booking, timeSlot TimeSlot, service Service, owner auth.User) (string, string) {
	// Generate plain text version
	plainTextContent := fmt.Sprintf(`
New Booking Alert - %s

Hello %s,

You have received a new booking for your service.

BOOKING DETAILS:
Service: %s
Date: %s
Time: %s
Duration: %d minutes
Booking Reference: %s

CLIENT INFORMATION:
Name: %s
Email: %s
`, 
	service.Name,
	owner.FirstName,
	service.Name,
	timeSlot.Date,
	timeSlot.Time,
	timeSlot.Duration,
	booking.BookingRef,
	booking.ClientName,
	booking.ClientEmail)

	// Add phone if available
	if booking.ClientPhone != "" {
		plainTextContent += fmt.Sprintf("Phone: %s\n", booking.ClientPhone)
	}

	// Add notes if available
	if booking.Notes != "" {
		plainTextContent += fmt.Sprintf("\nCLIENT NOTES:\n%s\n", booking.Notes)
	}

	plainTextContent += fmt.Sprintf(`
You can view the complete booking details in your dashboard:
%s/dashboard/bookings/%s

Thank you for using our booking system.

Best regards,
The Team
`, os.Getenv("APP_URL"), booking.BookingRef)

	// Generate HTML version
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
        }
        .header {
            background-color: #3182ce;
            padding: 20px;
            text-align: center;
            color: white;
        }
        .notification-icon {
            font-size: 24px;
            margin-bottom: 10px;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
        }
        .booking-details, .client-info {
            background-color: white;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .notes {
            background-color: #f0f4f8;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 5px;
        }
        .action-button {
            display: inline-block;
            background-color: #3182ce;
            color: white;
            text-decoration: none;
            padding: 12px 20px;
            border-radius: 5px;
            font-weight: bold;
            margin: 15px 0;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 14px;
            color: #666;
        }
        h2 {
            color: #2b6cb0;
            margin-top: 0;
        }
        table {
            width: 100%%;
        }
        td {
            padding: 8px 0;
        }
        .label {
            font-weight: bold;
            width: 40%%;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="notification-icon">🔔</div>
        <h1>New Booking Alert</h1>
    </div>
    
    <div class="content">
        <p>Hello <strong>%s</strong>,</p>
        
        <p>You have received a new booking for your service.</p>
        
        <div class="booking-details">
            <h2>Booking Details</h2>
            <table>
                <tr>
                    <td class="label">Service:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Date:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Time:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Duration:</td>
                    <td>%d minutes</td>
                </tr>
                <tr>
                    <td class="label">Booking Reference:</td>
                    <td>%s</td>
                </tr>
            </table>
        </div>
        
        <div class="client-info">
            <h2>Client Information</h2>
            <table>
                <tr>
                    <td class="label">Name:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Email:</td>
                    <td><a href="mailto:%s">%s</a></td>
                </tr>`,
        owner.FirstName,
        service.Name,
        timeSlot.Date,
        timeSlot.Time,
        timeSlot.Duration,
        booking.BookingRef,
        booking.ClientName,
        booking.ClientEmail,
        booking.ClientEmail)

    // Add phone if available
    if booking.ClientPhone != "" {
        htmlContent += fmt.Sprintf(`
                <tr>
                    <td class="label">Phone:</td>
                    <td>%s</td>
                </tr>`, booking.ClientPhone)
    }

    htmlContent += `
            </table>
        </div>`

    // Add notes if available
    if booking.Notes != "" {
        htmlContent += fmt.Sprintf(`
        <div class="notes">
            <h2>Client Notes</h2>
            <p>%s</p>
        </div>`, booking.Notes)
    }

    // Add action button to view booking details
    htmlContent += fmt.Sprintf(`
        <p>You can view the complete booking details in your dashboard:</p>
        <a href="%s/dashboard/bookings/%s" class="action-button">View Booking Details</a>
        
        <p>Thank you for using our booking system.</p>
        
        <p>Best regards,<br>The Team</p>
    </div>
    
    <div class="footer">
        <p>&copy; 2025 Our Service. All rights reserved.</p>
    </div>
</body>
</html>
`, os.Getenv("APP_URL"), booking.BookingRef)

	return plainTextContent, htmlContent
}

// Function to send notification email to service owner
func sendOwnerBookingNotificationEmail(booking Booking, timeSlot TimeSlot, service Service, owner auth.User, email string) error {
    from := mail.NewEmail("Booking System", email)
    subject := fmt.Sprintf("New Booking Alert - %s", service.Name)
    to := mail.NewEmail(owner.FirstName, owner.Email)
    
    plainTextContent, htmlContent := generateOwnerBookingNotificationEmail(booking, timeSlot, service, owner)
    
    message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
    client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
    
    response, err := client.Send(message)
		fmt.Println(response)
    if err != nil {
        slog.Error("Failed to send owner notification email", "booking_ref", booking.BookingRef, "err", err)
        return err
    }
    
    slog.Info("Owner notification email sent successfully", 
        "booking_ref", booking.BookingRef, 
        "owner_email", owner.Email,
        "status_code", response.StatusCode)
    
    return nil
}

// generateReminderEmail creates the content for booking reminder emails
func generateReminderEmail(booking Booking, timeSlot TimeSlot, service Service) (string, string) {
	// Generate plain text version
	plainTextContent := fmt.Sprintf(`
Booking Reminder - %s

Dear %s,

This is a friendly reminder about your upcoming appointment tomorrow.

BOOKING DETAILS:
Service: %s
Date: %s
Time: %s
Duration: %d minutes
Booking Reference: %s

If you need to reschedule or cancel your appointment, please contact us as soon as possible with your booking reference number.

We look forward to seeing you tomorrow!

Best regards,
The Team
`, 
	service.Name,
	booking.ClientName,
	service.Name,
	timeSlot.Date,
	timeSlot.Time,
	timeSlot.Duration,
	booking.BookingRef)

	// Generate HTML version
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
        }
        .header {
            background-color: #38a169;
            padding: 20px;
            text-align: center;
            color: white;
        }
        .content {
            padding: 20px;
            background-color: #f9f9f9;
        }
        .booking-details {
            background-color: white;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 14px;
            color: #666;
        }
        h2 {
            color: #2f855a;
            margin-top: 0;
        }
        table {
            width: 100%%;
        }
        td {
            padding: 8px 0;
        }
        .label {
            font-weight: bold;
            width: 40%%;
        }
        .reminder-icon {
            font-size: 24px;
            margin-bottom: 10px;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="reminder-icon">⏰</div>
        <h1>Appointment Reminder</h1>
    </div>
    
    <div class="content">
        <p>Dear <strong>%s</strong>,</p>
        
        <p>This is a friendly reminder about your upcoming appointment tomorrow.</p>
        
        <div class="booking-details">
            <h2>Booking Details</h2>
            <table>
                <tr>
                    <td class="label">Service:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Date:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Time:</td>
                    <td>%s</td>
                </tr>
                <tr>
                    <td class="label">Duration:</td>
                    <td>%d minutes</td>
                </tr>
                <tr>
                    <td class="label">Booking Reference:</td>
                    <td>%s</td>
                </tr>
            </table>
        </div>
        
        <p>If you need to reschedule or cancel your appointment, please contact us as soon as possible with your booking reference number.</p>
        
        <p>We look forward to seeing you tomorrow!</p>
        
        <p>Best regards,<br>The Team</p>
    </div>
    
    <div class="footer">
        <p>&copy; 2025 Our Service. All rights reserved.</p>
    </div>
</body>
</html>
`,
	booking.ClientName,
	service.Name,
	timeSlot.Date,
	timeSlot.Time,
	timeSlot.Duration,
	booking.BookingRef)

	return plainTextContent, htmlContent
}

// sendReminderEmail sends a reminder email for an upcoming booking
func sendReminderEmail(booking Booking, timeSlot TimeSlot, service Service, email string) error {
    from := mail.NewEmail("Hyvinvointikeskusluxus", email)
    subject := fmt.Sprintf("Reminder: Your Appointment Tomorrow - %s", service.Name)
    to := mail.NewEmail(booking.ClientName, booking.ClientEmail)
    
    plainTextContent, htmlContent := generateReminderEmail(booking, timeSlot, service)
    message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
    client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
    
    response, err := client.Send(message)
		fmt.Println(response)
    if err != nil {
        slog.Error("Failed to send reminder email", "booking_ref", booking.BookingRef, "err", err)
        return err
    }
    
    slog.Info("Reminder email sent successfully", 
        "booking_ref", booking.BookingRef, 
        "status_code", response.StatusCode)
    
    return nil
}
