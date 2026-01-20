package db

import (
	"database/sql"
	"embed"
	"log"
	"os"

	"github.com/anthdm/superkit/db"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// By default this is a pre-configured Gorm DB instance.
// Change this type based on the database package of your likings.
var dbInstance *gorm.DB

// Get returns the instantiated DB instance.
func Get() *gorm.DB {
	return dbInstance
}

func init() {
	// Load .env file if present, as this init runs before main.init
	if err := godotenv.Load(); err != nil {
		// handle error or ignore if .env is not required/expected in all environments
		// log.Println("db init: .env not found or error loading")
	}

	var err error

	// Create a default *sql.DB exposed by the superkit/db package
	// based on the given configuration.
	config := db.Config{
		Driver:   os.Getenv("DB_DRIVER"),
		Name:     os.Getenv("DB_NAME"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		Host:     os.Getenv("DB_HOST"),
	}

	var conn *sql.DB

	// Handle Turso/LibSQL specifically if driver is set to "libsql"
	// or if we decide to treat "sqlite" as libsql compatible.
	// For this task, we explicitly check for "libsql".
	if config.Driver == "libsql" {
		url := os.Getenv("DB_URL")
		if url == "" {
			log.Fatal("DB_URL environment variable is required for libsql driver")
		}

		conn, err = sql.Open("libsql", url)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Fallback to original Superkit behavior for other drivers
		conn, err = db.NewSQL(config)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Run Migrations
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(conn, "migrations"); err != nil {
		log.Fatal(err)
	}

	// Based on the driver create the corresponding DB instance.
	// By default, the SuperKit boilerplate comes with a pre-configured
	// ORM called Gorm. https://gorm.io.
	//
	// You can change this to any other DB interaction tool
	// of your liking. EG:
	// - uptrace bun -> https://bun.uptrace.dev
	// - SQLC -> https://github.com/sqlc-dev/sqlc
	// - gojet -> https://github.com/go-jet/jet
	switch config.Driver {
	case "libsql":
		// GORM with sqlite driver works for libsql
		dbInstance, err = gorm.Open(sqlite.New(sqlite.Config{
			Conn: conn,
		}))
	case db.DriverSqlite3:
		dbInstance, err = gorm.Open(sqlite.New(sqlite.Config{
			Conn: conn,
		}))
	case db.DriverMysql:
		// ...
	default:
		log.Fatal("invalid driver:", config.Driver)
	}
	if err != nil {
		log.Fatal(err)
	}
}
