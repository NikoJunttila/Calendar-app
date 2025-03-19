package calendar

import (
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
func CreateCalendar(name string, work bool, avgHours float64) (Calendar, error) {
	//get index number automatically?
	calendar := Calendar{
		Name:           name,
		IndexNumber:    1,
		Work:           work,
		DailyWorkHours: avgHours,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	result := db.Get().Create(&calendar)
	return calendar, result.Error
}

// GetCalendar retrieves a calendar by its ID
func GetCalendar(id uint) (Calendar, error) {
	var calendar Calendar
	result := db.Get().First(&calendar, id)
	return calendar, result.Error
}

// ListCalendars returns all calendars ordered by index_number
func ListCalendars() ([]Calendar, error) {
	var calendars []Calendar
	result := db.Get().Order("index_number asc").Find(&calendars)
	return calendars, result.Error
}

func GetCalendarWithEntries(id uint) (Calendar, error) {
	var calendar Calendar
	result := db.Get().Preload("Entries.WorkResource").First(&calendar, id)
	return calendar, result.Error
}

// GetCalendarWithEntriesByMonth retrieves a calendar by ID along with its entries for a specific month and year
// while also preloading the WorkResource for each entry
func GetCalendarWithEntriesByMonth(calendarID uint, year, month int) (Calendar, error) {
	var calendar Calendar

	// First, fetch the calendar
	if err := db.Get().First(&calendar, calendarID).Error; err != nil {
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

// CreateCalendarEntry creates a new calendar entry
func CreateCalendarEntry(calendarID uint, date time.Time, text string, hours float64, id uint) (CalendarEntry, error) {
	entry := CalendarEntry{
		CalendarID:     calendarID,
		Date:           date,
		Year:           date.Year(),
		Month:          int(date.Month()),
		Week:           getISOWeek(date),
		Hours:          hours,
		Text:           text,
		WorkResourceID: id,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	result := db.Get().Create(&entry)
	return entry, result.Error
}

// GetCalendarEntry retrieves a calendar entry by its ID
func GetCalendarEntry(id uint) (CalendarEntry, error) {
	var entry CalendarEntry
	result := db.Get().First(&entry, id)
	return entry, result.Error
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

// getISOWeek returns the ISO 8601 week number for a given date
func getISOWeek(date time.Time) int {
	year, week := date.ISOWeek()
	_ = year // We don't use the year returned by ISOWeek as we already have it
	return week
}
