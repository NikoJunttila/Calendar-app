package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	srcPath := os.Getenv("DB_NAME")
	if srcPath == "" {
		srcPath = "local.db"
		log.Println("DB_NAME not set, defaulting to local.db as source")
	}

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		log.Fatalf("Source database file '%s' does not exist", srcPath)
	}

	dstURL := os.Getenv("DB_URL")
	if dstURL == "" {
		log.Fatal("DB_URL environment variable is required (destination turso url)")
	}

	srcDB, err := sql.Open("sqlite3", srcPath)
	if err != nil {
		log.Fatal("Failed to open source database:", err)
	}
	defer srcDB.Close()

	dstDB, err := sql.Open("libsql", dstURL)
	if err != nil {
		log.Fatal("Failed to open destination database:", err)
	}
	defer dstDB.Close()

	// Define table migration order to satisfy foreign key constraints
	// Parent tables MUST come before child tables
	orderedTables := []string{
		"users",               // No dependencies
		"helloworld_messages", // Likely no dependencies
		"services",            // No dependencies
		"calendars",           // May depend on services or users
		"settings",            // No dependencies
		"work_resources",      // Depends on calendars
		"business_hours",      // Depends on calendars
		"special_dates",       // Depends on calendars
		"calendar_entries",    // Depends on calendars
		"sessions",            // Depends on users
		"books",               // No dependencies
		"book_progress",       // Depends on books and users
		"api_keys",            // Depends on users
		"time_slots",          // Depends on calendar_entries or work_resources
		"bookings",            // Depends on time_slots, calendars, services
		"webhooks",            // No dependencies
	}

	// Disable foreign keys on destination for the migration process
	if _, err := dstDB.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		log.Printf("Warning: Failed to disable foreign keys on destination: %v\n", err)
	}

	// DO NOT clear tables - we want to skip duplicates instead
	// Removed the table clearing logic

	for _, table := range orderedTables {
		fmt.Printf("Migrating table: %s\n", table)
		if err := migrateTable(srcDB, dstDB, table); err != nil {
			log.Printf("Failed to migrate table %s: %v\n", table, err)
			// Continue with other tables instead of stopping
		}
	}

	// Re-enable foreign keys
	if _, err := dstDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Printf("Warning: Failed to re-enable foreign keys on destination: %v\n", err)
	}

	fmt.Println("Migration completed successfully!")
}

func migrateTable(src, dst *sql.DB, table string) error {
	rows, err := src.Query(fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		return fmt.Errorf("failed to query source: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Get primary key columns for duplicate checking
	pkCols, err := getPrimaryKeyColumns(src, table)
	if err != nil {
		log.Printf("Warning: Could not get primary key for %s, using first column: %v\n", table, err)
		if len(cols) > 0 {
			pkCols = []string{cols[0]}
		}
	}

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(cols, ","), strings.Join(placeholders, ","))

	var insertedCount, skippedCount int

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Check if record already exists
		exists, err := recordExists(dst, table, cols, values, pkCols)
		if err != nil {
			log.Printf("Warning: Could not check if record exists in %s: %v\n", table, err)
			// Continue with insert attempt
		} else if exists {
			skippedCount++
			continue
		}

		// Convert byte slices to strings for compatibility
		processedValues := make([]interface{}, len(values))
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				processedValues[i] = string(b)
			} else {
				processedValues[i] = v
			}
		}

		if _, err := dst.Exec(insertQuery, processedValues...); err != nil {
			return fmt.Errorf("failed to execute SQL:\n%w", err)
		}
		insertedCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	fmt.Printf("  ✓ Inserted: %d, Skipped: %d (already exist)\n", insertedCount, skippedCount)
	return nil
}

func getPrimaryKeyColumns(db *sql.DB, tableName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pkColumns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue interface{}

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}

		if pk > 0 {
			pkColumns = append(pkColumns, name)
		}
	}

	if len(pkColumns) == 0 {
		return nil, fmt.Errorf("no primary key found")
	}

	return pkColumns, nil
}

func recordExists(db *sql.DB, tableName string, columns []string, values []interface{}, pkColumns []string) (bool, error) {
	if len(pkColumns) == 0 {
		return false, nil // Can't check without primary key
	}

	// Build WHERE clause using primary key columns
	var conditions []string
	var pkValues []interface{}

	for _, pkCol := range pkColumns {
		found := false
		for i, col := range columns {
			if col == pkCol {
				// Convert byte slice to string if needed
				val := values[i]
				if b, ok := val.([]byte); ok {
					val = string(b)
				}

				conditions = append(conditions, fmt.Sprintf("%s = ?", pkCol))
				pkValues = append(pkValues, val)
				found = true
				break
			}
		}
		if !found {
			return false, fmt.Errorf("primary key column %s not found in result set", pkCol)
		}
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s",
		tableName, strings.Join(conditions, " AND "))

	var count int
	err := db.QueryRow(query, pkValues...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
