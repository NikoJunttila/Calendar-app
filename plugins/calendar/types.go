package calendar

import (
	"fmt"
	"gothstack/app/db"
	"gothstack/plugins/auth"
	"time"

	"gorm.io/gorm"
)

// Calendar represents the calendars table in the database
type Calendar struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	IndexNumber    int    `gorm:"column:index_number"`
	OwnerID        uint   `gorm:"not null"`
	Work           bool
	DailyWorkHours float64
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relationship fields
	Entries []CalendarEntry `gorm:"foreignKey:CalendarID"`
	User    auth.User       `gorm:"foreignKey:OwnerID"`
}

// CalendarEntry represents the calendar_entrys table in the database
type CalendarEntry struct {
	ID             uint `gorm:"primaryKey"`
	CalendarID     uint `gorm:"not null"`
	WorkResourceID uint
	Date           time.Time `gorm:"not null"`
	Year           int       `gorm:"not null"`
	Month          int       `gorm:"not null"`
	Week           int       `gorm:"not null"`
	Hours          float64
	Text           string         `gorm:"not null"`
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relationship field
	Calendar     Calendar     `gorm:"foreignKey:CalendarID"`
	WorkResource WorkResource `gorm:"foreignKey:WorkResourceID"`
}

// Event name constants
const (
	CalendarCreatedEvent      = "calendar.created"
	CalendarEntryCreatedEvent = "calendar.entry.created"
)

// CreateCalendar creates a new calendar with the given name and index number
func CreateCalendar(name string, work bool, avgHours float64, owner_id uint) (Calendar, error) {
	//get index number automatically?
	calendar := Calendar{
		Name:           name,
		IndexNumber:    1,
		Work:           work,
		DailyWorkHours: avgHours,
		OwnerID:        owner_id,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	result := db.Get().Create(&calendar)
	return calendar, result.Error
}

// GetCalendar retrieves a calendar by its ID
func GetCalendar(id, ownerID uint) (Calendar, error) {
	var calendar Calendar
	result := db.Get().Where("id = ? AND owner_id = ?", id, ownerID).First(&calendar)
	return calendar, result.Error
}

func ListCalendars(ownerID uint) ([]Calendar, error) {
	var calendars []Calendar
	result := db.Get().Where("owner_id = ?", ownerID).Order("index_number asc").Find(&calendars)
	return calendars, result.Error
}

func GetCalendarWithEntries(id uint, ownerID uint) (Calendar, error) {
	var calendar Calendar

	// Step 1: Fetch the calendar with owner check
	// This ensures only the legitimate owner can access their calendar
	// We use First() instead of Find() because we're looking for a single record by primary key
	// The Where clause adds security by checking both the ID and ownerID match
	if err := db.Get().Where("id = ? AND owner_id = ?", id, ownerID).First(&calendar).Error; err != nil {
		// If no calendar is found or another error occurs, return early with the empty calendar and error
		return calendar, err
	}

	// Step 2: Fetch the entries for this calendar in a separate query
	// We do this in two steps rather than one complex query with preloading because:
	// - It gives us more control over how entries are loaded (sorting, filtering)
	// - It's clearer what's happening in each step
	// - It allows us to return immediately if the calendar doesn't exist or belong to the user
	if err := db.Get().
		// Only get entries for this specific calendar
		Where("calendar_id = ?", id).
		// Sort the entries by date in ascending order (oldest to newest)
		// This makes the data more useful for display and processing
		Order("date asc").
		// Eager load the WorkResource relationship for each entry
		// This prevents the N+1 query problem by loading all related resources in one go
		Preload("WorkResource").
		// Populate the Entries slice of our calendar struct
		Find(&calendar.Entries).Error; err != nil {
		// If an error occurs while fetching entries, return what we have so far with the error
		return calendar, err
	}

	// Return the calendar with its entries (sorted by date) and no error
	return calendar, nil
}

func GetCalendarWithEntriesByMonth(calendarID uint, ownerID uint, year, month int) (Calendar, error) {
	var calendar Calendar
	// First, fetch the calendar with owner check
	if err := db.Get().Where("id = ? AND owner_id = ?", calendarID, ownerID).First(&calendar).Error; err != nil {
		return calendar, err
	}

	// Then, fetch the entries for this calendar for the specified month and year
	// while preloading the WorkResource for each entry
	if err := db.Get().
		Where("calendar_id = ? AND year = ? AND month = ?", calendarID, year, month).
		Order("date asc").
		Preload("WorkResource").
		Find(&calendar.Entries).Error; err != nil {
		return calendar, err
	}

	return calendar, nil
}

// ListCalendarEntries returns all entries for a specific calendar
func ListCalendarEntries(calendarID uint) ([]CalendarEntry, error) {
	var entries []CalendarEntry
	result := db.Get().Where("calendar_id = ?", calendarID).Order("date desc").Find(&entries)
	return entries, result.Error
}

// GetEntriesByDateRange returns calendar entries within a specific date range
func GetEntriesByDateRange(calendarID uint, startDate, endDate time.Time) ([]CalendarEntry, error) {
	var entries []CalendarEntry
	result := db.Get().Where("calendar_id = ? AND date BETWEEN ? AND ?", calendarID, startDate, endDate).Order("date asc").Find(&entries)
	return entries, result.Error
}

// GetEntriesByYearMonth returns calendar entries for a specific year and month
func GetEntriesByYearMonth(calendarID uint, year, month int) ([]CalendarEntry, error) {
	var entries []CalendarEntry
	result := db.Get().Where("calendar_id = ? AND year = ? AND month = ?", calendarID, year, month).Order("date asc").Find(&entries)
	return entries, result.Error
}

// GetCalendarEntry retrieves a calendar entry by its ID
func GetCalendarEntry(entryID uint) (CalendarEntry, error) {
	var entry CalendarEntry
	result := db.Get().First(&entry, entryID)
	if result.Error != nil {
		return entry, fmt.Errorf("failed to retrieve calendar entry: %w", result.Error)
	}
	return entry, nil
}

// CreateCalendarEntry creates a new calendar entry
func CreateCalendarEntry(calendarID uint, date time.Time, text string, hours float64, workResourceID uint) (CalendarEntry, error) {
	entry := CalendarEntry{
		CalendarID:     calendarID,
		Date:           date,
		Year:           date.Year(),
		Month:          int(date.Month()),
		Week:           getISOWeek(date),
		Hours:          hours,
		Text:           text,
		WorkResourceID: workResourceID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	result := db.Get().Create(&entry)
	if result.Error != nil {
		return entry, fmt.Errorf("failed to create calendar entry: %w", result.Error)
	}
	return entry, nil
}

// UpdateCalendarEntry updates an existing calendar entry
func UpdateCalendarEntry(entryID uint, date time.Time, text string, hours float64, workResourceID uint) (CalendarEntry, error) {
	var entry CalendarEntry
	result := db.Get().First(&entry, entryID)
	if result.Error != nil {
		return entry, fmt.Errorf("failed to retrieve calendar entry: %w", result.Error)
	}

	// Update the entry fields
	entry.Date = date
	entry.Year = date.Year()
	entry.Month = int(date.Month())
	entry.Week = getISOWeek(date)
	entry.Text = text
	entry.Hours = hours
	entry.WorkResourceID = workResourceID
	entry.UpdatedAt = time.Now()

	// Save the updated entry
	result = db.Get().Save(&entry)
	if result.Error != nil {
		return entry, fmt.Errorf("failed to update calendar entry: %w", result.Error)
	}
	return entry, nil
}

// DeleteCalendarEntry deletes a calendar entry
func DeleteCalendarEntry(entryID uint) error {
	result := db.Get().Delete(&CalendarEntry{}, entryID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete calendar entry: %w", result.Error)
	}
	return nil
}

// getISOWeek returns the ISO 8601 week number for a given date
func getISOWeek(date time.Time) int {
	year, week := date.ISOWeek()
	_ = year // We don't use the year returned by ISOWeek as we already have it
	return week
}
