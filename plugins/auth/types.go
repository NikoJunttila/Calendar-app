package auth

import (
	"database/sql"
	"gothstack/app/db"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Event name constants
const (
	UserSignupEvent         = "auth.signup"
	ResendVerificationEvent = "auth.resend.verification"
)

// UserWithVerificationToken is a struct that will be sent over the
// auth.signup event. It holds the User struct and the Verification token string.
type UserWithVerificationToken struct {
	User  User
	Token string
}

type Auth struct {
	UserID   uint
	Email    string
	LoggedIn bool
}

func (auth Auth) Check() bool {
	return auth.LoggedIn
}

type User struct {
	gorm.Model

	Email           string
	FirstName       string
	LastName        string
	PasswordHash    string
	Role            string
	EmailVerifiedAt sql.NullTime
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func createUserFromFormValues(values SignupFormValues) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(values.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	user := User{
		Email:        values.Email,
		FirstName:    values.FirstName,
		LastName:     values.LastName,
		Role:         "user",
		PasswordHash: string(hash),
	}
	result := db.Get().Create(&user)
	return user, result.Error
}

type Session struct {
	gorm.Model

	UserID    uint
	Token     string
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	CreatedAt time.Time
	User      User
}

// APIKey represents the api_keys table in the database
type APIKey struct {
	gorm.Model
	UserID     uint           `gorm:"not null;index"` // Foreign key to users table
	APIKey     string         `gorm:"unique;not null"`
	Name       string         `gorm:"not null"`
	Scopes     sql.NullString // Use sql.NullString for nullable TEXT fields
	LastUsedAt sql.NullTime   // Use sql.NullTime for nullable DATETIME fields
	ExpiresAt  sql.NullTime   // Use sql.NullTime for nullable DATETIME fields
	User       User           `gorm:"foreignKey:UserID"` // Define the relationship to the User struct
}
