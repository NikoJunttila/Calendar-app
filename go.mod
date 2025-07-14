module gothstack

go 1.23.0

toolchain go1.24.0

// uncomment for local development on the superkit core.
// replace github.com/anthdm/superkit => ../

require (
	github.com/a-h/templ v0.3.898
	github.com/anthdm/superkit v0.0.0-20240701091803-e7f8e0aad3e9
	github.com/go-chi/chi/v5 v5.2.2
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/goodsign/monday v1.0.2
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/mailgun/mailgun-go/v5 v5.4.2
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/robfig/cron/v3 v3.0.0
	golang.org/x/crypto v0.35.0
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)

require (
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailgun/errors v0.4.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	golang.org/x/text v0.22.0 // indirect
)
