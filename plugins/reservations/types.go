package reservations

import (
	"fmt"
	"gothstack/app/db" // Added for database access
	"time"

	// Import your user type if it exists elsewhere, e.g.:
	// "gothstack/plugins/auth"
	"gorm.io/gorm" // Uncommented for GORM operations
	//"gothstack/app/db" // Already added above
)

// --- TimeSlot ---

// TimeSlot represents an available time slot for booking.
type TimeSlot struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"column:user_id;not null;index;index:idx_time_slots_user_date"` // User who owns this slot
	Date      string    `gorm:"column:date;type:text;not null;index"`                         // Format: YYYY-MM-DD
	Time      string    `gorm:"column:time;type:text;not null"`                               // Format: HH:MM
	Duration  int       `gorm:"column:duration;not null"`                                     // Duration in minutes
	IsBooked  bool      `gorm:"column:is_booked;not null;default:0;index"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User auth.User `gorm:"foreignKey:UserID"`
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
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"column:user_id;not null;index"`      // User who owns the booked slot
	TimeSlotID  uint      `gorm:"column:time_slot_id;not null;index"` // Link to the booked time slot
	ClientName  string    `gorm:"column:client_name;type:text;not null"`
	ClientEmail string    `gorm:"column:client_email;type:text;not null"`
	ClientPhone string    `gorm:"column:client_phone;type:text"` // Nullable
	BookingRef  string    `gorm:"column:booking_ref;type:text;not null;uniqueIndex"`
	Notes       string    `gorm:"column:notes;type:text"` // Nullable
	Status      string    `gorm:"column:status;type:text;not null;default:'confirmed';index"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`

	// Define GORM relationships (optional)
	// User     auth.User `gorm:"foreignKey:UserID"`
	TimeSlot TimeSlot `gorm:"foreignKey:TimeSlotID"`
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
func CreateTimeSlot(userID uint, dateStr, timeStr string, duration int) (TimeSlot, error) {
	slot := TimeSlot{
		UserID:   userID,
		Date:     dateStr,
		Time:     timeStr,
		Duration: duration,
		IsBooked: false, // Default to not booked
		// CreatedAt and UpdatedAt are handled by GORM defaults if configured,
		// otherwise set them explicitly:
		// CreatedAt: time.Now(),
		// UpdatedAt: time.Now(),
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
	result := db.Get().First(&slot, id)
	if result.Error != nil {
		return slot, fmt.Errorf("failed to retrieve time slot %d: %w", id, result.Error)
	}
	return slot, nil
}

// ListTimeSlots retrieves all time slots for a specific user.
func ListTimeSlots(userID uint) ([]TimeSlot, error) {
	var slots []TimeSlot
	result := db.Get().Where("user_id = ?", userID).Order("date asc, time asc").Find(&slots)
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

// --- Booking CRUD ---

// CreateBooking creates a new booking and marks the associated time slot as booked.
func CreateBooking(userID, timeSlotID uint, clientName, clientEmail, clientPhone, bookingRef, notes string) (Booking, error) {
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
	result := db.Get().Preload("TimeSlot").Where("id = ? AND user_id = ?", id, userID).First(&booking)
	if result.Error != nil {
		return booking, fmt.Errorf("failed to retrieve booking %d for user %d: %w", id, userID, result.Error)
	}
	return booking, nil
}

// GetBookingByRef retrieves a booking by its unique reference.
func GetBookingByRef(bookingRef string) (Booking, error) {
	var booking Booking
	result := db.Get().Preload("TimeSlot").Where("booking_ref = ?", bookingRef).First(&booking)
	if result.Error != nil {
		return booking, fmt.Errorf("failed to retrieve booking with ref %s: %w", bookingRef, result.Error)
	}
	return booking, nil
}

// ListBookings retrieves all bookings for a specific user.
func ListBookings(userID uint) ([]Booking, error) {
	var bookings []Booking
	result := db.Get().Preload("TimeSlot").Where("user_id = ?", userID).Order("created_at desc").Find(&bookings)
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

// CreateBusinessHour creates a new business hour record.
func CreateBusinessHour(userID uint, dayOfWeek int, startTime, endTime string, isWorking bool) (BusinessHour, error) {
	bh := BusinessHour{
		UserID:       userID,
		DayOfWeek:    dayOfWeek,
		StartTime:    startTime,
		EndTime:      endTime,
		IsWorkingDay: isWorking,
		// CreatedAt/UpdatedAt handled by GORM defaults
	}
	// Use FirstOrCreate or handle potential unique constraint violation (user_id, day_of_week)
	result := db.Get().Where(BusinessHour{UserID: userID, DayOfWeek: dayOfWeek}).Assign(bh).FirstOrCreate(&bh)
	// Or just Create and let it fail if exists: result := db.Get().Create(&bh)
	if result.Error != nil {
		return bh, fmt.Errorf("failed to create or find business hour for user %d, day %d: %w", userID, dayOfWeek, result.Error)
	}
	return bh, nil
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
