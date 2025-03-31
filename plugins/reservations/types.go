package reservations

import (
	"database/sql" // Added for sql.NullInt64/NullString/NullFloat64
	"errors"
	"fmt"
	"gothstack/app/db"       // Added for database access
	"gothstack/plugins/auth" // Added for User relation
	"time"

	// Import your user type if it exists elsewhere, e.g.:
	// "gothstack/plugins/auth"
	"gorm.io/gorm" // Uncommented for GORM operations
	//"gothstack/app/db" // Already added above
)

// --- Service ---

// Service represents a bookable service offered by a user.
type Service struct {
	ID           uint            `gorm:"primaryKey"`
	UserID       uint            `gorm:"column:user_id;not null;index"`
	Name         string          `gorm:"column:name;type:text;not null"`
	Description  sql.NullString  `gorm:"column:description;type:text"`    // Nullable TEXT
	Duration     int             `gorm:"column:duration;not null"`        // Duration in minutes
	Price        sql.NullFloat64 `gorm:"column:price;type:decimal(10,2)"` // Nullable DECIMAL
	Color        sql.NullString  `gorm:"column:color;type:text"`          // Nullable TEXT
	IsActive     bool            `gorm:"column:is_active;not null;default:1;index"`
	BufferBefore int             `gorm:"column:buffer_before;default:0"` // In minutes
	BufferAfter  int             `gorm:"column:buffer_after;default:0"`  // In minutes
	MaxAttendees int             `gorm:"column:max_attendees;default:1"`
	Location     sql.NullString  `gorm:"column:location;type:text"` // Nullable TEXT
	CreatedAt    time.Time       `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships
	User auth.User `gorm:"foreignKey:UserID"`
	// Add relationships to TimeSlots and Bookings if needed (usually inverse)
	// TimeSlots []TimeSlot `gorm:"foreignKey:ServiceID"`
	// Bookings []Booking `gorm:"foreignKey:ServiceID"`
}

// TableName specifies the table name for GORM.
func (Service) TableName() string {
	return "services"
}

// --- TimeSlot ---

// TimeSlot represents an available time slot for booking.
type TimeSlot struct {
	ID        uint          `gorm:"primaryKey"`
	UserID    uint          `gorm:"column:user_id;not null;index;index:idx_time_slots_user_date"` // User who owns this slot
	ServiceID sql.NullInt64 `gorm:"column:service_id;index"`                                      // Added reference to Service (nullable)
	Date      string        `gorm:"column:date;type:text;not null;index"`                         // Format: YYYY-MM-DD
	Time      string        `gorm:"column:time;type:text;not null"`                               // Format: HH:MM
	Duration  int           `gorm:"column:duration;not null"`                                     // Duration in minutes
	IsBooked  bool          `gorm:"column:is_booked;not null;default:0;index"`
	CreatedAt time.Time     `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time     `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User auth.User `gorm:"foreignKey:UserID"`
	Service  Service   `gorm:"foreignKey:ServiceID"`  // Added Service relationship
	Bookings []Booking `gorm:"foreignKey:TimeSlotID"` // A time slot can have one booking (or potentially more if logic allows?)
}

// TableName specifies the table name for GORM.
func (TimeSlot) TableName() string {
	return "time_slots"
}

// --- Booking ---

// BookingStatus defines possible statuses for a booking.
type BookingStatus string

const (
	StatusBookingConfirmed BookingStatus = "confirmed"
	StatusBookingCancelled BookingStatus = "canceled" // Note: SQL uses 'canceled'
	StatusBookingCompleted BookingStatus = "completed"
)

// Booking represents a confirmed booking for a specific time slot.
type Booking struct {
	ID          uint          `gorm:"primaryKey"`
	UserID      uint          `gorm:"column:user_id;not null;index"`      // User who owns the booked slot
	TimeSlotID  uint          `gorm:"column:time_slot_id;not null;index"` // Link to the booked time slot
	ServiceID   sql.NullInt64 `gorm:"column:service_id;index"`            // Added reference to Service (nullable)
	ClientName  string        `gorm:"column:client_name;type:text;not null"`
	ClientEmail string        `gorm:"column:client_email;type:text;not null"`
	ClientPhone string        `gorm:"column:client_phone;type:text"` // Nullable
	BookingRef  string        `gorm:"column:booking_ref;type:text;not null;uniqueIndex"`
	Notes       string        `gorm:"column:notes;type:text"` // Nullable
	Status      string        `gorm:"column:status;type:text;not null;default:'confirmed';index"`
	CreatedAt   time.Time     `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time     `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User     auth.User `gorm:"foreignKey:UserID"`
	TimeSlot TimeSlot `gorm:"foreignKey:TimeSlotID"`
	Service  Service  `gorm:"foreignKey:ServiceID"` // Added Service relationship
}

// TableName specifies the table name for GORM.
func (Booking) TableName() string {
	return "bookings"
}

// --- Setting ---

// Setting represents user-specific preferences for the reservation system.
type Setting struct {
	ID                   uint      `gorm:"primaryKey"`
	UserID               uint      `gorm:"column:user_id;not null;uniqueIndex"` // Each user has one settings row
	Timezone             string    `gorm:"column:timezone;type:text;not null;default:'UTC'"`
	NotificationEmail    bool      `gorm:"column:notification_email;not null;default:1"`
	NotificationSMS      bool      `gorm:"column:notification_sms;not null;default:0"`
	CalendarView         string    `gorm:"column:calendar_view;type:text;not null;default:'week'"` // week, month, day
	MinSchedulingNotice  int       `gorm:"column:min_scheduling_notice;not null;default:24"`       // hours
	MaxSchedulingAdvance int       `gorm:"column:max_scheduling_advance;not null;default:60"`      // days
	CreatedAt            time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt            time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User auth.User `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for GORM.
func (Setting) TableName() string {
	return "settings"
}

// --- BusinessHour ---

// BusinessHour defines the regular working hours for a specific day of the week.
type BusinessHour struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"column:user_id;not null;uniqueIndex:idx_user_day"`     // Combined unique index
	DayOfWeek    int       `gorm:"column:day_of_week;not null;uniqueIndex:idx_user_day"` // 0-6 (Sunday-Saturday)
	StartTime    string    `gorm:"column:start_time;type:text;not null"`                 // Format: HH:MM
	EndTime      string    `gorm:"column:end_time;type:text;not null"`                   // Format: HH:MM
	IsWorkingDay bool      `gorm:"column:is_working_day;not null;default:1"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User auth.User `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for GORM.
func (BusinessHour) TableName() string {
	return "business_hours"
}

// --- SpecialDate ---

// SpecialDate defines exceptions to regular business hours (e.g., holidays, days off).
type SpecialDate struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"column:user_id;not null;uniqueIndex:idx_user_date"`        // Combined unique index
	Date         string    `gorm:"column:date;type:text;not null;uniqueIndex:idx_user_date"` // Format: YYYY-MM-DD
	Description  string    `gorm:"column:description;type:text;not null"`
	IsWorkingDay bool      `gorm:"column:is_working_day;not null;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User auth.User `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for GORM.
func (SpecialDate) TableName() string {
	return "special_dates"
}

// --- Webhook ---

// Webhook defines endpoints to be notified on certain reservation events.
type Webhook struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"column:user_id;not null"`
	URL       string    `gorm:"column:url;type:text;not null"`
	Events    string    `gorm:"column:events;type:text;not null"` // Comma-separated list
	IsActive  bool      `gorm:"column:is_active;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User auth.User `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for GORM.
func (Webhook) TableName() string {
	return "webhooks"
}

// --- TimeSlot CRUD ---

// CreateTimeSlot creates a new time slot.
func CreateTimeSlot(userID uint, serviceID sql.NullInt64, dateStr, timeStr string, duration int) (TimeSlot, error) {
	slot := TimeSlot{
		UserID:    userID,
		ServiceID: serviceID, // Assign service ID
		Date:      dateStr,
		Time:      timeStr,
		Duration:  duration,
		IsBooked:  false,
	}
	result := db.Get().Create(&slot)
	if result.Error != nil {
		return slot, fmt.Errorf("failed to create time slot: %w", result.Error)
	}
	return slot, nil
}

// GetTimeSlot retrieves a time slot by its ID.
func GetTimeSlot(id uint) (TimeSlot, error) {
	var slot TimeSlot
	// Add UserID check if necessary for security: .Where("user_id = ?", userID)
	result := db.Get().Preload("Service").First(&slot, id) // Preload Service
	if result.Error != nil {
		return slot, fmt.Errorf("failed to retrieve time slot %d: %w", id, result.Error)
	}
	return slot, nil
}

// ListTimeSlots retrieves all time slots for a specific user.
func ListTimeSlots(userID uint) ([]TimeSlot, error) {
	var slots []TimeSlot
	result := db.Get().Preload("Service").Where("user_id = ?", userID).Order("date asc, time asc").Find(&slots) // Preload Service
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list time slots for user %d: %w", userID, result.Error)
	}
	return slots, nil
}

// ListAvailableTimeSlots retrieves available (not booked) time slots for a user within a date range.
func ListAvailableTimeSlots(userID uint, startDate, endDate time.Time) ([]TimeSlot, error) {
	var slots []TimeSlot
	// Ensure date strings are formatted correctly for the query (YYYY-MM-DD)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	result := db.Get().
		Preload("Service"). // Preload Service
		Where("user_id = ? AND is_booked = ? AND date BETWEEN ? AND ?", userID, false, startDateStr, endDateStr).
		Order("date asc, time asc").
		Find(&slots)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list available time slots for user %d: %w", userID, result.Error)
	}
	return slots, nil
}

// UpdateTimeSlot updates an existing time slot (e.g., mark as booked).
func UpdateTimeSlot(slot TimeSlot) (TimeSlot, error) {
	// Ensure UpdatedAt is set
	slot.UpdatedAt = time.Now()
	result := db.Get().Save(&slot)
	if result.Error != nil {
		return slot, fmt.Errorf("failed to update time slot %d: %w", slot.ID, result.Error)
	}
	return slot, nil
}

// DeleteTimeSlot deletes a time slot by its ID.
func DeleteTimeSlot(id uint) error {
	// Add UserID check if necessary for security: .Where("user_id = ?", userID)
	result := db.Get().Delete(&TimeSlot{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete time slot %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("time slot %d not found for deletion", id) // Or use gorm.ErrRecordNotFound
	}
	return nil
}

// GenerateTimeSlotsFromBusinessHours creates available time slots for a given date range
// based on the user's business hours and service duration.
func GenerateTimeSlotsFromBusinessHours(userID uint, serviceID uint, startDate, endDate time.Time, serviceDuration int) ([]TimeSlot, error) {
	// 1. Get business hours for the user
	businessHours, err := GetBusinessHours(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business hours: %w", err)
	}

	// Create a map for easier access to business hours by day of week
	businessHoursByDay := make(map[int]BusinessHour)
	for _, bh := range businessHours {
		businessHoursByDay[bh.DayOfWeek] = bh
	}

	// 2. Get special dates that might override business hours
	specialDates, err := GetSpecialDates(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get special dates: %w", err)
	}

	// Create a map of special dates for easier lookup
	specialDateMap := make(map[string]SpecialDate)
	for _, sd := range specialDates {
		specialDateMap[sd.Date] = sd
	}

	var slots []TimeSlot
	currentDate := startDate

	// Iterate through each day in the date range
	for !currentDate.After(endDate) {
		dateStr := currentDate.Format("2006-01-02")

		// Check if it's a special date
		if specialDate, exists := specialDateMap[dateStr]; exists {
			if !specialDate.IsWorkingDay {
				currentDate = currentDate.AddDate(0, 0, 1)
				continue
			}
			// Could add special hours handling here if needed
		}

		// Get business hours for this day of week
		dayOfWeek := int(currentDate.Weekday())
		businessHour, exists := businessHoursByDay[dayOfWeek]
		if !exists || !businessHour.IsWorkingDay {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Parse business hours
		startTime, _ := time.Parse("15:04", businessHour.StartTime)
		endTime, _ := time.Parse("15:04", businessHour.EndTime)

		// Adjust times to current date
		startDateTime := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, currentDate.Location(),
		)
		endDateTime := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0, currentDate.Location(),
		)

		// Generate time slots for this day
		currentSlot := startDateTime
		for currentSlot.Add(time.Duration(serviceDuration)*time.Minute).Before(endDateTime) ||
			currentSlot.Add(time.Duration(serviceDuration)*time.Minute).Equal(endDateTime) {

			// Skip if this slot is in the past
			if currentSlot.Before(time.Now()) {
				currentSlot = currentSlot.Add(30 * time.Minute) // 30-minute intervals
				continue
			}

			slots = append(slots, TimeSlot{
				UserID:    userID,
				ServiceID: sql.NullInt64{Int64: int64(serviceID), Valid: true},
				Date:      dateStr,
				Time:      currentSlot.Format("15:04"),
				Duration:  serviceDuration,
				IsBooked:  false,
			})

			currentSlot = currentSlot.Add(30 * time.Minute) // 30-minute intervals
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return slots, nil
}

// ListAvailableTimeSlotsForService retrieves available time slots for a service,
// taking into account business hours and existing bookings.
func ListAvailableTimeSlotsForService(userID uint, serviceID uint, startDate, endDate time.Time) ([]TimeSlot, error) {
	// 1. Get the service to check duration
	service, err := GetService(serviceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service details: %w", err)
	}

	// 2. Generate potential time slots based on business hours
	slots, err := GenerateTimeSlotsFromBusinessHours(userID, serviceID, startDate, endDate, service.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate time slots: %w", err)
	}

	// 3. Get existing bookings to filter out unavailable slots
	var existingSlots []TimeSlot
	err = db.Get().
		Where("user_id = ? AND date BETWEEN ? AND ? AND is_booked = ?",
			userID,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			true).
		Find(&existingSlots).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check existing bookings: %w", err)
	}

	// Create a map of booked slots for efficient lookup
	bookedSlots := make(map[string]bool)
	for _, slot := range existingSlots {
		key := fmt.Sprintf("%s-%s", slot.Date, slot.Time)
		bookedSlots[key] = true
	}

	// Filter out booked slots
	var availableSlots []TimeSlot
	for _, slot := range slots {
		key := fmt.Sprintf("%s-%s", slot.Date, slot.Time)
		if !bookedSlots[key] {
			availableSlots = append(availableSlots, slot)
		}
	}

	return availableSlots, nil
}

// --- Booking CRUD ---

// CreateBooking creates a new booking and marks the associated time slot as booked.
func CreateBooking(userID, timeSlotID uint, serviceID sql.NullInt64, clientName, clientEmail, clientPhone, bookingRef, notes string) (Booking, error) {
	// Use a transaction to ensure atomicity
	tx := db.Get().Begin()
	if tx.Error != nil {
		return Booking{}, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// 1. Check if the time slot exists and is available
	var slot TimeSlot
	if err := tx.Where("id = ? AND user_id = ?", timeSlotID, userID).First(&slot).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return Booking{}, fmt.Errorf("time slot %d not found or does not belong to user %d", timeSlotID, userID)
		}
		return Booking{}, fmt.Errorf("failed to retrieve time slot %d: %w", timeSlotID, err)
	}
	if slot.IsBooked {
		tx.Rollback()
		return Booking{}, fmt.Errorf("time slot %d is already booked", timeSlotID)
	}

	// 2. Create the booking
	booking := Booking{
		UserID:      userID,
		TimeSlotID:  timeSlotID,
		ServiceID:   serviceID, // Assign service ID
		ClientName:  clientName,
		ClientEmail: clientEmail,
		ClientPhone: clientPhone, // Assumes empty string if not provided
		BookingRef:  bookingRef,
		Notes:       notes, // Assumes empty string if not provided
		Status:      string(StatusBookingConfirmed),
		// CreatedAt/UpdatedAt handled by GORM defaults
	}
	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		return Booking{}, fmt.Errorf("failed to create booking: %w", err)
	}

	// 3. Mark the time slot as booked
	slot.IsBooked = true
	slot.UpdatedAt = time.Now() // Explicitly set UpdatedAt
	if err := tx.Save(&slot).Error; err != nil {
		tx.Rollback()
		return Booking{}, fmt.Errorf("failed to mark time slot %d as booked: %w", timeSlotID, err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return Booking{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Optionally reload the booking with the TimeSlot preloaded
	// tx.Preload("TimeSlot").First(&booking, booking.ID)

	return booking, nil
}

// GetBooking retrieves a booking by its ID, ensuring it belongs to the user.
func GetBooking(id, userID uint) (Booking, error) {
	var booking Booking
	result := db.Get().Preload("TimeSlot").Preload("Service").Where("id = ? AND user_id = ?", id, userID).First(&booking) // Preload Service
	if result.Error != nil {
		return booking, fmt.Errorf("failed to retrieve booking %d for user %d: %w", id, userID, result.Error)
	}
	return booking, nil
}

// GetBookingByRef retrieves a booking by its unique reference.
func GetBookingByRef(bookingRef string) (Booking, error) {
	var booking Booking
	result := db.Get().Preload("TimeSlot").Preload("Service").Where("booking_ref = ?", bookingRef).First(&booking) // Preload Service
	if result.Error != nil {
		return booking, fmt.Errorf("failed to retrieve booking with ref %s: %w", bookingRef, result.Error)
	}
	return booking, nil
}

// ListBookings retrieves all bookings for a specific user.
func ListBookings(userID uint) ([]Booking, error) {
	var bookings []Booking
	result := db.Get().Preload("TimeSlot").Preload("Service").Where("user_id = ?", userID).Order("created_at desc").Find(&bookings) // Preload Service
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list bookings for user %d: %w", userID, result.Error)
	}
	return bookings, nil
}

// UpdateBookingStatus updates the status of an existing booking.
func UpdateBookingStatus(id, userID uint, status BookingStatus) (Booking, error) {
	var booking Booking
	// Find the booking first to ensure it belongs to the user
	result := db.Get().Where("id = ? AND user_id = ?", id, userID).First(&booking)
	if result.Error != nil {
		return booking, fmt.Errorf("failed to retrieve booking %d for update: %w", id, result.Error)
	}

	booking.Status = string(status)
	booking.UpdatedAt = time.Now()

	// Save the updated booking
	result = db.Get().Save(&booking)
	if result.Error != nil {
		return booking, fmt.Errorf("failed to update booking status %d: %w", id, result.Error)
	}
	return booking, nil
}

// CancelBooking cancels a booking and makes the time slot available again.
func CancelBooking(id, userID uint) error {
	tx := db.Get().Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// 1. Find the booking
	var booking Booking
	if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&booking).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to retrieve booking %d for cancellation: %w", id, err)
	}

	// 2. Update booking status to canceled
	booking.Status = string(StatusBookingCancelled)
	booking.UpdatedAt = time.Now()
	if err := tx.Save(&booking).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update booking status to canceled for %d: %w", id, err)
	}

	// 3. Mark the associated time slot as available
	var slot TimeSlot
	if err := tx.First(&slot, booking.TimeSlotID).Error; err == nil {
		slot.IsBooked = false
		slot.UpdatedAt = time.Now()
		if err := tx.Save(&slot).Error; err != nil {
			// Log this error but don't necessarily fail the whole cancel operation
			fmt.Printf("Warning: failed to mark time slot %d as available during cancellation: %v\n", slot.ID, err)
			// Consider if you should rollback here based on requirements
		}
	} else {
		// Log error if timeslot cannot be found, but proceed with cancellation
		fmt.Printf("Warning: could not find associated time slot %d during cancellation: %v\n", booking.TimeSlotID, err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit cancellation transaction: %w", err)
	}

	return nil
}

// DeleteBooking deletes a booking by its ID. Consider using CancelBooking instead for history.
func DeleteBooking(id, userID uint) error {
	// WARNING: This permanently removes the booking record.
	// Usually, changing the status (e.g., 'canceled') is preferred.
	// If deleting, ensure the associated TimeSlot is handled appropriately (e.g., made available).
	// A transaction is highly recommended here if you also modify the TimeSlot.

	result := db.Get().Where("id = ? AND user_id = ?", id, userID).Delete(&Booking{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete booking %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("booking %d not found for deletion or user mismatch", id)
	}
	return nil
}

// --- Settings CRUD ---

// GetSettings retrieves the settings for a specific user.
func GetSettings(userID uint) (Setting, error) {
	var settings Setting
	// Use FirstOrCreate to handle cases where settings might not exist yet for a user
	result := db.Get().Where(Setting{UserID: userID}).FirstOrCreate(&settings)
	if result.Error != nil {
		return settings, fmt.Errorf("failed to get or create settings for user %d: %w", userID, result.Error)
	}
	return settings, nil
}

// UpdateSettings updates the settings for a specific user.
func UpdateSettings(settings Setting) (Setting, error) {
	// Ensure the UserID is set correctly and UpdatedAt is updated
	if settings.UserID == 0 {
		return settings, fmt.Errorf("cannot update settings without a UserID")
	}
	settings.UpdatedAt = time.Now()

	// Use Updates to only update specified fields, preventing accidental overwrites
	// or use Save to replace the whole record (ensure ID is present if using Save)
	result := db.Get().Model(&Setting{}).Where("user_id = ?", settings.UserID).Updates(settings)

	if result.Error != nil {
		return settings, fmt.Errorf("failed to update settings for user %d: %w", settings.UserID, result.Error)
	}
	if result.RowsAffected == 0 {
		// This might happen if the record didn't exist and FirstOrCreate wasn't called before
		// Or if the data being sent is identical to the existing data
		// You might want to retrieve the settings again to be sure.
		return GetSettings(settings.UserID) // Retrieve potentially created/existing settings
	}

	// Retrieve the full updated record as Updates doesn't return the full object easily
	return GetSettings(settings.UserID)
}

// Note: Create for Settings is handled by GetSettings (FirstOrCreate).
// Note: Delete for Settings might not be desirable; perhaps disable user instead.

// --- BusinessHour CRUD ---

func CreateBusinessHour(userID uint, dayOfWeek int, startTime, endTime string, isWorking bool) (BusinessHour, error) {
	// First try to find an existing record
	var existingBH BusinessHour
	findResult := db.Get().Where("user_id = ? AND day_of_week = ?", userID, dayOfWeek).First(&existingBH)

	// If record exists, update it
	if findResult.Error == nil {
		// Update existing record with new values
		existingBH.StartTime = startTime
		existingBH.EndTime = endTime
		existingBH.IsWorkingDay = isWorking

		updateResult := db.Get().Save(&existingBH)
		if updateResult.Error != nil {
			return existingBH, fmt.Errorf("failed to update business hour for user %d, day %d: %w", userID, dayOfWeek, updateResult.Error)
		}
		return existingBH, nil
	}

	// If record doesn't exist or another error occurred
	if findResult.Error != nil && !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
		return BusinessHour{}, fmt.Errorf("error checking for existing business hour for user %d, day %d: %w", userID, dayOfWeek, findResult.Error)
	}

	// Create new record
	newBH := BusinessHour{
		UserID:       userID,
		DayOfWeek:    dayOfWeek,
		StartTime:    startTime,
		EndTime:      endTime,
		IsWorkingDay: isWorking,
		// CreatedAt/UpdatedAt handled by GORM defaults
	}

	createResult := db.Get().Create(&newBH)
	if createResult.Error != nil {
		return newBH, fmt.Errorf("failed to create business hour for user %d, day %d: %w", userID, dayOfWeek, createResult.Error)
	}

	return newBH, nil
}

// GetBusinessHours retrieves all business hours for a specific user.
func GetBusinessHours(userID uint) ([]BusinessHour, error) {
	var hours []BusinessHour
	result := db.Get().Where("user_id = ?", userID).Order("day_of_week asc").Find(&hours)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list business hours for user %d: %w", userID, result.Error)
	}
	return hours, nil
}

// UpdateBusinessHour updates an existing business hour record.
// It finds the record by UserID and DayOfWeek.
func UpdateBusinessHour(userID uint, dayOfWeek int, startTime, endTime string, isWorking bool) (BusinessHour, error) {
	var bh BusinessHour
	// Find the existing record
	findResult := db.Get().Where("user_id = ? AND day_of_week = ?", userID, dayOfWeek).First(&bh)
	if findResult.Error != nil {
		return bh, fmt.Errorf("failed to find business hour for user %d, day %d to update: %w", userID, dayOfWeek, findResult.Error)
	}

	// Update fields
	bh.StartTime = startTime
	bh.EndTime = endTime
	bh.IsWorkingDay = isWorking
	bh.UpdatedAt = time.Now()

	// Save changes
	saveResult := db.Get().Save(&bh)
	if saveResult.Error != nil {
		return bh, fmt.Errorf("failed to update business hour for user %d, day %d: %w", userID, dayOfWeek, saveResult.Error)
	}
	return bh, nil
}

// DeleteBusinessHour deletes a business hour record by user and day.
func DeleteBusinessHour(userID uint, dayOfWeek int) error {
	result := db.Get().Where("user_id = ? AND day_of_week = ?", userID, dayOfWeek).Delete(&BusinessHour{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete business hour for user %d, day %d: %w", userID, dayOfWeek, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("business hour not found for user %d, day %d", userID, dayOfWeek)
	}
	return nil
}

// --- SpecialDate CRUD ---

// CreateSpecialDate creates a new special date record.
func CreateSpecialDate(userID uint, dateStr, description string, isWorking bool) (SpecialDate, error) {
	sd := SpecialDate{
		UserID:       userID,
		Date:         dateStr, // Ensure YYYY-MM-DD format
		Description:  description,
		IsWorkingDay: isWorking,
		// CreatedAt/UpdatedAt handled by GORM defaults
	}
	// Use FirstOrCreate or handle unique constraint violation (user_id, date)
	result := db.Get().Where(SpecialDate{UserID: userID, Date: dateStr}).Assign(sd).FirstOrCreate(&sd)
	if result.Error != nil {
		return sd, fmt.Errorf("failed to create or find special date for user %d, date %s: %w", userID, dateStr, result.Error)
	}
	return sd, nil
}

// GetSpecialDates retrieves all special dates for a specific user.
func GetSpecialDates(userID uint) ([]SpecialDate, error) {
	var dates []SpecialDate
	result := db.Get().Where("user_id = ?", userID).Order("date asc").Find(&dates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list special dates for user %d: %w", userID, result.Error)
	}
	return dates, nil
}

// GetSpecialDate retrieves a specific special date by user and date string.
func GetSpecialDate(userID uint, dateStr string) (SpecialDate, error) {
	var sd SpecialDate
	result := db.Get().Where("user_id = ? AND date = ?", userID, dateStr).First(&sd)
	if result.Error != nil {
		return sd, fmt.Errorf("failed to retrieve special date for user %d, date %s: %w", userID, dateStr, result.Error)
	}
	return sd, nil
}

// UpdateSpecialDate updates an existing special date record.
func UpdateSpecialDate(userID uint, dateStr, description string, isWorking bool) (SpecialDate, error) {
	var sd SpecialDate
	// Find by UserID and Date
	findResult := db.Get().Where("user_id = ? AND date = ?", userID, dateStr).First(&sd)
	if findResult.Error != nil {
		return sd, fmt.Errorf("failed to find special date for user %d, date %s to update: %w", userID, dateStr, findResult.Error)
	}

	// Update fields
	sd.Description = description
	sd.IsWorkingDay = isWorking
	sd.UpdatedAt = time.Now()

	// Save changes
	saveResult := db.Get().Save(&sd)
	if saveResult.Error != nil {
		return sd, fmt.Errorf("failed to update special date for user %d, date %s: %w", userID, dateStr, saveResult.Error)
	}
	return sd, nil
}

// DeleteSpecialDate deletes a special date record by user and date.
func DeleteSpecialDate(userID uint, dateStr string) error {
	result := db.Get().Where("user_id = ? AND date = ?", userID, dateStr).Delete(&SpecialDate{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete special date for user %d, date %s: %w", userID, dateStr, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("special date not found for user %d, date %s", userID, dateStr)
	}
	return nil
}

// --- Webhook CRUD ---

// CreateWebhook creates a new webhook.
func CreateWebhook(userID uint, url, events string, isActive bool) (Webhook, error) {
	wh := Webhook{
		UserID:   userID,
		URL:      url,
		Events:   events, // Comma-separated string
		IsActive: isActive,
		// CreatedAt/UpdatedAt handled by GORM defaults
	}
	result := db.Get().Create(&wh)
	if result.Error != nil {
		return wh, fmt.Errorf("failed to create webhook for user %d: %w", userID, result.Error)
	}
	return wh, nil
}

// GetWebhook retrieves a webhook by its ID, ensuring it belongs to the user.
func GetWebhook(id, userID uint) (Webhook, error) {
	var wh Webhook
	result := db.Get().Where("id = ? AND user_id = ?", id, userID).First(&wh)
	if result.Error != nil {
		return wh, fmt.Errorf("failed to retrieve webhook %d for user %d: %w", id, userID, result.Error)
	}
	return wh, nil
}

// ListWebhooks retrieves all webhooks for a specific user.
func ListWebhooks(userID uint) ([]Webhook, error) {
	var hooks []Webhook
	result := db.Get().Where("user_id = ?", userID).Order("created_at desc").Find(&hooks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list webhooks for user %d: %w", userID, result.Error)
	}
	return hooks, nil
}

// UpdateWebhook updates an existing webhook.
func UpdateWebhook(hook Webhook) (Webhook, error) {
	// Ensure UserID is present and UpdatedAt is set
	if hook.UserID == 0 {
		return hook, fmt.Errorf("cannot update webhook without UserID")
	}
	if hook.ID == 0 {
		return hook, fmt.Errorf("cannot update webhook without ID")
	}
	hook.UpdatedAt = time.Now()

	// Use Save to update the record by ID, implicitly includes a WHERE id = ? clause
	// Ensure the hook object passed in has the correct ID.
	result := db.Get().Save(&hook)
	if result.Error != nil {
		return hook, fmt.Errorf("failed to update webhook %d: %w", hook.ID, result.Error)
	}
	return hook, nil
}

// DeleteWebhook deletes a webhook by its ID, ensuring it belongs to the user.
func DeleteWebhook(id, userID uint) error {
	result := db.Get().Where("id = ? AND user_id = ?", id, userID).Delete(&Webhook{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete webhook %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook %d not found for deletion or user mismatch", id)
	}
	return nil
}

// Note: The APIKeys table and the additions to the Users table are not included here,
// as they might belong in different plugin type files (e.g., auth plugin).
// Adjust UserID types (uint) if your user ID is different.

// --- Service CRUD ---

// CreateService creates a new service for a user.
// Note: Pass nullable fields using sql.NullString, sql.NullFloat64, etc.
func CreateService(service Service) (Service, error) {
	// Ensure UserID is set
	if service.UserID == 0 {
		return Service{}, fmt.Errorf("cannot create service without UserID")
	}
	// GORM handles CreatedAt/UpdatedAt defaults
	result := db.Get().Create(&service)
	if result.Error != nil {
		return service, fmt.Errorf("failed to create service for user %d: %w", service.UserID, result.Error)
	}
	return service, nil
}

// GetService retrieves a specific service by ID, ensuring it belongs to the user.
func GetService(id, userID uint) (Service, error) {
	var service Service
	result := db.Get().Where("id = ? AND user_id = ?", id, userID).First(&service)
	if result.Error != nil {
		return service, fmt.Errorf("failed to retrieve service %d for user %d: %w", id, userID, result.Error)
	}
	return service, nil
}

// ListServices retrieves all services for a specific user (optionally filter by active status).
func ListServices(userID uint, activeOnly bool) ([]Service, error) {
	var services []Service
	query := db.Get().Where("user_id = ?", userID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	result := query.Order("name asc").Find(&services)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list services for user %d: %w", userID, result.Error)
	}
	return services, nil
}

// UpdateService updates an existing service.
// Ensure the passed 'service' struct has the correct ID and UserID set.
func UpdateService(service Service) (Service, error) {
	if service.ID == 0 || service.UserID == 0 {
		return Service{}, fmt.Errorf("cannot update service without ID and UserID")
	}
	// GORM automatically handles UpdatedAt if not explicitly set
	service.UpdatedAt = time.Now()

	// Use Save to update all fields, ensure caller provides the full object retrieved first.
	// Or use Model().Where().Updates() to update specific fields.
	result := db.Get().Save(&service) // Requires fetching the service first to have all fields
	// Example using Updates:
	// result := db.Get().Model(&Service{}).Where("id = ? AND user_id = ?", service.ID, service.UserID).Updates(service)

	if result.Error != nil {
		return service, fmt.Errorf("failed to update service %d for user %d: %w", service.ID, service.UserID, result.Error)
	}
	if result.RowsAffected == 0 {
		// Could mean record not found or data was identical
		// Re-fetch to confirm and return correct state
		return GetService(service.ID, service.UserID)
	}
	return service, nil // Save returns the updated object if it runs Hooks, otherwise might need re-fetch
}

// DeleteService deletes a service by its ID, ensuring it belongs to the user.
// Consider implications: associated TimeSlots/Bookings might need handling.
func DeleteService(id, userID uint) error {
	// Add checks here if needed: e.g., check if service is linked to future bookings.
	result := db.Get().Where("id = ? AND user_id = ?", id, userID).Delete(&Service{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete service %d for user %d: %w", id, userID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("service %d not found for user %d or already deleted", id, userID)
	}
	return nil
}
