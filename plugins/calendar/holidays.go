package calendar

import "time"

// FinnishHoliday represents a public holiday in Finland
type FinnishHoliday struct {
	Date        time.Time
	Name        string
	Description string
}

// IsFinnishHoliday checks if a given date is a Finnish public holiday
func IsFinnishHoliday(date time.Time) (bool, FinnishHoliday) {
	// Normalize the date to remove time component
	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	// Get all Finnish holidays for the year
	holidays := GetFinnishHolidays(date.Year())

	// Check if the date is in the list of holidays
	for _, holiday := range holidays {
		if holiday.Date.Equal(normalizedDate) {
			return true, holiday
		}
	}

	return false, FinnishHoliday{}
}

// GetFinnishHolidays returns all Finnish public holidays for a given year
func GetFinnishHolidays(year int) []FinnishHoliday {
	holidays := []FinnishHoliday{}

	// Fixed date holidays
	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local),
		Name:        "New Year's Day",
		Description: "Uudenvuodenpäivä",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.January, 6, 0, 0, 0, 0, time.Local),
		Name:        "Epiphany",
		Description: "Loppiainen",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.May, 1, 0, 0, 0, 0, time.Local),
		Name:        "May Day",
		Description: "Vappu",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.June, 24, 0, 0, 0, 0, time.Local), // Midsummer's Eve falls on different dates, this is a simplification
		Name:        "Midsummer's Eve",
		Description: "Juhannusaatto",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.December, 6, 0, 0, 0, 0, time.Local),
		Name:        "Independence Day",
		Description: "Itsenäisyyspäivä",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.December, 24, 0, 0, 0, 0, time.Local),
		Name:        "Christmas Eve",
		Description: "Jouluaatto",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.December, 25, 0, 0, 0, 0, time.Local),
		Name:        "Christmas Day",
		Description: "Joulupäivä",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        time.Date(year, time.December, 26, 0, 0, 0, 0, time.Local),
		Name:        "St. Stephen's Day",
		Description: "Tapaninpäivä",
	})

	// Calculate Easter Sunday (using the algorithm)
	easterDate := calculateEasterDate(year)

	// Easter and related holidays
	holidays = append(holidays, FinnishHoliday{
		Date:        easterDate.AddDate(0, 0, -2), // Good Friday
		Name:        "Good Friday",
		Description: "Pitkäperjantai",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        easterDate, // Easter Sunday
		Name:        "Easter Sunday",
		Description: "Pääsiäispäivä",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        easterDate.AddDate(0, 0, 1), // Easter Monday
		Name:        "Easter Monday",
		Description: "2. pääsiäispäivä",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        easterDate.AddDate(0, 0, 39), // Ascension Day
		Name:        "Ascension Day",
		Description: "Helatorstai",
	})

	holidays = append(holidays, FinnishHoliday{
		Date:        easterDate.AddDate(0, 0, 49), // Pentecost
		Name:        "Pentecost",
		Description: "Helluntaipäivä",
	})

	// Calculate Midsummer (Saturday between June 20-26)
	midsummerDate := calculateMidsummerDate(year)
	holidays = append(holidays, FinnishHoliday{
		Date:        midsummerDate,
		Name:        "Midsummer Day",
		Description: "Juhannuspäivä",
	})

	// All Saints' Day (Saturday between Oct 31 and Nov 6)
	allSaintsDate := calculateAllSaintsDate(year)
	holidays = append(holidays, FinnishHoliday{
		Date:        allSaintsDate,
		Name:        "All Saints' Day",
		Description: "Pyhäinpäivä",
	})

	return holidays
}

// calculateEasterDate calculates Easter Sunday for a given year
// This uses the Butcher's algorithm
func calculateEasterDate(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}

// calculateMidsummerDate calculates the Midsummer Day for a given year
// In Finland, Midsummer Day is the Saturday that falls between June 20-26
func calculateMidsummerDate(year int) time.Time {
	// Start with June 20
	date := time.Date(year, time.June, 20, 0, 0, 0, 0, time.Local)

	// Find the next Saturday
	daysUntilSaturday := (int(time.Saturday) - int(date.Weekday()) + 7) % 7
	return date.AddDate(0, 0, daysUntilSaturday)
}

// calculateAllSaintsDate calculates All Saints' Day for a given year
// In Finland, All Saints' Day is the Saturday that falls between Oct 31 and Nov 6
func calculateAllSaintsDate(year int) time.Time {
	// Start with October 31
	date := time.Date(year, time.October, 31, 0, 0, 0, 0, time.Local)

	// Find the next Saturday
	daysUntilSaturday := (int(time.Saturday) - int(date.Weekday()) + 7) % 7
	return date.AddDate(0, 0, daysUntilSaturday)
}
