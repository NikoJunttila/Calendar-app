package helloworld

import (
	"gothstack/app/db"
	"time"

	"gorm.io/gorm"
)

// Event name constants
const (
	HelloworldNewEvent = "helloworld.newEvent"
)

// Gorm looks for a table named helloworld_messages
type HelloworldMessage struct {
	gorm.Model

	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func createMessage(message string) (HelloworldMessage, error) {
	hello := HelloworldMessage{
		Message: message,
	}
	result := db.Get().Create(&hello)
	return hello, result.Error
}
func listMessages() ([]HelloworldMessage, error) {
	var messages []HelloworldMessage
	result := db.Get().Order("created_at desc").Find(&messages)
	return messages, result.Error
}

// using goose to init table
/* func initialize() {
	db.Get().AutoMigrate(&HelloworldMessage{})
} */
